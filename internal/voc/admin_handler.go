package voc

import "github.com/gofiber/fiber/v2"

// AdminHandler exposes CRUD over VOC cases for status changes/closure —
// separate from Handler (the public/operational voc group) since it lets
// an operator move status directly, bypassing the public tracking flow.
// There is no public "list all cases" endpoint by design (see spec), so
// listing only exists here.
type AdminHandler struct {
	pg *PgClient
}

func NewAdminHandler(pg *PgClient) *AdminHandler {
	return &AdminHandler{pg: pg}
}

func (h *AdminHandler) ListStatuses(c *fiber.Ctx) error {
	statuses, err := h.pg.ListStatuses(c.Context())
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(statuses)
}

func (h *AdminHandler) ListCases(c *fiber.Ctx) error {
	cases, err := h.pg.ListCases(c.Context())
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(cases)
}

func (h *AdminHandler) GetCase(c *fiber.Ctx) error {
	kase, err := h.pg.GetCase(c.Context(), c.Params("caseId"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(kase)
}

func (h *AdminHandler) UpdateCase(c *fiber.Ctx) error {
	var req UpdateCaseRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, &ApiError{Status: fiber.StatusBadRequest, Code: ErrInvalidInput, Message: "ข้อมูลไม่ถูกต้อง"})
	}

	kase, err := h.pg.UpdateCase(c.Context(), c.Params("caseId"), req)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(kase)
}

func (h *AdminHandler) DeleteCase(c *fiber.Ctx) error {
	if err := h.pg.DeleteCase(c.Context(), c.Params("caseId")); err != nil {
		return writeError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
