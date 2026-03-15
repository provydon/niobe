package tools

import (
	"fmt"
	"strings"
	"unicode"

	"niobe/agent/store"
)

// BuildDefinitions returns tool definitions for the LLM from a Waitress's tools config (mirrors Laravel).
func BuildDefinitions(waitress *store.Waitress) []Definition {
	if len(waitress.Tools) == 0 {
		return nil
	}
	out := make([]Definition, 0, len(waitress.Tools))
	for i, tool := range waitress.Tools {
		def := descriptorForTool(tool, i)
		out = append(out, def)
	}
	return out
}

func descriptorForTool(tool store.Tool, index int) Definition {
	typ := strings.TrimSpace(tool.Type)
	if typ == "" {
		typ = "action"
	}
	displayName := strings.TrimSpace(tool.Name)
	if displayName == "" {
		displayName = headline(typ)
	}
	target := strings.TrimSpace(tool.Target)
	slug := snake(displayName)
	if slug == "" {
		slug = "action"
	}
	name := fmt.Sprintf("niobe_%d_%s", index+1, slug)

	switch typ {
	case "send_email":
		return sendEmailDescriptor(name, displayName, target, typ)
	case "send_webhook_event":
		return Definition{
			Name:        name,
			DisplayName: displayName,
			Type:        typ,
			Target:      target,
			Description: descriptionForTool(displayName, typ),
			Enabled:     true,
			ParametersJSONSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"summary": map[string]any{
						"type":        "string",
						"description": "Short summary of the confirmed request.",
					},
					"payload": map[string]any{
						"type":        "object",
						"description": "Structured details for the confirmed request.",
					},
				},
				"required": []string{"summary"},
			},
		}
	default:
		return Definition{
			Name:                 name,
			DisplayName:          displayName,
			Type:                 typ,
			Target:               target,
			Description:          fmt.Sprintf("Configured action %q is not executable yet.", displayName),
			Enabled:              false,
			Reason:               fmt.Sprintf("Action type %q is not implemented yet.", typ),
			ParametersJSONSchema: map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{}},
		}
	}
}

func sendEmailDescriptor(name, displayName, target, typ string) Definition {
	isOrder := isOrderTool(displayName)
	desc := Definition{
		Name:        name,
		DisplayName: displayName,
		Type:        typ,
		Target:      target,
		Description: descriptionForTool(displayName, typ),
		Enabled:     true,
	}
	if isOrder {
		desc.ParametersJSONSchema = map[string]any{
			"type":                 "object",
			"additionalProperties": true,
			"properties": map[string]any{
				"order_details": map[string]any{
					"type":        "string",
					"description": "Required. Pass the confirmed order as a single string every time you call this tool. Example: '1 Chinese Fried Rice. Total: 6,690'. Do not call the tool without this parameter or with empty arguments.",
				},
			},
			"required": []string{"order_details"},
		}
		desc.Description = desc.Description + " When calling this tool, always pass order_details with the order summary (items and total)."
		return desc
	}
	desc.ParametersJSONSchema = map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"subject": map[string]any{"type": "string", "description": "Short internal label for the request. Keep it concise."},
			"body":    map[string]any{"type": "string", "description": "Complete request details to send. Do not omit this."},
		},
		"required": []string{"subject", "body"},
	}
	return desc
}

func descriptionForTool(displayName, typ string) string {
	name := strings.ToLower(strings.TrimSpace(displayName))
	if strings.Contains(name, "order") || strings.Contains(name, "menu") ||
		strings.Contains(name, "meal") || strings.Contains(name, "food") || strings.Contains(name, "receipt") {
		return "Place the customer order after confirming the details. Use this only when the user clearly wants to place the order. You must pass order_details (a string with the order summary and total) every time you call this tool."
	}
	if strings.Contains(name, "issue") || strings.Contains(name, "support") ||
		strings.Contains(name, "repair") || strings.Contains(name, "fix") || strings.Contains(name, "problem") {
		return "Send the confirmed issue to the appropriate team. Use this only when the user clearly wants help with that request."
	}
	switch typ {
	case "send_email", "send_webhook_event":
		return "Send the confirmed request to the appropriate team after confirming the details."
	default:
		return fmt.Sprintf("Configured action %q is not executable yet.", displayName)
	}
}

func isOrderTool(displayName string) bool {
	name := strings.ToLower(strings.TrimSpace(displayName))
	keywords := []string{"order", "menu", "meal", "food", "receipt"}
	for _, k := range keywords {
		if strings.Contains(name, k) {
			return true
		}
	}
	return false
}

func headline(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	words := strings.Fields(s)
	for i, w := range words {
		if w == "" {
			continue
		}
		r := []rune(w)
		r[0] = unicode.ToUpper(r[0])
		for j := 1; j < len(r); j++ {
			r[j] = unicode.ToLower(r[j])
		}
		words[i] = string(r)
	}
	return strings.Join(words, " ")
}

func snake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r - 'A' + 'a')
		} else if r == ' ' || r == '-' {
			b.WriteByte('_')
		} else if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}
