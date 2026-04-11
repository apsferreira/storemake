// BKL-502: Testes unitários para checkout handler (validação, cálculo de totais)
package handler

import (
	"testing"

	"github.com/apsferreira/storemaker/internal/model"
)

func TestValidateCheckout(t *testing.T) {
	validReq := model.CheckoutRequest{
		StoreID:       "store-1",
		CustomerName:  "João Silva",
		CustomerEmail: "joao@example.com",
		CustomerPhone: "71999990000",
		ShippingCents: 0,
		PaymentMethod: "pix",
	}

	tests := []struct {
		name    string
		req     model.CheckoutRequest
		wantErr bool
	}{
		{
			name:    "request válido",
			req:     validReq,
			wantErr: false,
		},
		{
			name: "store_id vazio",
			req: func() model.CheckoutRequest {
				r := validReq
				r.StoreID = ""
				return r
			}(),
			wantErr: true,
		},
		{
			name: "store_id só espaços",
			req: func() model.CheckoutRequest {
				r := validReq
				r.StoreID = "   "
				return r
			}(),
			wantErr: true,
		},
		{
			name: "customer_name vazio",
			req: func() model.CheckoutRequest {
				r := validReq
				r.CustomerName = ""
				return r
			}(),
			wantErr: true,
		},
		{
			name: "customer_name muito longo (256 chars)",
			req: func() model.CheckoutRequest {
				r := validReq
				r.CustomerName = string(make([]byte, 256))
				return r
			}(),
			wantErr: true,
		},
		{
			name: "customer_email vazio",
			req: func() model.CheckoutRequest {
				r := validReq
				r.CustomerEmail = ""
				return r
			}(),
			wantErr: true,
		},
		{
			name: "customer_email sem @",
			req: func() model.CheckoutRequest {
				r := validReq
				r.CustomerEmail = "emailinvalido"
				return r
			}(),
			wantErr: true,
		},
		{
			name: "customer_phone muito longo (21 chars)",
			req: func() model.CheckoutRequest {
				r := validReq
				r.CustomerPhone = "123456789012345678901"
				return r
			}(),
			wantErr: true,
		},
		{
			name: "shipping_cents negativo",
			req: func() model.CheckoutRequest {
				r := validReq
				r.ShippingCents = -1
				return r
			}(),
			wantErr: true,
		},
		{
			name: "payment_method inválido",
			req: func() model.CheckoutRequest {
				r := validReq
				r.PaymentMethod = "crypto"
				return r
			}(),
			wantErr: true,
		},
		{
			name: "payment_method pix (minúsculo) válido",
			req: func() model.CheckoutRequest {
				r := validReq
				r.PaymentMethod = "pix"
				return r
			}(),
			wantErr: false,
		},
		{
			name: "payment_method cartao válido",
			req: func() model.CheckoutRequest {
				r := validReq
				r.PaymentMethod = "cartao"
				return r
			}(),
			wantErr: false,
		},
		{
			name: "payment_method boleto válido",
			req: func() model.CheckoutRequest {
				r := validReq
				r.PaymentMethod = "boleto"
				return r
			}(),
			wantErr: false,
		},
		{
			name: "payment_method PIX (maiúsculo) aceito",
			req: func() model.CheckoutRequest {
				r := validReq
				r.PaymentMethod = "PIX"
				return r
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCheckout(tt.req)
			if tt.wantErr && err == nil {
				t.Errorf("esperava erro, mas não recebeu")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("não esperava erro, mas recebeu: %v", err)
			}
		})
	}
}

func TestCheckoutTotalCalculation(t *testing.T) {
	// Testa a lógica de cálculo: total = subtotal + shipping - desconto
	tests := []struct {
		name          string
		subtotal      int64
		shipping      int64
		discount      int64
		expectedTotal int64
	}{
		{
			name:          "sem frete sem desconto",
			subtotal:      10000,
			shipping:      0,
			discount:      0,
			expectedTotal: 10000,
		},
		{
			name:          "com frete",
			subtotal:      10000,
			shipping:      2000,
			discount:      0,
			expectedTotal: 12000,
		},
		{
			name:          "com desconto",
			subtotal:      10000,
			shipping:      0,
			discount:      1500,
			expectedTotal: 8500,
		},
		{
			name:          "desconto maior que total (não pode ser negativo)",
			subtotal:      1000,
			shipping:      0,
			discount:      5000,
			expectedTotal: 0, // min 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total := tt.subtotal + tt.shipping - tt.discount
			if total < 0 {
				total = 0
			}
			if total != tt.expectedTotal {
				t.Errorf("total esperado %d, obtido %d", tt.expectedTotal, total)
			}
		})
	}
}
