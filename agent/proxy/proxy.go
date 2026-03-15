package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"niobe/agent/live"
	"niobe/agent/tools"
)

type ToolExecutor interface {
	Definitions() []tools.Definition
	Lookup(toolName string) (tools.Definition, bool)
	Execute(ctx context.Context, toolName string, arguments map[string]any) (tools.CallResponse, error)
}

type toolEvent struct {
	CallID     string `json:"callId"`
	ToolName   string `json:"toolName"`
	Status     string `json:"status"`
	Title      string `json:"title"`
	Subtitle   string `json:"subtitle,omitempty"`
	Message    string `json:"message,omitempty"`
	ActionType string `json:"actionType,omitempty"`
}

type conversationEvent struct {
	Phase  string `json:"phase"`
	Status string `json:"status"`
	Tone   string `json:"tone"`
}

type guardedTurn struct {
	messages        []*live.ServerMessage
	userTranscript  strings.Builder
	transcript      strings.Builder
	actionSucceeded bool
	farewellStyle   string
	guardRequested  bool
}

type sessionState struct {
	awaitingConfirmation bool
	awaitingFollowup     bool
	confirmationTurn     int
	userTurnCount        int
	userTurnInProgress   bool
	lastFarewellStyle    string
	mu sync.Mutex
}

// Run bridges the browser WebSocket and the live session until one side closes.
func Run(c *websocket.Conn, session live.Session, toolExecutor ToolExecutor) {
	state := &sessionState{}
	sessionDone := make(chan struct{})
	go receiveFromModel(c, session, toolExecutor, state, sessionDone)
	sendToModel(c, session, state)
	<-sessionDone
}

func receiveFromModel(c *websocket.Conn, session live.Session, toolExecutor ToolExecutor, state *sessionState, done chan struct{}) {
	defer close(done)
	defer c.Close()
	guardEnabled := hasExecutableTools(toolExecutor)
	turn := &guardedTurn{}

	for {
		msg, err := session.Receive()
		if err != nil {
			if err == io.EOF {
				log.Printf("[live] session ended (model closed stream)")
			} else {
				log.Printf("[live] receive from model: %v", err)
			}
			return
		}
		logReceive(msg)

		if msg.ToolCall != nil && toolExecutor != nil {
			handled, succeeded, farewellStyle := handleToolCalls(c, session, toolExecutor, state, turn, msg.ToolCall)
			if handled {
				turn.actionSucceeded = turn.actionSucceeded || succeeded
				if succeeded && farewellStyle != "" {
					turn.farewellStyle = farewellStyle
				}
				turn.guardRequested = true
				continue
			}
		}

		if guardEnabled && msg.ServerContent != nil {
			trackTurnState(turn, msg)
			if handleFollowupState(c, state, turn) {
				return
			}
			if turn.guardRequested {
				bufferTurnMessage(turn, msg)
				if msg.ServerContent.TurnComplete {
					if shouldBlockOptimisticClaim(turn) {
						log.Printf("[live] blocked model action claim without tool call: user=%q output=%q", turn.userTranscript.String(), turn.transcript.String())
						writeJSON(c, map[string]any{
							"toolEvent": toolEvent{
								CallID:   fmt.Sprintf("guard-%d", time.Now().UnixNano()),
								ToolName: "action_guard",
								Status:   "error",
								Title:    "Request not completed",
								Message:  "I couldn't verify that your request was completed.",
							},
							"conversationEvent": conversationEvent{
								Phase:  "error",
								Status: "I couldn't verify that your request was completed.",
								Tone:   "error",
							},
						})
					} else {
						flushTurn(c, turn)
						if turn.actionSucceeded {
							state.markActionHandled(turn.farewellStyle)
							writeJSON(c, map[string]any{
								"conversationEvent": conversationEvent{
									Phase:  "awaiting_followup",
									Status: "Done. Waiting for the next request.",
									Tone:   "connected",
								},
							})
						}
					}

					turn = &guardedTurn{}
				}

				continue
			}

			emitTranscriptIfPresent(c, msg)
			raw, _ := json.Marshal(msg)
			if err := c.WriteMessage(websocket.TextMessage, raw); err != nil {
				return
			}

			if msg.ServerContent != nil && msg.ServerContent.TurnComplete {
				turn = &guardedTurn{}
			}

			continue
		}

		emitTranscriptIfPresent(c, msg)
		raw, _ := json.Marshal(msg)
		if err := c.WriteMessage(websocket.TextMessage, raw); err != nil {
			return
		}
	}
}

func handleToolCalls(c *websocket.Conn, session live.Session, toolExecutor ToolExecutor, state *sessionState, turn *guardedTurn, toolCall *live.ToolCall) (bool, bool, string) {
	if toolCall == nil || len(toolCall.FunctionCalls) == 0 {
		return false, false, ""
	}

	// Confirmation is handled in conversation by the model (system instruction: confirm then call tool).
	// Running the tool when the model calls it.

	responses := make([]*live.FunctionResponse, 0, len(toolCall.FunctionCalls))
	anySuccess := false
	farewellStyle := ""

	for _, call := range toolCall.FunctionCalls {
		log.Printf("[live] running tool call: name=%s callId=%s", call.Name, call.ID)

		definition, hasDefinition := toolExecutor.Lookup(call.Name)
		title := "Working on your request"
		subtitle := ""
		actionType := ""
		if hasDefinition {
			title = customerFacingToolTitle(definition)
			actionType = definition.Type
		}

		writeJSON(c, map[string]any{
			"toolEvent": toolEvent{
				CallID:     call.ID,
				ToolName:   call.Name,
				Status:     "running",
				Title:      title,
				Subtitle:   subtitle,
				Message:    "Working on it...",
				ActionType: actionType,
			},
			"conversationEvent": conversationEvent{
				Phase:  "executing_action",
				Status: "Working on that now...",
				Tone:   "default",
			},
		})

		callResponse, err := toolExecutor.Execute(context.Background(), call.Name, call.Args)
		responsePayload := map[string]any{
			"error": map[string]any{
				"message": "Action failed.",
			},
		}
		event := toolEvent{
			CallID:     call.ID,
			ToolName:   call.Name,
			Status:     "error",
			Title:      title,
			Subtitle:   subtitle,
			ActionType: actionType,
		}

		if err != nil {
			log.Printf("[live] tool call finished: name=%s success=false err=%v", call.Name, err)
			event.Message = err.Error()
			responsePayload["error"] = map[string]any{
				"message": err.Error(),
			}
		} else {
			success := synchronousActionSucceeded(callResponse.Result)
			log.Printf("[live] tool call finished: name=%s success=%v", call.Name, success)
			responsePayload = callResponse.Result

			if display, ok := displayFromResult(callResponse.Result); ok {
				event.Title = display.Title
				event.Subtitle = display.Subtitle
				event.Status = normalizeEventStatus(display.Status, callResponse.Result)
			}

			event.Message = actionResponseMessage(callResponse.Result)
			if event.Message == "" {
				event.Message = "Action finished."
			}

			if success {
				anySuccess = true
				farewellStyle = mergeFarewellStyle(farewellStyle, farewellStyleForDefinition(definition))
			}
		}

		responses = append(responses, &live.FunctionResponse{
			ID:       call.ID,
			Name:     call.Name,
			Response: responsePayload,
		})

		tone := "error"
		if event.Status == "success" {
			tone = "connected"
		}

		writeJSON(c, map[string]any{
			"toolEvent": event,
			"conversationEvent": conversationEvent{
				Phase:  "reporting_result",
				Status: event.Message,
				Tone:   tone,
			},
		})
	}

	if len(responses) == 0 {
		return false, anySuccess, farewellStyle
	}

	if err := session.SendToolResponse(responses); err != nil {
		writeJSON(c, map[string]any{
			"conversationEvent": conversationEvent{
				Phase:  "error",
				Status: "Failed to return tool result to the model.",
				Tone:   "error",
			},
		})
		log.Printf("[live] send tool response to model failed: %v", err)
	} else {
		log.Printf("[live] tool response sent to model (count=%d)", len(responses))
	}

	return true, anySuccess, farewellStyle
}

func logReceive(msg *live.ServerMessage) {
	if msg != nil && msg.ToolCall != nil && len(msg.ToolCall.FunctionCalls) > 0 {
		names := make([]string, 0, len(msg.ToolCall.FunctionCalls))
		for _, c := range msg.ToolCall.FunctionCalls {
			names = append(names, c.Name)
		}
		log.Printf("[live] model requested tool call: %v", names)
	}
}

type displayPayload struct {
	Title    string
	Subtitle string
	Status   string
}

func displayFromResult(result map[string]any) (displayPayload, bool) {
	raw, ok := result["display"].(map[string]any)
	if !ok {
		return displayPayload{}, false
	}

	return displayPayload{
		Title:    stringFromMap(raw, "title"),
		Subtitle: stringFromMap(raw, "subtitle"),
		Status:   stringFromMap(raw, "status"),
	}, true
}

func stringFromMap(input map[string]any, key string) string {
	value, _ := input[key].(string)
	return value
}

func writeJSON(c *websocket.Conn, payload map[string]any) {
	if err := c.WriteJSON(payload); err != nil {
		log.Printf("[live] write client event failed: %v", err)
	}
}

// emitTranscriptIfPresent sends transcript events for user/agent speech so the client can show a live transcript.
func emitTranscriptIfPresent(c *websocket.Conn, msg *live.ServerMessage) {
	if c == nil || msg == nil || msg.ServerContent == nil {
		return
	}
	sc := msg.ServerContent
	if sc.InputTranscription != nil {
		if text := strings.TrimSpace(sc.InputTranscription.Text); text != "" {
			writeJSON(c, map[string]any{
				"transcript": map[string]any{"role": "user", "text": text},
			})
		}
	}
	if sc.OutputTranscription != nil {
		if text := strings.TrimSpace(sc.OutputTranscription.Text); text != "" {
			writeJSON(c, map[string]any{
				"transcript": map[string]any{"role": "agent", "text": text},
			})
		}
	}
}

func handleFollowupState(c *websocket.Conn, state *sessionState, turn *guardedTurn) bool {
	if state == nil || !state.isAwaitingFollowup() || turn == nil {
		return false
	}

	userText := strings.TrimSpace(turn.userTranscript.String())
	if userText == "" {
		return false
	}

	if looksLikeFollowupYes(userText) {
		state.clearAwaitingFollowup()
		writeJSON(c, map[string]any{
			"conversationEvent": conversationEvent{
				Phase:  "listening",
				Status: "Listening for your next request.",
				Tone:   "connected",
			},
		})
	}

	return false
}

func hasExecutableTools(toolExecutor ToolExecutor) bool {
	if toolExecutor == nil {
		return false
	}

	for _, definition := range toolExecutor.Definitions() {
		if definition.Enabled {
			return true
		}
	}

	return false
}

func trackTurnState(turn *guardedTurn, msg *live.ServerMessage) {
	if msg == nil || msg.ServerContent == nil {
		return
	}

	if msg.ServerContent.InputTranscription != nil {
		text := strings.TrimSpace(msg.ServerContent.InputTranscription.Text)
		if text != "" {
			if turn.userTranscript.Len() > 0 {
				turn.userTranscript.WriteString(" ")
			}
			turn.userTranscript.WriteString(text)
			if looksLikeActionRequest(turn.userTranscript.String()) {
				turn.guardRequested = true
			}
		}
	}

	if msg.ServerContent.OutputTranscription != nil {
		text := strings.TrimSpace(msg.ServerContent.OutputTranscription.Text)
		if text != "" {
			if turn.transcript.Len() > 0 {
				turn.transcript.WriteString(" ")
			}
			turn.transcript.WriteString(text)
			if looksLikeSuccessClaim(turn.transcript.String()) {
				turn.guardRequested = true
			}
		}
	}
}

func bufferTurnMessage(turn *guardedTurn, msg *live.ServerMessage) {
	turn.messages = append(turn.messages, msg)
}

func flushTurn(c *websocket.Conn, turn *guardedTurn) {
	for _, buffered := range turn.messages {
		raw, _ := json.Marshal(buffered)
		if err := c.WriteMessage(websocket.TextMessage, raw); err != nil {
			log.Printf("[live] flush buffered turn failed: %v", err)
			return
		}
	}
}

func looksLikeSuccessClaim(transcript string) bool {
	text := strings.ToLower(strings.TrimSpace(transcript))
	if text == "" {
		return false
	}
	for _, phrase := range successClaimPhrases() {
		if strings.Contains(text, phrase) {
			return true
		}
	}
	return false
}

func successClaimPhrases() []string {
	return []string{
		"order placed",
		"your order has been placed",
		"i've sent",
		"i have sent",
		"it has been sent",
		"your order has been sent",
		"sent the order",
		"placed the order",
		"i've placed the order",
		"i have placed the order",
		"email sent",
		"webhook sent",
		"done for you",
		"completed that",
		"it's done",
		"it is done",
	}
}

func shouldBlockOptimisticClaim(turn *guardedTurn) bool {
	if turn == nil || turn.actionSucceeded {
		return false
	}

	transcript := strings.ToLower(strings.TrimSpace(turn.transcript.String()))
	if transcript == "" {
		return false
	}

	for _, phrase := range successClaimPhrases() {
		if strings.Contains(transcript, phrase) {
			return true
		}
	}

	return false
}

func synchronousActionSucceeded(result map[string]any) bool {
	ok, _ := result["ok"].(bool)
	return ok
}

func synchronousActionFailureMessage(result map[string]any) string {
	if rawError, ok := result["error"].(map[string]any); ok {
		if message, _ := rawError["message"].(string); strings.TrimSpace(message) != "" {
			return message
		}
	}

	message := strings.TrimSpace(stringFromMap(result, "message"))
	if message == "" {
		return "Laravel did not confirm that the action completed."
	}

	if synchronousActionSucceeded(result) {
		return message
	}

	return fmt.Sprintf("%s Laravel did not confirm that the action completed.", message)
}

func actionResponseMessage(result map[string]any) string {
	if ok, _ := result["ok"].(bool); ok {
		return strings.TrimSpace(stringFromMap(result, "message"))
	}

	return synchronousActionFailureMessage(result)
}

func normalizeEventStatus(status string, result map[string]any) string {
	normalized := strings.ToLower(strings.TrimSpace(status))
	if normalized == "" && synchronousActionSucceeded(result) {
		return "success"
	}
	if normalized == "queued" && synchronousActionSucceeded(result) {
		return "success"
	}
	return normalized
}

func customerFacingToolTitle(definition tools.Definition) string {
	name := strings.ToLower(strings.TrimSpace(definition.DisplayName))
	if strings.Contains(name, "order") {
		return "Placing your order"
	}
	if strings.Contains(name, "issue") || strings.Contains(name, "support") || strings.Contains(name, "repair") || strings.Contains(name, "fix") {
		return "Sending this to the right team"
	}

	switch definition.Type {
	case "send_email", "send_webhook_event":
		return "Sending this to the right team"
	default:
		return "Working on your request"
	}
}

func farewellStyleForDefinition(definition tools.Definition) string {
	name := strings.ToLower(strings.TrimSpace(definition.DisplayName))
	if strings.Contains(name, "order") || strings.Contains(name, "menu") || strings.Contains(name, "meal") || strings.Contains(name, "food") {
		return "meal"
	}
	if strings.Contains(name, "issue") || strings.Contains(name, "support") || strings.Contains(name, "repair") || strings.Contains(name, "fix") || strings.Contains(name, "problem") {
		return "service"
	}

	switch definition.Type {
	case "send_email", "send_webhook_event":
		return "service"
	default:
		return "general"
	}
}

func mergeFarewellStyle(current, next string) string {
	if next == "" {
		return current
	}
	if current == "" || current == "general" {
		return next
	}
	return current
}

func looksLikeActionRequest(input string) bool {
	text := strings.ToLower(strings.TrimSpace(input))
	if text == "" {
		return false
	}

	actionPhrases := []string{
		"send",
		"email",
		"mail",
		"order",
		"place the order",
		"book",
		"confirm",
		"submit",
		"notify",
		"message",
		"text ",
		"sms",
		"whatsapp",
		"webhook",
		"trigger",
		"call them",
		"reach out",
	}

	for _, phrase := range actionPhrases {
		if strings.Contains(text, phrase) {
			return true
		}
	}

	return false
}

func looksLikeConfirmation(input string) bool {
	text := strings.ToLower(strings.TrimSpace(input))
	if text == "" {
		return false
	}

	confirmationPhrases := []string{
		"yes",
		"confirm",
		"go ahead",
		"proceed",
		"do it",
		"send it",
		"place it",
		"that's correct",
		"that is correct",
		"correct",
		"approved",
		"you can send",
		"you can place",
		"continue",
	}

	for _, phrase := range confirmationPhrases {
		if strings.Contains(text, phrase) {
			return true
		}
	}

	return false
}

func looksLikeFollowupYes(input string) bool {
	text := strings.ToLower(strings.TrimSpace(input))
	if text == "" {
		return false
	}

	phrases := []string{
		"yes",
		"yeah",
		"yep",
		"another",
		"something else",
		"more",
		"continue",
		"one more",
		"new request",
	}

	for _, phrase := range phrases {
		if strings.Contains(text, phrase) {
			return true
		}
	}

	return false
}

func confirmationRequired(state *sessionState, turn *guardedTurn) bool {
	if state == nil {
		return true
	}

	userText := ""
	if turn != nil {
		userText = turn.userTranscript.String()
	}

	awaitingConfirmation, confirmationTurn, userTurnCount, userTurnInProgress := state.confirmationSnapshot()
	if awaitingConfirmation {
		if looksLikeConfirmation(userText) {
			return false
		}

		if userTurnInProgress && userTurnCount >= confirmationTurn {
			return false
		}

		return userTurnCount <= confirmationTurn
	}

	return true
}

func buildConfirmationResponses(toolCall *live.ToolCall) []*live.FunctionResponse {
	if toolCall == nil {
		return nil
	}
	responses := make([]*live.FunctionResponse, 0, len(toolCall.FunctionCalls))
	for _, call := range toolCall.FunctionCalls {
		responses = append(responses, &live.FunctionResponse{
			ID:   call.ID,
			Name: call.Name,
			Response: map[string]any{
				"ok":                    false,
				"confirmation_required": true,
				"message":               "Please confirm the request details with the user before you run this.",
			},
		})
	}
	return responses
}

func sendToModel(c *websocket.Conn, session live.Session, state *sessionState) {
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			break
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(message, &raw); err == nil {
			if _, ok := raw["audioStreamEnd"]; ok {
				state.markUserTurnComplete()
			} else if _, ok := raw["media"]; ok {
				state.markUserTurnStarted()
			}
		}
		if err := session.SendRealtimeInput(message); err != nil {
			log.Printf("[live] send to model failed: %v", err)
			break
		}
	}
}

func (s *sessionState) requestConfirmation() {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.awaitingConfirmation = true
	s.confirmationTurn = s.userTurnCount
	s.userTurnInProgress = false
}

func (s *sessionState) confirmationSatisfied() {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.awaitingConfirmation = false
	s.confirmationTurn = 0
	s.userTurnInProgress = false
}

func (s *sessionState) markActionHandled(farewellStyle string) {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.awaitingConfirmation = false
	s.awaitingFollowup = true
	s.confirmationTurn = 0
	s.userTurnInProgress = false
	if strings.TrimSpace(farewellStyle) == "" {
		farewellStyle = "general"
	}
	s.lastFarewellStyle = farewellStyle
}

func (s *sessionState) clearAwaitingFollowup() {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.awaitingFollowup = false
}

func (s *sessionState) isAwaitingFollowup() bool {
	if s == nil {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return s.awaitingFollowup
}

func (s *sessionState) markUserTurnComplete() {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.userTurnCount++
	s.userTurnInProgress = false
}

func (s *sessionState) markUserTurnStarted() {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.userTurnInProgress = true
}

func (s *sessionState) confirmationSnapshot() (bool, int, int, bool) {
	if s == nil {
		return false, 0, 0, false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return s.awaitingConfirmation, s.confirmationTurn, s.userTurnCount, s.userTurnInProgress
}

