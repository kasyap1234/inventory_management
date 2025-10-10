package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"agromart2/internal/common"
	"agromart2/internal/validation"

	"github.com/labstack/echo/v4"
)

func HTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	if details := validation.ExtractErrors(err); len(details) > 0 {
		respond(c, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", details)
		return
	}

	var he *echo.HTTPError
	if errors.As(err, &he) {
		if details := extractValidationDetails(he.Message); len(details) > 0 {
			respond(c, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", details)
			return
		}

		message := extractMessage(he)
		statusCode := he.Code
		errorCode := classifyErrorCode(statusCode)

		if statusCode >= http.StatusInternalServerError {
			c.Logger().Error(err)
			message = "Internal server error"
		}

		respond(c, statusCode, errorCode, message, nil)
		return
	}

	c.Logger().Error(err)
	respond(c, http.StatusInternalServerError, "SERVER_ERROR", "Internal server error", nil)
}

func respond(c echo.Context, status int, code, message string, details map[string]string) {
	if err := c.JSON(status, common.CreateErrorResponse(code, message, details)); err != nil {
		c.Logger().Error(err)
	}
}

func extractMessage(he *echo.HTTPError) string {
	switch msg := he.Message.(type) {
	case string:
		return msg
	case map[string]interface{}:
		if text, ok := msg["message"].(string); ok {
			return text
		}
	case map[string]string:
		if text, ok := msg["message"]; ok {
			return text
		}
	case []byte:
		return string(msg)
	case fmt.Stringer:
		return msg.String()
	case nil:
		return http.StatusText(he.Code)
	}
	return fmt.Sprint(he.Message)
}

func classifyErrorCode(status int) string {
	switch status {
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusTooManyRequests:
		return "RATE_LIMITED"
	}
	if status >= 500 {
		return "SERVER_ERROR"
	}
	return "CLIENT_ERROR"
}

func extractValidationDetails(message interface{}) map[string]string {
	if err, ok := message.(error); ok {
		return validation.ExtractErrors(err)
	}
	return nil
}
