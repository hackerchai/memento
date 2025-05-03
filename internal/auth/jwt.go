package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Payload contains the payload data of the token
type Payload struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"` // Added Role field
	jwt.RegisteredClaims
}

// JWTMaker is a JSON Web Token maker
type JWTMaker struct {
	secretKey string
}

// NewJWTMaker creates a new JWTMaker
func NewJWTMaker(secretKey string) (*JWTMaker, error) {
	if len(secretKey) < 32 {
		return nil, fmt.Errorf("invalid key size: must be at least 32 characters")
	}
	return &JWTMaker{secretKey}, nil
}

// CreateToken creates a new token for a specific user ID, role, and duration
func (maker *JWTMaker) CreateToken(userID uuid.UUID, role string, duration time.Duration) (string, error) {
	payload := &Payload{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(), // Keep subject for standard compatibility if needed
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "memento", // Optional: Identify the issuer
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return jwtToken.SignedString([]byte(maker.secretKey))
}

// VerifyToken checks if the token is valid or not and returns the payload
func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(maker.secretKey), nil
	}

	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("token has expired")
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, fmt.Errorf("token not yet valid")
		}
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	payload, ok := jwtToken.Claims.(*Payload)
	if !ok || !jwtToken.Valid {
		return nil, fmt.Errorf("invalid token claims or token invalid")
	}

	return payload, nil
}

// Remove ExtractUserIDFromToken as VerifyToken now returns the full payload
/*
func (maker *JWTMaker) ExtractUserIDFromToken(token string) (string, error) {
	payload, err := maker.VerifyToken(token)
	if err != nil {
		return "", err
	}
	// Assuming VerifyToken returned *Payload
	return payload.UserID.String(), nil // Or payload.Subject
}
*/
