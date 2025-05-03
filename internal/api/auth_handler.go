package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/hackerchai/memento/internal/entity"
	"github.com/hackerchai/memento/internal/errmsg"
	"github.com/hackerchai/memento/internal/response"
	"github.com/hackerchai/memento/internal/service"
	"github.com/hackerchai/memento/pkg/xlog"
)

// AuthHandler handles authentication related API requests.
type AuthHandler struct {
	userService *service.UserService
	validate    *validator.Validate
	logger      *xlog.Logger
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(userService *service.UserService, logger *xlog.Logger) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		validate:    validator.New(),
		logger:      logger.With().Str("handler", "AuthHandler").Logger(), // Base logger for the handler
	}
}

// Register godoc
// @Summary Register a new user
// @Description Creates a new user account based on the provided details.
// @Tags auth
// @Accept json
// @Produce json
// @Param user body entity.RegisterRequest true "User registration details"
// @Success 201 {object} response.SuccessResponse{data=entity.UserPublic} "Successfully registered user"
// @Failure 400 {object} response.ErrorResponse{details=map[string]string} "Validation error (code: 00004) or invalid JSON (code: 00003)"
// @Failure 409 {object} response.ErrorResponse "Email already registered (code: 01003)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001 or specific service error)"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req entity.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrBindJSON)
	}

	// Validate request body
	if err := h.validate.Struct(&req); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	user, err := h.userService.Register(c.Context(), &req)
	if err != nil {
		return response.HandleError(c, h.logger, err)
	}

	return response.RespondCreated(c, user)
}

// Login godoc
// @Summary Log in a user
// @Description Authenticates a user and returns a JWT token upon successful login.
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body entity.LoginRequest true "User login credentials"
// @Success 200 {object} response.SuccessResponse{data=entity.LoginResponse} "Successfully logged in"
// @Failure 400 {object} response.ErrorResponse{details=map[string]string} "Validation error (code: 00004) or invalid JSON (code: 00003)"
// @Failure 401 {object} response.ErrorResponse "Invalid email or password (code: 01002)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001 or specific service error)"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req entity.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrBindJSON)
	}

	// Validate request body
	if err := h.validate.Struct(&req); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	resp, err := h.userService.Login(c.Context(), &req)
	if err != nil {
		return response.HandleError(c, h.logger, err)
	}

	return response.Respond(c, resp)
}
