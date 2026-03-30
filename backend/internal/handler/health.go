package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/apsferreira/storemaker/internal/pkg/database"
)

func HealthCheck(c *fiber.Ctx) error {
	dbStatus := "ok"
	if database.DB != nil {
		if err := database.DB.Ping(); err != nil {
			dbStatus = "unhealthy"
		}
	} else {
		dbStatus = "disconnected"
	}

	status := fiber.StatusOK
	if dbStatus != "ok" {
		status = fiber.StatusServiceUnavailable
	}

	return c.Status(status).JSON(fiber.Map{
		"service":  "storemaker",
		"status":   "running",
		"database": dbStatus,
	})
}
