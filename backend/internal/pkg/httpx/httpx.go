// Package httpx provides the standard JSON envelope, error mapping, and
// pagination helpers shared by all handlers.
package httpx

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AppError is the canonical API error. Map it to a response with WriteError.
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
	Status  int    `json:"-"`
}

func (e *AppError) Error() string { return e.Code + ": " + e.Message }

// NewError builds an AppError with the given HTTP status.
func NewError(status int, code, message string) *AppError {
	return &AppError{Code: code, Message: message, Status: status}
}

// WithDetails attaches structured details for validation errors etc.
func (e *AppError) WithDetails(d any) *AppError { e.Details = d; return e }

// Common errors — keep codes stable; clients depend on them.
var (
	ErrUnauthenticated = NewError(http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required")
	ErrForbidden       = NewError(http.StatusForbidden, "FORBIDDEN", "you do not have permission to perform this action")
	ErrNotFound        = NewError(http.StatusNotFound, "NOT_FOUND", "resource not found")
	ErrConflict        = NewError(http.StatusConflict, "CONFLICT", "resource state conflicts with the request")
	ErrValidation      = NewError(http.StatusBadRequest, "VALIDATION_ERROR", "request validation failed")
	ErrRateLimited     = NewError(http.StatusTooManyRequests, "RATE_LIMITED", "too many requests, try again later")
	ErrInternal        = NewError(http.StatusInternalServerError, "INTERNAL_ERROR", "something went wrong")
)

// WriteJSON writes the standard success envelope {"data": ...}.
func WriteJSON(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data})
}

// WriteError maps any error to the error envelope. Unknown errors become 500
// (their message is never leaked to the client).
func WriteError(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.Status, gin.H{"error": appErr})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternal})
}

// Page is the pagination metadata returned with list endpoints.
type Page struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

// Pagination carries parsed page/limit params.
type Pagination struct {
	Page  int
	Limit int
}

// ParsePagination reads ?page= and ?limit= with sane defaults and caps.
func ParsePagination(c *gin.Context) Pagination {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return Pagination{Page: page, Limit: limit}
}

// Offset returns the SQL offset for the parsed pagination.
func (p Pagination) Offset() int { return (p.Page - 1) * p.Limit }

// WriteList writes a paginated list envelope.
func WriteList(c *gin.Context, items any, p Pagination, total int64) {
	WriteJSON(c, http.StatusOK, gin.H{
		"items": items,
		"meta":  Page{Page: p.Page, Limit: p.Limit, Total: total},
	})
}
