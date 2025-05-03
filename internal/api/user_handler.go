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

// UserHandler handles HTTP requests related to user operations.
type UserHandler struct {
	userService *service.UserService
	validate    *validator.Validate
	logger      *xlog.Logger
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService *service.UserService, logger *xlog.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		validate:    validator.New(),
		logger:      logger.With().Str("handler", "UserHandler").Logger(),
	}
}

// UpdatePassword godoc
// @Summary Update user password
// @Description Updates the authenticated user's password after verifying the old password.
// @Tags users
// @Accept json
// @Produce json
// @Param password body entity.UpdatePasswordRequest true "Old and new password details"
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse "Password updated successfully"
// @Failure 400 {object} response.ErrorResponse{details=map[string]string} "Validation error (code: 00004) or invalid JSON (code: 00003)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001) or incorrect old password (code: 01004)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001 or specific service error)"
// @Router /users/password [put]
func (h *UserHandler) UpdatePassword(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c, h.logger)
	if err != nil {
		// GetUserIDFromContext already logs the specific error
		return response.HandleError(c, h.logger, errmsg.ErrUnauthorized.WithDetails("Invalid user session (middleware check failed)"))
	}

	var req entity.UpdatePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrBindJSON)
	}

	// Validate request body
	if err := h.validate.Struct(&req); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	if err := h.userService.UpdatePassword(c.Context(), userID, &req); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	return response.Respond(c, nil) // Use standard response for success
}

// UpdateEmail godoc
// @Summary Update user email
// @Description Updates the authenticated user's email address. Checks for email uniqueness.
// @Tags users
// @Accept json
// @Produce json
// @Param email body entity.UpdateEmailRequest true "New email address"
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse "Email updated successfully"
// @Failure 400 {object} response.ErrorResponse{details=map[string]string} "Validation error (code: 00004) or invalid JSON (code: 00003)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001)"
// @Failure 409 {object} response.ErrorResponse "Email already taken (code: 01005)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001 or specific service error)"
// @Router /users/email [put]
func (h *UserHandler) UpdateEmail(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c, h.logger)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrUnauthorized.WithDetails("Invalid user session (middleware check failed)"))
	}

	var req entity.UpdateEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrBindJSON)
	}

	// Validate request body
	if err := h.validate.Struct(&req); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	if err := h.userService.UpdateEmail(c.Context(), userID, &req); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	return response.Respond(c, nil) // Use standard response for success
}

// UpdateName godoc
// @Summary Update user name
// @Description Updates the authenticated user's display name.
// @Tags users
// @Accept json
// @Produce json
// @Param name body entity.UpdateNameRequest true "New display name"
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse "Name updated successfully"
// @Failure 400 {object} response.ErrorResponse{details=map[string]string} "Validation error (code: 00004) or invalid JSON (code: 00003)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001 or specific service error)"
// @Router /users/name [put]
func (h *UserHandler) UpdateName(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c, h.logger)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrUnauthorized.WithDetails("Invalid user session (middleware check failed)"))
	}

	var req entity.UpdateNameRequest
	if err := c.BodyParser(&req); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrBindJSON)
	}

	// Validate request body
	if err := h.validate.Struct(&req); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	if err := h.userService.UpdateName(c.Context(), userID, &req); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	return response.Respond(c, nil) // Use standard response for success
}

// CreateUser godoc
// @Summary Create a new user (Root only)
// @Description Creates a new user account. This endpoint requires root privileges.
// @Tags users,root
// @Accept json
// @Produce json
// @Param user body entity.CreateUserRequest true "New user details"
// @Security BearerAuth
// @Success 201 {object} response.SuccessResponse{data=entity.UserPublic} "User created successfully"
// @Failure 400 {object} response.ErrorResponse{details=map[string]string} "Validation error (code: 00004) or invalid JSON (code: 00003)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (not root user - code: 01006)"
// @Failure 409 {object} response.ErrorResponse "Email already registered (code: 01003)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001 or specific service error)"
// @Router /users/root [post]
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req entity.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrBindJSON)
	}

	// Validate request body
	if err := h.validate.Struct(&req); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	user, err := h.userService.CreateUser(c.Context(), &req)
	if err != nil {
		return response.HandleError(c, h.logger, err)
	}

	return response.RespondCreated(c, user)
}

// DeleteUser godoc
// @Summary Delete a user (Root only)
// @Description Deletes a user account by ID. This endpoint requires root privileges.
// @Tags users,root
// @Accept json
// @Produce json
// @Param id path string true "User ID to delete" format(uuid)
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse "User deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid user ID format (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (not root user - code: 01006)"
// @Failure 404 {object} response.ErrorResponse "User not found (code: 01007)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001 or specific service error)"
// @Router /users/root/{id} [delete]
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		// Use validator error handling for consistency, although UUID parsing is specific
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"id": "invalid user ID format"}))
	}

	if err := h.userService.DeleteUser(c.Context(), userID); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	return response.Respond(c, nil) // Use standard response for success
}

// GetProfile godoc
// @Summary Get own user profile
// @Description Retrieves the profile information (ID, name, email, role, created_at) for the currently authenticated user.
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=entity.UserSelf} "Successfully retrieved user profile"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001)"
// @Failure 404 {object} response.ErrorResponse "User not found (code: 01007)" // Should not happen if token is valid
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001 or specific service error)"
// @Router /users/profile [get]
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c, h.logger)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrUnauthorized.WithDetails("Invalid user session (middleware check failed)"))
	}

	profile, err := h.userService.GetUserProfile(c.Context(), userID)
	if err != nil {
		return response.HandleError(c, h.logger, err)
	}

	return response.Respond(c, profile)
}

// GetUserList godoc
// @Summary Get list of all users (Root only)
// @Description Retrieves a paginated list of all users. This endpoint requires root privileges.
// @Tags users,root
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param per_page query int false "Items per page" default(10) minimum(1) maximum(100)
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=response.PaginationResponse{data=[]entity.UserPublic}} "Successfully retrieved user list"
// @Failure 400 {object} response.ErrorResponse{details=map[string]string} "Invalid pagination parameters (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (not root user - code: 01006)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001 or specific service error)"
// @Router /users/root [get]
func (h *UserHandler) GetUserList(c *fiber.Ctx) error {
	var pagination response.PaginationRequest
	if err := c.QueryParser(&pagination); err != nil {
		// Use validation error for query parsing issues
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"query": "invalid pagination parameters"}))
	}

	// Validate pagination parameters
	if err := h.validate.Struct(&pagination); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	userList, err := h.userService.GetUserList(c.Context(), &pagination)
	if err != nil {
		return response.HandleError(c, h.logger, err)
	}

	return response.Respond(c, userList)
}
