package handlers

import (
	"net/http"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// ErrorHandler provides standardized error handling for HTTP handlers
type ErrorHandler struct {
	logger *log.Logger
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *log.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// sanitizeErrorDetails removes sensitive information from error messages
func (eh *ErrorHandler) sanitizeErrorDetails(err error) string {
	if err == nil {
		return ""
	}
	
	errorMsg := err.Error()
	
	// Remove potentially sensitive information
	sensitivePatterns := []string{
		// File paths
		"/home/",
		"/tmp/",
		"/var/",
		"/usr/",
		"/etc/",
		// Database connection strings
		"password=",
		"pwd=",
		"passwd=",
		// API keys and tokens
		"token=",
		"key=",
		"secret=",
		// IP addresses and internal hostnames
		"127.0.0.1",
		"localhost",
		"internal.",
		".local",
	}
	
	for _, pattern := range sensitivePatterns {
		if strings.Contains(strings.ToLower(errorMsg), strings.ToLower(pattern)) {
			return "Internal system error occurred"
		}
	}
	
	// Truncate very long error messages
	if len(errorMsg) > 200 {
		return "Internal system error occurred"
	}
	
	return errorMsg
}

// BadRequest handles 400 Bad Request errors
func (eh *ErrorHandler) BadRequest(c *gin.Context, message string, err error) {
	if err != nil {
		eh.logError(c, "Bad Request", err)
	}
	
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error:   message,
		Code:    "BAD_REQUEST",
		Details: eh.sanitizeErrorDetails(err),
	})
}

// NotFound handles 404 Not Found errors
func (eh *ErrorHandler) NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, ErrorResponse{
		Error: message,
		Code:  "NOT_FOUND",
	})
}

// InternalError handles 500 Internal Server errors
func (eh *ErrorHandler) InternalError(c *gin.Context, message string, err error) {
	eh.logError(c, "Internal Server Error", err)
	
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error:   message,
		Code:    "INTERNAL_ERROR",
		Details: eh.sanitizeErrorDetails(err),
	})
}

// ValidationError handles validation errors
func (eh *ErrorHandler) ValidationError(c *gin.Context, message string, err error) {
	if err != nil {
		eh.logError(c, "Validation Error", err)
	}
	
	c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
		Error:   message,
		Code:    "VALIDATION_ERROR",
		Details: eh.sanitizeErrorDetails(err),
	})
}

// ConflictError handles 409 Conflict errors
func (eh *ErrorHandler) ConflictError(c *gin.Context, message string, err error) {
	if err != nil {
		eh.logError(c, "Conflict Error", err)
	}
	
	c.JSON(http.StatusConflict, ErrorResponse{
		Error:   message,
		Code:    "CONFLICT",
		Details: eh.sanitizeErrorDetails(err),
	})
}

// logError logs the error with context information
func (eh *ErrorHandler) logError(c *gin.Context, errorType string, err error) {
	if eh.logger != nil && err != nil {
		eh.logger.Printf("[%s] %s %s - %v", 
			errorType, 
			c.Request.Method, 
			c.Request.URL.Path, 
			err,
		)
	}
}

// SuccessResponse represents a standardized success response
type SuccessResponse struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Success handles successful responses
func (eh *ErrorHandler) Success(c *gin.Context, data interface{}, message ...string) {
	response := SuccessResponse{
		Data: data,
	}
	
	if len(message) > 0 {
		response.Message = message[0]
	}
	
	c.JSON(http.StatusOK, response)
}

// Created handles 201 Created responses
func (eh *ErrorHandler) Created(c *gin.Context, data interface{}, message ...string) {
	response := SuccessResponse{
		Data: data,
	}
	
	if len(message) > 0 {
		response.Message = message[0]
	}
	
	c.JSON(http.StatusCreated, response)
}