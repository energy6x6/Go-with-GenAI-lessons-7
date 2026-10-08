package repository

import (
	"errors"
	"sort"
	"sync"
	"time"

	"homework/internal/model"
)

var ErrNotFound = errors.New("book not found")

type BookRepository interface {
	Create(model.Book) model.Book
	FindByID(uint) (model.Book, error)
	List(category string, page, limit uint64) []model.Book
	Update(uint, model.Book) (model.Book, error)
	Delete(uint) error
}

type inMemoryBookRepository struct {
	mu     sync.RWMutex
	books  map[uint]*model.Book
	nextID uint
}

func NewBookRepository() BookRepository {
	return &inMemoryBookRepository{books: make(map[uint]*model.Book), nextID: 1}
}

// clone also copies optional values so callers never hold storage pointers.
func clone(b model.Book) model.Book {
	if b.PublishedYear != nil {
		v := *b.PublishedYear
		b.PublishedYear = &v
	}
	if b.Description != nil {
		v := *b.Description
		b.Description = &v
	}
	return b
}

func (r *inMemoryBookRepository) Create(b model.Book) model.Book {
	r.mu.Lock()
	defer r.mu.Unlock()
	b = clone(b)
	b.ID = r.nextID
	r.nextID++
	b.CreatedAt = time.Now().UTC()
	b.UpdatedAt = b.CreatedAt
	r.books[b.ID] = &b
	return clone(b)
}

func (r *inMemoryBookRepository) FindByID(id uint) (model.Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.books[id]
	if !ok {
		return model.Book{}, ErrNotFound
	}
	return clone(*b), nil
}

func (r *inMemoryBookRepository) List(category string, page, limit uint64) []model.Book {
	r.mu.RLock()
	defer r.mu.RUnlock()
	books := make([]model.Book, 0)
	for _, b := range r.books {
		if category == "" || b.Category == category {
			books = append(books, clone(*b))
		}
	}
	sort.Slice(books, func(i, j int) bool { return books[i].ID < books[j].ID })
	if page == 0 || limit == 0 || len(books) == 0 {
		return []model.Book{}
	}
	if limit > 100 {
		limit = 100
	}
	// Compare before multiplying, so even huge page values cannot overflow.
	if page-1 > uint64(len(books)-1)/limit {
		return []model.Book{}
	}
	start := (page - 1) * limit
	end := start + limit
	if end > uint64(len(books)) {
		end = uint64(len(books))
	}
	return books[start:end]
}

func (r *inMemoryBookRepository) Update(id uint, b model.Book) (model.Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	old, ok := r.books[id]
	if !ok {
		return model.Book{}, ErrNotFound
	}
	b = clone(b)
	b.ID = id
	b.CreatedAt = old.CreatedAt
	b.UpdatedAt = time.Now().UTC()
	if !b.UpdatedAt.After(old.UpdatedAt) {
		b.UpdatedAt = old.UpdatedAt.Add(time.Nanosecond)
	}
	r.books[id] = &b
	return clone(b), nil
}

func (r *inMemoryBookRepository) Delete(id uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.books[id]; !ok {
		return ErrNotFound
	}
	delete(r.books, id)
	return nil
}
