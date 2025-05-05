package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hackerchai/memento/internal/auth"
	"github.com/hackerchai/memento/internal/entity"
	"github.com/hackerchai/memento/internal/repository"
	"github.com/hackerchai/memento/internal/response"
	"github.com/hackerchai/memento/pkg/xlog"
)

// Define specific errors for login failures
var (
	ErrUserNotFoundLogin = errors.New("login failed: user not found")
	ErrInvalidPassword   = errors.New("login failed: invalid password")
)

// UserService provides user-related business logic.
type UserService struct {
	userRepo      *repository.UserRepository
	appConfigRepo *repository.AppConfigRepository
	jwtMaker      *auth.JWTMaker
	tokenDuration time.Duration
	logger        *xlog.Logger
}

// NewUserService creates a new UserService.
func NewUserService(userRepo *repository.UserRepository, appConfigRepo *repository.AppConfigRepository, jwtMaker *auth.JWTMaker, tokenDuration time.Duration, logger *xlog.Logger) *UserService {
	return &UserService{
		userRepo:      userRepo,
		appConfigRepo: appConfigRepo,
		jwtMaker:      jwtMaker,
		tokenDuration: tokenDuration,
		logger:        logger.With().Str("service", "UserService").Logger(),
	}
}

// Register creates a new user account.
func (s *UserService) Register(ctx context.Context, req *entity.RegisterRequest) (*entity.UserPublic, error) {
	s.logger.DebugX(ctx).Msg("Registering user")

	// Hash the password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		s.logger.ErrorX(ctx).Err(err).Msg("Failed to hash password during registration")
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create the user entity
	user := &entity.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	}

	// Save the user to the database
	err = s.userRepo.Create(ctx, user)
	if err != nil {
		if err.Error() == "email already exists" { // TODO: Use a custom error type
			s.logger.WarnX(ctx).Str("email", req.Email).Msg("Attempt to register with existing email")
			return nil, errors.New("email already registered")
		}
		s.logger.ErrorX(ctx).Err(err).Str("email", req.Email).Msg("Failed to create user in database")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create default AppConfig for the new user
	defaultConfig := &entity.AppConfig{
		UserID:             user.ID,
		ScrapeImgOffline:   true,
		LLMAutoGenTags:     false,
		ExtractLinks:       false,
		LLMAutoGenAbstract: false,
		BypassRefer:        false,
		// Other fields are nil/zero by default
	}
	if configErr := s.appConfigRepo.Create(ctx, defaultConfig); configErr != nil {
		// Log the error, but don't fail the registration since user creation was successful
		// The user might need to manually configure or a background job could retry later.
		s.logger.ErrorX(ctx).Err(configErr).Stringer("userID", user.ID).Msg("Failed to create default app config for new user")
		// Optionally, you could implement a retry mechanism or alert system here.
	}

	s.logger.InfoX(ctx).Str("userID", user.ID.String()).Str("email", user.Email).Msg("User registered successfully")
	return user.ToPublic(), nil
}

// Login authenticates a user and returns a JWT token.
func (s *UserService) Login(ctx context.Context, req *entity.LoginRequest) (*entity.LoginResponse, error) {
	s.logger.DebugX(ctx).Str("email", req.Email).Msg("Attempting user login")

	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.WarnX(ctx).Str("email", req.Email).Msg("Login attempt failed: user not found")
			return nil, ErrUserNotFoundLogin
		}
		s.logger.ErrorX(ctx).Err(err).Str("email", req.Email).Msg("Failed to find user by email during login")
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Compare the provided password with the stored hash
	match, err := auth.ComparePasswordAndHash(req.Password, user.Password)
	if err != nil {
		// Log internal error, return generic message
		s.logger.ErrorX(ctx).Err(err).Str("email", req.Email).Msg("Error comparing password during login")
		return nil, errors.New("internal server error during login") // Avoid exposing internal details
	}
	if !match {
		s.logger.WarnX(ctx).Str("email", req.Email).Msg("Login attempt failed: invalid password")
		return nil, ErrInvalidPassword
	}

	// Generate JWT token
	token, err := s.jwtMaker.CreateToken(user.ID, user.RoleString(), s.tokenDuration)
	if err != nil {
		s.logger.ErrorX(ctx).Err(err).Str("userID", user.ID.String()).Msg("Failed to create JWT token during login")
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	s.logger.InfoX(ctx).Str("userID", user.ID.String()).Str("email", user.Email).Msg("User logged in successfully")
	return &entity.LoginResponse{
		Token: token,
		User:  user.ToPublic(),
	}, nil
}

// UpdatePassword updates a user's password.
func (s *UserService) UpdatePassword(ctx context.Context, userID uuid.UUID, req *entity.UpdatePasswordRequest) error {
	s.logger.DebugX(ctx).Str("userID", userID.String()).Msg("Updating user password")

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		s.logger.ErrorX(ctx).Err(err).Str("userID", userID.String()).Msg("Failed to find user")
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Verify old password
	match, err := auth.ComparePasswordAndHash(req.OldPassword, user.Password)
	if err != nil {
		s.logger.ErrorX(ctx).Err(err).Str("userID", userID.String()).Msg("Error comparing password")
		return errors.New("internal server error")
	}
	if !match {
		return errors.New("incorrect old password")
	}

	// Hash new password
	hashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		s.logger.ErrorX(ctx).Err(err).Str("userID", userID.String()).Msg("Failed to hash new password")
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.Password = hashedPassword
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	s.logger.InfoX(ctx).Str("userID", userID.String()).Msg("Password updated successfully")
	return nil
}

// UpdateEmail updates a user's email.
func (s *UserService) UpdateEmail(ctx context.Context, userID uuid.UUID, req *entity.UpdateEmailRequest) error {
	s.logger.DebugX(ctx).Str("userID", userID.String()).Msg("Updating user email")

	// Check if email is already taken
	exists, err := s.userRepo.CheckEmailExists(ctx, req.Email, userID)
	if err != nil {
		return fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return errors.New("email already taken")
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	user.Email = req.Email
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update email: %w", err)
	}

	s.logger.InfoX(ctx).Str("userID", userID.String()).Str("email", req.Email).Msg("Email updated successfully")
	return nil
}

// UpdateName updates a user's name.
func (s *UserService) UpdateName(ctx context.Context, userID uuid.UUID, req *entity.UpdateNameRequest) error {
	s.logger.DebugX(ctx).Str("userID", userID.String()).Msg("Updating user name")

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	user.Name = req.Name
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update name: %w", err)
	}

	s.logger.InfoX(ctx).Str("userID", userID.String()).Str("name", req.Name).Msg("Name updated successfully")
	return nil
}

// CreateUser creates a new user (root only).
func (s *UserService) CreateUser(ctx context.Context, req *entity.CreateUserRequest) (*entity.UserPublic, error) {
	s.logger.DebugX(ctx).Msg("Creating new user by root")

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &entity.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     req.Role,
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		if err.Error() == "email already exists" {
			return nil, errors.New("email already registered")
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.InfoX(ctx).Str("userID", user.ID.String()).Str("email", user.Email).Msg("User created successfully by root")
	return user.ToPublic(), nil
}

// DeleteUser deletes a user (root only).
func (s *UserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	s.logger.DebugX(ctx).Str("userID", userID.String()).Msg("Deleting user")

	// Check if user exists
	_, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to find user: %w", err)
	}

	err = s.userRepo.Delete(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	s.logger.InfoX(ctx).Str("userID", userID.String()).Msg("User deleted successfully")
	return nil
}

// GetUserProfile retrieves the profile of a specific user by ID.
func (s *UserService) GetUserProfile(ctx context.Context, userID uuid.UUID) (*entity.UserSelf, error) {
	s.logger.DebugX(ctx).Str("userID", userID.String()).Msg("Getting user profile")

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Consider returning a specific 'not found' error
			return nil, errors.New("user not found")
		}
		s.logger.ErrorX(ctx).Err(err).Str("userID", userID.String()).Msg("Failed to find user for profile")
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return user.ToSelf(), nil
}

// GetUserList retrieves a paginated list of users (root only).
func (s *UserService) GetUserList(ctx context.Context, pagination *response.PaginationRequest) (*response.PaginationResponse, error) {
	s.logger.DebugX(ctx).Msg("Getting user list by root")

	limit := pagination.GetLimit()
	offset := pagination.GetOffset()

	users, total, err := s.userRepo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user list: %w", err)
	}

	// Convert users to public representation
	publicUsers := make([]*entity.UserPublic, len(users))
	for i, user := range users {
		publicUsers[i] = user.ToPublic()
	}

	resp := &response.PaginationResponse{
		Total: total,
		Page:  pagination.Page,
		Data:  publicUsers,
	}

	return resp, nil
}
