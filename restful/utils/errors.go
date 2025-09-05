package utils

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error *APIError `json:"error"`
}

// APIError represents a structured API error
type APIError struct {
	Code    string                    `json:"code"`
	Message string                    `json:"message"`
	Params  map[string]ParameterError `json:"params,omitempty"`
}

// ParameterError represents a parameter-specific error
type ParameterError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// HTTPErrorHandler is a custom error handler for Echo
func HTTPErrorHandler(err error, c echo.Context) {
	var (
		code = http.StatusInternalServerError
		msg  interface{}
	)

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg = he.Message
	} else {
		msg = err.Error()
	}

	// Don't log client errors
	if code >= 500 {
		c.Logger().Error(err)
	}

	// Send response
	if c.Response().Committed {
		return
	}

	if c.Request().Method == http.MethodHead {
		err = c.NoContent(code)
	} else {
		errorResponse := ErrorResponse{
			Error: &APIError{
				Code:    getErrorCode(code),
				Message: getString(msg),
			},
		}
		err = c.JSON(code, errorResponse)
	}

	if err != nil {
		c.Logger().Error(err)
	}
}

// getErrorCode returns appropriate error code based on HTTP status
func getErrorCode(httpCode int) string {
	switch httpCode {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusConflict:
		return "conflict"
	case http.StatusUnprocessableEntity:
		return "validation_error"
	case http.StatusInternalServerError:
		return "internal_error"
	default:
		return "unknown_error"
	}
}

// getString safely converts interface{} to string
func getString(i interface{}) string {
	if s, ok := i.(string); ok {
		return s
	}
	return "Unknown error"
}