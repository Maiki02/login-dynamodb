package openai

type CreateThreadAndRunResponse struct {
	ID                  string             `json:"id"`
	Object              string             `json:"object"`
	CreatedAt           int                `json:"created_at"`
	AssistantID         string             `json:"assistant_id"`
	ThreadID            string             `json:"thread_id"`
	Status              string             `json:"status"`
	StartedAt           any                `json:"started_at"`
	ExpiresAt           int                `json:"expires_at"`
	CancelledAt         any                `json:"cancelled_at"`
	FailedAt            any                `json:"failed_at"`
	CompletedAt         any                `json:"completed_at"`
	RequiredAction      any                `json:"required_action"`
	LastError           any                `json:"last_error"`
	Model               string             `json:"model"`
	Instructions        string             `json:"instructions"`
	Tools               []any              `json:"tools"`
	ToolResources       ToolResources      `json:"tool_resources"`
	Metadata            Metadata           `json:"metadata"`
	Temperature         float64            `json:"temperature"`
	TopP                float64            `json:"top_p"`
	MaxCompletionTokens any                `json:"max_completion_tokens"`
	MaxPromptTokens     any                `json:"max_prompt_tokens"`
	TruncationStrategy  TruncationStrategy `json:"truncation_strategy"`
	IncompleteDetails   any                `json:"incomplete_details"`
	Usage               any                `json:"usage"`
	ResponseFormat      string             `json:"response_format"`
	ToolChoice          string             `json:"tool_choice"`
	ParallelToolCalls   bool               `json:"parallel_tool_calls"`
}

type CreateMessageResponse struct {
	ID          string    `json:"id"`
	Object      string    `json:"object"`
	CreatedAt   int       `json:"created_at"`
	AssistantID string    `json:"assistant_id"`
	ThreadID    string    `json:"thread_id"`
	RunID       string    `json:"run_id"`
	Role        string    `json:"role"`
	Content     []Content `json:"content"`
	Attachments []any     `json:"attachments"`
	Metadata    Metadata  `json:"metadata"`
}

type CreateRunResponse struct {
	ID                  string             `json:"id"`
	Object              string             `json:"object"`
	CreatedAt           int                `json:"created_at"`
	AssistantID         string             `json:"assistant_id"`
	ThreadID            string             `json:"thread_id"`
	Status              string             `json:"status"`
	StartedAt           int                `json:"started_at"`
	ExpiresAt           any                `json:"expires_at"`
	CancelledAt         any                `json:"cancelled_at"`
	FailedAt            any                `json:"failed_at"`
	CompletedAt         int                `json:"completed_at"`
	LastError           any                `json:"last_error"`
	Model               string             `json:"model"`
	Instructions        any                `json:"instructions"`
	IncompleteDetails   any                `json:"incomplete_details"`
	Tools               []Tools            `json:"tools"`
	Metadata            Metadata           `json:"metadata"`
	Usage               any                `json:"usage"`
	Temperature         float64            `json:"temperature"`
	TopP                float64            `json:"top_p"`
	MaxPromptTokens     int                `json:"max_prompt_tokens"`
	MaxCompletionTokens int                `json:"max_completion_tokens"`
	TruncationStrategy  TruncationStrategy `json:"truncation_strategy"`
	ResponseFormat      string             `json:"response_format"`
	ToolChoice          string             `json:"tool_choice"`
	ParallelToolCalls   bool               `json:"parallel_tool_calls"`
}

type GetMessageListResponse struct {
	Object  string                  `json:"object"`
	Data    []CreateMessageResponse `json:"data"`
	FirstID string                  `json:"first_id"`
	LastID  string                  `json:"last_id"`
	HasMore bool                    `json:"has_more"`
}

type DeleteThreadResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

type ResponseGPT struct {
	Type     string    `json:"type"`
	Order    *OrderGPT `json:"order,omitempty"`
	Products []int32   `json:"products,omitempty"`
	Response string    `json:"response"`
}

type OrderGPT struct {
	DeliveryMethod     string  `json:"delivery_method"`
	Address            string  `json:"address,omitempty"`
	Products           []int32 `json:"products,omitempty"`
	IsConfirm          bool    `json:"is_confirm"`
	ClientObservations string  `json:"client_observations,omitempty"`
	PaymentMethod      int32   `json:"payment_method"`
}

//--------- Auxiliar Functions ---------\\

type Text struct {
	Value       string `json:"value"`
	Annotations []any  `json:"annotations"`
}
type Content struct {
	Type string `json:"type"`
	Text Text   `json:"text"`
}

type Tools struct {
	Type string `json:"type"`
}
type Metadata struct {
}
type TruncationStrategy struct {
	Type         string `json:"type"`
	LastMessages any    `json:"last_messages"`
}

type ToolResources struct {
}
