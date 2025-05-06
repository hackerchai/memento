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

// --- DTOs --- //

// TagResponse defines the data structure for tag API responses.
type TagResponse struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// ToResponseDTO converts a Tag entity to its TagResponse DTO representation.
func (t *Tag) ToResponseDTO() *TagResponse {
	if t == nil {
		return nil
	}
	return &TagResponse{
		Name: t.Name,
		Slug: t.Slug,
	}
}

// TagDetailResponse defines the data structure for detailed tag API responses.
type TagDetailResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToDetailResponseDTO converts a Tag entity to its TagDetailResponse DTO.
func (t *Tag) ToDetailResponseDTO() *TagDetailResponse {
	if t == nil {
		return nil
	}
	return &TagDetailResponse{
		ID:        t.ID,
		UserID:    t.UserID,
		Name:      t.Name,
		Slug:      t.Slug,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

// CreateTagRequest defines the input for creating a new tag.
type CreateTagRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
	// Slug is typically generated from Name.
}

// CreateTagRootRequest defines the input for creating a new tag by a root user for a target user.
type CreateTagRootRequest struct {
	TargetUserID uuid.UUID `json:"target_user_id" validate:"required"`
	Name         string    `json:"name" validate:"required,min=1,max=100"`
}
