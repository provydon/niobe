package live

import (
	"context"
)

// Session is the provider-agnostic interface for a live API session.
// Implementations (GenAI, future providers) adapt their SDK to this interface.
// Dependency Inversion: proxy and handler depend on Session, not on a specific SDK.
type Session interface {
	// SendRealtimeInput sends client payload (e.g. audio, events) to the model. raw is the JSON bytes from the client.
	SendRealtimeInput(raw []byte) error
	// SendToolResponse sends tool results back to the model.
	SendToolResponse(responses []*FunctionResponse) error
	// Receive returns the next message from the model, or an error when the stream ends.
	Receive() (*ServerMessage, error)
	Close() error
}

// SessionConfig is passed to Connector.Connect. Tools is provider-specific (e.g. GenAI uses []*genai.Tool).
type SessionConfig struct {
	SystemInstruction string
	Tools             any
}

// Connector creates a new live session. Implementations: google.go, openai.go, anthropic.go.
type Connector interface {
	Connect(ctx context.Context, cfg Config, sessionCfg SessionConfig) (Session, error)
}

// Config is the minimal config to establish a session (API key, etc.).
type Config interface {
	GetAPIKey() string
	GetUseVertex() bool
}
