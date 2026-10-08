package model

import "time"

// Book is a catalog entry. Author and Category are temporary plain-text fields;
// they will become relational in theme 8.
type Book struct {
	ID            uint      `json:"id"`
	Title         string    `json:"title"`
	ISBN          string    `json:"isbn"`
	Author        string    `json:"author"`
	Category      string    `json:"category"`
	PublishedYear *int      `json:"published_year"`
	Description   *string   `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateBookRequest struct {
	Title         string  `json:"title" binding:"required"`
	ISBN          string  `json:"isbn" binding:"required"`
	Author        string  `json:"author" binding:"required"`
	Category      string  `json:"category" binding:"required"`
	PublishedYear *int    `json:"published_year"`
	Description   *string `json:"description"`
}

// UpdateBookRequest replaces all editable fields, including omitted optionals.
type UpdateBookRequest = CreateBookRequest

func (r CreateBookRequest) Book() Book {
	return Book{Title: r.Title, ISBN: r.ISBN, Author: r.Author, Category: r.Category,
		PublishedYear: r.PublishedYear, Description: r.Description}
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}
