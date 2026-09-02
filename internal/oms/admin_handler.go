package oms

import "github.com/gofiber/fiber/v2"

// AdminHandler exposes CRUD over outage events for closing complaints/work
// orders — separate from Handler (the public/operational oms group) since
// it lets an operator change status directly, bypassing prepare/confirm.
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

func (h *AdminHandler) ListOutages(c *fiber.Ctx) error {
	entries, err := h.pg.ListAllOutages(c.Context())
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(entries)
}

func (h *AdminHandler) GetOutage(c *fiber.Ctx) error {
	event, err := h.pg.GetOutageEvent(c.Context(), c.Params("eventId"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(event)
}

func (h *AdminHandler) UpdateOutage(c *fiber.Ctx) error {
	var req UpdateOutageEventRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, &ApiError{Status: fiber.StatusBadRequest, Code: ErrInvalidInput, Message: "ข้อมูลไม่ถูกต้อง"})
	}

	event, err := h.pg.UpdateOutageEvent(c.Context(), c.Params("eventId"), req)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(event)
}

func (h *AdminHandler) DeleteOutage(c *fiber.Ctx) error {
	if err := h.pg.DeleteOutageEvent(c.Context(), c.Params("eventId")); err != nil {
		return writeError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AdminHandler) ListAnonymousReports(c *fiber.Ctx) error {
	reports, err := h.pg.ListAnonymousReports(c.Context())
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(reports)
}

func (h *AdminHandler) GetAnonymousReport(c *fiber.Ctx) error {
	report, err := h.pg.GetAnonymousReport(c.Context(), c.Params("reportId"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(report)
}

func (h *AdminHandler) UpdateAnonymousReport(c *fiber.Ctx) error {
	var req UpdateAnonymousReportRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, &ApiError{Status: fiber.StatusBadRequest, Code: ErrInvalidInput, Message: "ข้อมูลไม่ถูกต้อง"})
	}

	report, err := h.pg.UpdateAnonymousReport(c.Context(), c.Params("reportId"), req)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(report)
}

func (h *AdminHandler) DeleteAnonymousReport(c *fiber.Ctx) error {
	if err := h.pg.DeleteAnonymousReport(c.Context(), c.Params("reportId")); err != nil {
		return writeError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
