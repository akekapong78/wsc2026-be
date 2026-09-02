package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"wsc2026-be/internal/apikey"
	"wsc2026-be/internal/oms"
	"wsc2026-be/internal/voc"
)

func main() {
	_ = godotenv.Load() // ok if .env is missing (prod sets real env vars)

	dbstring := os.Getenv("GOOSE_DBSTRING")
	if dbstring == "" {
		log.Fatal("GOOSE_DBSTRING is not set")
	}
	pool, err := pgxpool.New(context.Background(), dbstring)
	if err != nil {
		log.Fatalf("open db pool: %v", err)
	}
	defer pool.Close()

	// Separate read-only GIS meter database (different Supabase project) —
	// optional: coordinate lookups just get skipped if it's not configured.
	var gisClient *oms.GisClient
	if gisDBString := os.Getenv("GIS_DBSTRING"); gisDBString != "" {
		gisPool, err := pgxpool.New(context.Background(), gisDBString)
		if err != nil {
			log.Fatalf("open gis db pool: %v", err)
		}
		defer gisPool.Close()
		gisClient = oms.NewGisClient(gisPool)
	} else {
		log.Println("GIS_DBSTRING not set — outage reports will have no location")
	}

	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Agent connectivity check: confirms both the BE process and the DB
	// connection are alive (unlike /health, which only proves BE is up).
	app.Get("/api/v1/ping", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "error",
				"db":     "unreachable",
			})
		}

		return c.JSON(fiber.Map{"status": "ok", "db": "ok"})
	})

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY is not set")
	}

	pgClient := oms.NewPgClient(pool, gisClient)

	api := app.Group("/api/v1")
	omsGroup := api.Group("/oms", apikey.Middleware(apiKey))
	oms.RegisterRoutes(omsGroup, pgClient)

	// Admin CRUD to close/re-open outage events directly (bypasses
	// prepare/confirm — same API_KEY guard as the operational group for now).
	adminGroup := api.Group("/oms/admin", apikey.Middleware(apiKey))
	oms.RegisterAdminRoutes(adminGroup, pgClient)

	vocPgClient := voc.NewPgClient(pool)
	vocGroup := api.Group("/voc", apikey.Middleware(apiKey))
	voc.RegisterRoutes(vocGroup, vocPgClient)

	vocAdminGroup := api.Group("/voc/admin", apikey.Middleware(apiKey))
	voc.RegisterAdminRoutes(vocAdminGroup, vocPgClient)

	// Dev-only Swagger UI test client — http://localhost:8080/test/swagger.html
	// (reads spec live from /spec/*.yaml, no rebuild needed on spec edits)
	app.Static("/test", "./web")
	app.Static("/spec", "./spec")

	// Keepalive: pinged once a day by a cron job (see scripts/keepalive.sh)
	// so the Supabase project registers activity and doesn't auto-pause.
	app.Post("/api/v1/ops/keepalive", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()

		var pingedAt time.Time
		err := pool.QueryRow(ctx,
			`INSERT INTO keepalive_pings (pinged_at) VALUES (now()) RETURNING pinged_at`,
		).Scan(&pingedAt)
		if err != nil {
			log.Printf("keepalive insert failed: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "keepalive insert failed",
			})
		}

		return c.JSON(fiber.Map{"pingedAt": pingedAt})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(app.Listen(":" + port))
}
