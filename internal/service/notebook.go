package service

import (
	"backend-ai-supernote/internal/dto"
	"backend-ai-supernote/internal/entity"
	"backend-ai-supernote/internal/repository"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type INotebookService interface {
	GetAllNotebook(ctx context.Context) ([]*dto.GetAllNotebookResponse, error)
	Create(ctx context.Context, req *dto.CreateNotebookRequest) (*dto.CreateNotebookResponse, error)
	Show(ctx context.Context, id uuid.UUID) (*dto.ShowNotebookResponse, error)
	Update(ctx context.Context, req *dto.UpdateNotebookRequest) (*dto.UpdateNotebookResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	MoveNotebook(ctx context.Context, req *dto.MoveNotebookRequest) (*dto.MoveNotebookResponse, error)
}

type notebookService struct {
	notebookRepository repository.INotebookRepository
	db                 *pgxpool.Pool
}

// GetAllNotebook implements [INotebookService].
func (c *notebookService) GetAllNotebook(ctx context.Context) ([]*dto.GetAllNotebookResponse, error) {
	notebooks, err := c.notebookRepository.GetAllNotebook(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.GetAllNotebookResponse, 0)
	for _, notebook := range notebooks {
		res := dto.GetAllNotebookResponse{
			Id:        notebook.Id,
			Name:      notebook.Name,
			ParentID:  notebook.ParentID,
			CreatedAt: notebook.CreatedAt,
			UpdatedAt: notebook.UpdatedAt,
		}

		result = append([]*dto.GetAllNotebookResponse{&res}, result...)
	}

	return result, nil
}

// MoveNotebook implements [INotebookService].
func (c *notebookService) MoveNotebook(ctx context.Context, req *dto.MoveNotebookRequest) (*dto.MoveNotebookResponse, error) {
	_, err := c.notebookRepository.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	if req.ParentId != nil {
		if _, err = c.notebookRepository.GetByID(ctx, *req.ParentId); err != nil {
			return nil, err
		}

	}
	if err = c.notebookRepository.UpdateParentId(ctx, req.Id, req.ParentId); err != nil {
		return nil, err
	}

	return &dto.MoveNotebookResponse{
		Id: req.Id,
	}, nil

}

// Delete implements [INotebookService].
func (c *notebookService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := c.notebookRepository.GetByID(ctx, id)
	if err != nil {
		return err
	}
	tx, err := c.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	notebookRepoTx := c.notebookRepository.UsingTx(ctx, tx)

	if err := notebookRepoTx.Delete(ctx, id); err != nil {
		return err
	}
	if err := notebookRepoTx.NullifyParentById(ctx, id); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
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

func NewNotebookService(notebookRepository repository.INotebookRepository, db *pgxpool.Pool) INotebookService {
	return &notebookService{
		notebookRepository: notebookRepository,
		db:                 db,
	}
}
