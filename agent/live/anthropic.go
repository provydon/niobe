package live

import (
	"context"
	"fmt"
)

// AnthropicConnector creates live sessions via the Anthropic API.
// All Anthropic-specific logic belongs in this file.
// Not implemented yet: Connect returns an error.
type AnthropicConnector struct {
	// APIKey, Model etc. can be added when implementing.
}

// Connect implements live.Connector.
func (AnthropicConnector) Connect(ctx context.Context, cfg Config, sessionCfg SessionConfig) (Session, error) {
	return nil, fmt.Errorf("anthropic: live connector not implemented yet")
}
