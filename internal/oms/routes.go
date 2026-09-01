package oms

import "github.com/gofiber/fiber/v2"

// RegisterRoutes mounts the OMS group under the given router, e.g.
//
//	oms.RegisterRoutes(api.Group("/oms"), oms.NewMockClient())
//
// maps 1:1 to spec/oms.openapi.yaml paths with the group prefix added.
func RegisterRoutes(router fiber.Router, client Client) {
	h := NewHandler(client)

	router.Get("/outages/by-ca/:caNumber", h.GetOutageByCA)
	router.Post("/outages", h.CreateOutage)
	router.Post("/outages/anonymous", h.CreateAnonymousOutage)
}
