package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/hackerchai/memento/internal/entity"
	"github.com/hackerchai/memento/internal/errmsg"
	"github.com/hackerchai/memento/internal/middleware"
	"github.com/hackerchai/memento/internal/response"
	"github.com/hackerchai/memento/internal/service"
	"github.com/hackerchai/memento/pkg/xlog"
)

// AppConfigHandler handles HTTP requests related to app configurations.
type AppConfigHandler struct {
	appConfigService *service.AppConfigService
	validate         *validator.Validate
	logger           *xlog.Logger
}

// NewAppConfigHandler creates a new AppConfigHandler.
func NewAppConfigHandler(appConfigService *service.AppConfigService, logger *xlog.Logger) *AppConfigHandler {
	return &AppConfigHandler{
		appConfigService: appConfigService,
		validate:         validator.New(),
		logger:           logger.With().Str("handler", "AppConfigHandler").Logger(),
	}
}

// GetOwnConfig godoc
// @Summary Get own app configuration
// @Description Retrieves the application configuration settings for the currently authenticated user.
// @Tags config
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=entity.AppConfigResponse} "Successfully retrieved app configuration"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001 or 00006)"
// @Router /config [get]
func (h *AppConfigHandler) GetOwnConfig(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c, h.logger)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrUnauthorized.WithDetails("Invalid user session"))
	}

	config, err := h.appConfigService.GetAppConfig(c.Context(), userID)
	if err != nil {
		return response.HandleError(c, h.logger, err) // Service layer handles specific errors
	}

	return response.Respond(c, config)
}

// UpdateOwnConfig godoc
// @Summary Update own app configuration
// @Description Updates the application configuration settings for the currently authenticated user. Only provided fields are updated.
// @Tags config
// @Accept json
// @Produce json
// @Param config body entity.UpdateAppConfigRequest true "Configuration fields to update"
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=entity.AppConfigResponse} "Successfully updated app configuration"
// @Failure 400 {object} response.ErrorResponse{details=map[string]string} "Validation error (code: 00004) or invalid JSON (code: 00003)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001 or 00006)"
// @Router /config [put]
func (h *AppConfigHandler) UpdateOwnConfig(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c, h.logger)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrUnauthorized.WithDetails("Invalid user session"))
	}

	var req entity.UpdateAppConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrBindJSON)
	}

	// Validate request body
	if err := h.validate.Struct(&req); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	updatedConfig, err := h.appConfigService.UpdateAppConfig(c.Context(), userID, &req)
	if err != nil {
		return response.HandleError(c, h.logger, err)
	}

	return response.Respond(c, updatedConfig)
}

// --- Root Operations ---

// GetConfigByUserIDRoot godoc
// @Summary Get user's app configuration (Root only)
// @Description Retrieves the application configuration settings for a specific user by their ID. Requires root privileges.
// @Tags config,root
// @Produce json
// @Param id path string true "Target User ID" format(uuid)
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=entity.AppConfigResponse} "Successfully retrieved user's app configuration"
// @Failure 400 {object} response.ErrorResponse "Invalid user ID format (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (not root user - code: 01011)"
// @Failure 404 {object} response.ErrorResponse "Configuration not found for user (code: 02001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001 or 00006)"
// @Router /users/root/config/{id} [get]
func (h *AppConfigHandler) GetConfigByUserIDRoot(c *fiber.Ctx) error {
	id := c.Params("id")
	targetUserID, err := uuid.Parse(id)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"id": "invalid user ID format"}))
	}

	config, err := h.appConfigService.GetAppConfigByUserIDRoot(c.Context(), targetUserID)
	if err != nil {
		return response.HandleError(c, h.logger, err)
	}

	return response.Respond(c, config)
}

// UpdateConfigByUserIDRoot godoc
// @Summary Update user's app configuration (Root only)
// @Description Updates the application configuration settings for a specific user by their ID. Requires root privileges. Only provided fields are updated.
// @Tags config,root
// @Accept json
// @Produce json
// @Param id path string true "Target User ID" format(uuid)
// @Param config body entity.UpdateAppConfigRequest true "Configuration fields to update"
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=entity.AppConfigResponse} "Successfully updated user's app configuration"
// @Failure 400 {object} response.ErrorResponse{details=map[string]string} "Validation error (code: 00004), invalid JSON (code: 00003), or invalid user ID format"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (not root user - code: 01011)"
// @Failure 404 {object} response.ErrorResponse "Configuration or user not found (code: 02001)" // Update might create if not found, but GET check happens first implicitly
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001 or 00006)"
// @Router /users/root/config/{id} [put]
func (h *AppConfigHandler) UpdateConfigByUserIDRoot(c *fiber.Ctx) error {
	id := c.Params("id")
	targetUserID, err := uuid.Parse(id)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"id": "invalid user ID format"}))
	}

	var req entity.UpdateAppConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrBindJSON)
	}

	// Validate request body
	if err := h.validate.Struct(&req); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	updatedConfig, err := h.appConfigService.UpdateAppConfigByUserIDRoot(c.Context(), targetUserID, &req)
	if err != nil {
		return response.HandleError(c, h.logger, err)
	}

	return response.Respond(c, updatedConfig)
}

// ListConfigsRoot godoc
// @Summary List all app configurations (Root only)
// @Description Retrieves a paginated list of all user application configurations. Requires root privileges.
// @Tags config,root
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param per_page query int false "Items per page" default(10) minimum(1) maximum(100)
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=response.PaginationResponse{data=[]entity.AppConfigResponse}} "Successfully retrieved configuration list"
// @Failure 400 {object} response.ErrorResponse{details=map[string]string} "Invalid pagination parameters (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (not root user - code: 01011)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001 or 00006)"
// @Router /users/root/config [get]
func (h *AppConfigHandler) ListConfigsRoot(c *fiber.Ctx) error {
	var pagination response.PaginationRequest
	if err := c.QueryParser(&pagination); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"query": "invalid pagination parameters"}))
	}

	// Validate pagination parameters
	if err := h.validate.Struct(&pagination); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	configList, err := h.appConfigService.ListAppConfigsRoot(c.Context(), &pagination)
	if err != nil {
		return response.HandleError(c, h.logger, err)
	}

	return response.Respond(c, configList)
}
