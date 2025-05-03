package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
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
func (r *CategoryRepository) Create(ctx context.Context, category *entity.Category) error {
	// Ensure UserID is set
	if category.UserID == uuid.Nil {
		return errors.New("user ID is required to create a category")
	}

	// Check if name already exists for this user
	exists, err := r.db.NewSelect().
		Model((*entity.Category)(nil)).
		Where("user_id = ? AND name = ?", category.UserID, category.Name).
		Exists(ctx)
	if err != nil {
		r.logger.ErrorX(ctx).Err(err).Str("name", category.Name).Stringer("userID", category.UserID).Msg("Failed check category existence")
		return err
	}
	if exists {
		// Consider returning the existing category ID or a specific error type
		return errors.New("category with this name already exists for the user")
	}

	_, err = r.db.NewInsert().Model(category).Exec(ctx)
	if err != nil {
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
// This is particularly useful for the "default" category logic.
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
		// ID, CreatedAt, UpdatedAt will be set by BeforeAppendModel hook or DB defaults
	}
	// Use Scan to get the newly created category back, including the generated ID
	err = r.db.NewInsert().Model(newCategory).Returning("*").Scan(ctx)
	if err != nil {
		// Handle potential race condition if another request created it between Find and Insert
		// Check for unique constraint violation errors (specific error strings might vary by DB driver)
		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "duplicate key value violates unique constraint") || strings.Contains(err.Error(), "Duplicate entry") {
			r.logger.WarnX(ctx).Str("name", name).Stringer("userID", userID).Msg("Race condition during FindOrCreateByName, category likely created concurrently")
			// Retry finding it once more
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
func (r *CategoryRepository) Update(ctx context.Context, category *entity.Category) error {
	if category.UserID == uuid.Nil || category.ID == uuid.Nil {
		return errors.New("category ID and user ID are required for update")
	}

	// Optional: Check if the new name already exists for this user (excluding the current category ID)
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

	// Only update specific fields (e.g., name and updated_at)
	// Bun automatically updates UpdatedAt if the hook/default is set correctly
	res, err := r.db.NewUpdate().
		Model(category).
		// Set("name = ?", category.Name). // Set only specific fields if needed
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
		r.logger.ErrorX(ctx).Err(err).Stringer("id", id).Stringer("userID", userID).Msg("Failed to delete category")
		return err
	}

	// Check if any row was actually affected
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		// This means the category with the given ID and UserID wasn't found
		return sql.ErrNoRows
	}

	r.logger.InfoX(ctx).Stringer("id", id).Stringer("userID", userID).Msg("Category deleted successfully")
	return nil
}
