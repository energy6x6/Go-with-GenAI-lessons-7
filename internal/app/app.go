// Package app wires the whole HTTP API together.
//
// The automatic tests talk to your API ONLY through NewRouter(), so you are free to
// organise the rest of the code (model, repository, handler, middleware) as you like.
package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "homework/docs"
	"homework/internal/handler"
	"homework/internal/middleware"
	"homework/internal/repository"
)

// NewRouter must return a fully configured HTTP handler for your resource.
//
// Requirements (see README.md for the full contract):
//   - every call returns a NEW router with its own EMPTY in-memory storage;
//   - it must be safe for concurrent requests;
//   - it can be built with net/http, gin, chi or any other router.
func NewRouter() http.Handler {
	r := gin.New()
	r.RedirectTrailingSlash = false
	r.HandleMethodNotAllowed = true
	r.Use(middleware.RecoveryMiddleware(), middleware.LoggingMiddleware(), middleware.CORSMiddleware())
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	h := handler.NewBookHandler(repository.NewBookRepository())
	api := r.Group("/api/v1")
	api.POST("/books", h.Create)
	api.GET("/books", h.List)
	api.GET("/books/:id", h.Get)
	api.PUT("/books/:id", h.Update)
	api.DELETE("/books/:id", h.Delete)
	r.NoRoute(func(c *gin.Context) { handler.Error(c, 404, "not_found", "Route not found") })
	r.NoMethod(func(c *gin.Context) { handler.Error(c, 405, "method_not_allowed", "Method not allowed") })
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return r
}
