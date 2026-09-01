package voc

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	client Client
}

func NewHandler(client Client) *Handler {
	return &Handler{client: client}
}

func (h *Handler) GetCatalog(c *fiber.Ctx) error {
	catalog, err := h.client.GetCatalog(c.Context())
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(catalog)
}

func (h *Handler) CreateCase(c *fiber.Ctx) error {
	idempotencyKey := c.Get("Idempotency-Key")
	if idempotencyKey == "" || len(idempotencyKey) > 128 {
		return writeError(c, &ApiError{Status: fiber.StatusBadRequest, Code: ErrInvalidInput, Message: "ต้องระบุ Idempotency-Key"})
	}

	var req CreateVocCaseRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, &ApiError{Status: fiber.StatusBadRequest, Code: ErrInvalidInput, Message: "ข้อมูลไม่ถูกต้อง"})
	}

	if _, apiErr := validateCreateCase(req); apiErr != nil {
		return writeError(c, apiErr)
	}

	resp, err := h.client.CreateCase(c.Context(), idempotencyKey, req)
	if err != nil {
		return writeError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *Handler) LookupCase(c *fiber.Ctx) error {
	var req CaseLookupRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, &ApiError{Status: fiber.StatusBadRequest, Code: ErrInvalidInput, Message: "ข้อมูลไม่ถูกต้อง"})
	}
	if !vocNumberRe.MatchString(req.VocNumber) || !keyCodeRe.MatchString(req.KeyCode) {
		return writeError(c, &ApiError{Status: fiber.StatusBadRequest, Code: ErrInvalidInput, Message: "ข้อมูลไม่ถูกต้อง"})
	}

	resp, err := h.client.LookupCase(c.Context(), req)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(resp)
}

// writeError translates a Client/validation error into the spec's error
// envelope.
func writeError(c *fiber.Ctx, err error) error {
	var apiErr *ApiError
	if !errors.As(err, &apiErr) {
		apiErr = &ApiError{Status: fiber.StatusInternalServerError, Code: ErrInternal, Message: "เกิดข้อผิดพลาดภายในระบบ VOC"}
	}
	errBody := fiber.Map{
		"code":    apiErr.Code,
		"message": apiErr.Message,
	}
	if len(apiErr.Fields) > 0 {
		errBody["fields"] = apiErr.Fields
	}
	return c.Status(apiErr.Status).JSON(fiber.Map{"simulation": true, "error": errBody})
}
