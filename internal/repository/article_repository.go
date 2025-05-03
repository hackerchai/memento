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
// Allows filtering by status and optionally loads relationships.
func (r *ArticleRepository) FindByUserID(ctx context.Context, userID uuid.UUID, status *int, limit, offset int, loadRelations bool) ([]entity.Article, int, error) {
	var articles []entity.Article
	query := r.db.NewSelect().
		Model(&articles).
		Where("a.user_id = ?", userID)

	if status != nil {
		query = query.Where("a.status = ?", *status)
	}

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

	// Optional: Check URL uniqueness if URL is being updated (and not the same as original)
	// This requires fetching the original article first, which adds complexity.
	// Simpler approach: rely on the database unique constraint to fail if URL is changed to an existing one.

	txErr := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Update the main article fields
		// Exclude relations from the update model data automatically
		res, err := tx.NewUpdate().
			Model(article).
			Where("id = ? AND user_id = ?", article.ID, article.UserID).
			Exec(ctx)
		if err != nil {
			r.logger.ErrorX(ctx).Err(err).Stringer("id", article.ID).Stringer("userID", article.UserID).Msg("Failed to update article in transaction")
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
