package service

import (
	"backend-ai-supernote/internal/dto"
	"backend-ai-supernote/internal/entity"
	"backend-ai-supernote/internal/repository"
	"backend-ai-supernote/pkg/embedding"
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/google/uuid"
)

type INoteService interface {
	Create(ctx context.Context, req *dto.CreateNoteRequest) (*dto.CreateNoteResponse, error)
	Show(ctx context.Context, id uuid.UUID) (*dto.ShowNoteResponse, error)
	Update(ctx context.Context, req *dto.UpdateNoteRequest) (*dto.UpdateNoteResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	MoveNote(ctx context.Context, req *dto.MoveNoteRequest) (*dto.MoveNoteResponse, error)
	SemanticSearch(ctx context.Context, search string) ([]*dto.SemanticSearchResponse, error)
}

type noteService struct {
	noteEmbeddingRepo repository.INoteEmbeddingRepository
	noteRepository    repository.INoteRepository
	publisherService  IPublisherService
}

// SemanticSearch implements [INoteService].
func (n *noteService) SemanticSearch(ctx context.Context, search string) ([]*dto.SemanticSearchResponse, error) {
	embeddingRes, err := embedding.GetGeminiEmbedding(
		os.Getenv("GEMINI_API_KEY"),
		search,
		"RETRIEVAL_QUERY",
	)
	if err != nil {
		return nil, err
	}

	noteEmbeddings, err := n.noteEmbeddingRepo.SemanticSearch(ctx, embeddingRes.Embedding.Values)
	if err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, 0)
	for _, noteEmbedding := range noteEmbeddings {
		ids = append(ids, noteEmbedding.NoteId)
	}

	notes, err := n.noteRepository.GetByIds(ctx, ids)
	if err != nil {
		return nil, err
	}

	res := make([]*dto.SemanticSearchResponse, 0)
	for _, noteEmbedding := range noteEmbeddings {
		for _, note := range notes {
			if noteEmbedding.NoteId == note.Id {
				res = append(res, &dto.SemanticSearchResponse{
					Id:         note.Id,
					Title:      note.Title,
					Content:    note.Content,
					NotebookId: note.NotebookId,
					CreatedAt:  note.CreatedAt,
					UpdatedAt:  note.UpdatedAt,
				})
			}
		}
	}
	return res, nil
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

	payload := dto.PublisherEmbedNoteMessage{
		NoteId: note.Id,
	}
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	err = n.publisherService.Publish(ctx, payloadJson)
	if err != nil {
		return nil, err
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

	message := dto.PublisherEmbedNoteMessage{
		NoteId: note.Id,
	}
	msgPayloadJson, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	if err = n.publisherService.Publish(ctx, msgPayloadJson); err != nil {
		return nil, err
	}

	return &dto.CreateNoteResponse{
		Id: note.Id,
	}, nil
}

func NewNoteService(
	noteRepository repository.INoteRepository,
	noteEmbeddingRepo repository.INoteEmbeddingRepository,
	publihserService IPublisherService,
) INoteService {
	return &noteService{
		noteRepository:    noteRepository,
		noteEmbeddingRepo: noteEmbeddingRepo,
		publisherService:  publihserService,
	}
}
