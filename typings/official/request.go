package official

type APIRequest struct {
	Messages          []api_message `json:"messages"`
	Stream            bool          `json:"stream"`
	Model             string        `json:"model"`
	Action            string        `json:"action,omitempty"`
	ConversationID    string        `json:"conversation_id,omitempty"`
	ParentMessageID   string        `json:"parent_message_id,omitempty"`
	MaxCompletionTokens int         `json:"max_completion_tokens,omitempty"`
	Prompt            string        `json:"prompt,omitempty"`     // Для continue.dev
	Stop              []string      `json:"stop,omitempty"`       // Стоп-слова для continue.dev
	Temperature       float64       `json:"temperature,omitempty"` // Температура для continue.dev
}

type api_message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type TTSAPIRequest struct {
	Input  string `json:"input"`
	Voice  string `json:"voice"`
	Format string `json:"response_format"`
}