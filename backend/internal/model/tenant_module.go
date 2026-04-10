package model

import "time"

// TenantModule representa a configuração de um módulo para um tenant. SPEC-006-B.
type TenantModule struct {
	TenantID  string    `json:"tenant_id" db:"tenant_id"`
	Module    string    `json:"module" db:"module"`
	Enabled   bool      `json:"enabled" db:"enabled"`
	Config    []byte    `json:"config" db:"config"` // JSONB raw
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ModuleName define os módulos disponíveis no StoreMake.
type ModuleName string

const (
	ModuleStorefront ModuleName = "storefront"
	ModuleInventory  ModuleName = "inventory"
	ModuleCRM        ModuleName = "crm"
	ModuleWhatsApp   ModuleName = "whatsapp"
)

// AllModules lista todos os módulos suportados.
var AllModules = []ModuleName{
	ModuleStorefront,
	ModuleInventory,
	ModuleCRM,
	ModuleWhatsApp,
}

// LojaType define o tipo de uma loja. SPEC-006-B.
type LojaType string

const (
	LojaTypeVirtual LojaType = "virtual" // loja online pública
	LojaTypeFilial  LojaType = "filial"  // filial física sem catálogo público
	LojaTypeMaster  LojaType = "master"  // controla estoque centralizado
)

// UpdateModuleRequest é usado para ativar/desativar um módulo.
type UpdateModuleRequest struct {
	Enabled bool            `json:"enabled"`
	Config  map[string]any  `json:"config,omitempty"`
}
