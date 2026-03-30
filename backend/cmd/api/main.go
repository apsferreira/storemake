package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog"

	"github.com/apsferreira/storemaker/internal/handler"
	"github.com/apsferreira/storemaker/internal/middleware"
	"github.com/apsferreira/storemaker/internal/pkg/config"
	"github.com/apsferreira/storemaker/internal/pkg/database"

	"time"
)

func main() {
	cfg := config.Load()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if cfg.Env == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if err := database.Connect(cfg.DatabaseURL); err != nil {
		log.Fatalf("falha ao conectar ao banco: %v", err)
	}
	defer database.Close()

	app := fiber.New(fiber.Config{
		AppName:      "StoreMaker API",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		BodyLimit:    10 * 1024 * 1024, // 10MB
	})

	app.Use(recover.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "limite de requisições excedido, tente novamente em 1 minuto",
			})
		},
	}))

	// Health check (público)
	app.Get("/health", handler.HealthCheck)

	// Rotas protegidas por JWT
	api := app.Group("/api/v1", middleware.JWTAuth(cfg.JWTSecret))
	_ = api // rotas serão adicionadas nos próximos PRs

	log.Printf("StoreMaker API rodando na porta %s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("falha ao iniciar servidor: %v", err)
	}
}
