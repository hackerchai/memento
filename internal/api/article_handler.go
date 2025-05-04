package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
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

// CreateArticleRequest defines the expected request body for creating an article.
type CreateArticleRequest struct {
	URL          string   `json:"url" validate:"required,url"`
	CategoryName string   `json:"category_name,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}

// CreateArticle godoc
// @Summary Submit a URL to save as an article
// @Description Accepts a URL, optionally a category name and tags. Creates a placeholder article immediately and processes the full content in the background.
// @Tags articles
// @Accept json
// @Produce json
// @Param article body CreateArticleRequest true "Article URL and optional metadata"
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

	var req CreateArticleRequest
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

// TODO: Add handlers for GetArticle(s), UpdateArticle, DeleteArticle later
