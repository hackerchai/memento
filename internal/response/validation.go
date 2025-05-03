package response

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// FormatValidationErrors converts validator.ValidationErrors into a map[string]string.
func FormatValidationErrors(err error) map[string]string {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		// Should not happen if called correctly, but handle gracefully
		return map[string]string{"error": "Invalid validation error type"}
	}

	errors := make(map[string]string)
	for _, fieldErr := range validationErrors {
		// Use JSON tag name if available, otherwise use field name
		fieldName := fieldErr.Field()
		// Attempt to get JSON name (this requires struct knowledge or tags, complex here)
		// For simplicity, we'll use the struct field name directly or a cleaned version.
		fieldName = strings.ToLower(fieldName) // Example: make it lowercase

		// Generate a user-friendly message based on the validation tag
		errors[fieldName] = validationMessageForTag(fieldErr.Tag(), fieldErr.Param())
	}
	return errors
}

// validationMessageForTag returns a user-friendly message for a given validation tag.
func validationMessageForTag(tag string, param string) string {
	switch tag {
	case "required":
		return "This field is required."
	case "email":
		return "Invalid email format."
	case "min":
		return fmt.Sprintf("Must be at least %s characters long.", param)
	case "max":
		return fmt.Sprintf("Must be at most %s characters long.", param)
	case "uuid":
		return "Invalid UUID format."
	case "url":
		return "Invalid URL format."
	// Add more cases for other validation tags used in your application
	default:
		return fmt.Sprintf("Validation failed for tag: %s", tag)
	}
}
