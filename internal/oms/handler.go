package oms

import (
	"errors"
	"regexp"

	"github.com/gofiber/fiber/v2"
)

var caNumberRe = regexp.MustCompile(`^[0-9]{12}$`)

type Handler struct {
	client Client
}

func NewHandler(client Client) *Handler {
	return &Handler{client: client}
}

func (h *Handler) GetOutageByCA(c *fiber.Ctx) error {
	caNumber := c.Params("caNumber")
	if !caNumberRe.MatchString(caNumber) {
		return writeError(c, &ApiError{Status: fiber.StatusBadRequest, Code: ErrInvalidCA, Message: "หมายเลขผู้ใช้ไฟต้องเป็นตัวเลข 12 หลัก"})
	}

	resp, err := h.client.GetOutageByCA(c.Context(), caNumber)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(resp)
}

func (h *Handler) CreateOutage(c *fiber.Ctx) error {
	var req CreateOutageRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, &ApiError{Status: fiber.StatusBadRequest, Code: ErrInvalidInput, Message: "ข้อมูลแจ้งเหตุไม่ครบถ้วนหรือไม่ถูกต้อง"})
	}

	if !caNumberRe.MatchString(req.CaNumber) {
		return writeError(c, &ApiError{Status: fiber.StatusBadRequest, Code: ErrInvalidCA, Message: "หมายเลขผู้ใช้ไฟต้องเป็นตัวเลข 12 หลัก"})
	}
	if len(req.Description) < 1 || len(req.Description) > 2000 {
		return writeError(c, &ApiError{Status: fiber.StatusBadRequest, Code: ErrInvalidInput, Message: "ข้อมูลแจ้งเหตุไม่ครบถ้วนหรือไม่ถูกต้อง"})
	}

	resp, err := h.client.CreateOutage(c.Context(), req)
	if err != nil {
		return writeError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *Handler) CreateAnonymousOutage(c *fiber.Ctx) error {
	var req CreateAnonymousOutageRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, &ApiError{Status: fiber.StatusBadRequest, Code: ErrInvalidInput, Message: "กรุณาระบุรายละเอียดเหตุ สถานที่ และเบอร์โทรติดต่อกลับ"})
	}

	if len(req.Description) < 1 || len(req.Location) < 1 || len(req.ContactPhone) < 8 {
		return writeError(c, &ApiError{Status: fiber.StatusBadRequest, Code: ErrInvalidInput, Message: "กรุณาระบุรายละเอียดเหตุ สถานที่ และเบอร์โทรติดต่อกลับ"})
	}

	resp, err := h.client.CreateAnonymousOutage(c.Context(), req)
	if err != nil {
		return writeError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

// writeError translates a Client/validation error into the spec's error
// envelope. ACTIVE_EVENT_EXISTS gets the conflict shape (with existingEventId).
func writeError(c *fiber.Ctx, err error) error {
	var apiErr *ApiError
	if !errors.As(err, &apiErr) {
		apiErr = &ApiError{Status: fiber.StatusInternalServerError, Code: ErrInternal, Message: "เกิดข้อผิดพลาดภายในระบบ OMS"}
	}

	if apiErr.Code == ErrActiveEventExist {
		return c.Status(apiErr.Status).JSON(fiber.Map{
			"error": fiber.Map{
				"code":            apiErr.Code,
				"message":         apiErr.Message,
				"existingEventId": apiErr.ExistingEventID,
			},
		})
	}

	return c.Status(apiErr.Status).JSON(fiber.Map{
		"error": fiber.Map{
			"code":    apiErr.Code,
			"message": apiErr.Message,
		},
	})
}
