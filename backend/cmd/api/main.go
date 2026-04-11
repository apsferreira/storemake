package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog"

	"github.com/apsferreira/storemaker/internal/handler"
	"github.com/apsferreira/storemaker/internal/middleware"
	"github.com/apsferreira/storemaker/internal/pkg/config"
	"github.com/apsferreira/storemaker/internal/pkg/database"
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

	// BKL-144: WhatsApp Business Cloud API (opcional — sem credenciais a integração fica inativa)
	handler.InitWhatsApp(cfg.WAPhoneNumberID, cfg.WAAccessToken)

	app := fiber.New(fiber.Config{
		AppName:      "StoreMaker API",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		BodyLimit:    10 * 1024 * 1024, // 10MB
	})

	app.Use(recover.New())
	app.Use(middleware.SecurityHeaders()) // BKL-455: X-Frame-Options, HSTS, CSP, etc.

	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Session-ID,X-Webhook-Signature",
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

	// Resolve loja pelo domínio customizado para requests de storefronts externos (BKL-143)
	app.Use(middleware.DomainResolver())

	// Servir uploads estáticos
	app.Static("/uploads", "./uploads")

	// Health check (público)
	app.Get("/health", handler.HealthCheck)

	// Catálogo público (sem auth) — SPEC-006-B: guard de módulo storefront via query param store_id
	// Nota: ModuleGuard no catálogo público usa store_id do domínio resolvido pelo DomainResolver
	app.Get("/api/v1/public/catalog", handler.PublicCatalog)

	// Carrinho (público, session-based)
	app.Post("/api/v1/cart/add", handler.AddToCart)
	app.Put("/api/v1/cart/update/:id", handler.UpdateCartItem)
	app.Delete("/api/v1/cart/remove/:id", handler.RemoveCartItem)
	app.Get("/api/v1/cart", handler.GetCart)

	// Checkout (público)
	app.Post("/api/v1/checkout", handler.Checkout)

	// BKL-422: Cálculo de frete via Melhor Envio (público — chamado no checkout)
	app.Post("/api/v1/shipping/calculate", handler.Calculate)

	// Validação de cupom (público por store)
	app.Post("/api/v1/coupons/validate", handler.ValidateCoupon)

	// Webhook de pagamento
	app.Post("/api/v1/webhooks/payment", handler.PaymentWebhook(cfg.WebhookSecret))

	// BKL-144: WhatsApp — webhook público (Meta faz GET+POST sem JWT)
	app.Get("/api/v1/whatsapp/webhook", handler.WAWebhookVerify(cfg.WAWebhookVerifyToken))
	app.Post("/api/v1/whatsapp/webhook", handler.WAWebhookInbound(cfg.WAAppSecret))

	// Rotas protegidas por JWT
	api := app.Group("/api/v1", middleware.JWTAuth(cfg.JWTSecret))

	// Categorias
	api.Post("/categories", handler.CreateCategory)
	api.Get("/categories", handler.ListCategories)
	api.Get("/categories/:id", handler.GetCategory)
	api.Put("/categories/:id", handler.UpdateCategory)
	api.Delete("/categories/:id", handler.DeleteCategory)

	// Produtos
	api.Post("/products", handler.CreateProduct)
	api.Get("/products", handler.ListProducts)
	api.Get("/products/:id", handler.GetProduct)
	api.Put("/products/:id", handler.UpdateProduct)
	api.Delete("/products/:id", handler.DeleteProduct)
	api.Put("/products/reorder", handler.ReorderProducts)
	api.Post("/products/:id/photos", handler.UploadProductPhotos)
	api.Post("/products/import", handler.ImportProductsCSV)

	// Domínio customizado por loja (BKL-143)
	api.Put("/stores/:id/domain", handler.SetStoreDomain)
	api.Post("/stores/:id/domain/verify", handler.VerifyStoreDomain)
	api.Delete("/stores/:id/domain", handler.RemoveStoreDomain)

	// Estoque legado
	api.Get("/stock/alerts", handler.GetLowStockAlert)

	// SPEC-006-B: Módulos (feature flags) do tenant
	api.Get("/modules", handler.ListModules)
	api.Put("/modules/:module", handler.UpdateModule)

	// BKL-900: Inventário multi-loja centralizado
	api.Post("/inventory", handler.CreateInventoryMaster)
	api.Get("/inventory/alerts", handler.ListInventoryAlerts)
	api.Post("/inventory/alerts/:id/acknowledge", handler.AcknowledgeInventoryAlert)
	api.Get("/inventory/orders", handler.ListSupplierOrders)
	api.Get("/inventory/:id", handler.GetInventoryMaster)
	api.Post("/inventory/:id/adjust", handler.AdjustInventoryQuantity)
	api.Put("/inventory/:id/allocations/:loja_id", handler.UpsertStoreAllocation)
	api.Get("/inventory/:id/movements", handler.ListInventoryMovements)
	api.Post("/inventory/:id/orders", handler.CreateSupplierOrder)

	// BKL-144: WhatsApp — status e envio manual (protegidos por JWT)
	api.Get("/whatsapp/status", handler.WAStatus)
	api.Post("/whatsapp/notify-order", handler.WANotifyOrder)

	// Pedidos (admin)
	api.Get("/orders", handler.ListOrders)
	api.Get("/orders/:id", handler.GetOrder)
	api.Put("/orders/:id/status", handler.UpdateOrderStatus)

	// Cupons (admin CRUD)
	api.Post("/coupons", handler.CreateCoupon)
	api.Get("/coupons", handler.ListCoupons)
	api.Get("/coupons/:id", handler.GetCoupon)
	api.Put("/coupons/:id", handler.UpdateCoupon)
	api.Delete("/coupons/:id", handler.DeleteCoupon)

	// CRM Admin — requer JWT + role admin/owner
	admin := app.Group("/admin", middleware.JWTAuth(cfg.JWTSecret), middleware.AdminOnly())

	// Clientes
	admin.Get("/customers", handler.ListCustomers)
	admin.Get("/customers/:id", handler.GetCustomerDetail)

	// Dashboard
	admin.Get("/dashboard", handler.GetDashboard)

	// Relatórios
	admin.Get("/reports/sales", handler.GetSalesReport)
	admin.Get("/reports/products", handler.GetTopProducts)
	admin.Get("/reports/stock-alerts", handler.GetStockAlerts)
	admin.Get("/reports/export", handler.ExportSalesCSV)

	log.Printf("StoreMaker API rodando na porta %s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("falha ao iniciar servidor: %v", err)
	}
}
