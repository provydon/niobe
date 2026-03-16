package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"regexp"
	"strings"
	"time"

	"niobe/agent/config"
	"niobe/agent/store"
)

// LocalNiobeTools implements tool definitions and execution in-process using the DB (no Laravel HTTP calls).
type LocalNiobeTools struct {
	waitress      *store.Waitress
	definitions   []Definition
	byName        map[string]Definition
	db            *sql.DB
	cfg           *config.Config
	tableFromURL  string // table number from talk URL (?table=5), used when model doesn't pass table_number
}

// NewLocalNiobeTools builds definitions from the Waitress and returns an executor that runs actions locally.
// tableFromURL is the table query param from the talk page (e.g. "5"); stored so we can save it on the order when the model omits it.
func NewLocalNiobeTools(waitress *store.Waitress, db *sql.DB, cfg *config.Config, tableFromURL string) *LocalNiobeTools {
	defs := BuildDefinitions(waitress)
	byName := make(map[string]Definition, len(defs))
	for _, d := range defs {
		byName[d.Name] = d
	}
	return &LocalNiobeTools{
		waitress:     waitress,
		definitions:  defs,
		byName:       byName,
		db:           db,
		cfg:          cfg,
		tableFromURL: strings.TrimSpace(tableFromURL),
	}
}

// Definitions returns the tool definitions for the LLM.
func (n *LocalNiobeTools) Definitions() []Definition {
	return n.definitions
}

// Lookup returns a definition by exact name or by niobe_N_ prefix.
func (n *LocalNiobeTools) Lookup(toolName string) (Definition, bool) {
	if d, ok := n.byName[toolName]; ok {
		return d, true
	}
	re := regexp.MustCompile(`^niobe_(\d+)_`)
	m := re.FindStringSubmatch(toolName)
	if m == nil {
		return Definition{}, false
	}
	prefix := "niobe_" + m[1] + "_"
	for _, d := range n.definitions {
		if strings.HasPrefix(d.Name, prefix) {
			return d, true
		}
	}
	return Definition{}, false
}

// Execute runs the tool locally: normalize args, validate, log to DB, send email or webhook, return result.
func (n *LocalNiobeTools) Execute(ctx context.Context, toolName string, arguments map[string]any) (CallResponse, error) {
	definition, ok := n.Lookup(toolName)
	if !ok {
		return CallResponse{Tool: toolName, Result: errorResult("Action failed", "Tool not found",
			fmt.Sprintf("The tool %q is not configured for this waitress.", toolName))}, nil
	}
	effectiveName := definition.Name
	if _, ok := n.byName[toolName]; !ok {
		toolName = effectiveName
	}

	if !definition.Enabled {
		return CallResponse{Tool: toolName, Result: errorResult("Action unavailable", definition.DisplayName,
			definition.Reason)}, nil
	}

	args := normalizeArguments(definition, arguments)
	if err := validateArguments(definition, args); err != nil {
		raw, _ := json.Marshal(arguments)
		log.Printf("[tools] validation failed: name=%s err=%v args=%s", toolName, err, raw)
		return CallResponse{Tool: toolName, Result: errorResult("Action failed", definition.DisplayName, err.Error())}, nil
	}

	// Insert action log (status queued, then we run and update)
	displayName := definition.DisplayName
	toolType := definition.Type
	target := definition.Target
	argsJSON, _ := json.Marshal(args)
	now := time.Now()

	var logID int64
	err := n.db.QueryRowContext(ctx,
		`INSERT INTO waitress_action_logs (waitress_id, tool_name, tool_type, display_name, target, status, arguments, queued_at, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, 'queued', $6, $7, $8, $8) RETURNING id`,
		n.waitress.ID, effectiveName, toolType, displayName, target, argsJSON, now, now,
	).Scan(&logID)
	if err != nil {
		log.Printf("[tools] insert action log: %v", err)
		return CallResponse{}, fmt.Errorf("failed to create action log: %w", err)
	}

	log.Printf("[tools] Waitress action queued. log_id=%d waitress_id=%d tool_name=%s", logID, n.waitress.ID, effectiveName)

	// Run synchronously
	started := time.Now()
	var runErr error
	var resultPayload map[string]any
	switch toolType {
	case "send_email":
		resultPayload, runErr = n.sendEmail(ctx, target, args)
	case "send_webhook_event":
		resultPayload, runErr = n.sendWebhook(ctx, target, displayName, args)
	default:
		runErr = fmt.Errorf("tool type %q is not implemented", toolType)
	}

	completed := time.Now()
	status := "succeeded"
	var resultJSON []byte
	var errMsg sql.NullString
	if runErr != nil {
		status = "failed"
		errMsg = sql.NullString{String: runErr.Error(), Valid: true}
	} else {
		resultJSON, _ = json.Marshal(resultPayload)
	}

	_, _ = n.db.ExecContext(ctx,
		`UPDATE waitress_action_logs SET status = $1, result = $2, error_message = $3, started_at = $4, completed_at = $5, updated_at = $5 WHERE id = $6`,
		status, resultJSON, errMsg, started, completed, logID,
	)

	if runErr != nil {
		return CallResponse{Tool: toolName, Result: errorResult("Action failed", definition.DisplayName, runErr.Error())}, nil
	}

	// Record order in orders table when an order tool succeeded
	if isOrderTool(displayName) && status == "succeeded" {
		orderSummary := strings.TrimSpace(strAny(args["body"]))
		if orderSummary == "" {
			orderSummary = strings.TrimSpace(strAny(args["order_details"]))
		}
		if orderSummary == "" {
			orderSummary = strings.TrimSpace(strAny(args["subject"]))
		}
		if orderSummary != "" {
			tableNum := strFromAny(args["table_number"])
			if tableNum == "" && n.tableFromURL != "" {
				tableNum = n.tableFromURL
			}
			customerName := strFromAny(args["customer_name"])
			var tableNumVal, customerNameVal any = nil, nil
			if tableNum != "" {
				tableNumVal = tableNum
			}
			if customerName != "" {
				customerNameVal = customerName
			}
			_, insErr := n.db.ExecContext(ctx,
				`INSERT INTO orders (waitress_id, waitress_action_log_id, order_summary, sent_to, sent_at, table_number, customer_name, created_at, updated_at)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $5, $5)`,
				n.waitress.ID, logID, orderSummary, displayName, completed, tableNumVal, customerNameVal,
			)
			if insErr != nil {
				log.Printf("[tools] insert order: %v", insErr)
			}
		}
	}

	title, message := successCopy(displayName, effectiveName, toolType, args)
	return CallResponse{
		Tool: toolName,
		Result: map[string]any{
			"ok":      true,
			"message": message,
			"output":  map[string]any{"log_id": logID, "status": status},
			"display": map[string]any{"title": title, "subtitle": "", "status": "success"},
		},
	}, nil
}

func errorResult(title, subtitle, message string) map[string]any {
	return map[string]any{
		"ok":      false,
		"message": message,
		"error":   map[string]any{"message": message},
		"display": map[string]any{"title": title, "subtitle": subtitle, "status": "error"},
	}
}

func normalizeArguments(def Definition, arguments map[string]any) map[string]any {
	if def.Type != "send_email" {
		return arguments
	}
	if !isOrderTool(def.DisplayName) {
		return arguments
	}

	// Accept camelCase from API (e.g. orderDetails, tableNumber)
	if v := arguments["orderDetails"]; v != nil && arguments["order_details"] == nil {
		arguments["order_details"] = v
	}
		if v := arguments["tableNumber"]; v != nil && arguments["table_number"] == nil {
			arguments["table_number"] = v
		}
		if v := arguments["customerName"]; v != nil && arguments["customer_name"] == nil {
			arguments["customer_name"] = v
		}

	subject := strings.TrimSpace(strAny(arguments["subject"]))
	body := strings.TrimSpace(strAny(arguments["body"]))
	orderDetails := arguments["order_details"]

	if orderDetails != nil && orderDetails != "" {
		if body == "" {
			switch v := orderDetails.(type) {
			case string:
				body = strings.TrimSpace(v)
			case map[string]any:
				b, _ := json.Marshal(v)
				body = string(b)
				if subject == "" {
					if s, ok := v["summary"].(string); ok && strings.TrimSpace(s) != "" {
						subject = strings.TrimSpace(s)
					}
				}
			case []any:
				b, _ := json.Marshal(v)
				body = string(b)
			default:
				body = fmt.Sprint(v)
			}
			if body != "" && subject == "" {
				subject = "Order request"
			}
		}
	}
	if subject == "" && body != "" {
		subject = "Order request"
	}
	// Fallback: try other keys for body (model may use different param names)
	if body == "" && subject == "" {
		for _, key := range []string{"details", "items", "text", "summary", "message", "order_summary", "order", "content", "description"} {
			if v := arguments[key]; v != nil {
				if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
					body = strings.TrimSpace(s)
					subject = "Order request"
					break
				}
				if m, ok := v.(map[string]any); ok && len(m) > 0 {
					b, _ := json.Marshal(m)
					body = string(b)
					subject = "Order request"
					break
				}
			}
		}
	}

	if subject != "" || body != "" {
		out := make(map[string]any)
		for k, v := range arguments {
			out[k] = v
		}
		if subject != "" {
			out["subject"] = subject
		}
		if body != "" {
			out["body"] = body
		}
		return out
	}
	return arguments
}

func strAny(v any) string {
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

// strFromAny returns a string from any value (number, string, etc.) so table_number/customer_name are saved even if the API sends a number.
func strFromAny(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(fmt.Sprint(v))
}

func validateArguments(def Definition, arguments map[string]any) error {
	switch def.Type {
	case "send_email":
		subject := strings.TrimSpace(strAny(arguments["subject"]))
		body := strings.TrimSpace(strAny(arguments["body"]))
		orderDetails := arguments["order_details"]
		if orderDetails == nil {
			orderDetails = arguments["orderDetails"] // camelCase from API
		}
		orderDetailsStr := ""
		if s, ok := orderDetails.(string); ok {
			orderDetailsStr = strings.TrimSpace(s)
		}
		isOrder := isOrderTool(def.DisplayName)
		if isOrder && subject == "" && body == "" {
			if orderDetails == nil || orderDetailsStr == "" {
				return fmt.Errorf("order details are required before the action can be sent: include a short summary of the order (e.g. items and total) in the order_details parameter")
			}
		}
		if subject == "" || body == "" {
			return fmt.Errorf("email subject and body are required before the action can be sent")
		}
	case "send_webhook_event":
		summary := strings.TrimSpace(strAny(arguments["summary"]))
		if summary == "" {
			return fmt.Errorf("webhook summary is required before the action can be sent")
		}
		if payload, ok := arguments["payload"]; ok && payload != nil {
			if _, ok := payload.(map[string]any); !ok {
				return fmt.Errorf("webhook payload must be a JSON object")
			}
		}
	}
	return nil
}

func successCopy(displayName, toolName, toolType string, args map[string]any) (title, message string) {
	ctx := strings.ToLower(displayName + " " + toolName + " " + strAny(args["subject"]) + " " + strAny(args["body"]) + " " + strAny(args["summary"]))
	if strings.Contains(ctx, "order") || strings.Contains(ctx, "receipt") {
		return "Order placed", "Your order has been placed."
	}
	if strings.Contains(ctx, "issue") || strings.Contains(ctx, "support") || strings.Contains(ctx, "repair") || strings.Contains(ctx, "fix") || strings.Contains(ctx, "problem") {
		return "Sent to the appropriate team", "I sent this to the appropriate team to fix."
	}
	switch toolType {
	case "send_email", "send_webhook_event":
		return "Sent to the appropriate team", "I sent this to the appropriate team."
	default:
		return "Done", "Your request has been handled."
	}
}

func (n *LocalNiobeTools) sendEmail(ctx context.Context, recipient string, args map[string]any) (map[string]any, error) {
	recipient = strings.TrimSpace(recipient)
	subject := strings.TrimSpace(strAny(args["subject"]))
	body := strings.TrimSpace(strAny(args["body"]))
	if recipient == "" || subject == "" || body == "" {
		return nil, fmt.Errorf("email target, subject, and body are required")
	}

	addr := n.cfg.MailHost + ":" + n.cfg.MailPort
	from := n.cfg.MailFrom
	if from == "" {
		from = "niobe@local"
	}
	msg := []byte("To: " + recipient + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" + body + "\r\n")

	var err error
	if n.cfg.MailUser != "" && n.cfg.MailPassword != "" {
		auth := smtp.PlainAuth("", n.cfg.MailUser, n.cfg.MailPassword, n.cfg.MailHost)
		err = smtp.SendMail(addr, auth, from, []string{recipient}, msg)
	} else {
		err = smtp.SendMail(addr, nil, from, []string{recipient}, msg)
	}
	if err != nil {
		return nil, fmt.Errorf("send mail: %w", err)
	}

	log.Printf("[tools] Waitress action email delivered. recipient=%s subject=%s", recipient, subject)
	return map[string]any{"recipient": recipient, "subject": subject, "body": body}, nil
}

func (n *LocalNiobeTools) sendWebhook(ctx context.Context, target, displayName string, args map[string]any) (map[string]any, error) {
	target = strings.TrimSpace(target)
	summary := strings.TrimSpace(strAny(args["summary"]))
	payload := args["payload"]
	if payload == nil {
		payload = map[string]any{}
	}
	payloadMap, ok := payload.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("webhook payload must be a JSON object")
	}
	if target == "" || summary == "" {
		return nil, fmt.Errorf("webhook target and summary are required")
	}

	body := map[string]any{
		"event": map[string]any{
			"type":    "niobe.action",
			"name":    displayName,
			"summary": summary,
		},
		"payload": payloadMap,
		"niobe": map[string]any{
			"id":   n.waitress.ID,
			"name": n.waitress.Name,
			"slug": n.waitress.Slug,
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("webhook returned HTTP %d", resp.StatusCode)
	}

	log.Printf("[tools] Waitress action webhook delivered. target=%s status=%d", target, resp.StatusCode)
	return map[string]any{"target": target, "summary": summary, "status": resp.StatusCode, "payload": payloadMap}, nil
}
