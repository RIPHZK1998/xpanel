// Package response provides standardized API response structures.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents the standard API response structure.
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Success sends a successful response with data.
func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error sends an error response.
func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Success: false,
		Error:   message,
	})
}

// OK sends a 200 OK response with data.
func OK(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusOK, message, data)
}

// Created sends a 201 Created response with data.
func Created(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusCreated, message, data)
}

// BadRequest sends a 400 Bad Request error response.
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// Unauthorized sends a 401 Unauthorized error response.
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

// Forbidden sends a 403 Forbidden error response.
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

// NotFound sends a 404 Not Found error response.
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// Conflict sends a 409 Conflict error response.
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, message)
}

// TooManyRequests sends a 429 Too Many Requests error response.
func TooManyRequests(c *gin.Context, message string) {
	Error(c, http.StatusTooManyRequests, message)
}

// InternalServerError sends a 500 Internal Server Error response.
func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}

// ValidationError sends a 422 Unprocessable Entity error response for validation failures.
func ValidationError(c *gin.Context, message string) {
	Error(c, http.StatusUnprocessableEntity, message)
}
