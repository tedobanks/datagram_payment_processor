// paystack_supabase_backend/pkg/utils/response_utils.go
package utils

import (
	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standard error JSON response structure.
// It's a common practice to have a consistent way to send error messages to the client.
type ErrorResponse struct {
	Error string `json:"error"` // The error message string
}

// RespondWithError is a helper function to send a JSON error response with a specific HTTP status code.
// c: The Gin context for the current HTTP request.
// code: The HTTP status code to send (e.g., http.StatusBadRequest, http.StatusInternalServerError).
// message: The error message to include in the JSON response.
func RespondWithError(c *gin.Context, code int, message string) {
	// c.AbortWithStatusJSON is useful if you want to stop further processing in the handler chain
	// and immediately send the response.
	// For simplicity, c.JSON is often sufficient if this is the final action in the handler.
	c.JSON(code, ErrorResponse{Error: message})
}

// RespondWithJSON is a helper function to send a JSON success response with a specific HTTP status code.
// c: The Gin context for the current HTTP request.
// code: The HTTP status code to send (e.g., http.StatusOK, http.StatusCreated).
// payload: The data to include as the JSON response body. This can be any struct, map, or other serializable type.
func RespondWithJSON(c *gin.Context, code int, payload interface{}) {
	c.JSON(code, payload)
}

// You can add more utility functions here as your application grows.
// For example:
// - Functions to parse request parameters.
// - Functions for generating unique IDs.
// - Helper functions for string manipulation or data validation if they are generic enough.
