package repository

import (
	"backend-ai-supernote/internal/entity"
	"backend-ai-supernote/internal/pkg/serverutils"
	"backend-ai-supernote/pkg/database"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type INotebookRepository interface {
	UsingTx(ctx context.Context, tx database.DatabaseQueryer) INotebookRepository
	Create(ctx context.Context, notebook *entity.Notebook) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Notebook, error)
	Update(ctx context.Context, notebook *entity.Notebook) error
}

type notebookRepository struct {
	db database.DatabaseQueryer
}

// Update implements [INotebookRepository].
func (n *notebookRepository) Update(ctx context.Context, notebook *entity.Notebook) error {
	_, err := n.db.Exec(
		ctx,
		`UPDATE notebook SET 
		name = $1,
		parent_id = $2,
		updated_at = $3
		WHERE id = $4`,
		notebook.Name,
		notebook.ParentID,
		notebook.UpdatedAt,
		notebook.Id,
	)

	if err != nil {
		return err
	}

	return nil
}

// Create implements [INotebookRepository].
func (n *notebookRepository) Create(ctx context.Context, notebook *entity.Notebook) error {
	_, err := n.db.Exec(
		ctx,
		`INSERT INTO notebook 
		(id, name, parent_id, created_at, updated_at, deleted_at, is_deleted) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		notebook.Id,
		notebook.Name,
		notebook.ParentID,
		notebook.CreatedAt,
		notebook.UpdatedAt,
		notebook.DeletedAt,
		notebook.IsDeleted,
	)

	if err != nil {
		return err
	}

	return nil
}

// GetByID implements [INotebookRepository].
func (n *notebookRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Notebook, error) {
	row := n.db.QueryRow(
		ctx,
		`SELECT id, name, parent_id, created_at, updated_at, deleted_at, is_deleted
		 FROM notebook n WHERE n.is_deleted=false AND n.id = $1`,
		id,
	)

	var notebook entity.Notebook
	err := row.Scan(
		&notebook.Id,
		&notebook.Name,
		&notebook.ParentID,
		&notebook.CreatedAt,
		&notebook.UpdatedAt,
		&notebook.DeletedAt,
		&notebook.IsDeleted,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, serverutils.ErrNotFound
		}
		return nil, err
	}

	return &notebook, nil
}

func (n *notebookRepository) UsingTx(ctx context.Context, tx database.DatabaseQueryer) INotebookRepository {
	return &notebookRepository{
		db: tx,
	}
}

func NewNotebookRepository(db *pgxpool.Pool) INotebookRepository {
	return &notebookRepository{
		db: db,
	}
}
