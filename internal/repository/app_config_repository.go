package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	// Import SQLite specific dialect for ON CONFLICT
	// Import pg specific driver for ON CONFLICT
	// MySQL driver is usually imported implicitly or handled by bun

	"github.com/hackerchai/memento/internal/entity"
	"github.com/hackerchai/memento/pkg/xlog"
)

// AppConfigRepository handles database operations for app configurations.
type AppConfigRepository struct {
	db     *bun.DB
	logger *xlog.Logger
}

// NewAppConfigRepository creates a new AppConfigRepository.
func NewAppConfigRepository(db *bun.DB, logger *xlog.Logger) *AppConfigRepository {
	return &AppConfigRepository{
		db:     db,
		logger: logger.With().Str("repository", "AppConfig").Logger(),
	}
}

// Create inserts a new app configuration for a user.
func (r *AppConfigRepository) Create(ctx context.Context, config *entity.AppConfig) error {
	if config.UserID == uuid.Nil {
		return errors.New("user ID is required to create app config")
	}

	// Check if config already exists for this user
	exists, err := r.db.NewSelect().
		Model((*entity.AppConfig)(nil)).
		Where("user_id = ?", config.UserID).
		Exists(ctx)
	if err != nil {
		r.logger.ErrorX(ctx).Err(err).Stringer("userID", config.UserID).Msg("Failed check app config existence")
		return err
	}
	if exists {
		return errors.New("app config already exists for this user")
	}

	_, err = r.db.NewInsert().Model(config).Exec(ctx)
	if err != nil {
		r.logger.ErrorX(ctx).Err(err).Stringer("userID", config.UserID).Msg("Failed to insert new app config")
		return err
	}
	r.logger.InfoX(ctx).Stringer("userID", config.UserID).Msg("App config created successfully")
	return nil
}

// GetByUserID retrieves the app configuration for a specific user.
// If no configuration exists for the user, it returns a default AppConfig object
// and a nil error, indicating that defaults should be used.
func (r *AppConfigRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.AppConfig, error) {
	config := new(entity.AppConfig)
	err := r.db.NewSelect().
		Model(config).
		Where("user_id = ?", userID).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No config found, return a default config struct
			r.logger.InfoX(ctx).Stringer("userID", userID).Msg("No app config found for user, returning defaults.")
			return &entity.AppConfig{UserID: userID}, nil // Return default struct with UserID set
		}
		// Handle other potential errors
		r.logger.ErrorX(ctx).Err(err).Stringer("userID", userID).Msg("Failed to get app config by user ID")
		return nil, err
	}

	r.logger.DebugX(ctx).Stringer("userID", userID).Msg("App config retrieved successfully.")
	return config, nil
}

// CreateOrUpdate creates a new app configuration for a user if one doesn't exist,
// or updates the existing one.
func (r *AppConfigRepository) CreateOrUpdate(ctx context.Context, config *entity.AppConfig) error {
	if config.UserID == uuid.Nil {
		return errors.New("user ID is required to create or update app config")
	}

	// Assign a new ID if the config doesn't have one (for the INSERT part of UPSERT)
	if config.ID == uuid.Nil {
		config.ID = uuid.New()
	}

	// Use ON CONFLICT (for PG/SQLite) or ON DUPLICATE KEY UPDATE (for MySQL)
	// Bun handles the dialect differences automatically with On("CONFLICT")
	_, err := r.db.NewInsert().
		Model(config).
		On("CONFLICT (user_id) DO UPDATE"). // Specify the conflict target (unique constraint)
		// Set the fields to update on conflict. Exclude pk (id) and user_id.
		Set("scrape_img_offline = EXCLUDED.scrape_img_offline").
		Set("llm_auto_gen_tags = EXCLUDED.llm_auto_gen_tags").
		Set("extract_links = EXCLUDED.extract_links").
		Set("llm_profile_id = EXCLUDED.llm_profile_id").
		Set("llm_provider = EXCLUDED.llm_provider").
		Set("llm_auto_gen_abstract = EXCLUDED.llm_auto_gen_abstract").
		Set("custom_user_agent = EXCLUDED.custom_user_agent").
		Set("custom_scrape_timeout_seconds = EXCLUDED.custom_scrape_timeout_seconds").
		Set("custom_scrape_retry_times = EXCLUDED.custom_scrape_retry_times").
		Set("custom_user_proxy = EXCLUDED.custom_user_proxy").
		Set("bypass_refer = EXCLUDED.bypass_refer").
		// UpdatedAt is typically handled by the database trigger/default
		// Set("updated_at = NOW()"). // Explicitly set if needed or DB doesn't handle it
		Exec(ctx)

	if err != nil {
		r.logger.ErrorX(ctx).Err(err).Stringer("userID", config.UserID).Msg("Failed to create or update app config")
		return err
	}

	r.logger.InfoX(ctx).Stringer("userID", config.UserID).Msg("App config created or updated successfully.")
	return nil
}

// Update updates an existing app configuration for a user.
func (r *AppConfigRepository) Update(ctx context.Context, config *entity.AppConfig) error {
	if config.UserID == uuid.Nil {
		return errors.New("user ID is required to update app config")
	}

	// We rely on `user_id` being the logical key for update operations from the service.
	res, err := r.db.NewUpdate().
		Model(config).
		// Exclude fields that shouldn't be updatable or are handled by DB
		Column("scrape_img_offline", "llm_auto_gen_tags", "extract_links", "llm_profile_id", "llm_provider", "llm_auto_gen_abstract", "custom_user_agent", "custom_scrape_timeout_seconds", "custom_scrape_retry_times", "custom_user_proxy", "bypass_refer"). // Explicitly list columns to update
		Where("user_id = ?", config.UserID).
		Exec(ctx)

	if err != nil {
		r.logger.ErrorX(ctx).Err(err).Stringer("userID", config.UserID).Msg("Failed to update app config")
		return err
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		// This could mean the config didn't exist for the user
		return sql.ErrNoRows // Or a more specific error
	}

	r.logger.InfoX(ctx).Stringer("userID", config.UserID).Msg("App config updated successfully")
	return nil
}
