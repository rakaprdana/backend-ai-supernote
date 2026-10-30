package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateNoteRequest struct {
	Title      string    `json:"title" validate:"required"`
	Content    string    `json:"content"`
	NotebookId uuid.UUID `json:"notebook_id" validate:"required"`
}
type UpdateNoteRequest struct {
	Id         uuid.UUID
	Title      string    `json:"title" validate:"required"`
	Content    string    `json:"content"`
	NotebookId uuid.UUID `json:"notebook_id" validate:"required"`
}

type MoveNoteRequest struct {
	Id         uuid.UUID
	NotebookID uuid.UUID `json:"notebook_id" validate:"required"`
}

type CreateNoteResponse struct {
	Id uuid.UUID `json:"id"`
}

type ShowNoteResponse struct {
	Id         uuid.UUID  `json:"id"`
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	NotebookId uuid.UUID  `json:"notebook_id"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at"`
}
type UpdateNoteResponse struct {
	Id uuid.UUID `json:"id"`
}

type MoveNoteResponse struct {
	Id uuid.UUID `json:"id"`
}
