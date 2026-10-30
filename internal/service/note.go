package service

import (
	"backend-ai-supernote/internal/dto"
	"backend-ai-supernote/internal/entity"
	"backend-ai-supernote/internal/repository"
	"context"
	"time"

	"github.com/google/uuid"
)

type INoteService interface {
	Create(ctx context.Context, req *dto.CreateNoteRequest) (*dto.CreateNoteResponse, error)
	Show(ctx context.Context, id uuid.UUID) (*dto.ShowNoteResponse, error)
	Update(ctx context.Context, req *dto.UpdateNoteRequest) (*dto.UpdateNoteResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	MoveNote(ctx context.Context, req *dto.MoveNoteRequest) (*dto.MoveNoteResponse, error)
}

type noteService struct {
	noteRepository repository.INoteRepository
}

// MoveNotebook implements [INoteService].
func (n *noteService) MoveNote(ctx context.Context, req *dto.MoveNoteRequest) (*dto.MoveNoteResponse, error) {
	note, err := n.noteRepository.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	note.NotebookId = req.NotebookID
	note.UpdatedAt = &now

	if err = n.noteRepository.Update(ctx, note); err != nil {
		return nil, err
	}

	return &dto.MoveNoteResponse{
		Id: note.Id,
	}, nil
}

// Delete implements [INoteService].
func (n *noteService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := n.noteRepository.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := n.noteRepository.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}

// Update implements [INoteService].
func (n *noteService) Update(ctx context.Context, req *dto.UpdateNoteRequest) (*dto.UpdateNoteResponse, error) {
	note, err := n.noteRepository.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	now := time.Now()

	note.Title = req.Title
	note.Content = req.Content
	note.NotebookId = req.NotebookId
	note.UpdatedAt = &now

	if err := n.noteRepository.Update(ctx, note); err != nil {
		return nil, err
	}
	res := dto.UpdateNoteResponse{
		Id: note.Id,
	}

	return &res, nil
}

// Show implements [INoteService].
func (n *noteService) Show(ctx context.Context, id uuid.UUID) (*dto.ShowNoteResponse, error) {
	note, err := n.noteRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	res := dto.ShowNoteResponse{
		Id:         note.Id,
		Title:      note.Title,
		Content:    note.Content,
		NotebookId: note.NotebookId,
		CreatedAt:  note.CreatedAt,
		UpdatedAt:  note.UpdatedAt,
	}

	return &res, nil
}

// Create implements [INoteService].
func (n *noteService) Create(ctx context.Context, req *dto.CreateNoteRequest) (*dto.CreateNoteResponse, error) {
	note := entity.Note{
		Id:         uuid.New(),
		Title:      req.Title,
		Content:    req.Content,
		NotebookId: req.NotebookId,
		CreatedAt:  time.Now(),
	}
	err := n.noteRepository.Create(ctx, &note)
	if err != nil {
		return nil, err
	}

	return &dto.CreateNoteResponse{
		Id: note.Id,
	}, nil
}

func NewNoteService(noteRepository repository.INoteRepository) INoteService {
	return &noteService{
		noteRepository: noteRepository,
	}
}
