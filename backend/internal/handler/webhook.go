package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/apsferreira/storemaker/internal/model"
	"github.com/apsferreira/storemaker/internal/repository"
)

// PaymentWebhook recebe notificações de pagamento do checkout-service.
func PaymentWebhook(webhookSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Verificar assinatura do webhook
		signature := c.Get("X-Webhook-Signature")
		if webhookSecret != "" && signature != "" {
			body := c.Body()
			mac := hmac.New(sha256.New, []byte(webhookSecret))
			mac.Write(body)
			expected := hex.EncodeToString(mac.Sum(nil))
			if !hmac.Equal([]byte(signature), []byte(expected)) {
				log.Warn().Msg("webhook: assinatura inválida")
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "assinatura inválida"})
			}
		} else if webhookSecret != "" && signature == "" {
			log.Warn().Msg("webhook: assinatura ausente")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "assinatura ausente"})
		}

		var payload model.PaymentWebhookPayload
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payload inválido"})
		}

		if payload.OrderID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "order_id é obrigatório"})
		}

		log.Info().
			Str("event", payload.Event).
			Str("order_id", payload.OrderID).
			Str("payment_status", payload.Status).
			Msg("webhook de pagamento recebido")

		// Buscar pedido
		order, err := repository.GetOrderByID(payload.OrderID)
		if err != nil {
			log.Error().Err(err).Str("order_id", payload.OrderID).Msg("erro ao buscar pedido no webhook")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro interno"})
		}
		if order == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "pedido não encontrado"})
		}

		// Processar conforme status do pagamento
		switch payload.Status {
		case "approved", "confirmed", "paid":
			// Atualizar pedido para pago
			if err := repository.UpdateOrderPayment(order.ID, payload.PaymentID, "pago"); err != nil {
				log.Error().Err(err).Msg("erro ao atualizar pagamento")
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao processar pagamento"})
			}

			// Baixa automática de estoque
			if err := deductOrderStock(order.ID); err != nil {
				log.Error().Err(err).Str("order_id", order.ID).Msg("erro na baixa de estoque")
				// Não retorna erro ao gateway — pedido já foi confirmado
			}

			// BKL-144: notificação WhatsApp (non-fatal — não bloqueia resposta ao gateway)
			NotifyOrderConfirmedViaWA(order.CustomerPhone, order.CustomerName, order.ID, order.TotalCents)

			log.Info().Str("order_id", order.ID).Msg("pagamento confirmado, estoque baixado")

		case "refunded", "cancelled", "failed":
			if err := repository.UpdateOrderPayment(order.ID, payload.PaymentID, "cancelado"); err != nil {
				log.Error().Err(err).Msg("erro ao cancelar pedido")
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao processar cancelamento"})
			}
			log.Info().Str("order_id", order.ID).Msg("pagamento cancelado/reembolsado")

		default:
			log.Warn().Str("status", payload.Status).Msg("status de pagamento desconhecido")
		}

		return c.JSON(fiber.Map{"received": true})
	}
}

// deductOrderStock baixa o estoque dos itens do pedido.
func deductOrderStock(orderID string) error {
	items, err := repository.GetOrderItems(orderID)
	if err != nil {
		return err
	}

	for _, item := range items {
		var variantID *string
		if item.Variant != "" {
			// O variant aqui é o nome formatado, precisamos do ID
			// Usar variacao_id diretamente da tabela
			variantID = getVariantIDFromOrderItem(orderID, item.ProductID)
		}
		if err := repository.DeductStock(item.ProductID, item.Quantity, variantID); err != nil {
			log.Error().Err(err).
				Str("product_id", item.ProductID).
				Int("quantity", item.Quantity).
				Msg("erro ao baixar estoque do item")
			// Continua com os outros itens mesmo se um falhar
		}
	}

	return nil
}

// getVariantIDFromOrderItem busca o variacao_id do pedido_itens.
func getVariantIDFromOrderItem(orderID, productID string) *string {
	var variantID *string
	_ = repository.GetOrderItemVariantID(orderID, productID, &variantID)
	return variantID
}
