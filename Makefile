.PHONY: run build test migrate-up migrate-down help

GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m

PG=shared-postgres
DB=storemaker_db

help: ## Mostra comandos disponíveis
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

run: ## Inicia o servidor em modo dev
	@echo "$(GREEN)Iniciando StoreMaker API...$(NC)"
	cd backend && go run cmd/api/main.go

build: ## Compila o binário
	@echo "$(GREEN)Compilando StoreMaker...$(NC)"
	cd backend && CGO_ENABLED=0 go build -o ../bin/storemaker cmd/api/main.go

test: ## Executa testes
	@echo "$(GREEN)Executando testes...$(NC)"
	cd backend && go test ./...

migrate-up: ## Executa migrations (up)
	@echo "$(GREEN)Executando migrations...$(NC)"
	docker exec -i $(PG) psql -U postgres -d $(DB) < backend/migrations/001_create_lojas.up.sql
	docker exec -i $(PG) psql -U postgres -d $(DB) < backend/migrations/002_create_categorias.up.sql
	docker exec -i $(PG) psql -U postgres -d $(DB) < backend/migrations/003_create_produtos.up.sql
	docker exec -i $(PG) psql -U postgres -d $(DB) < backend/migrations/004_create_produto_fotos.up.sql
	docker exec -i $(PG) psql -U postgres -d $(DB) < backend/migrations/005_create_produto_variacoes.up.sql
	docker exec -i $(PG) psql -U postgres -d $(DB) < backend/migrations/006_create_clientes.up.sql
	docker exec -i $(PG) psql -U postgres -d $(DB) < backend/migrations/007_create_pedidos.up.sql
	docker exec -i $(PG) psql -U postgres -d $(DB) < backend/migrations/008_create_pedido_itens.up.sql
	docker exec -i $(PG) psql -U postgres -d $(DB) < backend/migrations/009_create_cupons.up.sql
	@echo "$(GREEN)Migrations concluídas!$(NC)"

migrate-down: ## Rollback migrations
	@echo "$(YELLOW)Revertendo migrations...$(NC)"
	docker exec -i $(PG) psql -U postgres -d $(DB) < backend/migrations/009_create_cupons.down.sql
	docker exec -i $(PG) psql -U postgres -d $(DB) < backend/migrations/008_create_pedido_itens.down.sql
	docker exec -i $(PG) psql -U postgres -d $(DB) < backend/migrations/007_create_pedidos.down.sql
	docker exec -i $(PG) psql -U postgres -d $(DB) < backend/migrations/006_create_clientes.down.sql
	docker exec -i $(PG) psql -U postgres -d $(DB) < backend/migrations/005_create_produto_variacoes.down.sql
	docker exec -i $(PG) psql -U postgres -d $(DB) < backend/migrations/004_create_produto_fotos.down.sql
	docker exec -i $(PG) psql -U postgres -d $(DB) < backend/migrations/003_create_produtos.down.sql
	docker exec -i $(PG) psql -U postgres -d $(DB) < backend/migrations/002_create_categorias.down.sql
	docker exec -i $(PG) psql -U postgres -d $(DB) < backend/migrations/001_create_lojas.down.sql
	@echo "$(GREEN)Rollback concluído!$(NC)"
