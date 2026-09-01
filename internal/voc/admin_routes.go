package voc

import "github.com/gofiber/fiber/v2"

// RegisterAdminRoutes mounts the admin CRUD group, e.g.
//
//	voc.RegisterAdminRoutes(api.Group("/voc/admin", apikey.Middleware(key)), pgClient)
func RegisterAdminRoutes(router fiber.Router, pg *PgClient) {
	h := NewAdminHandler(pg)

	router.Get("/statuses", h.ListStatuses)
	router.Get("/cases", h.ListCases)
	router.Get("/cases/:caseId", h.GetCase)
	router.Patch("/cases/:caseId", h.UpdateCase)
	router.Delete("/cases/:caseId", h.DeleteCase)
}
