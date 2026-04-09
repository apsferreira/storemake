# ADR-001: Módulo de Estoque Multi-Loja dentro do StoreMake

**Data:** 2026-04-09
**Status:** Aceito
**Decisores:** Antonio Pedro Ferreira

---

## Contexto

Um cliente do StoreMake opera múltiplas lojas físicas/virtuais sob um mesmo grupo. Ele tem um estoque único centralizado que abastece todas as lojas. As dores principais são:

1. Não sabe o que tem em estoque em cada loja de forma consolidada
2. Não sabe o que está faltando / precisa pedir ao fornecedor
3. Cada loja repassa uma parte do lucro para a entidade master (holding)
4. Quer escalabilidade: futuramente pode ter mais lojas no grupo

### Opções consideradas

| Opção | Descrição | Pro | Contra |
|-------|-----------|-----|--------|
| **A) inventory-service** (novo serviço) | Go/Fiber standalone, banco próprio, API REST | Isolamento completo, escala independente | Só um cliente usa; overhead de infra; mais um K8s deploy; custo de manutenção |
| **B) Módulo no catalog-service** | Adicionar ao catálogo do IIT | Reutiliza catálogo | Catalog é vitrine pública do IIT, não de clientes. Domínio errado. |
| **C) Módulo no StoreMake** (escolhida) | Migration 016 + handlers Go/Fiber | Reutiliza lojas, produtos, pedidos existentes. Vendável como feature "Multi-loja". | Acopla mais responsabilidade ao storemake |

## Decisão

**Opção C: Módulo dentro do StoreMake.**

### Justificativas

1. **StoreMake já tem o contexto correto.** A tabela `lojas` existe, `produtos` já tem `stock_quantity` e `stock_alert_threshold`. A migration 016 expande naturalmente o que já existe.

2. **Vendável como feature de produto.** "Gestão Multi-loja com estoque centralizado" é um diferencial do StoreMake para tenants com múltiplas unidades — não é uma solução one-off. Outros clientes com redes de lojas podem usar.

3. **Sem overhead de novo serviço.** Evita mais um deployment K3s, mais um banco de dados, mais um Cloudflare tunnel, mais uma pipeline CI/CD.

4. **O catalog-service é inadequado.** É a vitrine de produtos *do IIT* (Libri, Nitro, MeuIngresso), não de produtos de clientes do StoreMake.

---

## Arquitetura do Módulo

### Entidades (migration 016)

```
inventory_masters          — SKU centralizado por tenant (holding)
├── quantity_total         — estoque global disponível
├── reorder_point          — gatilho de reposição
└── reorder_quantity       — quantidade sugerida de pedido

store_allocations          — cota por loja + repasse de lucro
├── quantity_allocated     — quanto a loja tem disponível
├── quantity_sold          — giro da loja (para relatórios)
└── profit_share_pct       — % do lucro que fica com a loja

inventory_movements        — auditoria completa de toda movimentação
├── entrada / saida_venda / saida_perda / transferencia / ajuste / devolucao
├── quantity_before/after  — snapshot para audit trail
└── created_by             — rastreabilidade

supplier_orders            — pedidos de reposição ao fornecedor
inventory_alerts           — alertas automáticos (trigger PG) quando stock <= reorder_point
```

### Trigger de alerta automático

Um trigger PostgreSQL dispara `inventory_alerts` automaticamente quando `quantity_total <= reorder_point`. Sem polling. Alertas podem ser consultados via API e futuramente notificados via notification-service.

### Repasse de lucro

```
lucro_venda = (preco_venda - custo_unitario) × qty_vendida
repasse_loja = lucro_venda × (profit_share_pct / 100)
retido_master = lucro_venda × (1 - profit_share_pct / 100)
```

Calculado por relatório periódico (endpoint `GET /inventory/masters/:id/profit-report?from=&to=`).

---

## Consequências

**Positivas:**
- Feature de produto disponível para todos os tenants multi-loja
- Audit trail completo de movimentações
- Alertas automáticos sem polling
- Reutiliza infraestrutura existente

**Negativas:**
- StoreMake fica mais complexo (aceitável — é o produto mais maduro)
- Se o módulo de inventário escalar muito (milhares de SKUs, alta frequência de movimentações), pode ser extraído como serviço separado futuramente

**Pendências de implementação:**
- [ ] Handlers Go/Fiber para CRUD de `inventory_masters`
- [ ] Handler de alocação para lojas (`POST /inventory/masters/:id/allocate`)
- [ ] Handler de movimentação manual (ajuste de estoque)
- [ ] Decrementar automaticamente quando pedido é confirmado (hook no handler de pedidos)
- [ ] Endpoint de relatório consolidado (`GET /inventory/dashboard`)
- [ ] Endpoint de relatório de repasse (`GET /inventory/profit-report`)
- [ ] Frontend admin: tela de dashboard de estoque
- [ ] Integração com notification-service para alertas de baixo estoque

---

## Referências

- Migration: `backend/migrations/016_add_inventory_multiloja.up.sql`
- BKL-900: pedido original do cliente
- Modelo de dados StoreMake: `migrations/001-015`
