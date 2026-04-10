package repository

// SPEC-006-B: Feature flags de módulos por tenant.

import (
	"encoding/json"
	"fmt"

	"github.com/apsferreira/storemaker/internal/model"
	"github.com/apsferreira/storemaker/internal/pkg/database"
)

// GetTenantModules retorna todos os módulos do tenant com seus estados.
func GetTenantModules(tenantID string) ([]model.TenantModule, error) {
	rows, err := database.DB.Query(
		`SELECT tenant_id, module, enabled, config, updated_at
		   FROM tenant_modules
		  WHERE tenant_id = $1
		  ORDER BY module`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var modules []model.TenantModule
	for rows.Next() {
		var m model.TenantModule
		if err := rows.Scan(&m.TenantID, &m.Module, &m.Enabled, &m.Config, &m.UpdatedAt); err != nil {
			return nil, err
		}
		modules = append(modules, m)
	}
	return modules, nil
}

// IsModuleEnabled verifica se um módulo específico está habilitado para o tenant.
func IsModuleEnabled(tenantID string, module model.ModuleName) (bool, error) {
	var enabled bool
	err := database.DB.QueryRow(
		`SELECT enabled FROM tenant_modules WHERE tenant_id = $1 AND module = $2`,
		tenantID, string(module),
	).Scan(&enabled)
	if err != nil {
		// Se não encontrado, assume habilitado (backward compat)
		return true, nil
	}
	return enabled, nil
}

// UpsertTenantModule cria ou atualiza a configuração de um módulo.
func UpsertTenantModule(tenantID string, module model.ModuleName, enabled bool, config map[string]any) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("erro ao serializar config: %w", err)
	}

	_, err = database.DB.Exec(
		`INSERT INTO tenant_modules (tenant_id, module, enabled, config)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (tenant_id, module) DO UPDATE
		 SET enabled = EXCLUDED.enabled, config = EXCLUDED.config, updated_at = NOW()`,
		tenantID, string(module), enabled, configJSON,
	)
	return err
}

// EnsureTenantModules inicializa os módulos padrão para um novo tenant.
// Deve ser chamado na criação de loja.
func EnsureTenantModules(tenantID string) error {
	for _, mod := range model.AllModules {
		_, err := database.DB.Exec(
			`INSERT INTO tenant_modules (tenant_id, module, enabled, config)
			 VALUES ($1, $2, true, '{}')
			 ON CONFLICT DO NOTHING`,
			tenantID, string(mod),
		)
		if err != nil {
			return err
		}
	}
	return nil
}
