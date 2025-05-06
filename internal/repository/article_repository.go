package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/hackerchai/memento/internal/entity"
	"github.com/hackerchai/memento/pkg/xlog"
)

// ArticleRepository handles database operations for articles.
type ArticleRepository struct {
	db     *bun.DB
	logger *xlog.Logger
}

// NewArticleRepository creates a new ArticleRepository.
func NewArticleRepository(db *bun.DB, logger *xlog.Logger) *ArticleRepository {
	return &ArticleRepository{
		db:     db,
		logger: logger.With().Str("repository", "Article").Logger(),
	}
}

// Create inserts a new article for a specific user, optionally associating tags.
// It checks if an article with the same URL already exists for the user.
func (r *ArticleRepository) Create(ctx context.Context, article *entity.Article, tags []*entity.Tag) error {
	if article.UserID == uuid.Nil {
		return errors.New("user ID is required to create an article")
	}
	if article.URL == "" {
		return errors.New("article URL is required")
	}

	// Check if URL already exists for this user
	exists, err := r.db.NewSelect().
		Model((*entity.Article)(nil)).
		Where("user_id = ? AND url = ?", article.UserID, article.URL).
		Exists(ctx)
	if err != nil {
		r.logger.ErrorX(ctx).Err(err).Str("url", article.URL).Stringer("userID", article.UserID).Msg("Failed check article existence by URL")
		return err
	}
	if exists {
		// Consider returning the existing article ID or a specific error type
		return errors.New("article with this URL already exists for the user")
	}

	// Use transaction to insert article and associate tags
	txErr := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Insert the article
		_, err := tx.NewInsert().Model(article).Exec(ctx)
		if err != nil {
			r.logger.ErrorX(ctx).Err(err).Str("url", article.URL).Stringer("userID", article.UserID).Msg("Failed to insert new article in transaction")
			return err
		}

		// Associate tags if provided
		if len(tags) > 0 {
			items := make([]entity.ArticleTag, len(tags))
			for i, tag := range tags {
				if tag.ID == uuid.Nil {
					// This should ideally not happen if tags are properly fetched/created before calling Create
					r.logger.ErrorX(ctx).Str("tag_name", tag.Name).Msg("Attempted to associate tag with nil ID")
					return errors.New("cannot associate tag with nil ID")
				}
				items[i] = entity.ArticleTag{
					ArticleID: article.ID,
					TagID:     tag.ID,
					UserID:    article.UserID, // Include UserID in join table
				}
			}
			_, err = tx.NewInsert().Model(&items).Exec(ctx)
			if err != nil {
				r.logger.ErrorX(ctx).Err(err).Stringer("articleID", article.ID).Msg("Failed to insert article tags association in transaction")
				return err
			}
		}
		return nil
	})

	if txErr != nil {
		// Handle potential race condition if unique constraint on URL fails during insert
		if strings.Contains(txErr.Error(), "UNIQUE constraint failed") || strings.Contains(txErr.Error(), "duplicate key value violates unique constraint") || strings.Contains(txErr.Error(), "Duplicate entry") {
			r.logger.WarnX(ctx).Str("url", article.URL).Stringer("userID", article.UserID).Msg("Race condition on article insert, likely created concurrently")
			// Return the specific error indicating duplication
			return errors.New("article with this URL already exists for the user")
		}
		return txErr // Return other transaction errors
	}

	r.logger.InfoX(ctx).Stringer("articleID", article.ID).Stringer("userID", article.UserID).Msg("Article created successfully")
	return nil
}

// FindByID retrieves an article by its ID for a specific user.
// Optionally loads relationships (Category, Tags).
func (r *ArticleRepository) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID, loadRelations bool) (*entity.Article, error) {
	article := new(entity.Article)
	query := r.db.NewSelect().
		Model(article).
		Where("a.id = ? AND a.user_id = ?", id, userID) // Use alias 'a' defined in model

	if loadRelations {
		query = query.Relation("Category").Relation("Tags")
	}

	err := query.Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		r.logger.ErrorX(ctx).Err(err).Stringer("id", id).Stringer("userID", userID).Msg("Failed to find article by ID")
		return nil, err
	}
	return article, nil
}

// FindByUserID retrieves a paginated list of articles for a specific user.
// Allows filtering by status, is_read, and is_starred. Optionally loads relationships.
func (r *ArticleRepository) FindByUserID(ctx context.Context, userID uuid.UUID, status *int, isRead, isStarred *bool, limit, offset int, loadRelations bool) ([]entity.Article, int, error) {
	var articles []entity.Article
	query := r.db.NewSelect().
		Model(&articles).
		Where("a.user_id = ?", userID)

	// Apply filters
	if status != nil {
		query = query.Where("a.status = ?", *status)
	}
	if isRead != nil {
		query = query.Where("a.is_read = ?", *isRead)
	}
	if isStarred != nil {
		query = query.Where("a.is_starred = ?", *isStarred)
	}

	// Load relations if requested
	if loadRelations {
		query = query.Relation("Category").Relation("Tags")
	}

	// Apply ordering, limit, and offset
	query = query.Order("a.created_at DESC").Limit(limit).Offset(offset)

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		// sql.ErrNoRows is not an error here, just means no articles found for the criteria
		if errors.Is(err, sql.ErrNoRows) {
			return []entity.Article{}, 0, nil
		}
		r.logger.ErrorX(ctx).Err(err).Stringer("userID", userID).Msg("Failed to find articles by user ID")
		return nil, 0, err
	}

	return articles, count, nil
}

// Update updates an existing article.
// Can update standard fields and also replace associated tags.
func (r *ArticleRepository) Update(ctx context.Context, article *entity.Article, newTags []*entity.Tag) error {
	if article.UserID == uuid.Nil || article.ID == uuid.Nil {
		return errors.New("article ID and user ID are required for update")
	}

	txErr := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Update the main article fields, explicitly listing columns to update.
		// Exclude URL, UserID, CreatedAt, and potentially ID itself from the SET clause.
		res, err := tx.NewUpdate().
			Model(article).
			// Specify columns to avoid updating URL/UserID/CreatedAt
			Set("title = ?, html = ?, author = ?, description = ?, plain_text = ?, llm_description = ?, og_image_url = ?, is_offline = ?, status = ?, is_read = ?, is_starred = ?, original_html = ?, updated_at = NOW()",
				article.Title, article.Html, article.Author, article.Description, article.PlainText, article.LLMDescription, article.OgImageURL, article.IsOffline, article.Status, article.IsRead, article.IsStarred, article.OriginalHtml).
			Where("id = ? AND user_id = ?", article.ID, article.UserID).
			Exec(ctx)
		if err != nil {
			r.logger.ErrorX(ctx).Err(err).Stringer("id", article.ID).Stringer("userID", article.UserID).Msg("Failed to update article core fields in transaction")
			return err
		}
		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			return sql.ErrNoRows // Article not found or doesn't belong to user
		}

		// --- Handle Tag Updates ---
		// 1. Delete existing tag associations for this article
		_, err = tx.NewDelete().
			Model((*entity.ArticleTag)(nil)).
			Where("article_id = ? AND user_id = ?", article.ID, article.UserID).
			Exec(ctx)
		if err != nil {
			r.logger.ErrorX(ctx).Err(err).Stringer("articleID", article.ID).Msg("Failed to delete old article tags association in transaction")
			return err
		}

		// 2. Insert new tag associations if provided
		if len(newTags) > 0 {
			items := make([]entity.ArticleTag, len(newTags))
			for i, tag := range newTags {
				if tag.ID == uuid.Nil {
					r.logger.ErrorX(ctx).Str("tag_name", tag.Name).Msg("Attempted to associate tag with nil ID during update")
					return errors.New("cannot associate tag with nil ID during update")
				}
				items[i] = entity.ArticleTag{
					ArticleID: article.ID,
					TagID:     tag.ID,
					UserID:    article.UserID,
				}
			}
			_, err = tx.NewInsert().Model(&items).Exec(ctx)
			if err != nil {
				r.logger.ErrorX(ctx).Err(err).Stringer("articleID", article.ID).Msg("Failed to insert new article tags association in transaction")
				return err
			}
		}
		return nil
	})

	if txErr != nil {
		// Log error but don't wrap sql.ErrNoRows if that's the case
		if !errors.Is(txErr, sql.ErrNoRows) {
			r.logger.ErrorX(ctx).Err(txErr).Stringer("id", article.ID).Msg("Article update transaction failed")
		}
		return txErr
	}

	r.logger.InfoX(ctx).Stringer("id", article.ID).Stringer("userID", article.UserID).Msg("Article updated successfully")
	return nil
}

// Delete removes an article by its ID, ensuring it belongs to the specified user.
// This also automatically removes associated rows in the article_tags table due to CASCADE DELETE
// if foreign keys were enabled. Since we removed FKs in MySQL/SQLite, the associations
// might remain orphaned if not deleted explicitly here (or by the service layer).
// The current implementation relies on the transaction in Update or explicit deletion if needed elsewhere.
func (r *ArticleRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// We need to delete the article and its tag associations in a transaction
	txErr := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// 1. Delete tag associations first (important if FKs are not used or ON DELETE CASCADE is not set)
		_, err := tx.NewDelete().
			Model((*entity.ArticleTag)(nil)).
			Where("article_id = ? AND user_id = ?", id, userID).
			Exec(ctx)
		if err != nil {
			// Log the error but potentially continue to delete the article itself?
			// Depending on desired behavior. Here we'll abort the transaction.
			r.logger.ErrorX(ctx).Err(err).Stringer("articleID", id).Msg("Failed to delete article tags associations during article deletion")
			return err
		}

		// 2. Delete the article itself
		res, err := tx.NewDelete().
			Model((*entity.Article)(nil)).
			Where("id = ? AND user_id = ?", id, userID).
			Exec(ctx)
		if err != nil {
			r.logger.ErrorX(ctx).Err(err).Stringer("id", id).Stringer("userID", userID).Msg("Failed to delete article in transaction")
			return err
		}
		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			// Article not found or doesn't belong to user
			// Return sql.ErrNoRows so the caller knows
			return sql.ErrNoRows
		}
		return nil
	})

	if txErr != nil {
		if errors.Is(txErr, sql.ErrNoRows) {
			r.logger.WarnX(ctx).Stringer("id", id).Stringer("userID", userID).Msg("Attempted to delete non-existent or unauthorized article")
			return sql.ErrNoRows // Propagate ErrNoRows
		}
		r.logger.ErrorX(ctx).Err(txErr).Stringer("id", id).Stringer("userID", userID).Msg("Failed to execute delete article transaction")
		return txErr
	}

	r.logger.InfoX(ctx).Stringer("id", id).Stringer("userID", userID).Msg("Article deleted successfully")
	return nil
}

// FindByURL retrieves an article by its URL for a specific user.
func (r *ArticleRepository) FindByURL(ctx context.Context, url string, userID uuid.UUID) (*entity.Article, error) {
	article := new(entity.Article)
	err := r.db.NewSelect().
		Model(article).
		Where("url = ? AND user_id = ?", url, userID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		r.logger.ErrorX(ctx).Err(err).Str("url", url).Stringer("userID", userID).Msg("Failed to find article by URL")
		return nil, err
	}
	return article, nil
}

// UpdateStatusFields updates only the is_read and is_starred fields of an article.
func (r *ArticleRepository) UpdateStatusFields(ctx context.Context, id uuid.UUID, userID uuid.UUID, isRead, isStarred *bool) error {
	log := r.logger.With().Stringer("id", id).Stringer("userID", userID).Logger()

	if isRead == nil && isStarred == nil {
		log.WarnX(ctx).Msg("UpdateStatusFields called with no fields to update")
		return nil // Nothing to update
	}

	query := r.db.NewUpdate().Model((*entity.Article)(nil))

	if isRead != nil {
		query = query.Set("is_read = ?", *isRead)
	}
	if isStarred != nil {
		query = query.Set("is_starred = ?", *isStarred)
	}

	// Always update the updated_at timestamp when modifying status
	query = query.Set("updated_at = NOW()")

	query = query.Where("id = ? AND user_id = ?", id, userID)

	res, err := query.Exec(ctx)
	if err != nil {
		log.ErrorX(ctx).Err(err).Msg("Failed to update article status fields")
		return err // Let the service layer handle error translation
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		log.WarnX(ctx).Msg("Article not found or not authorized for status update")
		return sql.ErrNoRows // Indicate not found
	}

	return nil
}

// FindByIDRoot retrieves an article by its ID, ignoring the caller's user ID (Root only).
// Optionally loads relationships (Category, Tags).
func (r *ArticleRepository) FindByIDRoot(ctx context.Context, id uuid.UUID, loadRelations bool) (*entity.Article, error) {
	article := new(entity.Article)
	query := r.db.NewSelect().
		Model(article).
		Where("a.id = ?", id) // Use alias 'a', only filter by article ID

	if loadRelations {
		query = query.Relation("Category").Relation("Tags")
	}

	err := query.Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows // Return specific error
		}
		r.logger.ErrorX(ctx).Err(err).Stringer("id", id).Msg("[Root] Failed to find article by ID")
		return nil, err // Return original db error
	}
	return article, nil
}

// FindByUserIDRoot retrieves articles for a specific user (Root only).
// Allows filtering by status, is_read, and is_starred. Optionally loads relationships.
func (r *ArticleRepository) FindByUserIDRoot(ctx context.Context, targetUserID uuid.UUID, status *int, isRead, isStarred *bool, limit, offset int, loadRelations bool) ([]entity.Article, int, error) {
	var articles []entity.Article
	// Use targetUserID in the WHERE clause
	query := r.db.NewSelect().
		Model(&articles).
		Where("a.user_id = ?", targetUserID)

	// Apply filters
	if status != nil {
		query = query.Where("a.status = ?", *status)
	}
	if isRead != nil {
		query = query.Where("a.is_read = ?", *isRead)
	}
	if isStarred != nil {
		query = query.Where("a.is_starred = ?", *isStarred)
	}

	// Load relations if requested
	if loadRelations {
		query = query.Relation("Category").Relation("Tags")
	}

	query = query.Order("a.created_at DESC").Limit(limit).Offset(offset)

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []entity.Article{}, 0, nil
		}
		r.logger.ErrorX(ctx).Err(err).Stringer("targetUserID", targetUserID).Msg("[Root] Failed to find articles by target user ID")
		return nil, 0, err
	}

	return articles, count, nil
}

// DeleteRoot removes an article by its ID, specifying the target user ID (Root only).
// Deletes associated tags for that user.
func (r *ArticleRepository) DeleteRoot(ctx context.Context, id uuid.UUID, targetUserID uuid.UUID) error {
	txErr := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// 1. Delete tag associations for the target user and article
		_, err := tx.NewDelete().
			Model((*entity.ArticleTag)(nil)).
			Where("article_id = ? AND user_id = ?", id, targetUserID). // Use targetUserID
			Exec(ctx)
		if err != nil {
			r.logger.ErrorX(ctx).Err(err).Stringer("articleID", id).Stringer("targetUserID", targetUserID).Msg("[Root] Failed to delete article tags associations during article deletion")
			return err
		}

		// 2. Delete the article itself, ensuring it belongs to the target user
		res, err := tx.NewDelete().
			Model((*entity.Article)(nil)).
			Where("id = ? AND user_id = ?", id, targetUserID). // Use targetUserID
			Exec(ctx)
		if err != nil {
			r.logger.ErrorX(ctx).Err(err).Stringer("id", id).Stringer("targetUserID", targetUserID).Msg("[Root] Failed to delete article in transaction")
			return err
		}
		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			return sql.ErrNoRows // Article not found for the specified user
		}
		return nil
	})

	if txErr != nil {
		if errors.Is(txErr, sql.ErrNoRows) {
			r.logger.WarnX(ctx).Stringer("id", id).Stringer("targetUserID", targetUserID).Msg("[Root] Attempted to delete non-existent article for target user")
			return sql.ErrNoRows // Propagate ErrNoRows
		}
		r.logger.ErrorX(ctx).Err(txErr).Stringer("id", id).Stringer("targetUserID", targetUserID).Msg("[Root] Failed to execute delete article transaction")
		return txErr
	}

	r.logger.InfoX(ctx).Stringer("id", id).Stringer("targetUserID", targetUserID).Msg("[Root] Article deleted successfully")
	return nil
}

// AddTags associates a list of tags with an article.
// It avoids creating duplicate associations.
func (r *ArticleRepository) AddTags(ctx context.Context, article *entity.Article, tagsToAdd []*entity.Tag) error {
	if len(tagsToAdd) == 0 {
		return nil
	}

	articleTags := make([]entity.ArticleTag, len(tagsToAdd))
	for i, tag := range tagsToAdd {
		articleTags[i] = entity.ArticleTag{
			ArticleID: article.ID,
			TagID:     tag.ID,
			UserID:    article.UserID, // Include UserID
			// CreatedAt defaults
		}
	}

	// Use ON CONFLICT DO NOTHING to silently ignore duplicates
	// This requires specifying the primary key columns of article_tags
	_, err := r.db.NewInsert().
		Model(&articleTags).
		On("CONFLICT (article_id, tag_id) DO NOTHING").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to add tags to article: %w", err)
	}

	// Update article's updated_at timestamp
	_, err = r.db.NewUpdate().
		Model((*entity.Article)(nil)).
		Set("updated_at = NOW()").
		Where("id = ?", article.ID).
		Exec(ctx)
	if err != nil {
		// Log this error but don't necessarily fail the whole operation
		fmt.Printf("Warning: Failed to update article timestamp after adding tags: %v\n", err)
	}

	return nil
}

// RemoveTags disassociates a list of tags from an article.
func (r *ArticleRepository) RemoveTags(ctx context.Context, article *entity.Article, tagsToRemove []*entity.Tag) error {
	if len(tagsToRemove) == 0 {
		return nil
	}

	tagIDsToRemove := make([]uuid.UUID, len(tagsToRemove))
	for i, tag := range tagsToRemove {
		tagIDsToRemove[i] = tag.ID
	}

	// Delete from the join table
	_, err := r.db.NewDelete().
		Model((*entity.ArticleTag)(nil)).
		Where("article_id = ?", article.ID).
		Where("tag_id IN (?)", bun.In(tagIDsToRemove)).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to remove tags from article: %w", err)
	}

	// Update article's updated_at timestamp
	_, err = r.db.NewUpdate().
		Model((*entity.Article)(nil)).
		Set("updated_at = NOW()").
		Where("id = ?", article.ID).
		Exec(ctx)
	if err != nil {
		fmt.Printf("Warning: Failed to update article timestamp after removing tags: %v\n", err)
	}

	return nil
}

// UpdateCategoryID changes the category_id of an article.
func (r *ArticleRepository) UpdateCategoryID(ctx context.Context, articleID, userID, newCategoryID uuid.UUID) error {
	res, err := r.db.NewUpdate().
		Model((*entity.Article)(nil)).
		Where("id = ? AND user_id = ?", articleID, userID).
		Set("category_id = ?", newCategoryID).
		Set("updated_at = NOW()"). // Also update timestamp
		Exec(ctx)

	if err != nil {
		return err // DB error
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows // Not found or wrong user
	}
	return nil
}

// FindArticlesByCategoryID retrieves articles belonging to a specific category for a user, with pagination.
// Articles are ordered by created_at DESC.
func (r *ArticleRepository) FindArticlesByCategoryID(ctx context.Context, categoryID uuid.UUID, userID uuid.UUID, limit, offset int) ([]entity.Article, int, error) {
	var articles []entity.Article
	query := r.db.NewSelect().
		Model(&articles).
		Where("a.category_id = ?", categoryID).
		Where("a.user_id = ?", userID).
		Order("a.created_at DESC").
		Limit(limit).
		Offset(offset).
		Relation("Category"). // Load category relation
		Relation("Tags")      // Load tags relation

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []entity.Article{}, 0, nil
		}
		r.logger.ErrorX(ctx).Err(err).Stringer("categoryID", categoryID).Stringer("userID", userID).Msg("Failed to find articles by category ID")
		return nil, 0, err
	}
	return articles, count, nil
}

// FindArticlesByTagID retrieves articles associated with a specific tag for a user, with pagination.
// Articles are ordered by created_at DESC.
func (r *ArticleRepository) FindArticlesByTagID(ctx context.Context, tagID uuid.UUID, userID uuid.UUID, limit, offset int) ([]entity.Article, int, error) {
	var articles []entity.Article

	query := r.db.NewSelect().
		Model(&articles).
		Join("JOIN article_tags AS at ON at.article_id = a.id").
		Where("at.tag_id = ?", tagID).
		Where("a.user_id = ?", userID).
		Order("a.created_at DESC").
		Limit(limit).
		Offset(offset).
		Relation("Category"). // Load category relation
		Relation("Tags")      // Load tags relation

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []entity.Article{}, 0, nil
		}
		r.logger.ErrorX(ctx).Err(err).Stringer("tagID", tagID).Stringer("userID", userID).Msg("Failed to find articles by tag ID")
		return nil, 0, err
	}
	return articles, count, nil
}
