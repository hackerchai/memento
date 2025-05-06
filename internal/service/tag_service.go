package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/hackerchai/memento/internal/entity"
	"github.com/hackerchai/memento/internal/errmsg"
	"github.com/hackerchai/memento/internal/repository"
	"github.com/hackerchai/memento/internal/response"
	"github.com/hackerchai/memento/pkg/xlog"
)

// TagService provides business logic for tag operations.
type TagService struct {
	tagRepo     *repository.TagRepository
	articleRepo *repository.ArticleRepository // Needed for fetching articles by tag
	logger      *xlog.Logger
}

// NewTagService creates a new TagService.
func NewTagService(tagRepo *repository.TagRepository, articleRepo *repository.ArticleRepository, logger *xlog.Logger) *TagService {
	return &TagService{
		tagRepo:     tagRepo,
		articleRepo: articleRepo,
		logger:      logger.With().Str("service", "TagService").Logger(),
	}
}

// ListTags retrieves a paginated list of tags for the authenticated user.
func (s *TagService) ListTags(ctx context.Context, userID uuid.UUID, pagination *response.PaginationRequest) (*response.PaginationResponse, error) {
	log := s.logger.With().Stringer("userID", userID).Logger()

	limit := pagination.GetLimit()
	offset := pagination.GetOffset()

	tags, totalCount, err := s.tagRepo.FindByUserID(ctx, userID, limit, offset)
	if err != nil {
		log.ErrorX(ctx).Err(err).Msg("Failed to find tags by user ID")
		return nil, errmsg.ErrDatabase
	}

	tagDTOs := make([]*entity.TagDetailResponse, len(tags))
	for i, tag := range tags {
		tagDTOs[i] = tag.ToDetailResponseDTO()
	}

	return &response.PaginationResponse{
		Total: totalCount,
		Page:  pagination.Page,
		Data:  tagDTOs,
	}, nil
}

// DeleteTag deletes a tag by its ID for the authenticated user.
// It also removes associations from article_tags table.
func (s *TagService) DeleteTag(ctx context.Context, tagID uuid.UUID, userID uuid.UUID) error {
	log := s.logger.With().Stringer("tagID", tagID).Stringer("userID", userID).Logger()

	// The repository method DeleteTagAndAssociations handles the transaction.
	err := s.tagRepo.DeleteTagAndAssociations(ctx, tagID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.WarnX(ctx).Msg("Tag not found or not authorized for deletion")
			// Consider a specific errmsg.ErrTagNotFound if desired, using ErrRecordNotFound for now as a generic
			return errmsg.ErrRecordNotFound.WithDetails("Tag not found")
		}
		log.ErrorX(ctx).Err(err).Msg("Failed to delete tag and associations")
		return errmsg.ErrDatabase
	}

	return nil
}

// GetArticlesByTag retrieves a paginated list of articles associated with a specific tag for the user.
// The tag can be identified by its ID or slug.
func (s *TagService) GetArticlesByTag(ctx context.Context, userID uuid.UUID, tagIdentifier string, pagination *response.PaginationRequest) (*response.PaginationResponse, error) {
	log := s.logger.With().Stringer("userID", userID).Str("tagIdentifier", tagIdentifier).Logger()

	// 1. Find the tag by ID or Slug to confirm it exists and belongs to the user.
	tag, err := s.tagRepo.FindByIDOrSlug(ctx, tagIdentifier, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.WarnX(ctx).Msg("Tag not found or not authorized")
			return nil, errmsg.ErrRecordNotFound.WithDetails("Tag not found") // Consistent error for tag not found
		}
		log.ErrorX(ctx).Err(err).Msg("Failed to find tag by identifier")
		return nil, errmsg.ErrDatabase
	}

	// 2. Fetch articles for this tag ID.
	limit := pagination.GetLimit()
	offset := pagination.GetOffset()

	articles, totalCount, err := s.articleRepo.FindArticlesByTagID(ctx, tag.ID, userID, limit, offset)
	if err != nil {
		log.ErrorX(ctx).Err(err).Stringer("tagID", tag.ID).Msg("Failed to find articles by tag ID")
		return nil, errmsg.ErrDatabase
	}

	articleDTOs := make([]*entity.ArticleResponse, len(articles))
	for i, article := range articles {
		articleDTOs[i] = article.ToResponseDTO() // Using existing ArticleResponse DTO
	}

	return &response.PaginationResponse{
		Total: totalCount,
		Page:  pagination.Page,
		Data:  articleDTOs,
	}, nil
}

// SearchTags searches tags by name or slug for the authenticated user.
func (s *TagService) SearchTags(ctx context.Context, userID uuid.UUID, query string, pagination *response.PaginationRequest) (*response.PaginationResponse, error) {
	log := s.logger.With().Stringer("userID", userID).Str("query", query).Logger()

	if strings.TrimSpace(query) == "" {
		log.WarnX(ctx).Msg("Search query is empty, returning empty result")
		return &response.PaginationResponse{
			Total: 0,
			Page:  pagination.Page,
			Data:  []*entity.TagDetailResponse{},
		}, nil
	}

	limit := pagination.GetLimit()
	offset := pagination.GetOffset()

	tags, totalCount, err := s.tagRepo.SearchByNameOrSlugForUser(ctx, userID, query, limit, offset)
	if err != nil {
		log.ErrorX(ctx).Err(err).Msg("Failed to search tags")
		return nil, errmsg.ErrDatabase
	}

	tagDTOs := make([]*entity.TagDetailResponse, len(tags))
	for i, tag := range tags {
		tagDTOs[i] = tag.ToDetailResponseDTO()
	}

	return &response.PaginationResponse{
		Total: totalCount,
		Page:  pagination.Page,
		Data:  tagDTOs,
	}, nil
}

// CreateTag creates a new tag for the authenticated user.
func (s *TagService) CreateTag(ctx context.Context, userID uuid.UUID, req *entity.CreateTagRequest) (*entity.TagDetailResponse, error) {
	log := s.logger.With().Stringer("userID", userID).Str("name", req.Name).Logger()

	newTag := &entity.Tag{
		UserID: userID,
		Name:   req.Name,
		// Slug will be generated by repository.Create
	}

	if err := s.tagRepo.Create(ctx, newTag); err != nil {
		// Check for specific uniqueness errors
		if strings.Contains(err.Error(), "already exists") { // Basic check
			log.WarnX(ctx).Err(err).Msg("Failed to create tag due to conflict")
			// TODO: Consider mapping to a more specific errmsg.ErrTagConflict or similar
			return nil, errmsg.ErrDatabase.WithDetails(err.Error()) // Or a specific conflict error
		}
		log.ErrorX(ctx).Err(err).Msg("Failed to create tag in repository")
		return nil, errmsg.ErrDatabase
	}

	return newTag.ToDetailResponseDTO(), nil
}

// --- Root Operations --- //

// ListTagsRoot retrieves a paginated list of tags for a specific target user (Root operation).
func (s *TagService) ListTagsRoot(ctx context.Context, targetUserID uuid.UUID, pagination *response.PaginationRequest) (*response.PaginationResponse, error) {
	log := s.logger.With().Stringer("targetUserID", targetUserID).Logger()
	limit := pagination.GetLimit()
	offset := pagination.GetOffset()

	tags, totalCount, err := s.tagRepo.FindByUserID(ctx, targetUserID, limit, offset)

	if err != nil {
		log.ErrorX(ctx).Err(err).Msg("[Root] Failed to list tags for target user")
		return nil, errmsg.ErrDatabase
	}

	tagDTOs := make([]*entity.TagDetailResponse, len(tags))
	for i, tag := range tags {
		tagDTOs[i] = tag.ToDetailResponseDTO()
	}

	return &response.PaginationResponse{
		Total: totalCount,
		Page:  pagination.Page,
		Data:  tagDTOs,
	}, nil
}

// DeleteTagRoot deletes a tag by its ID (Root operation).
// It finds the tag to determine its owner, then performs the deletion.
func (s *TagService) DeleteTagRoot(ctx context.Context, tagID uuid.UUID) error {
	log := s.logger.With().Stringer("tagID", tagID).Logger()

	// 1. Find the tag to get its UserID.
	tag, err := s.tagRepo.FindByIDRegardlessOfUser(ctx, tagID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.WarnX(ctx).Msg("[Root] Tag not found for deletion")
			return errmsg.ErrRecordNotFound.WithDetails("Tag not found")
		}
		log.ErrorX(ctx).Err(err).Msg("[Root] Failed to find tag by ID for deletion")
		return errmsg.ErrDatabase
	}

	// 2. Perform deletion using the tag's actual UserID.
	err = s.tagRepo.DeleteTagAndAssociations(ctx, tag.ID, tag.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // Should not happen if FindByIDRegardlessOfUser found it
			log.WarnX(ctx).Stringer("ownerUserID", tag.UserID).Msg("[Root] Tag not found during deletion by owner ID (race condition?)")
			return errmsg.ErrRecordNotFound.WithDetails("Tag not found")
		}
		log.ErrorX(ctx).Err(err).Stringer("ownerUserID", tag.UserID).Msg("[Root] Failed to delete tag and associations")
		return errmsg.ErrDatabase
	}

	return nil
}

// GetArticlesByTagRoot retrieves articles for a tag identifier (ID or Slug) (Root operation).
// It finds the tag regardless of user, then fetches its articles.
func (s *TagService) GetArticlesByTagRoot(ctx context.Context, tagIdentifier string, pagination *response.PaginationRequest) (*response.PaginationResponse, error) {
	log := s.logger.With().Str("tagIdentifier", tagIdentifier).Logger()

	// 1. Find the tag by ID or Slug, regardless of user.
	var tag *entity.Tag
	var err error

	parsedID, parseErr := uuid.Parse(tagIdentifier)
	if parseErr == nil {
		tag, err = s.tagRepo.FindByIDRegardlessOfUser(ctx, parsedID)
	} else {
		tag, err = s.tagRepo.FindBySlugRegardlessOfUser(ctx, tagIdentifier)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.WarnX(ctx).Msg("[Root] Tag not found by identifier")
			return nil, errmsg.ErrRecordNotFound.WithDetails("Tag not found")
		}
		log.ErrorX(ctx).Err(err).Msg("[Root] Failed to find tag by identifier")
		return nil, errmsg.ErrDatabase
	}

	// 2. Fetch articles for this tag using its ID and its UserID.
	limit := pagination.GetLimit()
	offset := pagination.GetOffset()

	articles, totalCount, err := s.articleRepo.FindArticlesByTagID(ctx, tag.ID, tag.UserID, limit, offset)
	if err != nil {
		log.ErrorX(ctx).Err(err).Stringer("tagID", tag.ID).Stringer("ownerUserID", tag.UserID).Msg("[Root] Failed to find articles by tag ID")
		return nil, errmsg.ErrDatabase
	}

	articleDTOs := make([]*entity.ArticleResponse, len(articles))
	for i, article := range articles {
		articleDTOs[i] = article.ToResponseDTO()
	}

	return &response.PaginationResponse{
		Total: totalCount,
		Page:  pagination.Page,
		Data:  articleDTOs,
	}, nil
}

// CreateTagRoot creates a new tag for a target user (Root operation).
func (s *TagService) CreateTagRoot(ctx context.Context, req *entity.CreateTagRootRequest) (*entity.TagDetailResponse, error) {
	log := s.logger.With().Stringer("targetUserID", req.TargetUserID).Str("name", req.Name).Logger()

	newTag := &entity.Tag{
		UserID: req.TargetUserID,
		Name:   req.Name,
		// Slug will be generated by repository.Create
	}

	if err := s.tagRepo.Create(ctx, newTag); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			log.WarnX(ctx).Err(err).Msg("[Root] Failed to create tag for target user due to conflict")
			return nil, errmsg.ErrDatabase.WithDetails(err.Error()) // Or a specific conflict error
		}
		log.ErrorX(ctx).Err(err).Msg("[Root] Failed to create tag for target user in repository")
		return nil, errmsg.ErrDatabase
	}

	return newTag.ToDetailResponseDTO(), nil
}
