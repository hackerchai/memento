package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/uptrace/bun"

	"github.com/hackerchai/memento/internal/entity"
	"github.com/hackerchai/memento/pkg/xlog"
)

// CategoryRepository handles database operations for categories.
type CategoryRepository struct {
	db     *bun.DB
	logger *xlog.Logger
}

// NewCategoryRepository creates a new CategoryRepository.
func NewCategoryRepository(db *bun.DB, logger *xlog.Logger) *CategoryRepository {
	return &CategoryRepository{
		db:     db,
		logger: logger.With().Str("repository", "Category").Logger(),
	}
}

// Create inserts a new category for a specific user.
// Checks if a category with the same name already exists for the user.
// Slug is generated here if not provided.
func (r *CategoryRepository) Create(ctx context.Context, category *entity.Category) error {
	// Ensure UserID is set
	if category.UserID == uuid.Nil {
		return errors.New("user ID is required to create a category")
	}

	// Generate slug if it's empty
	if category.Slug == "" && category.Name != "" {
		category.Slug = slug.Make(category.Name)
		// Optional: Add a check here to ensure generated slug is not empty
		if category.Slug == "" {
			return errors.New("generated slug is empty, possibly due to invalid name")
		}
	} else if category.Slug == "" {
		// If slug is empty and name is also empty, it's an error
		return errors.New("cannot create category with empty name and empty slug")
	}

	// Check if name already exists for this user
	exists, err := r.db.NewSelect().
		Model((*entity.Category)(nil)).
		Where("user_id = ? AND name = ?", category.UserID, category.Name).
		Exists(ctx)
	if err != nil {
		r.logger.ErrorX(ctx).Err(err).Str("name", category.Name).Stringer("userID", category.UserID).Msg("Failed check category name existence")
		return err
	}
	if exists {
		// Consider returning the existing category ID or a specific error type
		return errors.New("category with this name already exists for the user")
	}

	// Check slug uniqueness explicitly before insert
	exists, err = r.db.NewSelect().
		Model((*entity.Category)(nil)).
		Where("user_id = ? AND slug = ?", category.UserID, category.Slug).
		Exists(ctx)
	if err != nil {
		r.logger.ErrorX(ctx).Err(err).Str("slug", category.Slug).Stringer("userID", category.UserID).Msg("Failed check category slug existence")
		return err
	}
	if exists {
		return errors.New("category with this slug already exists for the user")
	}

	_, err = r.db.NewInsert().Model(category).Exec(ctx)
	if err != nil {
		// DB level unique constraint might still catch race conditions
		r.logger.ErrorX(ctx).Err(err).Str("name", category.Name).Stringer("userID", category.UserID).Msg("Failed to insert new category")
		return err
	}
	r.logger.InfoX(ctx).Str("name", category.Name).Stringer("userID", category.UserID).Msg("Category created successfully")
	return nil
}

// FindByID retrieves a category by its ID, ensuring it belongs to the specified user.
func (r *CategoryRepository) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.Category, error) {
	category := new(entity.Category)
	err := r.db.NewSelect().
		Model(category).
		Where("id = ? AND user_id = ?", id, userID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		r.logger.ErrorX(ctx).Err(err).Stringer("id", id).Stringer("userID", userID).Msg("Failed to find category by ID")
		return nil, err
	}
	return category, nil
}

// FindBySlug retrieves a category by its slug for a specific user.
func (r *CategoryRepository) FindBySlug(ctx context.Context, slug string, userID uuid.UUID) (*entity.Category, error) {
	category := new(entity.Category)
	err := r.db.NewSelect().
		Model(category).
		Where("slug = ? AND user_id = ?", slug, userID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		r.logger.ErrorX(ctx).Err(err).Str("slug", slug).Stringer("userID", userID).Msg("Failed to find category by slug")
		return nil, err
	}
	return category, nil
}

// FindByName retrieves a category by its name for a specific user.
func (r *CategoryRepository) FindByName(ctx context.Context, name string, userID uuid.UUID) (*entity.Category, error) {
	category := new(entity.Category)
	err := r.db.NewSelect().
		Model(category).
		Where("name = ? AND user_id = ?", name, userID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		r.logger.ErrorX(ctx).Err(err).Str("name", name).Stringer("userID", userID).Msg("Failed to find category by name")
		return nil, err
	}
	return category, nil
}

// FindOrCreateByName finds a category by name for a user, or creates it if it doesn't exist.
// Slug is generated here if creating.
func (r *CategoryRepository) FindOrCreateByName(ctx context.Context, name string, userID uuid.UUID) (*entity.Category, error) {
	category, err := r.FindByName(ctx, name, userID)
	if err == nil {
		// Found existing category
		return category, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		// An unexpected error occurred during find
		return nil, err
	}

	// Category not found, create it
	newCategory := &entity.Category{
		UserID: userID,
		Name:   name,
		// Generate slug before attempting insert
		Slug: slug.Make(name),
	}

	// Handle case where generated slug is empty
	if newCategory.Slug == "" {
		return nil, errors.New("generated slug is empty for category name: " + name)
	}

	// Explicitly check slug uniqueness before insert to handle race conditions better
	exists, err := r.db.NewSelect().
		Model((*entity.Category)(nil)).
		Where("user_id = ? AND slug = ?", newCategory.UserID, newCategory.Slug).
		Exists(ctx)
	if err != nil {
		r.logger.ErrorX(ctx).Err(err).Str("slug", newCategory.Slug).Stringer("userID", newCategory.UserID).Msg("Failed check category slug existence in FindOrCreate")
		return nil, err
	}
	if exists {
		// If slug exists but name didn't, something is inconsistent, maybe log warning?
		// Or maybe another request generated the slug just now. Retry FindBySlug.
		r.logger.WarnX(ctx).Str("name", name).Str("slug", newCategory.Slug).Stringer("userID", userID).Msg("Slug collision during FindOrCreateByName, attempting to find by slug")
		return r.FindBySlug(ctx, newCategory.Slug, userID)
	}

	// Use Returning("*") to get the full object back, including ID and timestamps
	err = r.db.NewInsert().Model(newCategory).Returning("*").Scan(ctx)
	if err != nil {
		// Check for unique constraint violation errors again (DB level check)
		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "duplicate key value violates unique constraint") || strings.Contains(err.Error(), "Duplicate entry") {
			r.logger.WarnX(ctx).Str("name", name).Str("slug", newCategory.Slug).Stringer("userID", userID).Msg("DB level unique constraint hit during FindOrCreateByName, category likely created concurrently")
			// Retry finding by name or slug
			return r.FindByName(ctx, name, userID)
		}
		r.logger.ErrorX(ctx).Err(err).Str("name", newCategory.Name).Stringer("userID", newCategory.UserID).Msg("Failed to insert new category during FindOrCreate")
		return nil, err
	}
	r.logger.InfoX(ctx).Str("name", newCategory.Name).Stringer("userID", newCategory.UserID).Msg("Category created successfully during FindOrCreate")
	return newCategory, nil
}

// FindByUserID retrieves all categories for a specific user.
func (r *CategoryRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Category, error) {
	var categories []entity.Category
	err := r.db.NewSelect().
		Model(&categories).
		Where("user_id = ?", userID).
		Order("name ASC"). // Order alphabetically
		Scan(ctx)
	if err != nil {
		// sql.ErrNoRows is not an error here, just means no categories found
		if errors.Is(err, sql.ErrNoRows) {
			return []entity.Category{}, nil
		}
		r.logger.ErrorX(ctx).Err(err).Stringer("userID", userID).Msg("Failed to find categories by user ID")
		return nil, err
	}
	return categories, nil
}

// Update updates an existing category's name for a specific user.
// Slug is considered immutable and is not updated.
func (r *CategoryRepository) Update(ctx context.Context, category *entity.Category) error {
	if category.UserID == uuid.Nil || category.ID == uuid.Nil {
		return errors.New("category ID and user ID are required for update")
	}

	// Check if the new name already exists for this user (excluding the current category ID)
	exists, err := r.db.NewSelect().
		Model((*entity.Category)(nil)).
		Where("user_id = ? AND name = ? AND id != ?", category.UserID, category.Name, category.ID).
		Exists(ctx)
	if err != nil {
		r.logger.ErrorX(ctx).Err(err).Str("name", category.Name).Stringer("userID", category.UserID).Msg("Failed check category name uniqueness on update")
		return err
	}
	if exists {
		return errors.New("another category with this name already exists for the user")
	}

	// Explicitly update only the Name field.
	// UpdatedAt is handled by the DB trigger/default or Bun hook.
	res, err := r.db.NewUpdate().
		Model((*entity.Category)(nil)). // Use nil model to avoid updating all fields
		Set("name = ?", category.Name).
		Where("id = ? AND user_id = ?", category.ID, category.UserID). // Ensure ownership
		Exec(ctx)

	if err != nil {
		r.logger.ErrorX(ctx).Err(err).Stringer("id", category.ID).Stringer("userID", category.UserID).Msg("Failed to update category")
		return err
	}

	// Check if any row was actually affected
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		// This means the category with the given ID and UserID wasn't found
		return sql.ErrNoRows
	}

	r.logger.InfoX(ctx).Stringer("id", category.ID).Stringer("userID", category.UserID).Msg("Category updated successfully")
	return nil
}

// Delete removes a category by its ID, ensuring it belongs to the specified user.
func (r *CategoryRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	res, err := r.db.NewDelete().
		Model((*entity.Category)(nil)).
		Where("id = ? AND user_id = ?", id, userID). // Ensure ownership
		Exec(ctx)

	if err != nil {
		r.logger.ErrorX(ctx).Err(err).Stringer("id", id).Stringer("userID", userID).Msg("Failed to delete category by ID")
		return err
	}

	// Check if any row was actually affected
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		// This means the category with the given ID and UserID wasn't found
		return sql.ErrNoRows
	}

	r.logger.InfoX(ctx).Stringer("id", id).Stringer("userID", userID).Msg("Category deleted successfully by ID")
	return nil
}

// DeleteBySlug removes a category by its slug, ensuring it belongs to the specified user.
func (r *CategoryRepository) DeleteBySlug(ctx context.Context, slug string, userID uuid.UUID) error {
	res, err := r.db.NewDelete().
		Model((*entity.Category)(nil)).
		Where("slug = ? AND user_id = ?", slug, userID). // Ensure ownership
		Exec(ctx)

	if err != nil {
		r.logger.ErrorX(ctx).Err(err).Str("slug", slug).Stringer("userID", userID).Msg("Failed to delete category by slug")
		return err
	}

	// Check if any row was actually affected
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	r.logger.InfoX(ctx).Str("slug", slug).Stringer("userID", userID).Msg("Category deleted successfully by slug")
	return nil
}
