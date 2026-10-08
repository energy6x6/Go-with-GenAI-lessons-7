package middleware

import (
	"bytes"
	"log"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRecoveryHidesDetailsAndLogsFailure(t *testing.T) {
	var logs bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&logs)
	defer log.SetOutput(previous)
	t.Setenv("ALLOWED_ORIGIN", "https://example.org")
	r := gin.New()
	r.Use(RecoveryMiddleware(), LoggingMiddleware(), CORSMiddleware())
	r.GET("/panic", func(c *gin.Context) { panic("private diagnostic") })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/panic", nil))
	if w.Code != 500 || !strings.Contains(w.Body.String(), `"code":"internal_server_error"`) || strings.Contains(w.Body.String(), "private diagnostic") {
		t.Fatalf("recovery: %d %s", w.Code, w.Body.String())
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.org" {
		t.Fatal("missing configured CORS on error")
	}
	if !strings.Contains(logs.String(), "private diagnostic") || !strings.Contains(logs.String(), "status=500") {
		t.Fatal("missing diagnostic/status log")
	}
}
