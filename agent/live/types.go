package live

// ServerMessage is a provider-agnostic live server message (tool calls, content, transcriptions).
// Implementations (e.g. GenAI) convert from their SDK type to this.
type ServerMessage struct {
	SetupComplete  *struct{}       `json:"setupComplete,omitempty"`
	ServerContent  *ServerContent  `json:"serverContent,omitempty"`
	ToolCall       *ToolCall       `json:"toolCall,omitempty"`
}

// ServerContent holds model turn, transcriptions, turn completion, and interruption.
type ServerContent struct {
	ModelTurn            *ModelTurn     `json:"modelTurn,omitempty"`
	TurnComplete         bool           `json:"turnComplete,omitempty"`
	Interrupted          bool           `json:"interrupted,omitempty"`
	InputTranscription   *Transcription `json:"inputTranscription,omitempty"`
	OutputTranscription  *Transcription `json:"outputTranscription,omitempty"`
}

// ModelTurn contains audio/text parts from the model.
type ModelTurn struct {
	Parts []Part `json:"parts,omitempty"`
}

// Part is one part of a model turn (e.g. inline audio).
type Part struct {
	InlineData *InlineData `json:"inlineData,omitempty"`
}

// InlineData holds MIME type and base64-encoded data.
type InlineData struct {
	MimeType string `json:"mimeType,omitempty"`
	Data     []byte `json:"data,omitempty"`
}

// Transcription is user or model text.
type Transcription struct {
	Text string `json:"text,omitempty"`
}

// ToolCall is a request from the model to run tools.
type ToolCall struct {
	FunctionCalls []FunctionCall `json:"functionCalls,omitempty"`
}

// FunctionCall is one tool invocation.
type FunctionCall struct {
	ID   string         `json:"id,omitempty"`
	Name string         `json:"name,omitempty"`
	Args map[string]any `json:"args,omitempty"`
}

// FunctionResponse is the result of a tool call sent back to the model.
type FunctionResponse struct {
	ID       string         `json:"id,omitempty"`
	Name     string         `json:"name,omitempty"`
	Response map[string]any `json:"response,omitempty"`
}
