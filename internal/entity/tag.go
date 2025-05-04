package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Tag represents a user-defined tag for articles.
type Tag struct {
	bun.BaseModel `bun:"table:tags,alias:t"`

	ID        uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `bun:"user_id,type:uuid,notnull" json:"user_id"` // Foreign key to users table
	Name      string    `bun:"name,notnull" json:"name"`
	Slug      string    `bun:"slug,notnull,unique:uk_tags_user_slug" json:"slug"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`

	// Relationships
	// User     *User      `bun:"rel:belongs-to,join:user_id=id"`
	Articles []*Article `bun:"m2m:article_tags,join:Tag=Article"` // Many-to-many relation
}
