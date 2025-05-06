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

// TagHandler handles HTTP requests related to tags.
type TagHandler struct {
	tagService *service.TagService
	validate   *validator.Validate
	logger     *xlog.Logger
}

// NewTagHandler creates a new TagHandler.
func NewTagHandler(tagService *service.TagService, logger *xlog.Logger) *TagHandler {
	return &TagHandler{
		tagService: tagService,
		validate:   validator.New(),
		logger:     logger.With().Str("handler", "TagHandler").Logger(),
	}
}

// ListTags godoc
// @Summary List tags
// @Description Retrieves a paginated list of tags for the authenticated user.
// @Tags tags
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param per_page query int false "Items per page" default(10) minimum(1) maximum(100)
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=response.PaginationResponse{data=[]entity.TagDetailResponse}} "Successfully retrieved tags"
// @Failure 400 {object} response.ErrorResponse "Invalid pagination parameters (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (code: 01001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001)"
// @Router /tags [get]
func (h *TagHandler) ListTags(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c, h.logger)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrUnauthorized.WithDetails("Invalid user session"))
	}

	var pagination response.PaginationRequest
	if err := c.QueryParser(&pagination); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"query": "invalid pagination parameters"}))
	}
	if err := h.validate.Struct(&pagination); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	tags, err := h.tagService.ListTags(c.Context(), userID, &pagination)
	if err != nil {
		return response.HandleError(c, h.logger, err)
	}
	return response.Respond(c, tags)
}

// DeleteTag godoc
// @Summary Delete a tag
// @Description Deletes a tag by its ID for the authenticated user. Associated article-tag links will also be removed.
// @Tags tags
// @Accept json
// @Produce json
// @Param id path string true "Tag ID" format(uuid)
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse "Tag deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid tag ID format (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (code: 01001)"
// @Failure 404 {object} response.ErrorResponse "Tag not found (code: 02001 with details)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001)"
// @Router /tags/{id} [delete]
func (h *TagHandler) DeleteTag(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c, h.logger)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrUnauthorized.WithDetails("Invalid user session"))
	}

	tagIDStr := c.Params("id")
	tagID, err := uuid.Parse(tagIDStr)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"id": "invalid tag ID format"}))
	}

	if err := h.tagService.DeleteTag(c.Context(), tagID, userID); err != nil {
		return response.HandleError(c, h.logger, err)
	}
	return response.Respond(c, nil)
}

// GetArticlesByTag godoc
// @Summary Get articles by tag
// @Description Retrieves a paginated list of articles associated with a specific tag (by ID or slug) for the authenticated user. Articles are ordered by creation date (newest first).
// @Tags tags,articles
// @Accept json
// @Produce json
// @Param identifier path string true "Tag ID or Slug"
// @Param page query int false "Page number" default(1) minimum(1)
// @Param per_page query int false "Items per page" default(10) minimum(1) maximum(100)
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=response.PaginationResponse{data=[]entity.ArticleResponse}} "Successfully retrieved articles"
// @Failure 400 {object} response.ErrorResponse "Invalid pagination parameters or tag identifier (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (code: 01001)"
// @Failure 404 {object} response.ErrorResponse "Tag not found (code: 02001 with details)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001)"
// @Router /tags/{identifier}/articles [get]
func (h *TagHandler) GetArticlesByTag(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c, h.logger)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrUnauthorized.WithDetails("Invalid user session"))
	}

	tagIdentifier := c.Params("identifier")
	if tagIdentifier == "" {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"identifier": "tag identifier cannot be empty"}))
	}

	var pagination response.PaginationRequest
	if err := c.QueryParser(&pagination); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"query": "invalid pagination parameters"}))
	}
	if err := h.validate.Struct(&pagination); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	articles, err := h.tagService.GetArticlesByTag(c.Context(), userID, tagIdentifier, &pagination)
	if err != nil {
		return response.HandleError(c, h.logger, err)
	}
	return response.Respond(c, articles)
}

// SearchTags godoc
// @Summary Search tags
// @Description Searches tags by name or slug for the authenticated user. Case-insensitive.
// @Tags tags
// @Accept json
// @Produce json
// @Param q query string true "Search query term"
// @Param page query int false "Page number" default(1) minimum(1)
// @Param per_page query int false "Items per page" default(10) minimum(1) maximum(100)
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=response.PaginationResponse{data=[]entity.TagDetailResponse}} "Successfully retrieved tags matching the query"
// @Failure 400 {object} response.ErrorResponse "Invalid pagination parameters or missing query (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (code: 01001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001)"
// @Router /tags/search [get]
func (h *TagHandler) SearchTags(c *fiber.Ctx) error {
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

	tags, err := h.tagService.SearchTags(c.Context(), userID, searchQuery, &pagination)
	if err != nil {
		return response.HandleError(c, h.logger, err)
	}
	return response.Respond(c, tags)
}

// CreateTag godoc
// @Summary Create a tag
// @Description Creates a new tag for the authenticated user.
// @Tags tags
// @Accept json
// @Produce json
// @Param tag body entity.CreateTagRequest true "Tag details"
// @Security BearerAuth
// @Success 201 {object} response.SuccessResponse{data=entity.TagDetailResponse} "Tag created successfully"
// @Failure 400 {object} response.ErrorResponse "Validation error or invalid JSON (code: 00004 or 00003)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (code: 01001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error or conflict (code: 00001 or 00006)"
// @Router /tags [post]
func (h *TagHandler) CreateTag(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c, h.logger)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrUnauthorized.WithDetails("Invalid user session"))
	}

	var req entity.CreateTagRequest
	if err := c.BodyParser(&req); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrBindJSON)
	}
	if err := h.validate.Struct(&req); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	tag, err := h.tagService.CreateTag(c.Context(), userID, &req)
	if err != nil {
		return response.HandleError(c, h.logger, err) // Service layer handles specific errors
	}

	return response.RespondCreated(c, tag)
}

// --- Root Operations --- //

// ListTagsRoot godoc
// @Summary List tags for a user (Root)
// @Description Retrieves a paginated list of tags for a specified target user. Requires root privileges.
// @Tags tags,root
// @Accept json
// @Produce json
// @Param target_user_id query string true "Target User ID to filter tags for" format(uuid)
// @Param page query int false "Page number" default(1) minimum(1)
// @Param per_page query int false "Items per page" default(10) minimum(1) maximum(100)
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=response.PaginationResponse{data=[]entity.TagDetailResponse}} "Successfully retrieved tags"
// @Failure 400 {object} response.ErrorResponse "Invalid pagination or missing/invalid target_user_id (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (code: 01001)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (not root user - code: 01011)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001)"
// @Router /users/root/tags [get]
func (h *TagHandler) ListTagsRoot(c *fiber.Ctx) error {
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

	tags, err := h.tagService.ListTagsRoot(c.Context(), parsedID, &pagination)
	if err != nil {
		return response.HandleError(c, h.logger, err)
	}
	return response.Respond(c, tags)
}

// DeleteTagRoot godoc
// @Summary Delete a tag (Root)
// @Description Deletes a tag by its ID. Requires root privileges.
// @Tags tags,root
// @Accept json
// @Produce json
// @Param id path string true "Tag ID to delete" format(uuid)
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse "Tag deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid tag ID format (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (code: 01001)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (not root user - code: 01011)"
// @Failure 404 {object} response.ErrorResponse "Tag not found (code: 02001 with details)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001)"
// @Router /users/root/tags/{id} [delete]
func (h *TagHandler) DeleteTagRoot(c *fiber.Ctx) error {
	tagIDStr := c.Params("id")
	tagID, err := uuid.Parse(tagIDStr)
	if err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"id": "invalid tag ID format"}))
	}

	if err := h.tagService.DeleteTagRoot(c.Context(), tagID); err != nil {
		return response.HandleError(c, h.logger, err)
	}
	return response.Respond(c, nil)
}

// GetArticlesByTagRoot godoc
// @Summary Get articles by tag (Root)
// @Description Retrieves a paginated list of articles for a specific tag (by ID or slug). Requires root privileges.
// @Tags tags,articles,root
// @Accept json
// @Produce json
// @Param identifier path string true "Tag ID or Slug"
// @Param page query int false "Page number" default(1) minimum(1)
// @Param per_page query int false "Items per page" default(10) minimum(1) maximum(100)
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=response.PaginationResponse{data=[]entity.ArticleResponse}} "Successfully retrieved articles"
// @Failure 400 {object} response.ErrorResponse "Invalid pagination or tag identifier (code: 00004)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (code: 01001)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (not root user - code: 01011)"
// @Failure 404 {object} response.ErrorResponse "Tag not found (code: 02001 with details)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (code: 00001)"
// @Router /users/root/tags/{identifier}/articles [get]
func (h *TagHandler) GetArticlesByTagRoot(c *fiber.Ctx) error {
	tagIdentifier := c.Params("identifier")
	if tagIdentifier == "" {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"identifier": "tag identifier cannot be empty"}))
	}

	var pagination response.PaginationRequest
	if err := c.QueryParser(&pagination); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrValidation.WithDetails(map[string]string{"query": "invalid pagination parameters"}))
	}
	if err := h.validate.Struct(&pagination); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	articles, err := h.tagService.GetArticlesByTagRoot(c.Context(), tagIdentifier, &pagination)
	if err != nil {
		return response.HandleError(c, h.logger, err)
	}
	return response.Respond(c, articles)
}

// CreateTagRoot godoc
// @Summary Create a tag for a user (Root)
// @Description Creates a new tag for a specified target user. Requires root privileges.
// @Tags tags,root
// @Accept json
// @Produce json
// @Param tag body entity.CreateTagRootRequest true "Tag details including target user ID"
// @Security BearerAuth
// @Success 201 {object} response.SuccessResponse{data=entity.TagDetailResponse} "Tag created successfully"
// @Failure 400 {object} response.ErrorResponse "Validation error or invalid JSON (code: 00004 or 00003)"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (code: 01001)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (not root user - code: 01011)"
// @Failure 500 {object} response.ErrorResponse "Internal server error or conflict (code: 00001 or 00006)"
// @Router /users/root/tags [post]
func (h *TagHandler) CreateTagRoot(c *fiber.Ctx) error {
	var req entity.CreateTagRootRequest
	if err := c.BodyParser(&req); err != nil {
		return response.HandleError(c, h.logger, errmsg.ErrBindJSON)
	}
	if err := h.validate.Struct(&req); err != nil {
		return response.HandleError(c, h.logger, err)
	}

	tag, err := h.tagService.CreateTagRoot(c.Context(), &req)
	if err != nil {
		return response.HandleError(c, h.logger, err)
	}

	return response.RespondCreated(c, tag)
}
