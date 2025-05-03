package response

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/hackerchai/memento/internal/errmsg"
	"github.com/hackerchai/memento/pkg/xlog"
)

// SuccessResponse defines the standard success response structure.
type SuccessResponse struct {
	Code    string      `json:"code" example:"00000"`
	Message string      `json:"message" example:"success"`
	Data    interface{} `json:"data,omitempty"`
}

// Pagination defines the structure for pagination metadata.
type Pagination struct {
	Page       int   `json:"page" example:"1"`
	PageSize   int   `json:"pageSize" example:"10"`
	TotalItems int64 `json:"totalItems" example:"100"`
	TotalPages int   `json:"totalPages" example:"10"`
}

// PaginatedData wraps list items with pagination info.
type PaginatedData struct {
	Items      interface{} `json:"items"`
	Pagination Pagination  `json:"pagination"`
}

// PaginatedResponse defines the standard success response for paginated data.
type PaginatedResponse struct {
	Code    string        `json:"code" example:"00000"`
	Message string        `json:"message" example:"success"`
	Data    PaginatedData `json:"data"`
}

// ErrorResponse defines the standard error response structure.
type ErrorResponse struct {
	Code    string      `json:"code" example:"00001"`
	Message string      `json:"message" example:"Error message"`
	Details interface{} `json:"details,omitempty"`
}

const SuccessCode = "00000"

// Respond sends a standard success JSON response (HTTP 200 OK).
func Respond(c *fiber.Ctx, data interface{}) error {
	return c.Status(http.StatusOK).JSON(SuccessResponse{
		Code:    SuccessCode,
		Message: "success",
		Data:    data,
	})
}

// RespondCreated sends a standard success JSON response (HTTP 201 Created).
func RespondCreated(c *fiber.Ctx, data interface{}) error {
	return c.Status(http.StatusCreated).JSON(SuccessResponse{
		Code:    SuccessCode,
		Message: "success",
		Data:    data,
	})
}

// RespondWithPagination sends a standard paginated success JSON response (HTTP 200 OK).
func RespondWithPagination(c *fiber.Ctx, items interface{}, pagination Pagination) error {
	return c.Status(http.StatusOK).JSON(PaginatedResponse{
		Code:    SuccessCode,
		Message: "success",
		Data: PaginatedData{
			Items:      items,
			Pagination: pagination,
		},
	})
}

// HandleError logs the error and sends a standard error JSON response.
// It first checks for validation errors, then errmsg.ErrorInfo, then treats as internal error.
func HandleError(c *fiber.Ctx, log *xlog.Logger, err error) error {
	var errInfo *errmsg.ErrorInfo
	var validationErrors validator.ValidationErrors
	var statusCode int
	var resp ErrorResponse
	ctx := c.UserContext()

	// 1. Check for validation errors first
	if errors.As(err, &validationErrors) {
		errInfo = errmsg.ErrValidation // Use the predefined validation error info
		statusCode = errInfo.HTTPStatus
		resp = ErrorResponse{
			Code:    errInfo.Code,
			Message: errInfo.Message,
			Details: FormatValidationErrors(validationErrors), // Add formatted details
		}
		// Log validation errors as WARN
		log.WarnX(ctx). // Use context logger
				Str("error_code", errInfo.Code).
				Interface("validation_details", resp.Details).
				Msg("Validation error occurred")

		// 2. Check for our custom application errors (errmsg.ErrorInfo)
	} else if errors.As(err, &errInfo) {
		statusCode = errInfo.HTTPStatus
		resp = ErrorResponse{
			Code:    errInfo.Code,
			Message: errInfo.Message,
		}
		// Log known application errors as WARN
		log.WarnX(ctx). // Use context logger
				Str("error_code", errInfo.Code).
				Str("error_message", errInfo.Message).
				Msg("Application error occurred")

		// 3. Handle as unexpected internal server error
	} else {
		errInfo = errmsg.ErrServer // Use the predefined internal server error
		statusCode = errInfo.HTTPStatus
		resp = ErrorResponse{
			Code:    errInfo.Code,
			Message: errInfo.Message,
		}
	}

	return c.Status(statusCode).JSON(resp)
}
