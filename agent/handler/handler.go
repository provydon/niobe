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

			if len(waitress.Tools) > 0 {
				toolExecutor = tools.NewLocalNiobeTools(waitress, db, appCfg)
				sessionCfg.Tools = tools.ToGenAITools(toolExecutor.Definitions())
			}

			sessionCfg.SystemInstruction = buildNiobeInstruction(waitress, toolExecutor)
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

func buildNiobeInstruction(waitress *store.Waitress, toolExecutor *tools.LocalNiobeTools) string {
	parts := []string{
		strings.TrimSpace(systemInstructionPreamble),
		fmt.Sprintf("Niobe name: %s", waitress.Name),
		"The Niobe name above is your own assistant name and identity, not the user's name.",
		"Do not address the user as " + waitress.Name + " unless the user explicitly tells you that is their name.",
		"If you do not know the user's name, do not guess one. Address them naturally without using a name.",
		"CRITICAL - Actions and tools: When the user confirms an action (e.g. says Yes, correct, go ahead), you MUST call the matching tool immediately. Do NOT say the action is done, placed, or completed until you have called the tool and received its success response. Never say phrases like 'order has been placed', 'I have placed your order', 'done', or 'your order has been sent' before you have actually invoked the tool. Step 1: User confirms. Step 2: You call the tool (no success wording yet). Step 3: You get the tool result. Step 4: Only then say it is done.",
		"For place-order (or any order) tools: every time you call the tool you MUST pass the order_details parameter with a string that summarizes the confirmed order (e.g. order_details: '1 Chinese Fried Rice. Total: 6,690'). Never call the tool with empty arguments {} or the action will fail.",
		"When stating a price to the customer, always say it with the menu currency (e.g. 10 dollars, 35 euros, 6,690 naira).",
		"Use this flow for action requests: confirm the details, wait for user confirmation (Yes, etc.), then call the tool with order_details set to the order summary, then after the tool returns success tell the user the request is done and ask if they need anything else.",
		"The user ends the conversation when they choose (e.g. by clicking end call). Do not assume the conversation is over or try to end it yourself.",
		"Do not mention internal queueing, jobs, background processing, acceptance states, or implementation details to the user.",
		"If a tool call fails, explain the failure plainly and do not pretend the action happened.",
	}

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
