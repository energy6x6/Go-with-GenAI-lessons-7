package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"homework/internal/model"
	"homework/internal/repository"
)

type BookHandler struct{ repo repository.BookRepository }

func NewBookHandler(repo repository.BookRepository) *BookHandler { return &BookHandler{repo: repo} }

func Error(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, model.ErrorResponse{Error: model.ErrorBody{Code: code, Message: message}})
}

func (h *BookHandler) repositoryError(c *gin.Context, err error) {
	if errors.Is(err, repository.ErrNotFound) {
		Error(c, 404, "not_found", "Book not found")
		return
	}
	Error(c, 500, "internal_server_error", "Internal server error")
}

func parseID(c *gin.Context) (uint, bool) {
	s := c.Param("id")
	n, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		Error(c, 400, "invalid_id", "ID must be an unsigned 32-bit integer")
		return 0, false
	}
	return uint(n), true
}

func readBook(c *gin.Context) (model.Book, bool) {
	const maxBody = 1 << 20
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBody+1))
	trimmed := bytes.TrimSpace(body)
	if err != nil || len(body) > maxBody || !json.Valid(body) || len(trimmed) == 0 || trimmed[0] != '{' {
		Error(c, 422, "validation_error", "Body must contain one valid JSON object (at most 1 MiB)")
		return model.Book{}, false
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	var req model.CreateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 422, "validation_error", "Body must be a valid book JSON object")
		return model.Book{}, false
	}
	for _, v := range []string{req.Title, req.ISBN, req.Author, req.Category} {
		if strings.TrimSpace(v) == "" {
			Error(c, 422, "validation_error", "Title, ISBN, author and category must be non-blank strings")
			return model.Book{}, false
		}
	}
	return req.Book(), true
}

// Create godoc
// @Summary Create a book
// @Tags books
// @Accept json
// @Produce json
// @Param book body model.CreateBookRequest true "Book fields"
// @Success 201 {object} model.Book
// @Failure 422 {object} model.ErrorResponse
// @Router /books [post]
func (h *BookHandler) Create(c *gin.Context) {
	b, ok := readBook(c)
	if !ok {
		return
	}
	c.JSON(http.StatusCreated, h.repo.Create(b))
}

// Get godoc
// @Summary Get a book by ID
// @Tags books
// @Produce json
// @Param id path int true "Book ID (uint32)"
// @Success 200 {object} model.Book
// @Failure 400,404 {object} model.ErrorResponse
// @Router /books/{id} [get]
func (h *BookHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	b, err := h.repo.FindByID(id)
	if err != nil {
		h.repositoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, b)
}

// Update godoc
// @Summary Replace a book (omitted optional fields become null)
// @Tags books
// @Accept json
// @Produce json
// @Param id path int true "Book ID (uint32)"
// @Param book body model.UpdateBookRequest true "Replacement fields"
// @Success 200 {object} model.Book
// @Failure 400,404,422 {object} model.ErrorResponse
// @Router /books/{id} [put]
func (h *BookHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	b, ok := readBook(c)
	if !ok {
		return
	}
	b, err := h.repo.Update(id, b)
	if err != nil {
		h.repositoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, b)
}

// Delete godoc
// @Summary Delete a book
// @Tags books
// @Param id path int true "Book ID (uint32)"
// @Success 204 "No content"
// @Failure 400,404 {object} model.ErrorResponse
// @Router /books/{id} [delete]
func (h *BookHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.repo.Delete(id); err != nil {
		h.repositoryError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func pagination(c *gin.Context, key string, fallback uint64) (uint64, bool) {
	s, exists := c.GetQuery(key)
	// GetQuery treats empty values as absent; an explicit empty value is invalid.
	if _, present := c.Request.URL.Query()[key]; !present {
		return fallback, true
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if !exists || err != nil || n == 0 {
		Error(c, 400, "invalid_pagination", "Page and limit must be positive integers")
		return 0, false
	}
	return n, true
}

// List godoc
// @Summary List books ordered by ID
// @Tags books
// @Produce json
// @Param category query string false "Exact case-sensitive category"
// @Param page query int false "Page (starts at 1)" default(1) minimum(1)
// @Param limit query int false "Page size (clamped to 100)" default(10) minimum(1)
// @Success 200 {array} model.Book
// @Failure 400 {object} model.ErrorResponse
// @Router /books [get]
func (h *BookHandler) List(c *gin.Context) {
	page, ok := pagination(c, "page", 1)
	if !ok {
		return
	}
	limit, ok := pagination(c, "limit", 10)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, h.repo.List(c.Query("category"), page, limit))
}
