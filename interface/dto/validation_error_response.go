package dto

import "github.com/ilhamtubagus/urlShortener/utils"

// A ValidationError is an error that is used when the required input fails validation.
// swagger:response validationError
type ValidationErrorResponse struct {
	// in:body
	// required: true
	Body ValidationErrorResponseBody
}

type ValidationErrorResponseBody struct {
	// The message
	Message string `json:"message"`
	// Field errors with its messages
	Errors *[]utils.ValidationError `json:"errors"`
	// Error code
	Code int32 `json:"code"`
}

func NewValidationError(msg string, err *[]utils.ValidationError, code int32) *ValidationErrorResponseBody {
	return &ValidationErrorResponseBody{msg, err, code}
}