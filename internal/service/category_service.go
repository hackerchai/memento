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

// CategoryService provides business logic for category operations.
type CategoryService struct {
	categoryRepo *repository.CategoryRepository
	articleRepo  *repository.ArticleRepository // Needed for unlinking articles
	logger       *xlog.Logger
}

// NewCategoryService creates a new CategoryService.
func NewCategoryService(categoryRepo *repository.CategoryRepository, articleRepo *repository.ArticleRepository, logger *xlog.Logger) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
		articleRepo:  articleRepo, // Store articleRepo
		logger:       logger.With().Str("service", "CategoryService").Logger(),
	}
}

// ListCategories retrieves a paginated list of categories for the authenticated user.
func (s *CategoryService) ListCategories(ctx context.Context, userID uuid.UUID, pagination *response.PaginationRequest) (*response.PaginationResponse, error) {
	log := s.logger.With().Stringer("userID", userID).Logger()

	limit := pagination.GetLimit()
	offset := pagination.GetOffset()

	categories, totalCount, err := s.categoryRepo.FindByUserID(ctx, userID, limit, offset)
	if err != nil {
		log.ErrorX(ctx).Err(err).Msg("Failed to find categories by user ID")
		return nil, errmsg.ErrDatabase
	}

	categoryDTOs := make([]*entity.CategoryDetailResponse, len(categories))
	for i, category := range categories {
		categoryDTOs[i] = category.ToDetailResponseDTO()
	}

	return &response.PaginationResponse{
		Total: totalCount,
		Page:  pagination.Page,
		Data:  categoryDTOs,
	}, nil
}

// DeleteCategory deletes a category by its ID for the authenticated user.
// It also unlinks articles from this category (sets their category_id to NULL).
func (s *CategoryService) DeleteCategory(ctx context.Context, categoryID uuid.UUID, userID uuid.UUID) error {
	log := s.logger.With().Stringer("categoryID", categoryID).Stringer("userID", userID).Logger()

	// The repository method DeleteAndUnlinkArticles handles the transaction.
	err := s.categoryRepo.DeleteAndUnlinkArticles(ctx, categoryID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.WarnX(ctx).Msg("Category not found or not authorized for deletion")
			return errmsg.ErrCategoryNotFound // Use specific error for category not found
		}
		log.ErrorX(ctx).Err(err).Msg("Failed to delete category and unlink articles")
		return errmsg.ErrDatabase
	}

	log.InfoX(ctx).Msg("Category deleted successfully")
	return nil
}

// GetArticlesByCategory retrieves a paginated list of articles associated with a specific category for the user.
// The category can be identified by its ID or slug.
func (s *CategoryService) GetArticlesByCategory(ctx context.Context, userID uuid.UUID, categoryIdentifier string, pagination *response.PaginationRequest) (*response.PaginationResponse, error) {
	log := s.logger.With().Stringer("userID", userID).Str("categoryIdentifier", categoryIdentifier).Logger()

	// 1. Find the category by ID or Slug to confirm it exists and belongs to the user.
	category, err := s.categoryRepo.FindByIDOrSlug(ctx, categoryIdentifier, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.WarnX(ctx).Msg("Category not found or not authorized")
			return nil, errmsg.ErrCategoryNotFound
		}
		log.ErrorX(ctx).Err(err).Msg("Failed to find category by identifier")
		return nil, errmsg.ErrDatabase
	}

	// 2. Fetch articles for this category ID.
	limit := pagination.GetLimit()
	offset := pagination.GetOffset()

	articles, totalCount, err := s.articleRepo.FindArticlesByCategoryID(ctx, category.ID, userID, limit, offset)
	if err != nil {
		// sql.ErrNoRows is handled by returning an empty list, not an error here by the repo.
		log.ErrorX(ctx).Err(err).Stringer("categoryID", category.ID).Msg("Failed to find articles by category ID")
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

// SearchCategories searches categories by name or slug for the authenticated user.
func (s *CategoryService) SearchCategories(ctx context.Context, userID uuid.UUID, query string, pagination *response.PaginationRequest) (*response.PaginationResponse, error) {
	log := s.logger.With().Stringer("userID", userID).Str("query", query).Logger()

	if strings.TrimSpace(query) == "" {
		log.WarnX(ctx).Msg("Search query is empty, returning empty result")
		return &response.PaginationResponse{
			Total: 0,
			Page:  pagination.Page,
			Data:  []*entity.CategoryDetailResponse{},
		}, nil
	}

	limit := pagination.GetLimit()
	offset := pagination.GetOffset()

	categories, totalCount, err := s.categoryRepo.SearchByNameOrSlugForUser(ctx, userID, query, limit, offset)
	if err != nil {
		log.ErrorX(ctx).Err(err).Msg("Failed to search categories")
		return nil, errmsg.ErrDatabase
	}

	categoryDTOs := make([]*entity.CategoryDetailResponse, len(categories))
	for i, category := range categories {
		categoryDTOs[i] = category.ToDetailResponseDTO()
	}

	return &response.PaginationResponse{
		Total: totalCount,
		Page:  pagination.Page,
		Data:  categoryDTOs,
	}, nil
}

// CreateCategory creates a new category for the authenticated user.
func (s *CategoryService) CreateCategory(ctx context.Context, userID uuid.UUID, req *entity.CreateCategoryRequest) (*entity.CategoryDetailResponse, error) {
	log := s.logger.With().Stringer("userID", userID).Str("name", req.Name).Logger()

	newCategory := &entity.Category{
		UserID: userID,
		Name:   req.Name,
		// Slug will be generated by repository.Create
	}

	if err := s.categoryRepo.Create(ctx, newCategory); err != nil {
		// Check for specific uniqueness errors if the repository returns them discernibly
		if strings.Contains(err.Error(), "already exists") { // Basic check
			log.WarnX(ctx).Err(err).Msg("Failed to create category due to conflict")
			// TODO: Consider mapping to a more specific errmsg.ErrCategoryConflict or similar
			return nil, errmsg.ErrDatabase.WithDetails(err.Error()) // Or a specific conflict error
		}
		log.ErrorX(ctx).Err(err).Msg("Failed to create category in repository")
		return nil, errmsg.ErrDatabase
	}

	return newCategory.ToDetailResponseDTO(), nil
}

// --- Root Operations --- //

// ListCategoriesRoot retrieves a paginated list of categories for a specific target user (Root operation).
func (s *CategoryService) ListCategoriesRoot(ctx context.Context, targetUserID uuid.UUID, pagination *response.PaginationRequest) (*response.PaginationResponse, error) {
	log := s.logger.With().Stringer("targetUserID", targetUserID).Logger()
	limit := pagination.GetLimit()
	offset := pagination.GetOffset()

	categories, totalCount, err := s.categoryRepo.FindByUserID(ctx, targetUserID, limit, offset)

	if err != nil {
		log.ErrorX(ctx).Err(err).Msg("[Root] Failed to list categories for target user")
		return nil, errmsg.ErrDatabase
	}

	categoryDTOs := make([]*entity.CategoryDetailResponse, len(categories))
	for i, category := range categories {
		categoryDTOs[i] = category.ToDetailResponseDTO()
	}

	return &response.PaginationResponse{
		Total: totalCount,
		Page:  pagination.Page,
		Data:  categoryDTOs,
	}, nil
}

// DeleteCategoryRoot deletes a category by its ID (Root operation).
// It finds the category to determine its owner, then performs the deletion.
func (s *CategoryService) DeleteCategoryRoot(ctx context.Context, categoryID uuid.UUID) error {
	log := s.logger.With().Stringer("categoryID", categoryID).Logger()

	// 1. Find the category to get its UserID for the actual delete operation.
	category, err := s.categoryRepo.FindByIDRegardlessOfUser(ctx, categoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.WarnX(ctx).Msg("[Root] Category not found for deletion")
			return errmsg.ErrCategoryNotFound
		}
		log.ErrorX(ctx).Err(err).Msg("[Root] Failed to find category by ID for deletion")
		return errmsg.ErrDatabase
	}

	// 2. Perform deletion using the category's actual UserID.
	err = s.categoryRepo.DeleteAndUnlinkArticles(ctx, category.ID, category.UserID)
	if err != nil {
		// DeleteAndUnlinkArticles already returns sql.ErrNoRows if the final delete fails to find the record
		if errors.Is(err, sql.ErrNoRows) { // Should ideally not happen if FindByIDRegardlessOfUser found it, but handle defensively.
			log.WarnX(ctx).Stringer("ownerUserID", category.UserID).Msg("[Root] Category not found during deletion by owner ID (race condition?)")
			return errmsg.ErrCategoryNotFound
		}
		log.ErrorX(ctx).Err(err).Stringer("ownerUserID", category.UserID).Msg("[Root] Failed to delete category and unlink articles")
		return errmsg.ErrDatabase
	}

	return nil
}

// GetArticlesByCategoryRoot retrieves articles for a category identifier (ID or Slug) (Root operation).
// It finds the category regardless of user, then fetches its articles.
func (s *CategoryService) GetArticlesByCategoryRoot(ctx context.Context, categoryIdentifier string, pagination *response.PaginationRequest) (*response.PaginationResponse, error) {
	log := s.logger.With().Str("categoryIdentifier", categoryIdentifier).Logger()

	// 1. Find the category by ID or Slug, regardless of user.
	var category *entity.Category
	var err error

	parsedID, parseErr := uuid.Parse(categoryIdentifier)
	if parseErr == nil {
		category, err = s.categoryRepo.FindByIDRegardlessOfUser(ctx, parsedID)
	} else {
		category, err = s.categoryRepo.FindBySlugRegardlessOfUser(ctx, categoryIdentifier)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.WarnX(ctx).Msg("[Root] Category not found by identifier")
			return nil, errmsg.ErrCategoryNotFound
		}
		log.ErrorX(ctx).Err(err).Msg("[Root] Failed to find category by identifier")
		return nil, errmsg.ErrDatabase
	}

	// 2. Fetch articles for this category using its ID and its UserID.
	limit := pagination.GetLimit()
	offset := pagination.GetOffset()

	articles, totalCount, err := s.articleRepo.FindArticlesByCategoryID(ctx, category.ID, category.UserID, limit, offset)
	if err != nil {
		log.ErrorX(ctx).Err(err).Stringer("categoryID", category.ID).Stringer("ownerUserID", category.UserID).Msg("[Root] Failed to find articles by category ID")
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

// CreateCategoryRoot creates a new category for a target user (Root operation).
func (s *CategoryService) CreateCategoryRoot(ctx context.Context, req *entity.CreateCategoryRootRequest) (*entity.CategoryDetailResponse, error) {
	log := s.logger.With().Stringer("targetUserID", req.TargetUserID).Str("name", req.Name).Logger()

	newCategory := &entity.Category{
		UserID: req.TargetUserID,
		Name:   req.Name,
		// Slug will be generated by repository.Create
	}

	if err := s.categoryRepo.Create(ctx, newCategory); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			log.WarnX(ctx).Err(err).Msg("[Root] Failed to create category for target user due to conflict")
			return nil, errmsg.ErrDatabase.WithDetails(err.Error()) // Or a specific conflict error
		}
		log.ErrorX(ctx).Err(err).Msg("[Root] Failed to create category for target user in repository")
		return nil, errmsg.ErrDatabase
	}

	return newCategory.ToDetailResponseDTO(), nil
}
