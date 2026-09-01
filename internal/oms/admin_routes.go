package oms

import "github.com/gofiber/fiber/v2"

// RegisterAdminRoutes mounts the admin CRUD group, e.g.
//
//	oms.RegisterAdminRoutes(api.Group("/oms/admin", apikey.Middleware(key)), pgClient)
func RegisterAdminRoutes(router fiber.Router, pg *PgClient) {
	h := NewAdminHandler(pg)

	router.Get("/statuses", h.ListStatuses)
	router.Get("/outages", h.ListOutages)
	router.Get("/outages/:eventId", h.GetOutage)
	router.Patch("/outages/:eventId", h.UpdateOutage)
	router.Delete("/outages/:eventId", h.DeleteOutage)
}
