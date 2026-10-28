package controller

import (
	"backend-ai-supernote/internal/dto"
	"backend-ai-supernote/internal/pkg/serverutils"
	"backend-ai-supernote/internal/service"

	"github.com/google/uuid"

	"github.com/gofiber/fiber/v3"
)

type INotebookController interface {
	RegisterRoutes(r fiber.Router)
	Create(ctx fiber.Ctx) error
	Show(ctx fiber.Ctx) error
	Update(ctx fiber.Ctx) error
}

type notebookController struct {
	service service.INotebookService
}

// Updated implements [INotebookController].
func (c *notebookController) Update(ctx fiber.Ctx) error {
	idParams := ctx.Params("id")
	id, _ := uuid.Parse(idParams)

	var req dto.UpdateNotebookRequest
	if err := ctx.Bind().Body(&req); err != nil {
		return err
	}

	req.Id = id
	err := serverutils.ValidateRequest(req)
	if err != nil {
		return err
	}

	res, err := c.service.Update(ctx.Context(), &req)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Success update notebook", res))
}

// Show implements [INotebookController].
func (c *notebookController) Show(ctx fiber.Ctx) error {
	idParams := ctx.Params("id")
	id, _ := uuid.Parse(idParams)

	res, err := c.service.Show(ctx.Context(), id)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Success show notebook", res))
}

// Create implements [INotebookController].
func (c *notebookController) Create(ctx fiber.Ctx) error {
	var req dto.CreateNotebookRequest
	if err := ctx.Bind().Body(&req); err != nil {
		return err
	}

	err := serverutils.ValidateRequest(req)
	if err != nil {
		return err
	}

	res, err := c.service.Create(ctx.Context(), &req)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Success create notebook", res))
}

// RegisterRoutes implements [INotebookController].
func (c *notebookController) RegisterRoutes(r fiber.Router) {
	h := r.Group("/notebook/v1")
	h.Post("", c.Create)
	h.Get(":id", c.Show)
	h.Put(":id", c.Update)
}

func NewNotebookController(service service.INotebookService) INotebookController {
	return &notebookController{service: service}
}
