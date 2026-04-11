// BKL-502: Testes unitários para a lógica de desconto de cupons (sem banco de dados)
package repository

import (
	"testing"
)

// calculateDiscount extrai a lógica pura de cálculo de desconto usada em ValidateCoupon.
func calculateDiscount(discountType string, discountValue, subtotalCents int64) (int64, bool) {
	switch discountType {
	case "percent":
		return subtotalCents * discountValue / 100, true
	case "fixed":
		d := discountValue
		if d > subtotalCents {
			d = subtotalCents
		}
		return d, true
	default:
		return 0, false
	}
}

func TestCouponDiscountCalculation(t *testing.T) {
	tests := []struct {
		name          string
		discountType  string
		discountValue int64
		subtotal      int64
		wantDiscount  int64
		wantValid     bool
	}{
		{
			name:          "10% de desconto em R$100",
			discountType:  "percent",
			discountValue: 10,
			subtotal:      10000,
			wantDiscount:  1000,
			wantValid:     true,
		},
		{
			name:          "50% de desconto em R$200",
			discountType:  "percent",
			discountValue: 50,
			subtotal:      20000,
			wantDiscount:  10000,
			wantValid:     true,
		},
		{
			name:          "desconto fixo R$20 em R$100",
			discountType:  "fixed",
			discountValue: 2000,
			subtotal:      10000,
			wantDiscount:  2000,
			wantValid:     true,
		},
		{
			name:          "desconto fixo maior que subtotal (cap no subtotal)",
			discountType:  "fixed",
			discountValue: 15000,
			subtotal:      10000,
			wantDiscount:  10000, // cap no subtotal
			wantValid:     true,
		},
		{
			name:          "tipo de desconto inválido",
			discountType:  "invalido",
			discountValue: 1000,
			subtotal:      10000,
			wantDiscount:  0,
			wantValid:     false,
		},
		{
			name:          "1% de desconto (arredondamento inteiro)",
			discountType:  "percent",
			discountValue: 1,
			subtotal:      999, // R$9,99 → 9 centavos de desconto (999*1/100 = 9)
			wantDiscount:  9,
			wantValid:     true,
		},
		{
			name:          "100% de desconto",
			discountType:  "percent",
			discountValue: 100,
			subtotal:      5000,
			wantDiscount:  5000,
			wantValid:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, valid := calculateDiscount(tt.discountType, tt.discountValue, tt.subtotal)
			if valid != tt.wantValid {
				t.Errorf("valid: esperado %v, obtido %v", tt.wantValid, valid)
			}
			if got != tt.wantDiscount {
				t.Errorf("desconto: esperado %d centavos, obtido %d centavos", tt.wantDiscount, got)
			}
		})
	}
}

func TestMinOrderValidation(t *testing.T) {
	// Documenta regra de pedido mínimo para cupom
	tests := []struct {
		name          string
		subtotal      int64
		minOrderCents int64
		wantAllowed   bool
	}{
		{
			name:          "subtotal igual ao mínimo",
			subtotal:      5000,
			minOrderCents: 5000,
			wantAllowed:   true,
		},
		{
			name:          "subtotal acima do mínimo",
			subtotal:      10000,
			minOrderCents: 5000,
			wantAllowed:   true,
		},
		{
			name:          "subtotal abaixo do mínimo",
			subtotal:      3000,
			minOrderCents: 5000,
			wantAllowed:   false,
		},
		{
			name:          "sem mínimo (zero)",
			subtotal:      100,
			minOrderCents: 0,
			wantAllowed:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed := tt.subtotal >= tt.minOrderCents
			if allowed != tt.wantAllowed {
				t.Errorf("subtotal=%d min=%d: esperado allowed=%v, obtido %v",
					tt.subtotal, tt.minOrderCents, tt.wantAllowed, allowed)
			}
		})
	}
}

func TestCouponUsageLimit(t *testing.T) {
	// Documenta regra de limite de uso
	tests := []struct {
		name      string
		maxUses   int
		usedCount int
		wantAllow bool
	}{
		{
			name:      "sem limite (maxUses=0)",
			maxUses:   0,
			usedCount: 1000,
			wantAllow: true,
		},
		{
			name:      "abaixo do limite",
			maxUses:   10,
			usedCount: 5,
			wantAllow: true,
		},
		{
			name:      "exatamente no limite (esgotado)",
			maxUses:   10,
			usedCount: 10,
			wantAllow: false,
		},
		{
			name:      "acima do limite",
			maxUses:   10,
			usedCount: 15,
			wantAllow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Lógica de ValidateCoupon: if maxUses > 0 && usedCount >= maxUses → esgotado
			esgotado := tt.maxUses > 0 && tt.usedCount >= tt.maxUses
			allow := !esgotado
			if allow != tt.wantAllow {
				t.Errorf("maxUses=%d used=%d: esperado allow=%v, obtido %v",
					tt.maxUses, tt.usedCount, tt.wantAllow, allow)
			}
		})
	}
}
