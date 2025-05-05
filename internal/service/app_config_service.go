package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/hackerchai/memento/internal/entity"
	"github.com/hackerchai/memento/internal/errmsg"
	"github.com/hackerchai/memento/internal/repository"
	"github.com/hackerchai/memento/internal/response"
	"github.com/hackerchai/memento/pkg/xlog"
)

// AppConfigService provides business logic for application configurations.
type AppConfigService struct {
	appConfigRepo *repository.AppConfigRepository
	logger        *xlog.Logger
}

// NewAppConfigService creates a new AppConfigService.
func NewAppConfigService(appConfigRepo *repository.AppConfigRepository, logger *xlog.Logger) *AppConfigService {
	return &AppConfigService{
		appConfigRepo: appConfigRepo,
		logger:        logger.With().Str("service", "AppConfigService").Logger(),
	}
}

// Helper function to convert entity.AppConfig to entity.AppConfigResponse DTO
func toAppConfigResponse(config *entity.AppConfig) *entity.AppConfigResponse {
	if config == nil {
		return nil
	}
	return &entity.AppConfigResponse{
		ID:                         config.ID,
		UserID:                     config.UserID,
		ScrapeImgOffline:           config.ScrapeImgOffline,
		LLMAutoGenTags:             config.LLMAutoGenTags,
		ExtractLinks:               config.ExtractLinks,
		LLMProfileID:               config.LLMProfileID,
		LLMProvider:                config.LLMProvider,
		LLMAutoGenAbstract:         config.LLMAutoGenAbstract,
		CustomUserAgent:            config.CustomUserAgent,
		CustomScrapeTimeoutSeconds: config.CustomScrapeTimeoutSeconds,
		CustomScrapeRetryTimes:     config.CustomScrapeRetryTimes,
		CustomUserProxy:            config.CustomUserProxy,
		BypassRefer:                config.BypassRefer,
		Locale:                     config.Locale,
		CreatedAt:                  config.CreatedAt,
		UpdatedAt:                  config.UpdatedAt,
	}
}

// GetAppConfig retrieves the configuration for the specified user as a DTO.
func (s *AppConfigService) GetAppConfig(ctx context.Context, userID uuid.UUID) (*entity.AppConfigResponse, error) {
	s.logger.DebugX(ctx).Stringer("userID", userID).Msg("Getting app config")
	config, err := s.appConfigRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.ErrorX(ctx).Err(err).Stringer("userID", userID).Msg("Failed to get app config from repository")
		return nil, errmsg.ErrDatabase.WithDetails("Failed to retrieve app configuration")
	}
	return toAppConfigResponse(config), nil
}

// UpdateAppConfig updates the configuration for the specified user and returns the updated DTO.
func (s *AppConfigService) UpdateAppConfig(ctx context.Context, userID uuid.UUID, req *entity.UpdateAppConfigRequest) (*entity.AppConfigResponse, error) {
	s.logger.DebugX(ctx).Stringer("userID", userID).Msg("Attempting to update app config")

	// Get the current entity config
	configEntity, err := s.appConfigRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			configEntity = &entity.AppConfig{UserID: userID, Locale: "en-US"}
		} else {
			return nil, errmsg.ErrDatabase.WithDetails("Failed to retrieve current configuration for update")
		}
	}

	// Apply updates to the entity
	updated := false
	if req.ScrapeImgOffline != nil {
		configEntity.ScrapeImgOffline = *req.ScrapeImgOffline
		updated = true
	}
	if req.LLMAutoGenTags != nil {
		configEntity.LLMAutoGenTags = *req.LLMAutoGenTags
		updated = true
	}
	if req.ExtractLinks != nil {
		configEntity.ExtractLinks = *req.ExtractLinks
		updated = true
	}
	if req.LLMProfileID != nil {
		configEntity.LLMProfileID = req.LLMProfileID
		updated = true
	}
	if req.LLMProvider != nil {
		configEntity.LLMProvider = req.LLMProvider
		updated = true
	}
	if req.LLMAutoGenAbstract != nil {
		configEntity.LLMAutoGenAbstract = *req.LLMAutoGenAbstract
		updated = true
	}
	if req.CustomUserAgent != nil {
		configEntity.CustomUserAgent = req.CustomUserAgent
		updated = true
	}
	if req.CustomScrapeTimeoutSeconds != nil {
		configEntity.CustomScrapeTimeoutSeconds = req.CustomScrapeTimeoutSeconds
		updated = true
	}
	if req.CustomScrapeRetryTimes != nil {
		configEntity.CustomScrapeRetryTimes = req.CustomScrapeRetryTimes
		updated = true
	}
	if req.CustomUserProxy != nil {
		configEntity.CustomUserProxy = req.CustomUserProxy
		updated = true
	}
	if req.BypassRefer != nil {
		configEntity.BypassRefer = *req.BypassRefer
		updated = true
	}
	if req.Locale != nil {
		configEntity.Locale = *req.Locale
		updated = true
	}

	if !updated {
		s.logger.InfoX(ctx).Stringer("userID", userID).Msg("No fields provided for update.")
		return toAppConfigResponse(configEntity), nil // Return current config DTO if nothing changed
	}

	// Save the updated entity
	err = s.appConfigRepo.CreateOrUpdate(ctx, configEntity)
	if err != nil {
		s.logger.ErrorX(ctx).Err(err).Stringer("userID", userID).Msg("Failed to update app config in repository")
		return nil, errmsg.ErrDatabase.WithDetails("Failed to save configuration changes")
	}

	s.logger.InfoX(ctx).Stringer("userID", userID).Msg("App config updated successfully")
	// Return the DTO representation of the updated entity
	return toAppConfigResponse(configEntity), nil
}

// --- Root Operations ---

// GetAppConfigByUserIDRoot retrieves the DTO configuration for a specific user (Root only).
func (s *AppConfigService) GetAppConfigByUserIDRoot(ctx context.Context, targetUserID uuid.UUID) (*entity.AppConfigResponse, error) {
	s.logger.DebugX(ctx).Stringer("targetUserID", targetUserID).Msg("Getting app config by user ID (Root)")
	configEntity, err := s.appConfigRepo.GetByUserID(ctx, targetUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errmsg.ErrRecordNotFound.WithDetails(fmt.Sprintf("App configuration not found for user ID %s", targetUserID))
		}
		s.logger.ErrorX(ctx).Err(err).Stringer("targetUserID", targetUserID).Msg("Failed to get app config by user ID (Root)")
		return nil, errmsg.ErrDatabase.WithDetails("Failed to retrieve app configuration")
	}
	return toAppConfigResponse(configEntity), nil
}

// UpdateAppConfigByUserIDRoot updates the configuration for a specific user and returns the updated DTO (Root only).
func (s *AppConfigService) UpdateAppConfigByUserIDRoot(ctx context.Context, targetUserID uuid.UUID, req *entity.UpdateAppConfigRequest) (*entity.AppConfigResponse, error) {
	s.logger.DebugX(ctx).Stringer("targetUserID", targetUserID).Msg("Updating app config by user ID (Root)")
	// Reuse the core update logic, which now returns a DTO
	return s.UpdateAppConfig(ctx, targetUserID, req)
}

// ListAppConfigsRoot retrieves a paginated list of all user DTO configurations (Root only).
func (s *AppConfigService) ListAppConfigsRoot(ctx context.Context, pagination *response.PaginationRequest) (*response.PaginationResponse, error) {
	s.logger.DebugX(ctx).Msg("Listing all app configs (Root)")

	limit := pagination.GetLimit()
	offset := pagination.GetOffset()

	configEntities, total, err := s.appConfigRepo.FindAll(ctx, limit, offset)
	if err != nil {
		s.logger.ErrorX(ctx).Err(err).Msg("Failed to list app configs (Root)")
		return nil, errmsg.ErrDatabase.WithDetails("Failed to retrieve configuration list")
	}

	// Convert entity list to DTO list
	configDTOs := make([]*entity.AppConfigResponse, len(configEntities))
	for i, entity := range configEntities {
		configDTOs[i] = toAppConfigResponse(entity)
	}

	resp := &response.PaginationResponse{
		Total: int(total),
		Page:  pagination.Page,
		Data:  configDTOs, // Use the DTO list
	}

	return resp, nil
}
