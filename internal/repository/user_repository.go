package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/hackerchai/memento/internal/entity"
	"github.com/hackerchai/memento/pkg/xlog"
)

// UserRepository handles database operations for users.
type UserRepository struct {
	db     *bun.DB
	logger *xlog.Logger
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *bun.DB, logger *xlog.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger.With().Str("repository", "User").Logger(),
	}
}

// Create inserts a new user into the database.
// It checks if the email already exists.
func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	// Check if email already exists
	exists, err := r.db.NewSelect().Model((*entity.User)(nil)).Where("email = ?", user.Email).Exists(ctx)
	if err != nil {
		r.logger.ErrorX(ctx).Err(err).Str("email", user.Email).Msg("Failed check user existence")
		return err
	}
	if exists {
		return errors.New("email already exists")
	}

	_, err = r.db.NewInsert().Model(user).Exec(ctx)
	if err != nil {
		r.logger.ErrorX(ctx).Err(err).Str("email", user.Email).Msg("Failed to insert new user")
		return err
	}
	return nil
}

// FindByEmail retrieves a user by their email address.
// Returns sql.ErrNoRows if the user is not found.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	user := new(entity.User)
	err := r.db.NewSelect().Model(user).Where("email = ?", email).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		r.logger.ErrorX(ctx).Err(err).Str("email", email).Msg("Failed to find user by email")
		return nil, err
	}
	return user, nil
}

// FindByID retrieves a user by their ID.
// Returns sql.ErrNoRows if the user is not found.
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user := new(entity.User)
	err := r.db.NewSelect().Model(user).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		r.logger.ErrorX(ctx).Err(err).Str("userID", id.String()).Msg("Failed to find user by ID")
		return nil, err
	}
	return user, nil
}

// Update updates a user's information in the database.
func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	_, err := r.db.NewUpdate().Model(user).WherePK().Exec(ctx)
	if err != nil {
		r.logger.ErrorX(ctx).Err(err).Str("userID", user.ID.String()).Msg("Failed to update user")
		return err
	}
	return nil
}

// Delete deletes a user from the database.
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.NewDelete().Model((*entity.User)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		r.logger.ErrorX(ctx).Err(err).Str("userID", id.String()).Msg("Failed to delete user")
		return err
	}
	return nil
}

// CheckEmailExists checks if an email exists for any user except the given ID.
func (r *UserRepository) CheckEmailExists(ctx context.Context, email string, excludeID uuid.UUID) (bool, error) {
	exists, err := r.db.NewSelect().
		Model((*entity.User)(nil)).
		Where("email = ? AND id != ?", email, excludeID).
		Exists(ctx)
	if err != nil {
		r.logger.ErrorX(ctx).Err(err).Str("email", email).Msg("Failed to check email existence")
		return false, err
	}
	return exists, nil
}

// FindAll retrieves a paginated list of users.
func (r *UserRepository) FindAll(ctx context.Context, limit, offset int) ([]entity.User, int, error) {
	var users []entity.User
	count, err := r.db.NewSelect().
		Model(&users).
		Order("created_at DESC"). // Or order by name, email, etc.
		Limit(limit).
		Offset(offset).
		ScanAndCount(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []entity.User{}, 0, nil // Return empty list if no users found
		}
		r.logger.ErrorX(ctx).Err(err).Msg("Failed to find all users")
		return nil, 0, err
	}
	return users, count, nil
}
