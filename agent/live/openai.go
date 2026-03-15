package live

import (
	"context"
	"fmt"
)

// OpenAIConnector creates live sessions via the OpenAI API (e.g. Realtime API).
// All OpenAI-specific logic belongs in this file.
// Not implemented yet: Connect returns an error.
type OpenAIConnector struct {
	// APIKey, BaseURL, Model etc. can be added when implementing.
}

// Connect implements live.Connector.
func (OpenAIConnector) Connect(ctx context.Context, cfg Config, sessionCfg SessionConfig) (Session, error) {
	return nil, fmt.Errorf("openai: live connector not implemented yet")
}
