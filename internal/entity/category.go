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

// CategoryDetailResponse defines the data structure for detailed category API responses.
type CategoryDetailResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToDetailResponseDTO converts a Category entity to its CategoryDetailResponse DTO.
func (c *Category) ToDetailResponseDTO() *CategoryDetailResponse {
	if c == nil {
		return nil
	}
	return &CategoryDetailResponse{
		ID:        c.ID,
		UserID:    c.UserID,
		Name:      c.Name,
		Slug:      c.Slug,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// CreateCategoryRequest defines the input for creating a new category.
type CreateCategoryRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
	// Slug is typically generated from Name.
}

// CreateCategoryRootRequest defines the input for creating a new category by a root user for a target user.
type CreateCategoryRootRequest struct {
	TargetUserID uuid.UUID `json:"target_user_id" validate:"required"`
	Name         string    `json:"name" validate:"required,min=1,max=100"`
}
