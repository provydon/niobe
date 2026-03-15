package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"google.golang.org/genai"
)

type Client struct {
	baseURL    string
	secret     string
	httpClient *http.Client
}

type Definition struct {
	Name                 string         `json:"name"`
	DisplayName          string         `json:"display_name"`
	Type                 string         `json:"type"`
	Target               string         `json:"target"`
	Description          string         `json:"description"`
	Enabled              bool           `json:"enabled"`
	Reason               string         `json:"reason"`
	ParametersJSONSchema map[string]any `json:"parameters_json_schema"`
}

type CallResponse struct {
	Tool   string         `json:"tool"`
	Result map[string]any `json:"result"`
}

type NiobeTools struct {
	client      *Client
	slug        string
	definitions []Definition
	byName      map[string]Definition
}

func NewClient(baseURL, secret string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		secret:  secret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Load(ctx context.Context, slug string) (*NiobeTools, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint(slug, "/tools"), nil)
	if err != nil {
		return nil, err
	}
	c.decorate(req)

	var response struct {
		Tools []Definition `json:"tools"`
	}

	if err := c.doJSON(req, &response); err != nil {
		return nil, err
	}

	handle := &NiobeTools{
		client:      c,
		slug:        slug,
		definitions: response.Tools,
		byName:      make(map[string]Definition, len(response.Tools)),
	}

	for _, definition := range response.Tools {
		handle.byName[definition.Name] = definition
	}

	return handle, nil
}

func (n *NiobeTools) Definitions() []Definition {
	return n.definitions
}

func (n *NiobeTools) Lookup(toolName string) (Definition, bool) {
	definition, ok := n.byName[toolName]
	return definition, ok
}

func (n *NiobeTools) Execute(ctx context.Context, toolName string, arguments map[string]any) (CallResponse, error) {
	body, err := json.Marshal(map[string]any{
		"tool":      toolName,
		"arguments": arguments,
	})
	if err != nil {
		return CallResponse{}, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		n.client.endpoint(n.slug, "/tools/call"),
		bytes.NewReader(body),
	)
	if err != nil {
		return CallResponse{}, err
	}

	n.client.decorate(req)
	req.Header.Set("Content-Type", "application/json")

	var response CallResponse
	if err := n.client.doJSON(req, &response); err != nil {
		return CallResponse{}, err
	}

	return response, nil
}

func ToGenAITools(definitions []Definition) []*genai.Tool {
	functions := make([]*genai.FunctionDeclaration, 0, len(definitions))

	for _, definition := range definitions {
		if !definition.Enabled {
			continue
		}

		functions = append(functions, &genai.FunctionDeclaration{
			Name:                 definition.Name,
			Description:          definition.Description,
			ParametersJsonSchema: definition.ParametersJSONSchema,
		})
	}

	if len(functions) == 0 {
		return nil
	}

	return []*genai.Tool{{
		FunctionDeclarations: functions,
	}}
}

func HumanSummary(definitions []Definition) string {
	if len(definitions) == 0 {
		return "No actions are configured."
	}

	lines := make([]string, 0, len(definitions))
	for _, definition := range definitions {
		line := fmt.Sprintf("- %s", capabilitySummary(definition))
		if !definition.Enabled && definition.Reason != "" {
			line += fmt.Sprintf(" (unavailable: %s)", definition.Reason)
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func capabilitySummary(definition Definition) string {
	name := strings.ToLower(strings.TrimSpace(definition.DisplayName))
	if strings.Contains(name, "order") {
		return "Can place an order after confirming the details"
	}
	if strings.Contains(name, "issue") || strings.Contains(name, "support") || strings.Contains(name, "repair") || strings.Contains(name, "fix") {
		return "Can send the request to the appropriate team after confirming the details"
	}

	switch definition.Type {
	case "send_email", "send_webhook_event":
		return "Can send the request to the appropriate team after confirming the details"
	default:
		return definition.DisplayName
	}
}

func (c *Client) endpoint(slug, suffix string) string {
	return fmt.Sprintf("%s/api/agent/niobes/%s%s", c.baseURL, url.PathEscape(slug), suffix)
}

func (c *Client) decorate(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	if c.secret != "" {
		req.Header.Set("X-Agent-Secret", c.secret)
	}
}

func (c *Client) doJSON(req *http.Request, target any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		var apiErr struct {
			Message string `json:"message"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&apiErr)
		if strings.TrimSpace(apiErr.Message) != "" {
			return fmt.Errorf("tool api: %s", apiErr.Message)
		}
		return fmt.Errorf("tool api returned status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return err
	}

	return nil
}
