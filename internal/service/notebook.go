package service

import (
	"backend-ai-supernote/internal/dto"
	"backend-ai-supernote/internal/entity"
	"backend-ai-supernote/internal/repository"
	"context"
	"time"

	"github.com/google/uuid"
)

type INotebookService interface {
	Create(ctx context.Context, req *dto.CreateNotebookRequest) (*dto.CreateNotebookResponse, error)
	Show(ctx context.Context, id uuid.UUID) (*dto.ShowNotebookResponse, error)
	Update(ctx context.Context, req *dto.UpdateNotebookRequest) (*dto.UpdateNotebookResponse, error)
}

type notebookService struct {
	notebookRepository repository.INotebookRepository
}

// Update implements [INotebookService].
func (c *notebookService) Update(ctx context.Context, req *dto.UpdateNotebookRequest) (*dto.UpdateNotebookResponse, error) {
	notebook, err := c.notebookRepository.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	notebook.Name = req.Name
	notebook.UpdatedAt = &now
	if err := c.notebookRepository.Update(ctx, notebook); err != nil {
		return nil, err
	}
	res := dto.UpdateNotebookResponse{
		Id: notebook.Id,
	}

	return &res, nil
}

// Show implements [INotebookService].
func (c *notebookService) Show(ctx context.Context, id uuid.UUID) (*dto.ShowNotebookResponse, error) {
	notebook, err := c.notebookRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	res := dto.ShowNotebookResponse{
		Id:        notebook.Id,
		Name:      notebook.Name,
		ParentID:  notebook.ParentID,
		CreatedAt: notebook.CreatedAt,
		UpdatedAt: notebook.UpdatedAt,
	}

	return &res, nil
}

// Create implements [INotebookService].
func (c *notebookService) Create(ctx context.Context, req *dto.CreateNotebookRequest) (*dto.CreateNotebookResponse, error) {
	notebook := entity.Notebook{
		Id:        uuid.New(),
		Name:      req.Name,
		ParentID:  req.ParentID,
		CreatedAt: time.Now(),
	}
	err := c.notebookRepository.Create(ctx, &notebook)
	if err != nil {
		return nil, err
	}

	return &dto.CreateNotebookResponse{
		Id: notebook.Id,
	}, nil
}

func NewNotebookService(notebookRepository repository.INotebookRepository) INotebookService {
	return &notebookService{
		notebookRepository: notebookRepository,
	}
}
