package voc

import "github.com/gofiber/fiber/v2"

// RegisterRoutes mounts the VOC group under the given router, e.g.
//
//	voc.RegisterRoutes(api.Group("/voc"), voc.NewMockClient())
//
// maps 1:1 to spec/voc.openapi.yaml paths with the group prefix added.
func RegisterRoutes(router fiber.Router, client Client) {
	h := NewHandler(client)

	router.Get("/catalog", h.GetCatalog)
	router.Post("/cases", h.CreateCase)
	router.Post("/cases/lookup", h.LookupCase)
}
