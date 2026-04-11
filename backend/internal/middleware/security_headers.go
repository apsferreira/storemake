package middleware

// BKL-455: Security headers — X-Frame-Options, X-Content-Type-Options, HSTS, CSP.

import "github.com/gofiber/fiber/v2"

// SecurityHeaders adiciona headers de segurança a todas as respostas.
func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-XSS-Protection", "0") // browsers modernos ignoram; CSP é suficiente
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		c.Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'self'")
		return c.Next()
	}
}
