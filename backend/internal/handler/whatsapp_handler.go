package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/apsferreira/storemaker/internal/pkg/whatsapp"
)

// whatsappClient é o cliente singleton injetado via InitWhatsApp.
// É nil quando as credenciais não estiverem configuradas.
var whatsappClient *whatsapp.Client

// InitWhatsApp inicializa o cliente WhatsApp com as credenciais do ambiente.
// Deve ser chamado em main.go antes de registrar as rotas.
func InitWhatsApp(phoneNumberID, accessToken string) {
	whatsappClient = whatsapp.New(phoneNumberID, accessToken)
	if whatsappClient.IsConfigured() {
		log.Info().Msg("whatsapp: integração ativa")
	} else {
		log.Warn().Msg("whatsapp: credenciais ausentes — integração desabilitada")
	}
}

// WAStatus godoc
// GET /api/v1/whatsapp/status
// Retorna se a integração WhatsApp está configurada.
func WAStatus(c *fiber.Ctx) error {
	configured := whatsappClient != nil && whatsappClient.IsConfigured()
	return c.JSON(fiber.Map{
		"configured": configured,
	})
}

// WAWebhookVerify godoc
// GET /api/v1/whatsapp/webhook
// Verificação do webhook exigida pelo Meta para registrar o endpoint.
// Parâmetros de query: hub.mode, hub.verify_token, hub.challenge.
func WAWebhookVerify(verifyToken string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		mode := c.Query("hub.mode")
		token := c.Query("hub.verify_token")
		challenge := c.Query("hub.challenge")

		if mode == "subscribe" && token == verifyToken {
			log.Info().Msg("whatsapp: webhook verificado com sucesso")
			return c.SendString(challenge)
		}

		log.Warn().
			Str("mode", mode).
			Msg("whatsapp: falha na verificação do webhook — token inválido ou mode incorreto")
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "verificação inválida"})
	}
}

// validateMetaSignature verifica o header X-Hub-Signature-256 enviado pelo Meta.
// VUL-003: sem validação qualquer requisição pode injetar eventos falsos no webhook.
func validateMetaSignature(body []byte, signature string, appSecret string) bool {
	if appSecret == "" || signature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expected))
}

// WAWebhookInbound godoc
// POST /api/v1/whatsapp/webhook
// Recebe mensagens inbound e eventos de status da Cloud API do Meta.
// Valida X-Hub-Signature-256 antes de processar o payload (VUL-003).
func WAWebhookInbound(appSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		body := c.Body()
		sig := c.Get("X-Hub-Signature-256")

		if !validateMetaSignature(body, sig, appSecret) {
			log.Warn().
				Str("ip", c.IP()).
				Msg("whatsapp: webhook rejeitado — assinatura HMAC inválida ou ausente")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "assinatura inválida"})
		}

		log.Info().
			RawJSON("body", body).
			Msg("whatsapp: evento inbound recebido")
		// A Cloud API exige HTTP 200 imediato; processamento assíncrono no futuro.
		return c.JSON(fiber.Map{"received": true})
	}
}

// WANotifyOrder godoc
// POST /api/v1/whatsapp/notify-order
// Envia manualmente uma notificação de pedido via WhatsApp.
// Body: { "phone": "+5511999999999", "customer_name": "João", "order_id": "abc-123", "total_cents": 5000 }
func WANotifyOrder(c *fiber.Ctx) error {
	if whatsappClient == nil || !whatsappClient.IsConfigured() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "integração WhatsApp não configurada",
		})
	}

	var req struct {
		Phone        string `json:"phone"`
		CustomerName string `json:"customer_name"`
		OrderID      string `json:"order_id"`
		TotalCents   int64  `json:"total_cents"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}

	if req.Phone == "" || req.OrderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "phone e order_id são obrigatórios"})
	}

	if err := whatsappClient.SendOrderConfirmation(c.Context(), req.Phone, req.CustomerName, req.OrderID, req.TotalCents); err != nil {
		log.Error().Err(err).Str("order_id", req.OrderID).Msg("whatsapp: falha ao enviar notificação manual")
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "falha ao enviar mensagem WhatsApp"})
	}

	log.Info().Str("order_id", req.OrderID).Str("phone", req.Phone).Msg("whatsapp: notificação enviada com sucesso")
	return c.JSON(fiber.Map{"sent": true})
}

// NotifyOrderConfirmedViaWA envia notificação de confirmação de pedido de forma non-fatal.
// Deve ser chamado em goroutine quando um pedido for confirmado.
// Retorna imediatamente; erros são apenas logados.
func NotifyOrderConfirmedViaWA(phone, customerName, orderID string, totalCents int64) {
	if whatsappClient == nil || !whatsappClient.IsConfigured() {
		return
	}
	if phone == "" {
		return
	}

	go func() {
		ctx := context.Background()
		if err := whatsappClient.SendOrderConfirmation(ctx, phone, customerName, orderID, totalCents); err != nil {
			log.Error().
				Err(err).
				Str("order_id", orderID).
				Str("phone", phone).
				Msg("whatsapp: falha ao notificar confirmação de pedido")
		} else {
			log.Info().
				Str("order_id", orderID).
				Msg("whatsapp: confirmação de pedido enviada")
		}
	}()
}

// NotifyOrderStatusViaWA envia atualização de status de pedido de forma non-fatal.
func NotifyOrderStatusViaWA(phone, orderID, status string) {
	if whatsappClient == nil || !whatsappClient.IsConfigured() {
		return
	}
	if phone == "" {
		return
	}

	go func() {
		ctx := context.Background()
		if err := whatsappClient.SendOrderStatusUpdate(ctx, phone, orderID, status); err != nil {
			log.Error().
				Err(err).
				Str("order_id", orderID).
				Str("status", status).
				Msg("whatsapp: falha ao notificar atualização de status")
		} else {
			log.Info().
				Str("order_id", orderID).
				Str("status", status).
				Msg("whatsapp: atualização de status enviada")
		}
	}()
}
