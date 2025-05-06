package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/uptrace/bun"

	"github.com/hackerchai/memento/internal/entity"
	"github.com/hackerchai/memento/pkg/xlog"
)

// TagRepository handles database operations for tags.
type TagRepository struct {
	db     *bun.DB
	logger *xlog.Logger
}

// NewTagRepository creates a new TagRepository.
func NewTagRepository(db *bun.DB, logger *xlog.Logger) *TagRepository {
	return &TagRepository{
		db:     db,
		logger: logger.With().Str("repository", "Tag").Logger(),
	}
}

// Create inserts a new tag for a specific user.
// Checks if a tag with the same name already exists for the user.
// Slug is generated here if not provided.
func (r *TagRepository) Create(ctx context.Context, tag *entity.Tag) error {
	if tag.UserID == uuid.Nil {
		return errors.New("user ID is required to create a tag")
	}

	// Generate slug if it's empty
	if tag.Slug == "" && tag.Name != "" {
		tag.Slug = slug.Make(tag.Name)
		if tag.Slug == "" {
			return errors.New("generated slug is empty, possibly due to invalid name")
		}
	} else if tag.Slug == "" {
		return errors.New("cannot create tag with empty name and empty slug")
	}

	// Check for name uniqueness first
	exists, err := r.db.NewSelect().
		Model((*entity.Tag)(nil)).
		Where("user_id = ? AND name = ?", tag.UserID, tag.Name).
		Exists(ctx)
	if err != nil {
		r.logger.ErrorX(ctx).Err(err).Str("name", tag.Name).Stringer("userID", tag.UserID).Msg("Failed check tag name existence")
		return err
	}
	if exists {
		return errors.New("tag with this name already exists for the user")
	}

	// Check slug uniqueness explicitly before insert
	exists, err = r.db.NewSelect().
		Model((*entity.Tag)(nil)).
		Where("user_id = ? AND slug = ?", tag.UserID, tag.Slug).
		Exists(ctx)
	if err != nil {
		r.logger.ErrorX(ctx).Err(err).Str("slug", tag.Slug).Stringer("userID", tag.UserID).Msg("Failed check tag slug existence")
		return err
	}
	if exists {
		return errors.New("tag with this slug already exists for the user")
	}

	_, err = r.db.NewInsert().Model(tag).Exec(ctx)
	if err != nil {
		// DB level unique constraint might still catch race conditions
		r.logger.ErrorX(ctx).Err(err).Str("name", tag.Name).Stringer("userID", tag.UserID).Msg("Failed to insert new tag")
		return err
	}
	r.logger.InfoX(ctx).Str("name", tag.Name).Stringer("userID", tag.UserID).Msg("Tag created successfully")
	return nil
}

// FindByID retrieves a tag by its ID, ensuring it belongs to the specified user.
func (r *TagRepository) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.Tag, error) {
	tag := new(entity.Tag)
	err := r.db.NewSelect().
		Model(tag).
		Where("id = ? AND user_id = ?", id, userID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		r.logger.ErrorX(ctx).Err(err).Stringer("id", id).Stringer("userID", userID).Msg("Failed to find tag by ID")
		return nil, err
	}
	return tag, nil
}

// FindBySlug retrieves a tag by its slug for a specific user.
func (r *TagRepository) FindBySlug(ctx context.Context, slug string, userID uuid.UUID) (*entity.Tag, error) {
	tag := new(entity.Tag)
	err := r.db.NewSelect().
		Model(tag).
		Where("slug = ? AND user_id = ?", slug, userID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		r.logger.ErrorX(ctx).Err(err).Str("slug", slug).Stringer("userID", userID).Msg("Failed to find tag by slug")
		return nil, err
	}
	return tag, nil
}

// FindByName retrieves a tag by its name for a specific user.
func (r *TagRepository) FindByName(ctx context.Context, name string, userID uuid.UUID) (*entity.Tag, error) {
	tag := new(entity.Tag)
	err := r.db.NewSelect().
		Model(tag).
		Where("name = ? AND user_id = ?", name, userID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		r.logger.ErrorX(ctx).Err(err).Str("name", name).Stringer("userID", userID).Msg("Failed to find tag by name")
		return nil, err
	}
	return tag, nil
}

// FindOrCreateByName finds tags by name for a user, creating them if they don't exist.
// Slug is generated here if creating.
// Returns a map of name to Tag entity.
func (r *TagRepository) FindOrCreateByName(ctx context.Context, names []string, userID uuid.UUID) (map[string]*entity.Tag, error) {
	if len(names) == 0 {
		return map[string]*entity.Tag{}, nil
	}

	resultTags := make(map[string]*entity.Tag)
	var tagsToCreate []*entity.Tag
	existingTags := []*entity.Tag{}

	// Find existing tags first
	err := r.db.NewSelect().Model(&existingTags).
		Where("name IN (?) AND user_id = ?", bun.In(names), userID).
		Scan(ctx)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		r.logger.ErrorX(ctx).Err(err).Strs("names", names).Stringer("userID", userID).Msg("Failed to find existing tags by name")
		return nil, err
	}

	// Populate map with existing tags and identify names to create
	existingNames := make(map[string]bool)
	for _, tag := range existingTags {
		resultTags[tag.Name] = tag
		existingNames[tag.Name] = true
	}

	for _, name := range names {
		if !existingNames[name] {
			// Generate slug before adding to the create list
			generatedSlug := slug.Make(name)
			if generatedSlug == "" {
				// Handle or log error for tags with names that produce empty slugs
				r.logger.WarnX(ctx).Str("name", name).Msg("Skipping tag creation due to empty generated slug")
				continue // Skip this tag
			}
			tagsToCreate = append(tagsToCreate, &entity.Tag{Name: name, UserID: userID, Slug: generatedSlug})
		}
	}

	// Create non-existing tags
	if len(tagsToCreate) > 0 {
		// Before bulk insert, maybe check slug uniqueness for the batch?
		// This is complex due to potential race conditions. Relying on DB constraint for now.

		_, err = r.db.NewInsert().Model(&tagsToCreate).Returning("*").Exec(ctx)
		if err != nil {
			// Simplified error handling for batch insert, might need refinement for race conditions
			r.logger.ErrorX(ctx).Err(err).Int("create_count", len(tagsToCreate)).Stringer("userID", userID).Msg("Failed to bulk insert new tags during FindOrCreate")
			// If uniqueness constraint fails (name or slug), some tags might have been created concurrently.
			// A more robust solution might involve retrying the find for the failed names.
			return nil, err
		}
		// Add newly created tags to the result map
		for _, tag := range tagsToCreate {
			if tag.ID != uuid.Nil { // Check if Returning worked (might depend on DB/driver)
				resultTags[tag.Name] = tag
			} else {
				// Fallback if Returning("*") didn't populate the struct (e.g., some drivers/versions)
				// We might need to fetch them again here, or rely on the caller to handle potential missing IDs
				r.logger.WarnX(ctx).Str("tag_name", tag.Name).Msg("Tag ID not returned after bulk insert, might need refetch")
			}
		}
	}

	return resultTags, nil
}

// FindByUserID retrieves all tags for a specific user with pagination.
// This new version supports pagination and returns a total count.
func (r *TagRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]entity.Tag, int, error) {
	var tags []entity.Tag
	query := r.db.NewSelect().
		Model(&tags).
		Where("user_id = ?", userID).
		Order("name ASC"). // Or created_at DESC etc. name ASC is common for tags
		Limit(limit).
		Offset(offset)

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []entity.Tag{}, 0, nil // No error, just an empty list
		}
		r.logger.ErrorX(ctx).Err(err).Stringer("userID", userID).Msg("Failed to find tags by user ID")
		return nil, 0, err
	}
	return tags, count, nil
}

// Update updates an existing tag's name for a specific user.
// Slug is considered immutable and is not updated.
func (r *TagRepository) Update(ctx context.Context, tag *entity.Tag) error {
	if tag.UserID == uuid.Nil || tag.ID == uuid.Nil {
		return errors.New("tag ID and user ID are required for update")
	}

	// Check if the new name already exists for this user (excluding the current tag ID)
	exists, err := r.db.NewSelect().
		Model((*entity.Tag)(nil)).
		Where("user_id = ? AND name = ? AND id != ?", tag.UserID, tag.Name, tag.ID).
		Exists(ctx)
	if err != nil {
		r.logger.ErrorX(ctx).Err(err).Str("name", tag.Name).Stringer("userID", tag.UserID).Msg("Failed check tag name uniqueness on update")
		return err
	}
	if exists {
		return errors.New("another tag with this name already exists for the user")
	}

	// Explicitly update only allowed fields (Name).
	// UpdatedAt is handled by the DB trigger/default or Bun hook.
	res, err := r.db.NewUpdate().
		Model((*entity.Tag)(nil)). // Use nil model to avoid updating all fields
		Set("name = ?", tag.Name).
		// Bun might automatically add SET updated_at = NOW() or similar based on hooks/defaults
		Where("id = ? AND user_id = ?", tag.ID, tag.UserID).
		Exec(ctx)
	if err != nil {
		r.logger.ErrorX(ctx).Err(err).Stringer("id", tag.ID).Stringer("userID", tag.UserID).Msg("Failed to update tag")
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	r.logger.InfoX(ctx).Stringer("id", tag.ID).Stringer("userID", tag.UserID).Msg("Tag updated successfully")
	return nil
}

// Delete removes a tag by its ID, ensuring it belongs to the specified user.
func (r *TagRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	res, err := r.db.NewDelete().
		Model((*entity.Tag)(nil)).
		Where("id = ? AND user_id = ?", id, userID).
		Exec(ctx)
	if err != nil {
		r.logger.ErrorX(ctx).Err(err).Stringer("id", id).Stringer("userID", userID).Msg("Failed to delete tag by ID")
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	r.logger.InfoX(ctx).Stringer("id", id).Stringer("userID", userID).Msg("Tag deleted successfully by ID")
	return nil
}

// DeleteBySlug removes a tag by its slug, ensuring it belongs to the specified user.
func (r *TagRepository) DeleteBySlug(ctx context.Context, slug string, userID uuid.UUID) error {
	res, err := r.db.NewDelete().
		Model((*entity.Tag)(nil)).
		Where("slug = ? AND user_id = ?", slug, userID).
		Exec(ctx)
	if err != nil {
		r.logger.ErrorX(ctx).Err(err).Str("slug", slug).Stringer("userID", userID).Msg("Failed to delete tag by slug")
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	r.logger.InfoX(ctx).Str("slug", slug).Stringer("userID", userID).Msg("Tag deleted successfully by slug")
	return nil
}

// FindByNames finds existing tags by their names for a specific user.
func (r *TagRepository) FindByNames(ctx context.Context, names []string, userID uuid.UUID) ([]*entity.Tag, error) {
	if len(names) == 0 {
		return []*entity.Tag{}, nil
	}

	var tags []*entity.Tag
	err := r.db.NewSelect().
		Model(&tags).
		Where("name IN (?)", bun.In(names)).
		Where("user_id = ?", userID).
		Scan(ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) { // ErrNoRows is not an error, just means none found
		return nil, err
	}
	return tags, nil
}

// DeleteTagAndAssociations deletes a tag and all its associations from article_tags.
// This operation is performed in a transaction.
func (r *TagRepository) DeleteTagAndAssociations(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// 1. Delete associations from article_tags
		// We use the tag's ID. Since tags are user-specific, this implicitly handles user scoping for associations.
		// The user_id in article_tags is the article owner, not necessarily the tag creator for general system, but here tags are user-scoped.
		_, err := tx.NewDelete().
			Model((*entity.ArticleTag)(nil)). // Assuming ArticleTag is the model for article_tags join table
			Where("tag_id = ?", id).
			// If article_tags also has a user_id that needs to be checked (e.g. user who made the association)
			// and it's different from the tag's user_id scope, add: .Where("user_id = ?", userID)
			// For now, assume deleting a user's tag removes all its links.
			Exec(ctx)
		if err != nil {
			r.logger.ErrorX(ctx).Err(err).Stringer("tagID", id).Msg("Failed to delete article-tag associations")
			return fmt.Errorf("failed to delete article-tag associations: %w", err)
		}

		// 2. Delete the tag itself
		res, err := tx.NewDelete().
			Model((*entity.Tag)(nil)).
			Where("id = ? AND user_id = ?", id, userID).
			Exec(ctx)
		if err != nil {
			r.logger.ErrorX(ctx).Err(err).Stringer("id", id).Stringer("userID", userID).Msg("Failed to delete tag")
			return fmt.Errorf("failed to delete tag: %w", err)
		}

		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			return sql.ErrNoRows // Tag not found or not owned by user
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.WarnX(ctx).Stringer("id", id).Stringer("userID", userID).Msg("Tag not found for deletion")
			return sql.ErrNoRows
		}
		r.logger.ErrorX(ctx).Err(err).Stringer("id", id).Stringer("userID", userID).Msg("Transaction failed for deleting tag and associations")
		return err
	}

	r.logger.InfoX(ctx).Stringer("id", id).Stringer("userID", userID).Msg("Tag and associations deleted successfully")
	return nil
}

// FindByIDOrSlug finds a tag by its ID (if identifier is a UUID) or slug for a specific user.
func (r *TagRepository) FindByIDOrSlug(ctx context.Context, identifier string, userID uuid.UUID) (*entity.Tag, error) {
	parsedID, err := uuid.Parse(identifier)
	if err == nil {
		// Identifier is a valid UUID, try finding by ID
		return r.FindByID(ctx, parsedID, userID)
	}
	// Identifier is not a UUID, try finding by slug
	return r.FindBySlug(ctx, identifier, userID) // Assuming FindBySlug exists and is correct
}

// FindByIDRegardlessOfUser retrieves a tag by its ID, without user scoping (Root operation).
func (r *TagRepository) FindByIDRegardlessOfUser(ctx context.Context, tagID uuid.UUID) (*entity.Tag, error) {
	tag := new(entity.Tag)
	err := r.db.NewSelect().
		Model(tag).
		Where("id = ?", tagID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		r.logger.ErrorX(ctx).Err(err).Stringer("id", tagID).Msg("[Root] Failed to find tag by ID regardless of user")
		return nil, err
	}
	return tag, nil
}

// FindBySlugRegardlessOfUser retrieves a tag by its slug, without user scoping (Root operation).
func (r *TagRepository) FindBySlugRegardlessOfUser(ctx context.Context, slug string) (*entity.Tag, error) {
	tag := new(entity.Tag)
	err := r.db.NewSelect().
		Model(tag).
		Where("slug = ?", slug).
		Limit(1). // Take the first one if multiple users have the same slug
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		r.logger.ErrorX(ctx).Err(err).Str("slug", slug).Msg("[Root] Failed to find tag by slug regardless of user")
		return nil, err
	}
	return tag, nil
}

// SearchByNameOrSlugForUser searches tags by name or slug for a specific user, with pagination.
// It performs a case-insensitive LIKE search.
func (r *TagRepository) SearchByNameOrSlugForUser(ctx context.Context, userID uuid.UUID, query string, limit, offset int) ([]entity.Tag, int, error) {
	var tags []entity.Tag
	searchPattern := "%%" + strings.ToLower(query) + "%%" // Prepare for case-insensitive LIKE

	selectQuery := r.db.NewSelect().
		Model(&tags).
		Where("user_id = ?", userID).
		WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("LOWER(name) LIKE ?", searchPattern). // Case-insensitive search for name
										WhereOr("LOWER(slug) LIKE ?", searchPattern) // Case-insensitive search for slug
		}).
		OrderExpr("CASE WHEN LOWER(name) = LOWER(?) THEN 0 WHEN LOWER(slug) = LOWER(?) THEN 1 WHEN LOWER(name) LIKE LOWER(?) THEN 2 WHEN LOWER(slug) LIKE LOWER(?) THEN 3 ELSE 4 END, name ASC", query, query, query+"%%", query+"%%"). // Prioritize exact matches, then prefix matches
		Limit(limit).
		Offset(offset)

	count, err := selectQuery.ScanAndCount(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []entity.Tag{}, 0, nil
		}
		r.logger.ErrorX(ctx).Err(err).Stringer("userID", userID).Str("query", query).Msg("Failed to search tags by name or slug for user")
		return nil, 0, err
	}
	return tags, count, nil
}
