package handler

// BKL-422: StoreMake — Integração com Melhor Envio API para cálculo de frete.
// Calcula frete automático para pedidos, compara transportadoras (Correios, Jadlog, etc.)
// e permite geração de etiquetas pós-pagamento.
//
// Endpoints:
//   POST /api/v1/shipping/calculate    — calcula frete para um carrinho (sem autenticação)
//   GET  /api/v1/shipping/options/:id  — retorna opções de frete para um pedido específico
//   POST /api/v1/shipping/label/:id    — gera etiqueta de envio para pedido pago (autenticado)

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ShippingOption representa uma opção de frete.
type ShippingOption struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`           // ex: "PAC", "SEDEX", "Jadlog"
	Company       string    `json:"company"`        // ex: "Correios", "Jadlog"
	Price         float64   `json:"price"`          // preço em BRL
	Discount      float64   `json:"discount"`       // desconto aplicado
	Currency      string    `json:"currency"`       // "BRL"
	DeliveryTime  int       `json:"delivery_time"`  // dias úteis
	DeliveryRange struct {
		Min int `json:"min"`
		Max int `json:"max"`
	} `json:"delivery_range"`
	Error *string `json:"error,omitempty"`
}

// ShippingCalculateRequest payload para cálculo de frete.
type ShippingCalculateRequest struct {
	FromPostalCode string              `json:"from_postal_code"` // CEP de origem (loja)
	ToPostalCode   string              `json:"to_postal_code"`   // CEP de destino (cliente)
	Products       []ShippingProduct   `json:"products"`
}

// ShippingProduct item para cálculo de frete.
type ShippingProduct struct {
	ID           string  `json:"id"`
	Width        float64 `json:"width"`        // cm
	Height       float64 `json:"height"`       // cm
	Length       float64 `json:"length"`       // cm
	Weight       float64 `json:"weight"`       // kg
	Insurance    float64 `json:"insurance"`    // valor declarado BRL
	Quantity     int     `json:"quantity"`
	IsFragile    bool    `json:"is_fragile,omitempty"`
}

// Calculate POST /api/v1/shipping/calculate
func Calculate(c *fiber.Ctx) error {
	var req ShippingCalculateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payload inválido"})
	}

	if req.ToPostalCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "to_postal_code obrigatório"})
	}
	if len(req.Products) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "products obrigatório"})
	}

	fromCEP := req.FromPostalCode
	if fromCEP == "" {
		fromCEP = os.Getenv("STORE_DEFAULT_CEP")
	}
	if fromCEP == "" {
		fromCEP = "40000000" // Salvador/BA como fallback
	}

	options, err := calculateShippingMelhorEnvio(fromCEP, req.ToPostalCode, req.Products)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("erro ao calcular frete: %v", err)})
	}

	return c.JSON(fiber.Map{
		"from_postal_code": fromCEP,
		"to_postal_code":   req.ToPostalCode,
		"options":          options,
		"calculated_at":    time.Now().UTC(),
	})
}

// calculateShippingMelhorEnvio chama a API do Melhor Envio para calcular frete.
// Usa o ambiente sandbox ou produção baseado em MELHOR_ENVIO_SANDBOX.
func calculateShippingMelhorEnvio(fromCEP, toCEP string, products []ShippingProduct) ([]ShippingOption, error) {
	token := os.Getenv("MELHOR_ENVIO_TOKEN")
	if token == "" {
		// Sem token: retorna estimativa mockada para não bloquear o checkout
		return mockShippingOptions(), nil
	}

	baseURL := "https://melhorenvio.com.br"
	if os.Getenv("MELHOR_ENVIO_SANDBOX") == "true" {
		baseURL = "https://sandbox.melhorenvio.com.br"
	}

	// Mapeia produtos para o formato da API
	type meProduct struct {
		ID          string  `json:"id"`
		Width       float64 `json:"width"`
		Height      float64 `json:"height"`
		Length      float64 `json:"length"`
		Weight      float64 `json:"weight"`
		Insurance   float64 `json:"insurance_value"`
		Quantity    int     `json:"quantity"`
	}
	var meProducts []meProduct
	for _, p := range products {
		meProducts = append(meProducts, meProduct{
			ID:        p.ID,
			Width:     p.Width,
			Height:    p.Height,
			Length:    p.Length,
			Weight:    p.Weight,
			Insurance: p.Insurance,
			Quantity:  maxInt(p.Quantity, 1),
		})
	}

	payload := map[string]interface{}{
		"from":     map[string]string{"postal_code": fromCEP},
		"to":       map[string]string{"postal_code": toCEP},
		"package":  meProducts[0], // simplificado: usa o primeiro pacote
		"options": map[string]interface{}{
			"receipt":           false,
			"own_hand":          false,
			"collect":           false,
			"reverse":           false,
			"non_commercial":    false,
		},
		// Todas as transportadoras disponíveis
		"services": "1,2,3,4,7,8,291,292,293,294", // IDs padrão ME: Correios PAC/SEDEX + Jadlog + etc
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/v2/me/shipment/calculate", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "StoreMake/1.0 (store@institutoitinerante.com.br)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Melhor Envio retorna array de opções
	var meOptions []struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Company  struct {
			Name string `json:"name"`
		} `json:"company"`
		Price        string `json:"price"`
		Discount     string `json:"discount,omitempty"`
		Currency     string `json:"currency"`
		DeliveryTime int    `json:"delivery_time"`
		DeliveryRange struct {
			Min int `json:"min"`
			Max int `json:"max"`
		} `json:"delivery_range"`
		Error *string `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&meOptions); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta ME: %w", err)
	}

	var options []ShippingOption
	for _, opt := range meOptions {
		price := 0.0
		fmt.Sscanf(opt.Price, "%f", &price) //nolint

		discount := 0.0
		if opt.Discount != "" {
			fmt.Sscanf(opt.Discount, "%f", &discount)
		}

		o := ShippingOption{
			ID:           fmt.Sprintf("%d", opt.ID),
			Name:         opt.Name,
			Company:      opt.Company.Name,
			Price:        price,
			Discount:     discount,
			Currency:     "BRL",
			DeliveryTime: opt.DeliveryTime,
			Error:        opt.Error,
		}
		o.DeliveryRange.Min = opt.DeliveryRange.Min
		o.DeliveryRange.Max = opt.DeliveryRange.Max
		options = append(options, o)
	}

	return options, nil
}

// mockShippingOptions retorna opções de frete mockadas quando não há token ME configurado.
// Permite testar o checkout sem integração real.
func mockShippingOptions() []ShippingOption {
	pac := ShippingOption{
		ID:           "1",
		Name:         "PAC",
		Company:      "Correios",
		Price:        15.90,
		Currency:     "BRL",
		DeliveryTime: 7,
	}
	pac.DeliveryRange.Min = 5
	pac.DeliveryRange.Max = 9

	sedex := ShippingOption{
		ID:           "2",
		Name:         "SEDEX",
		Company:      "Correios",
		Price:        28.50,
		Currency:     "BRL",
		DeliveryTime: 2,
	}
	sedex.DeliveryRange.Min = 1
	sedex.DeliveryRange.Max = 3

	return []ShippingOption{pac, sedex}
}

// maxInt retorna o maior de dois inteiros.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
