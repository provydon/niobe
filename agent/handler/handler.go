package handler

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"

	"niobe/agent/config"
	"niobe/agent/live"
	"niobe/agent/proxy"
	"niobe/agent/store"
	"niobe/agent/tools"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Health reports that the agent API is reachable.
func Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
			log.Printf("write health: %v", err)
		}
	}
}

// Live upgrades the connection to WebSocket, opens a live session via the connector,
// then delegates to the proxy. Tool execution runs in-process (Go + DB); no Laravel HTTP calls.
func Live(connector live.Connector, cfg live.Config, waitresses *store.WaitressRepository, db *sql.DB, appCfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("upgrade: %v", err)
			return
		}
		defer c.Close()

		ctx := context.Background()
		sessionCfg := live.SessionConfig{}
		slug := strings.TrimSpace(r.URL.Query().Get("niobe"))
		var toolExecutor *tools.LocalNiobeTools
		var waitress *store.Waitress
		if slug != "" {
			var findErr error
			waitress, findErr = waitresses.FindBySlug(r.Context(), slug)
			if findErr != nil {
				log.Printf("[live] waitress lookup failed: %v", findErr)
				sendWSError(c, fmt.Sprintf("load waitress: %v", findErr))
				return
			}

			tableFromURL := strings.TrimSpace(r.URL.Query().Get("table"))
			if len(waitress.Tools) > 0 {
				toolExecutor = tools.NewLocalNiobeTools(waitress, db, appCfg, tableFromURL)
				sessionCfg.Tools = tools.ToGenAITools(toolExecutor.Definitions())
			}
			sessionCfg.SystemInstruction = buildNiobeInstruction(waitress, toolExecutor, tableFromURL)
		}

		session, err := connector.Connect(ctx, cfg, sessionCfg)
		if err != nil {
			log.Printf("[live] connect failed: %v", err)
			sendWSError(c, fmt.Sprintf("connect to model: %v", err))
			return
		}
		defer session.Close()

		if waitress != nil {
			log.Printf("[live] session started waitress=%s slug=%s tools=%v", waitress.Name, slug, len(waitress.Tools))
		} else {
			log.Printf("[live] session started (no waitress)")
		}
		proxy.Run(c, session, toolExecutor)
	}
}

func sendWSError(c *websocket.Conn, text string) {
	_ = c.WriteJSON(map[string]string{"error": text})
}

// System instruction preamble: same for all providers. Only this file and buildNiobeInstruction define prompt logic.
const systemInstructionPreamble = `You are a friendly AI waitress. Take orders and help the customer.
Silently use the provided menu and instructions. Do not read them aloud unless asked.
Listen first, then respond naturally. Keep replies concise.
Focus on the primary speaker; ignore background noise.
If the user interrupts you while you are speaking, stop immediately and listen to what they say. Respond to their new input naturally; do not apologize for being interrupted or refer to your previous sentence.`

func buildNiobeInstruction(waitress *store.Waitress, toolExecutor *tools.LocalNiobeTools, tableFromURL string) string {
	parts := []string{
		strings.TrimSpace(systemInstructionPreamble),
		fmt.Sprintf("Niobe name: %s", waitress.Name),
		"The Niobe name above is your own assistant name and identity, not the user's name.",
		"Do not address the user as " + waitress.Name + " unless the user explicitly tells you that is their name.",
		"If you do not know the user's name, do not guess one. Address them naturally without using a name.",
		"CRITICAL - Actions and tools: When the user confirms an action, you MUST call the matching tool in that same turn. Treat any of these as confirmation and invoke the tool right away: yes, yeah, yep, no problem, go ahead, go on, sure, correct, that's right, that's correct, do it, send it, place it, ok, okay, sounds good, go for it, sure thing, absolutely, definitely, of course, please do, and similar short affirmations. When they confirm, your response must be to call the tool—do not respond with only speech. Wrong: user says 'Yes' and you reply 'Your order has been placed' without calling the tool. Right: user says 'Yes' and you call the place-order tool with order_details and table_number; only after the tool returns success do you say 'Your order has been placed. Anything else?'. Do NOT say the action is done, placed, or completed until you have called the tool and received its success response. Step 1: User confirms. Step 2: You call the tool immediately (no success wording yet). Step 3: You get the tool result. Step 4: Say the success message exactly once, then stop.",
		"For place-order (or any order) tools: every time you call the tool you MUST pass order_details, table_number, and customer_name (if the customer gave their name). Never call the tool with empty arguments {} or the action will fail. Extract table number and customer name from what the customer said and pass them in the tool call: e.g. if they said 'table 3' pass table_number: '3'; if they said 'I'm Sarah' or 'Sarah' pass customer_name: 'Sarah'. Do not omit table_number or customer_name when you know them from the conversation—the order is saved with these fields so the restaurant knows who the order is for.",
		"Do NOT call the place-order tool in the same turn where you ask 'Shall I place that?' or 'What table number?'. Only call it after the user has replied with a confirmation (yes, go ahead, table 1, etc.). If you are still asking for confirmation or table number, wait for the user's next message—then call the tool.",
		"After a successful order tool result: say 'Your order has been placed' (or similar) exactly ONE time, then ask if they need anything else. Never say it again in the same turn or the next turn.",
	}
	if tableFromURL != "" {
		parts = append(parts, fmt.Sprintf("The customer is at table %s. When you call the place-order tool, always pass table_number: %q and pass customer_name if the customer gave their name.", tableFromURL, tableFromURL))
	} else if waitress.TablesCount > 0 {
		parts = append(parts, fmt.Sprintf("When the customer places an order, ask which table they are at (1 to %d) and optionally their name. When you call the place-order tool, always pass table_number with the table they said and customer_name if they gave it.", waitress.TablesCount))
	} else {
		parts = append(parts, "For restaurant orders: ask which table they are at and optionally their name. When you call the place-order tool, always pass table_number and customer_name (if given) so the order is saved with them.")
	}
	parts = append(parts,
		"When stating a price to the customer, always say it with the menu currency (e.g. 10 dollars, 35 euros, 6,690 naira).",
		"Do not ask the customer if the price is correct. You are the waitress; you state the item and price to confirm what they want. Ask them to confirm they want to place the order (e.g. 'Shall I put that in?' or 'Would you like me to place that?'), not whether the price is correct.",
		"Use this flow for action requests: (1) Say what they want and the price, ask 'Shall I place that?' and if needed ask for table number. (2) Stop and wait for the user's reply. (3) Only when the user has replied with confirmation (yes, go ahead, etc.) and you have table number, call the tool. Do not call the tool in step 1. After the tool returns success, say one short sentence that the order is placed and ask if they need anything else.",
		"The user ends the conversation when they choose (e.g. by clicking end call). Do not assume the conversation is over or try to end it yourself.",
		"Do not mention internal queueing, jobs, background processing, acceptance states, or implementation details to the user.",
		"If a tool call fails, explain the failure plainly and do not pretend the action happened.",
	)

	if fullContext := waitress.FullContext(); fullContext != "" {
		parts = append(parts, "Niobe context:\n"+fullContext)
	}

	if toolExecutor != nil {
		parts = append(parts, "Configured actions:\n"+tools.HumanSummary(toolExecutor.Definitions()))
	} else {
		parts = append(parts, "Configured actions:\n"+waitress.ToolsPrompt())
	}

	return strings.Join(parts, "\n\n")
}
