package sse

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hackerchai/memento/internal/entity" // Assuming entity path
)

// EventType defines the type of SSE event.
type EventType string

const (
	EventTypeArticleProcessingComplete EventType = "article_processing_complete"
	EventTypeArticleProcessingFailed   EventType = "article_processing_failed"
	EventTypeArticleLLMSummaryComplete EventType = "article_llm_summary_complete" // Future use
	EventTypeArticleLLMTagsComplete    EventType = "article_llm_tags_complete"    // Future use
	EventTypeArticleProcessingUpdate   EventType = "article_processing_update"    // Optional: for progress updates
	EventTypeLLMChunk                  EventType = "llm_chunk"                    // For LLM streaming responses
	EventTypeError                     EventType = "error"                        // General error reporting
)

// EventData is the base structure for SSE event data.
type EventData struct {
	ArticleID uuid.UUID `json:"article_id"`
}

// ArticleProcessedData contains data for completion/failure events.
type ArticleProcessedData struct {
	EventData
	Status entity.ArticleStatus `json:"status"`          // e.g., completed, failed
	Title  string               `json:"title,omitempty"` // Send final title on completion
	Error  string               `json:"error,omitempty"` // Include error message on failure
}

// LLMSummaryData contains data for LLM summary completion.
type LLMSummaryData struct {
	EventData
	Summary string `json:"summary"`
}

// LLMTagsData contains data for LLM tag completion.
type LLMTagsData struct {
	EventData
	Tags []string `json:"tags"`
}

// LLMChunkData contains a piece of a streaming LLM response.
type LLMChunkData struct {
	EventData        // Associate chunk with an article/context
	Chunk     string `json:"chunk"`
	IsLast    bool   `json:"is_last"` // Indicate the final chunk
}

// ErrorData contains error details.
type ErrorData struct {
	EventData        // Optional: Associate error with an article if applicable
	Message   string `json:"message"`
	Code      string `json:"code,omitempty"`
}

// FormatSSEMessage formats data into an SSE message string.
func FormatSSEMessage(eventType EventType, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SSE data: %w", err)
	}
	// Format: event: <type>\ndata: <json>\n\n
	message := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(jsonData))
	return []byte(message), nil
}
