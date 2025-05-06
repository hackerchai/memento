package errmsg

import (
	"fmt"
	"net/http"
)

// ErrorInfo defines the structure for application errors.
type ErrorInfo struct {
	Code       string `json:"code"`              // Application-specific error code
	Message    string `json:"message"`           // User-friendly error message
	HTTPStatus int    `json:"-"`                 // Corresponding HTTP status code (omit from JSON response body)
	Details    any    `json:"details,omitempty"` // Optional additional details
}

// New creates a new ErrorInfo instance.
func New(httpStatus int, code string, msg string) *ErrorInfo {
	return &ErrorInfo{
		Code:       code,
		Message:    msg,
		HTTPStatus: httpStatus,
	}
}

// Error implements the error interface.
// This allows ErrorInfo to be used as an error type.
func (e *ErrorInfo) Error() string {
	// Include details in the error string if present, for logging purposes
	if e.Details != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Details)
	}
	return e.Message // Return the user-friendly message
}

// WithDetails adds details to the error message.
// It returns a *new* ErrorInfo instance to avoid modifying the original constant error.
func (e *ErrorInfo) WithDetails(details any) *ErrorInfo {
	// Create a copy
	newErr := *e
	newErr.Details = details
	return &newErr
}

// Predefined ErrorInfos
var (
	// OK is technically not an error, but useful for success cases if needed elsewhere.
	// OK = New(http.StatusOK, "00000", "Success")

	// --- 00xxx: General Errors ---
	ErrServer       = New(http.StatusInternalServerError, "00001", "Internal Server Error")
	ErrInvalidParam = New(http.StatusBadRequest, "00002", "Invalid Parameters")
	ErrBindJSON     = New(http.StatusBadRequest, "00003", "Failed to parse request data")
	ErrValidation   = New(http.StatusBadRequest, "00004", "Validation Failed")
	ErrConfigLoad   = New(http.StatusInternalServerError, "00005", "Failed to load configuration")
	ErrDatabase     = New(http.StatusInternalServerError, "00006", "Database operation failed")

	// --- 01xxx: Auth/User Errors ---
	ErrUnauthorized  = New(http.StatusUnauthorized, "01001", "Unauthorized or authentication failed")
	ErrInvalidCreds  = New(http.StatusUnauthorized, "01002", "Invalid email or password")
	ErrEmailConflict = New(http.StatusConflict, "01003", "Email is already registered")
	ErrUserNotFound  = New(http.StatusNotFound, "01004", "User not found")
	ErrPasswordHash  = New(http.StatusInternalServerError, "01005", "Password processing failed")
	ErrTokenCreation = New(http.StatusInternalServerError, "01006", "Failed to create token")

	// --- 02xxx: Database Specific Errors (Can be mapped from driver errors) ---
	ErrRecordNotFound = New(http.StatusNotFound, "02001", "Record not found")

	// Add more codes as needed for other modules (e.g., web scraping)

	// Auth Errors (01xxx)
	ErrTokenInvalid      = New(http.StatusUnauthorized, "01007", "Token is invalid")
	ErrTokenExpired      = New(http.StatusUnauthorized, "01008", "Token has expired")
	ErrIncorrectPassword = New(http.StatusUnauthorized, "01009", "Incorrect password")
	ErrEmailTaken        = New(http.StatusConflict, "01010", "Email already taken")
	ErrForbidden         = New(http.StatusForbidden, "01011", "Forbidden")

	// Article Errors (02xxx)
	ErrArticleConflict = New(http.StatusConflict, "02001", "Article already exists")
	ErrArticleNotFound = New(http.StatusNotFound, "02001", "Article not found")

	// Category/Tag Errors (03xxx)
	ErrCategoryNotFound = New(http.StatusNotFound, "03001", "Category not found")

	// Add more error types here
)
