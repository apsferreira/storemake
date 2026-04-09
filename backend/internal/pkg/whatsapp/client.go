// Package whatsapp fornece integração com a WhatsApp Business Cloud API (Meta).
// Documentação: https://developers.facebook.com/docs/whatsapp/cloud-api
package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const apiBaseURL = "https://graph.facebook.com"

// Client encapsula as credenciais e o http.Client para a Cloud API.
type Client struct {
	phoneNumberID string
	accessToken   string
	apiVersion    string
	httpClient    *http.Client
}

// New cria um Client configurado.
// phoneNumberID e accessToken são obtidos no Meta Business Dashboard.
// Se qualquer um estiver vazio, IsConfigured() retorna false e os métodos de envio
// retornam erro sem fazer chamadas de rede.
func New(phoneNumberID, accessToken string) *Client {
	return &Client{
		phoneNumberID: phoneNumberID,
		accessToken:   accessToken,
		apiVersion:    "v19.0",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// IsConfigured reporta se as credenciais estão presentes.
func (c *Client) IsConfigured() bool {
	return c.phoneNumberID != "" && c.accessToken != ""
}

// textPayload monta o corpo da requisição para mensagem de texto simples.
type textPayload struct {
	MessagingProduct string          `json:"messaging_product"`
	RecipientType    string          `json:"recipient_type"`
	To               string          `json:"to"`
	Type             string          `json:"type"`
	Text             textBodyPayload `json:"text"`
}

type textBodyPayload struct {
	PreviewURL bool   `json:"preview_url"`
	Body       string `json:"body"`
}

// apiErrorResponse representa o corpo de erro da API do Meta.
type apiErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}

// SendTextMessage envia uma mensagem de texto simples para um número no formato E.164 (ex: +5511999999999).
func (c *Client) SendTextMessage(ctx context.Context, to, text string) error {
	if !c.IsConfigured() {
		return fmt.Errorf("whatsapp: credenciais não configuradas")
	}

	payload := textPayload{
		MessagingProduct: "whatsapp",
		RecipientType:    "individual",
		To:               to,
		Type:             "text",
		Text:             textBodyPayload{PreviewURL: false, Body: text},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("whatsapp: falha ao serializar payload: %w", err)
	}

	url := fmt.Sprintf("%s/%s/%s/messages", apiBaseURL, c.apiVersion, c.phoneNumberID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("whatsapp: falha ao criar requisição: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("whatsapp: erro na chamada HTTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		var apiErr apiErrorResponse
		if jsonErr := json.Unmarshal(raw, &apiErr); jsonErr == nil && apiErr.Error.Message != "" {
			return fmt.Errorf("whatsapp API erro %d: %s", apiErr.Error.Code, apiErr.Error.Message)
		}
		return fmt.Errorf("whatsapp API HTTP %d", resp.StatusCode)
	}

	return nil
}

// SendOrderConfirmation envia mensagem de confirmação de pedido.
// phone deve estar no formato E.164 (ex: +5511999999999).
func (c *Client) SendOrderConfirmation(ctx context.Context, phone, customerName, orderID string, totalCents int64) error {
	totalBRL := float64(totalCents) / 100
	msg := fmt.Sprintf(
		"Olá, %s! Seu pedido #%s foi confirmado com sucesso.\nTotal: R$ %.2f\nAcompanhe pelo site ou entre em contato conosco. Obrigado!",
		customerName, orderID, totalBRL,
	)
	return c.SendTextMessage(ctx, phone, msg)
}

// SendOrderStatusUpdate envia atualização de status de pedido.
// phone deve estar no formato E.164.
// status deve ser um dos valores usados pelo storemake (pendente, pago, preparando, enviado, entregue, cancelado).
func (c *Client) SendOrderStatusUpdate(ctx context.Context, phone, orderID, status string) error {
	statusLabels := map[string]string{
		"pendente":   "pendente de pagamento",
		"pago":       "pagamento confirmado",
		"preparando": "em preparo",
		"enviado":    "enviado para entrega",
		"entregue":   "entregue",
		"cancelado":  "cancelado",
	}
	label, ok := statusLabels[status]
	if !ok {
		label = status
	}
	msg := fmt.Sprintf("Atualização do pedido #%s: %s.", orderID, label)
	return c.SendTextMessage(ctx, phone, msg)
}
