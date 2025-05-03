package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Article represents a saved article.
type Article struct {
	bun.BaseModel `bun:"table:articles,alias:a"`

	ID             uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	UserID         uuid.UUID  `bun:"user_id,type:uuid,notnull" json:"user_id"`
	CategoryID     *uuid.UUID `bun:"category_id,type:uuid" json:"category_id,omitempty"` // Pointer to allow NULL
	Title          string     `bun:"title,notnull" json:"title"`
	Html           *string    `bun:"html" json:"html,omitempty"` // Use pointer for potentially large text fields
	Author         *string    `bun:"author" json:"author,omitempty"`
	Description    *string    `bun:"description" json:"description,omitempty"`         // Renamed from 'desc'
	LLMDescription *string    `bun:"llm_description" json:"llm_description,omitempty"` // Renamed from 'llm_desc'
	PlainText      *string    `bun:"plain_text" json:"plain_text,omitempty"`           // Renamed from 'text'
	OgImageURL     *string    `bun:"og_image_url" json:"og_image_url,omitempty"`
	URL            string     `bun:"url,notnull" json:"url"`
	IsOffline      bool       `bun:"is_offline,notnull,default:false" json:"is_offline"`
	Status         int        `bun:"status,notnull,default:0" json:"status"` // Consider defining status constants (e.g., StatusUnread, StatusRead, StatusArchived)
	CreatedAt      time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt      time.Time  `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`

	// Relationships
	// User     *User     `bun:"rel:belongs-to,join:user_id=id"`
	Category *Category `bun:"rel:belongs-to,join:category_id=id"`
	Tags     []*Tag    `bun:"m2m:article_tags,join:Article=Tag"`
}

// ArticleTag is the join table model for the many-to-many relationship
// between Articles and Tags. Bun uses this implicitly for m2m relations,
// but defining it explicitly can be useful for certain queries or operations.
type ArticleTag struct {
	bun.BaseModel `bun:"table:article_tags,alias:at"`

	ArticleID uuid.UUID `bun:"article_id,pk,type:uuid"`
	TagID     uuid.UUID `bun:"tag_id,pk,type:uuid"`
	UserID    uuid.UUID `bun:"user_id,type:uuid,notnull"` // Included from migration
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`

	// Define relationships back to Article and Tag if needed for specific queries
	// Article *Article `bun:"rel:belongs-to,join:article_id=id"`
	// Tag     *Tag     `bun:"rel:belongs-to,join:tag_id=id"`
}
