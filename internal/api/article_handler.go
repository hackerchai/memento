package api

import (
	"strconv"

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

// ArticleHandler handles HTTP requests related to articles.
type ArticleHandler struct {
	articleService *service.ArticleService
	validate       *validator.Validate
	logger         *xlog.Logger
}

// NewArticleHandler creates a new ArticleHandler.
func NewArticleHandler(articleService *service.ArticleService, logger *xlog.Logger) *ArticleHandler {
	return &ArticleHandler{
		articleService: articleService,
		validate:       validator.New(),
		logger:         logger.With().Str("handler", "ArticleHandler").Logger(),
	}
}

// CreateArticle godoc
// @Summary Submit a URL to save as an article
// @Description Accepts a URL, optionally a category name and tags. Creates a placeholder article immediately and processes the full content in the background.
// @Tags articles
// @Accept json
// @Produce json
// @Param article body entity.CreateArticleRequest true "Article URL and optional metadata"
// @Security BearerAuth
// @Success 202 {object} response.SuccessResponse{data=entity.ArticleResponse} "Article accepted for processing, returns initial article data (with ID and processing status)"
// @Failure 400 {object} response.ErrorResponse{details=map[string]string} "Validation error (URL format, etc. - code: 00004) or invalid JSON (code: 00003)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001)"
// @Failure 409 {object} response.ErrorResponse "Conflict - Article with this URL already exists (code: 02001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (failed to create placeholder, etc. - code: 00001)"
// @Router /articles [post]
func (h *ArticleHandler) CreateArticle(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c, h.logger)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrUnauthorized.WithDetails("Invalid user session"))
	}

	var req entity.CreateArticleRequest
	if err := c.BodyParser(&req); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrBindJSON)
	}

	// Validate request body
	if err := h.validate.Struct(&req); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	// Prepare service input
	input := &service.SaveArticleFromURLInput{
		UserID:       userID,
		URL:          req.URL,
		CategoryName: req.CategoryName,
		TagNames:     req.Tags,
	}

	// Call the service to initiate saving
	initialArticleEntity, err := h.articleService.SaveArticleFromURL(c.Context(), input)
	if err != nil {
		// Handle specific errors like duplicates
		if err.Error() == "article already exists" { // TODO: Use typed errors from errmsg
			return response.HandleError(c, h.logger, errmsg.ErrArticleConflict)
		}
		// Handle other service errors
		return response.HandleError(c, h.logger, err)
	}

	// Map the initial entity to the DTO for the response using the method
	responseDTO := initialArticleEntity.ToResponseDTO()

	// Return 202 Accepted with the initial article DTO
	return response.RespondAccepted(c, responseDTO)
}

// GetArticle godoc
// @Summary Get a specific article
// @Description Retrieves details of a specific article belonging to the authenticated user, including category and tags.
// @Tags articles
// @Produce json
// @Param id path string true "Article ID" format(uuid)
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=entity.ArticleDetailResponse} "Successfully retrieved article with details"
// @Failure 400 {object} response.ErrorResponse "Invalid Article ID format (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001)"
// @Failure 404 {object} response.ErrorResponse "Article not found (code: 02001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001)"
// @Router /articles/{id} [get]
func (h *ArticleHandler) GetArticle(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c, h.logger)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrUnauthorized.WithDetails("Invalid user session"))
	}

	articleIDStr := c.Params("id")
	articleID, err := uuid.Parse(articleIDStr)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"id": "invalid article ID format"}))
	}

	articleEntity, err := h.articleService.GetArticle(c.Context(), articleID, userID)
	if err != nil {
		return response.HandleError(c, h.logger, err) // Service layer returns specific errors (NotFound, Internal)
	}

	responseDTO := articleEntity.ToDetailResponseDTO()
	return response.Respond(c, responseDTO)
}

// ListArticles godoc
// @Summary List user's articles
// @Description Retrieves a paginated list of articles belonging to the authenticated user. Allows filtering by read and starred status.
// @Tags articles
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param per_page query int false "Items per page" default(10) minimum(1) maximum(100)
// @Param is_read query boolean false "Filter by read status (true/false)"
// @Param is_starred query boolean false "Filter by starred status (true/false)"
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=response.PaginationResponse{data=[]entity.ArticleResponse}} "Successfully retrieved articles"
// @Failure 400 {object} response.ErrorResponse{details=map[string]string} "Invalid pagination or filter parameters (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001)"
// @Router /articles [get]
func (h *ArticleHandler) ListArticles(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c, h.logger)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrUnauthorized.WithDetails("Invalid user session"))
	}

	var pagination response.PaginationRequest
	if err := c.QueryParser(&pagination); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"query": "invalid pagination parameters"}))
	}

	// Validate pagination parameters
	if err := h.validate.Struct(&pagination); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	// Parse optional boolean filter parameters
	var isReadFilter *bool
	if isReadStr := c.Query("is_read"); isReadStr != "" {
		b, err := strconv.ParseBool(isReadStr)
		if err != nil {
			return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"is_read": "invalid boolean value"}))
		}
		isReadFilter = &b
	}

	var isStarredFilter *bool
	if isStarredStr := c.Query("is_starred"); isStarredStr != "" {
		b, err := strconv.ParseBool(isStarredStr)
		if err != nil {
			return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"is_starred": "invalid boolean value"}))
		}
		isStarredFilter = &b
	}

	// Pass filters to the service layer
	paginatedResult, err := h.articleService.ListArticles(c.Context(), userID, &pagination, isReadFilter, isStarredFilter)
	if err != nil {
		return response.HandleError(c, h.logger, err) // Service layer returns specific errors (Internal)
	}

	return response.Respond(c, paginatedResult) // Correct return for ListArticles
}

// DeleteArticle godoc
// @Summary Delete an article
// @Description Deletes a specific article belonging to the authenticated user.
// @Tags articles
// @Produce json
// @Param id path string true "Article ID" format(uuid)
// @Security BearerAuth
// @Success 204 "Successfully deleted article"
// @Failure 400 {object} response.ErrorResponse "Invalid Article ID format (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001)"
// @Failure 404 {object} response.ErrorResponse "Article not found (code: 02001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001)"
// @Router /articles/{id} [delete]
func (h *ArticleHandler) DeleteArticle(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c, h.logger)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrUnauthorized.WithDetails("Invalid user session"))
	}

	articleIDStr := c.Params("id")
	articleID, err := uuid.Parse(articleIDStr)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"id": "invalid article ID format"}))
	}

	err = h.articleService.DeleteArticle(c.Context(), articleID, userID)
	if err != nil {
		return response.HandleError(c, h.logger, err) // Service layer returns specific errors (NotFound, Internal)
	}

	return c.SendStatus(fiber.StatusNoContent) // Use Fiber's method for 204
}

// UpdateArticleStatus godoc
// @Summary Update article read/starred status
// @Description Updates the is_read and/or is_starred flags for a specific article.
// @Tags articles
// @Accept json
// @Produce json
// @Param id path string true "Article ID" format(uuid)
// @Param status body entity.UpdateArticleStatusRequest true "Fields to update (at least one required)"
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=entity.ArticleResponse} "Successfully updated article status"
// @Failure 400 {object} response.ErrorResponse{details=map[string]string} "Invalid Article ID format, invalid JSON, or no fields provided (code: 00004/00003)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001)"
// @Failure 404 {object} response.ErrorResponse "Article not found (code: 02001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001)"
// @Router /articles/{id} [patch]
func (h *ArticleHandler) UpdateArticleStatus(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c, h.logger)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrUnauthorized.WithDetails("Invalid user session"))
	}

	articleIDStr := c.Params("id")
	articleID, err := uuid.Parse(articleIDStr)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"id": "invalid article ID format"}))
	}

	var req entity.UpdateArticleStatusRequest // Use the DTO from entity package
	if err := c.BodyParser(&req); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrBindJSON)
	}

	// Basic validation: ensure at least one field is present (service layer also checks this)
	if req.IsRead == nil && req.IsStarred == nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"body": "at least one field (is_read or is_starred) must be provided"}))
	}

	updatedArticleDTO, err := h.articleService.UpdateArticleStatus(c.Context(), articleID, userID, &req)
	if err != nil {
		return response.HandleError(c, h.logger, err) // Service layer returns specific errors
	}

	return response.Respond(c, updatedArticleDTO)
}

// ReScrapeArticle godoc
// @Summary Re-scrape an article
// @Description Triggers the background process to fetch and parse the article content again for a specific article.
// @Tags articles
// @Produce json
// @Param id path string true "Article ID" format(uuid)
// @Security BearerAuth
// @Success 202 {object} response.SuccessResponse{data=entity.ArticleResponse} "Article accepted for re-processing, returns article data with pending status"
// @Failure 400 {object} response.ErrorResponse "Invalid Article ID format (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001)"
// @Failure 404 {object} response.ErrorResponse "Article not found (code: 02001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001)"
// @Router /articles/{id}/rescrape [post]
func (h *ArticleHandler) ReScrapeArticle(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c, h.logger)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrUnauthorized.WithDetails("Invalid user session"))
	}

	articleIDStr := c.Params("id")
	articleID, err := uuid.Parse(articleIDStr)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"id": "invalid article ID format"}))
	}

	pendingArticleDTO, err := h.articleService.ReScrapeArticle(c.Context(), articleID, userID)
	if err != nil {
		return response.HandleError(c, h.logger, err) // Service layer returns specific errors
	}

	// Return 202 Accepted with the article DTO showing pending status
	return response.RespondAccepted(c, pendingArticleDTO)
}

// --- Root Article Operations ---

// GetArticleRoot godoc
// @Summary Get any article by ID (Root only)
// @Description Retrieves details of a specific article by its ID, including category and tags. Requires root privileges.
// @Tags articles,root
// @Produce json
// @Param id path string true "Article ID" format(uuid)
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=entity.ArticleDetailResponse} "Successfully retrieved article with details"
// @Failure 400 {object} response.ErrorResponse "Invalid Article ID format (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (not root user - code: 01011)"
// @Failure 404 {object} response.ErrorResponse "Article not found (code: 02001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001)"
// @Router /users/root/articles/{id} [get]
func (h *ArticleHandler) GetArticleRoot(c *fiber.Ctx) error {
	// Root check middleware should run before this handler

	articleIDStr := c.Params("id")
	articleID, err := uuid.Parse(articleIDStr)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"id": "invalid article ID format"}))
	}

	articleEntity, err := h.articleService.GetArticleRoot(c.Context(), articleID)
	if err != nil {
		return response.HandleError(c, h.logger, err) // Service layer returns specific errors
	}

	responseDTO := articleEntity.ToResponseDTO()
	return response.Respond(c, responseDTO)
}

// ListUserArticlesRoot godoc
// @Summary List articles for a specific user (Root only)
// @Description Retrieves a paginated list of articles for a given user ID. Requires root privileges.
// @Tags articles,root
// @Produce json
// @Param user_id query string true "Target User ID" format(uuid)
// @Param page query int false "Page number" default(1) minimum(1)
// @Param per_page query int false "Items per page" default(10) minimum(1) maximum(100)
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=response.PaginationResponse{data=[]entity.ArticleResponse}} "Successfully retrieved user's articles"
// @Failure 400 {object} response.ErrorResponse{details=map[string]string} "Invalid User ID format or pagination parameters (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (not root user - code: 01011)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001)"
// @Router /users/root/articles/user [get] // Using query param
func (h *ArticleHandler) ListUserArticlesRoot(c *fiber.Ctx) error {
	// Root check middleware should run before this handler

	targetUserIDStr := c.Query("user_id")
	if targetUserIDStr == "" {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"user_id": "target user ID is required"}))
	}
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"user_id": "invalid target user ID format"}))
	}

	var pagination response.PaginationRequest
	if err := c.QueryParser(&pagination); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"query": "invalid pagination parameters"}))
	}
	if err := h.validate.Struct(&pagination); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	paginatedResult, err := h.articleService.ListArticlesRoot(c.Context(), targetUserID, &pagination)
	if err != nil {
		return response.HandleError(c, h.logger, err)
	}

	return response.Respond(c, paginatedResult)
}

// DeleteArticleRoot godoc
// @Summary Delete any article by ID (Root only)
// @Description Deletes a specific article by its ID. Requires root privileges.
// @Tags articles,root
// @Produce json
// @Param id path string true "Article ID" format(uuid)
// @Security BearerAuth
// @Success 204 "Successfully deleted article"
// @Failure 400 {object} response.ErrorResponse "Invalid Article ID format (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (not root user - code: 01011)"
// @Failure 404 {object} response.ErrorResponse "Article not found (code: 02001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001)"
// @Router /users/root/articles/{id} [delete]
func (h *ArticleHandler) DeleteArticleRoot(c *fiber.Ctx) error {
	// Root check middleware should run before this handler

	articleIDStr := c.Params("id")
	articleID, err := uuid.Parse(articleIDStr)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"id": "invalid article ID format"}))
	}

	// Fetch article first to get owner ID (needed by service/repo DeleteRoot)
	article, err := h.articleService.GetArticleRoot(c.Context(), articleID)
	if err != nil {
		// If GetArticleRoot returns RecordNotFound, HandleError will map it correctly to 404
		return response.HandleError(c, h.logger, err)
	}
	targetUserID := article.UserID

	err = h.articleService.DeleteArticleRoot(c.Context(), articleID, targetUserID) // Pass fetched targetUserID
	if err != nil {
		return response.HandleError(c, h.logger, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ReScrapeArticleRoot godoc
// @Summary Re-scrape any article (Root only)
// @Description Triggers the background process to fetch and parse content again for any article ID. Requires root privileges.
// @Tags articles,root
// @Produce json
// @Param id path string true "Article ID" format(uuid)
// @Security BearerAuth
// @Success 202 {object} response.SuccessResponse{data=entity.ArticleResponse} "Article accepted for re-processing, returns article data with pending status"
// @Failure 400 {object} response.ErrorResponse "Invalid Article ID format (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (not root user - code: 01011)"
// @Failure 404 {object} response.ErrorResponse "Article not found (code: 02001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001)"
// @Router /users/root/articles/{id}/rescrape [post]
func (h *ArticleHandler) ReScrapeArticleRoot(c *fiber.Ctx) error {
	// Root check middleware should run before this handler

	articleIDStr := c.Params("id")
	articleID, err := uuid.Parse(articleIDStr)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"id": "invalid article ID format"}))
	}

	// Service method handles finding the owner and triggering re-scrape
	pendingArticleDTO, err := h.articleService.ReScrapeArticleRoot(c.Context(), articleID)
	if err != nil {
		return response.HandleError(c, h.logger, err)
	}

	return response.RespondAccepted(c, pendingArticleDTO)
}

// --- New Handlers for Tag and Category Management ---

// AddTagsToArticle godoc
// @Summary Add tags to an article
// @Description Adds one or more tags to a specific article. If tags don't exist for the user, they are created.
// @Tags articles
// @Accept json
// @Produce json
// @Param id path string true "Article ID" format(uuid)
// @Param tags body entity.AddTagsRequest true "List of tag names to add"
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=entity.ArticleDetailResponse} "Successfully added tags and returned updated article details"
// @Failure 400 {object} response.ErrorResponse "Invalid Article ID format, invalid JSON, or validation error on tags (code: 00004/00003)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001)"
// @Failure 404 {object} response.ErrorResponse "Article not found (code: 02001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (failed to add tags, etc. - code: 00001)"
// @Router /articles/{id}/tags [post]
func (h *ArticleHandler) AddTagsToArticle(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c, h.logger)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrUnauthorized.WithDetails("Invalid user session"))
	}

	articleIDStr := c.Params("id")
	articleID, err := uuid.Parse(articleIDStr)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"id": "invalid article ID format"}))
	}

	var req entity.AddTagsRequest
	if err := c.BodyParser(&req); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrBindJSON)
	}

	// Validate request body
	if err := h.validate.Struct(&req); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	updatedArticleDTO, err := h.articleService.AddTagsToArticle(c.Context(), articleID, userID, req.Tags)
	if err != nil {
		return response.HandleError(c, h.logger, err) // Service layer returns specific errors
	}

	return response.Respond(c, updatedArticleDTO)
}

// RemoveTagsFromArticle godoc
// @Summary Remove tags from an article
// @Description Removes one or more tags from a specific article. Ignores tags that are not associated or do not exist.
// @Tags articles
// @Accept json
// @Produce json
// @Param id path string true "Article ID" format(uuid)
// @Param tags body entity.RemoveTagsRequest true "List of tag names to remove"
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=entity.ArticleDetailResponse} "Successfully removed tags (if found) and returned updated article details"
// @Failure 400 {object} response.ErrorResponse "Invalid Article ID format, invalid JSON, or validation error on tags (code: 00004/00003)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001)"
// @Failure 404 {object} response.ErrorResponse "Article not found (code: 02001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (failed to remove tags, etc. - code: 00001)"
// @Router /articles/{id}/tags [delete] // Using DELETE for removal
func (h *ArticleHandler) RemoveTagsFromArticle(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c, h.logger)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrUnauthorized.WithDetails("Invalid user session"))
	}

	articleIDStr := c.Params("id")
	articleID, err := uuid.Parse(articleIDStr)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"id": "invalid article ID format"}))
	}

	var req entity.RemoveTagsRequest
	if err := c.BodyParser(&req); err != nil {
		// Note: DELETE requests might not always have a body depending on client/standard interpretation.
		// Consider allowing empty body or using query params if this becomes an issue.
		return response.HandleError(c, h.logger, errmsg.ErrBindJSON)
	}

	// Validate request body
	if err := h.validate.Struct(&req); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	updatedArticleDTO, err := h.articleService.RemoveTagsFromArticle(c.Context(), articleID, userID, req.Tags)
	if err != nil {
		return response.HandleError(c, h.logger, err) // Service layer returns specific errors
	}

	return response.Respond(c, updatedArticleDTO)
}

// UpdateArticleCategory godoc
// @Summary Change an article's category
// @Description Updates the category associated with a specific article.
// @Tags articles
// @Accept json
// @Produce json
// @Param id path string true "Article ID" format(uuid)
// @Param category body entity.UpdateArticleCategoryRequest true "New category name"
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=entity.ArticleDetailResponse} "Successfully updated category and returned updated article details"
// @Failure 400 {object} response.ErrorResponse "Invalid Article ID format, invalid JSON, or validation error (e.g., category not found/invalid) (code: 00004/00003)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001)"
// @Failure 404 {object} response.ErrorResponse "Article not found or target category not found for user (code: 02001/00004)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001)"
// @Router /articles/{id}/category [patch] // Using PATCH as it partially modifies the article resource
func (h *ArticleHandler) UpdateArticleCategory(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c, h.logger)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrUnauthorized.WithDetails("Invalid user session"))
	}

	articleIDStr := c.Params("id")
	articleID, err := uuid.Parse(articleIDStr)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"id": "invalid article ID format"}))
	}

	var req entity.UpdateArticleCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrBindJSON)
	}

	// Validate request body
	if err := h.validate.Struct(&req); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	updatedArticleDTO, err := h.articleService.UpdateArticleCategory(c.Context(), articleID, userID, req.Category)
	if err != nil {
		return response.HandleError(c, h.logger, err) // Service layer returns specific errors (NotFound, Validation for category, DB)
	}

	return response.Respond(c, updatedArticleDTO)
}
