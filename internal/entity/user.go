package entity

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Role constants
const (
	RoleUser  = 0
	RoleAdmin = 1
	RoleRoot  = 2
)

// User represents the user model in the database.
type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID        uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	Name      string    `bun:"name,notnull" json:"name"`
	Email     string    `bun:"email,notnull,unique" json:"email"`
	Password  string    `bun:"password,notnull" json:"-"`       // Hashed password
	Role      int       `bun:"role,notnull,default:0" json:"-"` // Role: 0=User, 1=Admin, 2=Root
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`

	// Fields for future enhancements
	TOTPSecret         *string `bun:"totp_secret" json:"-"`          // For Time-based One-Time Password
	PasskeyData        []byte  `bun:"passkey_data" json:"-"`         // For Passkey authentication data (e.g., JSON)
	ThirdPartyProvider *string `bun:"third_party_provider" json:"-"` // e.g., "google", "github"
	ThirdPartyUserID   *string `bun:"third_party_user_id" json:"-"`  // User ID from the third-party provider
}

// BeforeAppendModel generates a UUID for the user if ID is zero.
var _ bun.BeforeAppendModelHook = (*User)(nil)

func (u *User) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	// Set default role if not explicitly set (though DB default should handle it)
	// if u.Role == 0 && u.ID != uuid.Nil { // Check if it's not a new user being created with default
	//  // No, default should be handled by DB. Keep Go struct simple.
	// }
	return nil
}

// Basic registration request structure
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=7,max=72"`
}

// Basic login request structure
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Basic login response structure
type LoginResponse struct {
	Token string      `json:"token"`
	User  *UserPublic `json:"user"`
}

// UserPublic is a subset of User fields safe to expose publicly.
type UserPublic struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// ToPublic converts a User to a UserPublic struct.
func (u *User) ToPublic() *UserPublic {
	if u == nil {
		return nil
	}
	return &UserPublic{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}

// RoleString returns the role as a string
func (u *User) RoleString() string {
	switch u.Role {
	case RoleUser:
		return "user"
	case RoleAdmin:
		return "admin"
	case RoleRoot:
		return "root"
	default:
		return "unknown"
	}
}

// UpdatePasswordRequest represents the request to update a user's password
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=7,max=72"`
}

// UpdateEmailRequest represents the request to update a user's email
type UpdateEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// UpdateNameRequest represents the request to update a user's name
type UpdateNameRequest struct {
	Name string `json:"name" validate:"required,min=2,max=50"`
}

// CreateUserRequest represents the request to create a new user by root
type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=7,max=72"`
	Role     int    `json:"role" validate:"oneof=0 1 2"`
}

// UserSelf represents the data returned for the self-profile endpoint.
type UserSelf struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"` // Role as string for clarity
	CreatedAt time.Time `json:"created_at"`
}

// ToSelf converts a User to a UserSelf struct.
func (u *User) ToSelf() *UserSelf {
	if u == nil {
		return nil
	}
	return &UserSelf{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      u.RoleString(),
		CreatedAt: u.CreatedAt,
	}
}
