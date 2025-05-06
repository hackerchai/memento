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

// CategoryHandler handles HTTP requests related to categories.
type CategoryHandler struct {
	categoryService *service.CategoryService
	validate        *validator.Validate
	logger          *xlog.Logger
}

// NewCategoryHandler creates a new CategoryHandler.
func NewCategoryHandler(categoryService *service.CategoryService, logger *xlog.Logger) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
		validate:        validator.New(),
		logger:          logger.With().Str("handler", "CategoryHandler").Logger(),
	}
}

// ListCategories godoc
// @Summary List categories
// @Description Retrieves a paginated list of categories for the authenticated user.
// @Tags categories
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param per_page query int false "Items per page" default(10) minimum(1) maximum(100)
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=response.PaginationResponse{data=[]entity.CategoryDetailResponse}} "Successfully retrieved categories"
// @Failure 400 {object} response.ErrorResponse "Invalid pagination parameters (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (code: 01001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001)"
// @Router /categories [get]
func (h *CategoryHandler) ListCategories(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c, h.logger)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrUnauthorized.WithDetails("Invalid user session"))
	}

	var pagination response.PaginationRequest
	if err := c.QueryParser(&pagination); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"query": "invalid pagination parameters"}))
	}
	if err := h.validate.Struct(&pagination); err != nil {
		return response.HandleError(c, h.logger, err) // validator errors are handled by HandleError
	}

	categories, err := h.categoryService.ListCategories(c.Context(), userID, &pagination)
	if err != nil {
		return response.HandleError(c, h.logger, err)
	}
	return response.Respond(c, categories)
}

// DeleteCategory godoc
// @Summary Delete a category
// @Description Deletes a category by its ID for the authenticated user. Associated articles will be unlinked (category set to null).
// @Tags categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID" format(uuid)
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse "Category deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid category ID format (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (code: 01001)"
// @Failure 404 {object} response.ErrorResponse "Category not found (code: 03001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001)"
// @Router /categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c, h.logger)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrUnauthorized.WithDetails("Invalid user session"))
	}

	categoryIDStr := c.Params("id")
	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"id": "invalid category ID format"}))
	}

	if err := h.categoryService.DeleteCategory(c.Context(), categoryID, userID); err != nil {
		return response.HandleError(c, h.logger, err)
	}
	return response.Respond(c, nil)
}

// GetArticlesByCategory godoc
// @Summary Get articles by category
// @Description Retrieves a paginated list of articles associated with a specific category (by ID or slug) for the authenticated user. Articles are ordered by creation date (newest first).
// @Tags categories,articles
// @Accept json
// @Produce json
// @Param identifier path string true "Category ID or Slug"
// @Param page query int false "Page number" default(1) minimum(1)
// @Param per_page query int false "Items per page" default(10) minimum(1) maximum(100)
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=response.PaginationResponse{data=[]entity.ArticleResponse}} "Successfully retrieved articles"
// @Failure 400 {object} response.ErrorResponse "Invalid pagination parameters or category identifier (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (code: 01001)"
// @Failure 404 {object} response.ErrorResponse "Category not found (code: 03001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001)"
// @Router /categories/{identifier}/articles [get]
func (h *CategoryHandler) GetArticlesByCategory(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c, h.logger)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrUnauthorized.WithDetails("Invalid user session"))
	}

	categoryIdentifier := c.Params("identifier")
	if categoryIdentifier == "" {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"identifier": "category identifier cannot be empty"}))
	}

	var pagination response.PaginationRequest
	if err := c.QueryParser(&pagination); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"query": "invalid pagination parameters"}))
	}
	if err := h.validate.Struct(&pagination); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	articles, err := h.categoryService.GetArticlesByCategory(c.Context(), userID, categoryIdentifier, &pagination)
	if err != nil {
		return response.HandleError(c, h.logger, err)
	}
	return response.Respond(c, articles)
}

// SearchCategories godoc
// @Summary Search categories
// @Description Searches categories by name or slug for the authenticated user. Case-insensitive.
// @Tags categories
// @Accept json
// @Produce json
// @Param q query string true "Search query term"
// @Param page query int false "Page number" default(1) minimum(1)
// @Param per_page query int false "Items per page" default(10) minimum(1) maximum(100)
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=response.PaginationResponse{data=[]entity.CategoryDetailResponse}} "Successfully retrieved categories matching the query"
// @Failure 400 {object} response.ErrorResponse "Invalid pagination parameters or missing query (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (code: 01001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001)"
// @Router /categories/search [get]
func (h *CategoryHandler) SearchCategories(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c, h.logger)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrUnauthorized.WithDetails("Invalid user session"))
	}

	searchQuery := c.Query("q")
	if searchQuery == "" { // Basic validation for query presence
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"q": "search query 'q' is required"}))
	}

	var pagination response.PaginationRequest
	if err := c.QueryParser(&pagination); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"query_parser": "invalid pagination parameters"}))
	}
	if err := h.validate.Struct(&pagination); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	categories, err := h.categoryService.SearchCategories(c.Context(), userID, searchQuery, &pagination)
	if err != nil {
		return response.HandleError(c, h.logger, err)
	}
	return response.Respond(c, categories)
}

// CreateCategory godoc
// @Summary Create a category
// @Description Creates a new category for the authenticated user.
// @Tags categories
// @Accept json
// @Produce json
// @Param category body entity.CreateCategoryRequest true "Category details"
// @Security BearerAuth
// @Success 201 {object} response.SuccessResponse{data=entity.CategoryDetailResponse} "Category created successfully"
// @Failure 400 {object} response.ErrorResponse "Validation error or invalid JSON (code: 00004 or 00003)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (code: 01001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error or conflict (code: 00001 or 00006)"
// @Router /categories [post]
func (h *CategoryHandler) CreateCategory(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c, h.logger)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrUnauthorized.WithDetails("Invalid user session"))
	}

	var req entity.CreateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrBindJSON)
	}
	if err := h.validate.Struct(&req); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	category, err := h.categoryService.CreateCategory(c.Context(), userID, &req)
	if err != nil {
		return response.HandleError(c, h.logger, err) // Service layer should return appropriate errmsg
	}

	return response.RespondCreated(c, category)
}

// --- Root Operations --- //

// ListCategoriesRoot godoc
// @Summary List categories for a user (Root)
// @Description Retrieves a paginated list of categories for a specified target user. Requires root privileges.
// @Tags categories,root
// @Accept json
// @Produce json
// @Param target_user_id query string true "Target User ID to filter categories for" format(uuid)
// @Param page query int false "Page number" default(1) minimum(1)
// @Param per_page query int false "Items per page" default(10) minimum(1) maximum(100)
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=response.PaginationResponse{data=[]entity.CategoryDetailResponse}} "Successfully retrieved categories"
// @Failure 400 {object} response.ErrorResponse "Invalid pagination or missing/invalid target_user_id (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (code: 01001)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (not root user - code: 01011)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001)"
// @Router /users/root/categories [get]
func (h *CategoryHandler) ListCategoriesRoot(c *fiber.Ctx) error {
	targetUserIDStr := c.Query("target_user_id")
	if targetUserIDStr == "" {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"target_user_id": "target_user_id is required"}))
	}
	parsedID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"target_user_id": "invalid target user ID format"}))
	}

	var pagination response.PaginationRequest
	if err := c.QueryParser(&pagination); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"query": "invalid pagination parameters"}))
	}
	if err := h.validate.Struct(&pagination); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	categories, err := h.categoryService.ListCategoriesRoot(c.Context(), parsedID, &pagination)
	if err != nil {
		return response.HandleError(c, h.logger, err)
	}
	return response.Respond(c, categories)
}

// DeleteCategoryRoot godoc
// @Summary Delete a category (Root)
// @Description Deletes a category by its ID. Requires root privileges.
// @Tags categories,root
// @Accept json
// @Produce json
// @Param id path string true "Category ID to delete" format(uuid)
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse "Category deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid category ID format (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (code: 01001)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (not root user - code: 01011)"
// @Failure 404 {object} response.ErrorResponse "Category not found (code: 03001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001)"
// @Router /users/root/categories/{id} [delete]
func (h *CategoryHandler) DeleteCategoryRoot(c *fiber.Ctx) error {
	categoryIDStr := c.Params("id")
	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"id": "invalid category ID format"}))
	}

	if err := h.categoryService.DeleteCategoryRoot(c.Context(), categoryID); err != nil {
		return response.HandleError(c, h.logger, err)
	}
	return response.Respond(c, nil)
}

// GetArticlesByCategoryRoot godoc
// @Summary Get articles by category (Root)
// @Description Retrieves a paginated list of articles for a specific category (by ID or slug). Requires root privileges.
// @Tags categories,articles,root
// @Accept json
// @Produce json
// @Param identifier path string true "Category ID or Slug"
// @Param page query int false "Page number" default(1) minimum(1)
// @Param per_page query int false "Items per page" default(10) minimum(1) maximum(100)
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=response.PaginationResponse{data=[]entity.ArticleResponse}} "Successfully retrieved articles"
// @Failure 400 {object} response.ErrorResponse "Invalid pagination or category identifier (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (code: 01001)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (not root user - code: 01011)"
// @Failure 404 {object} response.ErrorResponse "Category not found (code: 03001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001)"
// @Router /users/root/categories/{identifier}/articles [get]
func (h *CategoryHandler) GetArticlesByCategoryRoot(c *fiber.Ctx) error {
	categoryIdentifier := c.Params("identifier")
	if categoryIdentifier == "" {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"identifier": "category identifier cannot be empty"}))
	}

	var pagination response.PaginationRequest
	if err := c.QueryParser(&pagination); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"query": "invalid pagination parameters"}))
	}
	if err := h.validate.Struct(&pagination); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	articles, err := h.categoryService.GetArticlesByCategoryRoot(c.Context(), categoryIdentifier, &pagination)
	if err != nil {
		return response.HandleError(c, h.logger, err)
	}
	return response.Respond(c, articles)
}

// CreateCategoryRoot godoc
// @Summary Create a category for a user (Root)
// @Description Creates a new category for a specified target user. Requires root privileges.
// @Tags categories,root
// @Accept json
// @Produce json
// @Param category body entity.CreateCategoryRootRequest true "Category details including target user ID"
// @Security BearerAuth
// @Success 201 {object} response.SuccessResponse{data=entity.CategoryDetailResponse} "Category created successfully"
// @Failure 400 {object} response.ErrorResponse "Validation error or invalid JSON (code: 00004 or 00003)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (code: 01001)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (not root user - code: 01011)"
// @Failure 500 {object} response.ErrorResponse "Internal server error or conflict (code: 00001 or 00006)"
// @Router /users/root/categories [post]
func (h *CategoryHandler) CreateCategoryRoot(c *fiber.Ctx) error {
	var req entity.CreateCategoryRootRequest
	if err := c.BodyParser(&req); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrBindJSON)
	}
	if err := h.validate.Struct(&req); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	category, err := h.categoryService.CreateCategoryRoot(c.Context(), &req)
	if err != nil {
		return response.HandleError(c, h.logger, err)
	}

	return response.RespondCreated(c, category)
}
