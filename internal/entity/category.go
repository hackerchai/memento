package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Category represents a user-defined category for articles.
type Category struct {
	bun.BaseModel `bun:"table:categories,alias:cat"`

	ID        uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `bun:"user_id,type:uuid,notnull" json:"user_id"` // Foreign key to users table
	Name      string    `bun:"name,notnull" json:"name"`
	Slug      string    `bun:"slug,notnull,unique:uk_categories_user_slug" json:"slug"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`

	// Relationships (optional, depending on query needs)
	// User *User `bun:"rel:belongs-to,join:user_id=id"` // Example if User entity is defined
}

// --- DTOs --- //

// CategoryResponse defines the data structure for category API responses.
type CategoryResponse struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// ToResponseDTO converts a Category entity to its CategoryResponse DTO representation.
func (c *Category) ToResponseDTO() *CategoryResponse {
	if c == nil {
		return nil
	}
	return &CategoryResponse{
		Name: c.Name,
		Slug: c.Slug,
	}
}
