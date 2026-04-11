package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Plano representa um plano de preço do StoreMake
type Plano struct {
	ID             uuid.UUID          `json:"id"`
	Slug           string             `json:"slug"`            // 'free', 'starter', 'pro'
	Name           string             `json:"name"`            // 'Free', 'Starter', 'Pro'
	PriceCents     int                `json:"price_cents"`     // em centavos
	MaxProducts    int                `json:"max_products"`    // limite de produtos
	CustomDomain   bool               `json:"custom_domain"`   // permite domínio customizado
	SupportLevel   string             `json:"support_level"`   // 'community', 'email', 'priority'
	Features       json.RawMessage    `json:"features"`        // features adicionais em JSON
	IsActive       bool               `json:"is_active"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}

// PriceDisplay retorna o preço formatado para exibição
func (p *Plano) PriceDisplay() string {
	if p.PriceCents == 0 {
		return "Grátis"
	}
	return "R$ " + formatPrice(p.PriceCents)
}

// HasFeature verifica se um plano tem uma feature específica
func (p *Plano) HasFeature(featureName string) bool {
	if p.Features == nil || len(p.Features) == 0 {
		return false
	}
	var features map[string]interface{}
	if err := json.Unmarshal(p.Features, &features); err != nil {
		return false
	}
	if val, ok := features[featureName]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return false
}

// Converte centavos para formato de preço (ex: 7900 -> "79,00")
func formatPrice(cents int) string {
	reais := cents / 100
	centavos := cents % 100
	if centavos == 0 {
		return string(rune(reais)) + ",00"
	}
	return string(rune(reais)) + "," + string(rune(centavos))
}
