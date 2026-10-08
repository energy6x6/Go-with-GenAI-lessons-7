package repository

import (
	"errors"
	"math"
	"testing"

	"homework/internal/model"
)

func TestBookLifecycleAndCopies(t *testing.T) {
	r := NewBookRepository()
	year, description := 2026, "original"
	input := model.Book{Title: "Go", ISBN: "isbn", Author: "Author", Category: "programming", PublishedYear: &year, Description: &description}
	created := r.Create(input)
	year = 2000
	*created.Description = "changed outside storage"
	got, err := r.FindByID(created.ID)
	if err != nil || got.ID != 1 || *got.PublishedYear != 2026 || *got.Description != "original" {
		t.Fatalf("create/find or copy isolation: %+v, %v", got, err)
	}
	*got.PublishedYear = 1990
	listed := r.List("programming", 1, 10)
	if len(listed) != 1 || *listed[0].PublishedYear != 2026 {
		t.Fatal("find returned storage pointer")
	}
	*listed[0].Description = "list mutation"
	got, _ = r.FindByID(1)
	if *got.Description != "original" {
		t.Fatal("list returned storage pointer")
	}
	updated, err := r.Update(1, model.Book{Title: "New", ISBN: "new", Author: "New", Category: "fiction"})
	if err != nil || updated.ID != 1 || !updated.CreatedAt.Equal(got.CreatedAt) || !updated.UpdatedAt.After(got.UpdatedAt) || updated.Description != nil || updated.PublishedYear != nil {
		t.Fatalf("update: %+v %v", updated, err)
	}
	if len(r.List("programming", 1, 10)) != 0 || len(r.List("fiction", 1, 10)) != 1 {
		t.Fatal("category filter")
	}
	if err := r.Delete(1); err != nil {
		t.Fatal(err)
	}
	if _, err := r.FindByID(1); !errors.Is(err, ErrNotFound) {
		t.Fatal("missing find", err)
	}
	if _, err := r.Update(1, input); !errors.Is(err, ErrNotFound) {
		t.Fatal("missing update", err)
	}
	if err := r.Delete(1); !errors.Is(err, ErrNotFound) {
		t.Fatal("missing delete", err)
	}
	if r.Create(input).ID != 2 {
		t.Fatal("ID reused")
	}
}

func TestListBoundaries(t *testing.T) {
	r := NewBookRepository()
	for i := 0; i < 105; i++ {
		r.Create(model.Book{Category: "x"})
	}
	if got := r.List("x", 1, 1000); len(got) != 100 || got[0].ID != 1 || got[99].ID != 100 {
		t.Fatal("limit clamp or ordering")
	}
	for _, page := range []uint64{0, 12, math.MaxInt64, math.MaxUint64} {
		if got := r.List("x", page, 10); got == nil || len(got) != 0 {
			t.Fatalf("page %d: %+v", page, got)
		}
	}
	if got := r.List("X", 1, 10); got == nil || len(got) != 0 {
		t.Fatal("case-sensitive filter")
	}
}
