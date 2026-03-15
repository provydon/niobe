package live

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"google.golang.org/genai"
)

// Connect uses sessionCfg as-is. Prompt/logic live in handler; this is the API gateway only.

// GoogleConnector creates live sessions via the Google GenAI SDK (Gemini / Vertex).
// All Google-specific logic is in this file.
type GoogleConnector struct{}

// Connect implements live.Connector.
func (GoogleConnector) Connect(ctx context.Context, cfg Config, sessionCfg SessionConfig) (Session, error) {
	httpOpts := genai.HTTPOptions{APIVersion: "v1"}
	if cfg.GetUseVertex() {
		httpOpts.APIVersion = "v1beta1"
	} else {
		httpOpts.APIVersion = "v1beta"
	}
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:      cfg.GetAPIKey(),
		HTTPOptions: httpOpts,
	})
	if err != nil {
		return nil, err
	}
	model := "gemini-2.5-flash-native-audio-preview-12-2025"
	if cfg.GetUseVertex() {
		model = "gemini-2.0-flash-live-preview-04-09"
	}
	systemInstruction := strings.TrimSpace(sessionCfg.SystemInstruction)
	tools, _ := sessionCfg.Tools.([]*genai.Tool)
	connectConfig := &genai.LiveConnectConfig{
		InputAudioTranscription:  &genai.AudioTranscriptionConfig{},
		ResponseModalities:       []genai.Modality{genai.ModalityAudio},
		OutputAudioTranscription: &genai.AudioTranscriptionConfig{},
		SpeechConfig: &genai.SpeechConfig{
			VoiceConfig: &genai.VoiceConfig{
				PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
					VoiceName: "Aoede",
				},
			},
		},
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{genai.NewPartFromText(systemInstruction)},
			Role:  genai.RoleUser,
		},
		Tools: tools,
	}
	sess, err := client.Live.Connect(ctx, model, connectConfig)
	if err != nil {
		return nil, err
	}
	log.Printf("[live] google: connected to model %s", model)
	return &googleSession{sess}, nil
}

// googleSession adapts *genai.Session to live.Session.
type googleSession struct{ *genai.Session }

func (s *googleSession) SendRealtimeInput(raw []byte) error {
	var input genai.LiveRealtimeInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return err
	}
	return s.Session.SendRealtimeInput(input)
}

func (s *googleSession) SendToolResponse(responses []*FunctionResponse) error {
	genaiResponses := googleFunctionResponsesToGenAI(responses)
	return s.Session.SendToolResponse(genai.LiveToolResponseInput{
		FunctionResponses: genaiResponses,
	})
}

func (s *googleSession) Receive() (*ServerMessage, error) {
	msg, err := s.Session.Receive()
	if err != nil {
		return nil, err
	}
	return googleServerMessageToLive(msg), nil
}

func (s *googleSession) Close() error {
	return s.Session.Close()
}

func googleServerMessageToLive(msg *genai.LiveServerMessage) *ServerMessage {
	if msg == nil {
		return nil
	}
	out := &ServerMessage{}
	if msg.SetupComplete != nil {
		out.SetupComplete = &struct{}{}
	}
	if msg.ServerContent != nil {
		out.ServerContent = &ServerContent{
			TurnComplete: msg.ServerContent.TurnComplete,
			Interrupted:  msg.ServerContent.Interrupted,
		}
		if msg.ServerContent.ModelTurn != nil && len(msg.ServerContent.ModelTurn.Parts) > 0 {
			parts := make([]Part, 0, len(msg.ServerContent.ModelTurn.Parts))
			for _, p := range msg.ServerContent.ModelTurn.Parts {
				if p.InlineData != nil {
					parts = append(parts, Part{
						InlineData: &InlineData{
							MimeType: p.InlineData.MIMEType,
							Data:     p.InlineData.Data,
						},
					})
				}
			}
			out.ServerContent.ModelTurn = &ModelTurn{Parts: parts}
		}
		if msg.ServerContent.InputTranscription != nil {
			out.ServerContent.InputTranscription = &Transcription{Text: msg.ServerContent.InputTranscription.Text}
		}
		if msg.ServerContent.OutputTranscription != nil {
			out.ServerContent.OutputTranscription = &Transcription{Text: msg.ServerContent.OutputTranscription.Text}
		}
	}
	if msg.ToolCall != nil && len(msg.ToolCall.FunctionCalls) > 0 {
		calls := make([]FunctionCall, 0, len(msg.ToolCall.FunctionCalls))
		for _, c := range msg.ToolCall.FunctionCalls {
			if c == nil {
				continue
			}
			calls = append(calls, FunctionCall{
				ID:   c.ID,
				Name: c.Name,
				Args: c.Args,
			})
		}
		out.ToolCall = &ToolCall{FunctionCalls: calls}
	}
	return out
}

func googleFunctionResponsesToGenAI(responses []*FunctionResponse) []*genai.FunctionResponse {
	if len(responses) == 0 {
		return nil
	}
	out := make([]*genai.FunctionResponse, 0, len(responses))
	for _, r := range responses {
		if r == nil {
			continue
		}
		out = append(out, &genai.FunctionResponse{
			ID:       r.ID,
			Name:     r.Name,
			Response: r.Response,
		})
	}
	return out
}
