package controller

import (
	"backend-ai-supernote/internal/dto"
	"backend-ai-supernote/internal/pkg/serverutils"
	"backend-ai-supernote/internal/service"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type INoteController interface {
	RegisterRoutes(r fiber.Router)
	Create(ctx fiber.Ctx) error
	Show(ctx fiber.Ctx) error
	Update(ctx fiber.Ctx) error
	Delete(ctx fiber.Ctx) error
}

type noteController struct {
	service service.INoteService
}

// Delete implements [INoteController].
func (c *noteController) Delete(ctx fiber.Ctx) error {
	idParams := ctx.Params("id")
	id, _ := uuid.Parse(idParams)

	err := c.service.Delete(ctx.Context(), id)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse[any]("Success delete note", nil))
}

// Update implements [INoteController].
func (c *noteController) Update(ctx fiber.Ctx) error {
	idParams := ctx.Params("id")
	id, _ := uuid.Parse(idParams)

	var req dto.UpdateNoteRequest
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

	return ctx.JSON(serverutils.SuccessResponse("Success update note", res))
}

// Show implements [INoteController].
func (c *noteController) Show(ctx fiber.Ctx) error {
	idParams := ctx.Params("id")
	id, _ := uuid.Parse(idParams)

	res, err := c.service.Show(ctx.Context(), id)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Success show note", res))

}

// Create implements [INoteController].
func (c *noteController) Create(ctx fiber.Ctx) error {
	var req dto.CreateNoteRequest
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

	return ctx.JSON(serverutils.SuccessResponse("Success create note", res))
}

// RegisterRoutes implements [INoteController].
func (c *noteController) RegisterRoutes(r fiber.Router) {
	h := r.Group("/note/v1")
	h.Post("", c.Create)
	h.Get(":id", c.Show)
	h.Put(":id", c.Update)
	h.Delete(":id", c.Delete)

}

func NewNoteController(service service.INoteService) INoteController {
	return &noteController{service: service}
}
