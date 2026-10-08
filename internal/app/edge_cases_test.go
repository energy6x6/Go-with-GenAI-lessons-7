package app_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestBookValidationEdgeCases(t *testing.T) {
	e := newEnv(t)
	valid := `{"title":"Go","isbn":"id","author":"A","category":"programming"}`
	for _, body := range []string{"null", `"text"`, "123", valid + ` {}`, valid + " trailing", strings.Replace(valid, `"Go"`, `" \t "`, 1), strings.Replace(valid, `"Go"`, `null`, 1), strings.Replace(valid, `"Go"`, `42`, 1), strings.TrimSuffix(valid, "}") + `,"published_year":1.5}`, strings.TrimSuffix(valid, "}") + `,"description":123}`} {
		t.Run(body, func(t *testing.T) { e.wantError(e.do(http.MethodPost, e.base(), body), 422, "validation_error") })
	}
	if ids := e.listIDs(""); len(ids) != 0 {
		t.Fatal("invalid request wrote data")
	}
	created := e.create(e.valid())
	id := idOf(t, created)
	e.wantError(e.do(http.MethodPut, e.item(id), `{}`), 422, "validation_error")
	after := e.do(http.MethodGet, e.item(id), "").obj(t)
	beforeJSON, _ := json.Marshal(created)
	afterJSON, _ := json.Marshal(after)
	if string(beforeJSON) != string(afterJSON) {
		t.Fatal("invalid update changed stored resource")
	}
	for _, id := range []string{"4294967296", "1.5", "-1", "abc"} {
		e.wantError(e.do(http.MethodGet, e.item(id), ""), 400, "invalid_id")
	}
	for _, id := range []string{"0", "4294967295"} {
		e.wantError(e.do(http.MethodGet, e.item(id), ""), 404, "not_found")
	}
}

func TestPaginationAndFallbacks(t *testing.T) {
	e := newEnv(t)
	e.create(e.valid())
	for _, query := range []string{"?page=9223372036854775807", "?page=18446744073709551615"} {
		if ids := e.listIDs(query); len(ids) != 0 {
			t.Fatal("huge page must be empty")
		}
	}
	for _, query := range []string{"?page=", "?limit=", "?page=1.5"} {
		e.wantError(e.do(http.MethodGet, e.base()+query, ""), 400, "invalid_pagination")
	}
	for _, path := range []string{"/unknown", e.base()} {
		r := e.do(http.MethodPatch, path, "")
		if r.Code < 400 || r.Code >= 500 || !isJSONContent(r) || r.Header.Get("Access-Control-Allow-Origin") == "" {
			t.Fatalf("fallback contract: %+v", r)
		}
	}
}

func TestSwaggerDocument(t *testing.T) {
	e := newEnv(t)
	r := e.do(http.MethodGet, "/swagger/doc.json", "")
	e.expect(r, 200)
	var doc struct {
		Paths map[string]map[string]any `json:"paths"`
	}
	if err := json.Unmarshal(r.Body, &doc); err != nil {
		t.Fatal(err)
	}
	for path, methods := range map[string][]string{"/books": {"get", "post"}, "/books/{id}": {"get", "put", "delete"}} {
		for _, method := range methods {
			if doc.Paths[path][method] == nil {
				t.Errorf("Swagger missing %s %s", method, path)
			}
		}
	}
	e.expect(e.do(http.MethodGet, "/swagger/index.html", ""), 200)
}
