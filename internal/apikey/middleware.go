// Package apikey is a minimal shared-secret check via the X-API-Key header.
// Only guards domain sub-groups (e.g. /api/v1/oms/...) — the core agent
// contract (/api/v1/chat, actions, traces, reset) stays open because the
// existing static FE (wsc2026/web/) calls it directly from the browser
// with no secret storage.
package apikey

import (
	"crypto/subtle"

	"github.com/gofiber/fiber/v2"
)

func Middleware(key string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		got := c.Get("X-API-Key")
		if got == "" || subtle.ConstantTimeCompare([]byte(got), []byte(key)) != 1 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{"code": "UNAUTHORIZED", "message": "missing or invalid X-API-Key"},
			})
		}
		return c.Next()
	}
}
