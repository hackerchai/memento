package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// ArticleStatus defines the processing status of an article
type ArticleStatus int

const (
	StatusPending       ArticleStatus = 0  // Initial state or processing
	StatusCompleted     ArticleStatus = 1  // Successfully processed
	StatusFailed        ArticleStatus = -1 // Processing failed
	StatusLLMSummarized ArticleStatus = 2  // Successfully summarized by LLM
	StatusAutoTagged    ArticleStatus = 3  // Successfully auto-tagged by LLM
	// Add other statuses like StatusArchived if needed
)

// Article represents a saved article.
type Article struct {
	bun.BaseModel `bun:"table:articles,alias:a"`

	ID             uuid.UUID     `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	UserID         uuid.UUID     `bun:"user_id,type:uuid,notnull" json:"user_id"`
	CategoryID     *uuid.UUID    `bun:"category_id,type:uuid" json:"category_id,omitempty"` // Pointer to allow NULL
	Title          string        `bun:"title,notnull" json:"title"`
	Html           *string       `bun:"html" json:"html,omitempty"` // Use pointer for potentially large text fields
	Author         *string       `bun:"author" json:"author,omitempty"`
	Description    *string       `bun:"description" json:"description,omitempty"`         // Renamed from 'desc'
	LLMDescription *string       `bun:"llm_description" json:"llm_description,omitempty"` // Renamed from 'llm_desc'
	PlainText      *string       `bun:"plain_text" json:"plain_text,omitempty"`           // Renamed from 'text'
	OgImageURL     *string       `bun:"og_image_url" json:"og_image_url,omitempty"`
	URL            string        `bun:"url,notnull" json:"url"`
	IsOffline      bool          `bun:"is_offline,notnull,default:false" json:"is_offline"`
	Status         ArticleStatus `bun:"status,notnull,default:0" json:"status"` // Use the defined type
	CreatedAt      time.Time     `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt      time.Time     `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`

	// Relationships
	// User     *User     `bun:"rel:belongs-to,join:user_id=id"`
	Category *Category `bun:"rel:belongs-to,join:category_id=id"`
	Tags     []*Tag    `bun:"m2m:article_tags,join:Article=Tag"`
}

// ArticleTag is the join table model for the many-to-many relationship
// between Articles and Tags. Bun uses this implicitly for m2m relations,
// but defining it explicitly can be useful for certain queries or operations.
// It's also required by Bun's default m2m mapping logic to have fields
// corresponding to the related models.
type ArticleTag struct {
	bun.BaseModel `bun:"table:article_tags,alias:at"`

	ArticleID uuid.UUID `bun:"article_id,pk,type:uuid"`
	TagID     uuid.UUID `bun:"tag_id,pk,type:uuid"`
	UserID    uuid.UUID `bun:"user_id,type:uuid,notnull"` // Included from migration
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`

	// Define relationships back to Article and Tag for Bun's m2m mapping
	Article *Article `bun:"rel:belongs-to,join:article_id=id"`
	Tag     *Tag     `bun:"rel:belongs-to,join:tag_id=id"`
}

// --- DTOs and Mapping --- //

// ArticleResponse defines the data structure for article API responses.
// It omits the bun.BaseModel and potentially large fields for clarity.
type ArticleResponse struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	CategoryID *uuid.UUID `json:"category_id,omitempty"`
	URL        string     `json:"url"`
	Title      string     `json:"title"`
	// Html        *string              `json:"-"` // Explicitly exclude HTML from JSON/Swagger by default
	Author      *string `json:"author,omitempty"`
	Description *string `json:"description,omitempty"`
	// PlainText   *string              `json:"-"` // Explicitly exclude PlainText
	LLMDescription *string       `json:"llm_description,omitempty"`
	OgImageURL     *string       `json:"og_image_url,omitempty"`
	IsOffline      bool          `json:"is_offline"`
	Status         ArticleStatus `json:"status"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	// TODO: Consider adding simplified CategoryName and TagNames fields if needed
	// CategoryName string   `json:"category_name,omitempty"`
	// Tags         []string `json:"tags,omitempty"`
}

// ToResponseDTO converts an Article entity to its ArticleResponse DTO representation.
func (a *Article) ToResponseDTO() *ArticleResponse {
	if a == nil {
		return nil
	}
	return &ArticleResponse{
		ID:             a.ID,
		UserID:         a.UserID,
		CategoryID:     a.CategoryID,
		URL:            a.URL,
		Title:          a.Title,
		Author:         a.Author,
		Description:    a.Description,
		LLMDescription: a.LLMDescription,
		OgImageURL:     a.OgImageURL,
		IsOffline:      a.IsOffline,
		Status:         a.Status,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
		// Map Category/Tags here if added to DTO and loaded in the entity
	}
}
