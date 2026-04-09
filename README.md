# StoreMake

Vitrine digital multi-tenant do [Instituto Itinerante de Tecnologia](https://institutoitinerante.com.br). Permite que lojistas criem sua loja virtual em minutos com catalogo de produtos, carrinho de compras, checkout integrado e dashboard de vendas.

## Stack

- **Backend:** Go 1.24 + Fiber v2
- **Frontend (Web):** React + TypeScript + Vite (loja publica)
- **Frontend (Admin):** React + TypeScript + Vite (painel do lojista)
- **Database:** PostgreSQL 15/16
- **Pagamento:** Asaas (PIX + cartao via checkout-service)
- **Storage:** Uploads locais (migrar para MinIO)
- **Infra:** Docker + K3s + ArgoCD

## Arquitetura

```
                  ┌─────────────────┐
                  │   Loja Publica  │  (React SPA)
                  │  /web           │
                  └────────┬────────┘
                           │
                           ▼
┌──────────────┐  ┌─────────────────┐  ┌──────────────────┐
│ Admin Panel  │  │   StoreMake     │  │ checkout-service  │
│ /admin       │──│   Backend (Go)  │──│ (Asaas)           │
└──────────────┘  │   :3070         │  └──────────────────┘
                  └────────┬────────┘
                           │
                  ┌────────┴────────┐
                  │   PostgreSQL    │
                  │   storemake_db  │
                  └─────────────────┘
```

## Funcionalidades

### Lojista (Admin)
- Criar e gerenciar loja (nome, logo, cores da marca)
- CRUD de categorias
- CRUD de produtos com variacoes (tamanho, cor, etc)
- Upload de fotos de produtos (multiplas por produto)
- Importacao de produtos via CSV
- Cupons de desconto (% ou valor fixo)
- Gestao de estoque com alertas de estoque baixo
- Reordenacao de produtos (drag & drop)
- Dashboard de vendas (GMV, pedidos, top produtos)
- Relatorios: vendas por periodo, top produtos, alertas de estoque
- Exportacao de vendas em CSV
- Lista de clientes com historico

### Comprador (Web)
- Catalogo publico com filtros por categoria
- Detalhes do produto com fotos e variacoes
- Carrinho de compras (adicionar, atualizar quantidade, remover)
- Cupom de desconto no carrinho
- Checkout com pagamento via Asaas (PIX/cartao)
- Webhook de confirmacao de pagamento

## Endpoints

### Publicos (sem auth)
| Metodo | Rota | Descricao |
|--------|------|-----------|
| GET | `/health` | Health check |
| GET | `/api/v1/public/catalog` | Catalogo publico com filtros |
| POST | `/api/v1/cart/add` | Adicionar item ao carrinho |
| PUT | `/api/v1/cart/update/:id` | Atualizar quantidade |
| DELETE | `/api/v1/cart/remove/:id` | Remover item |
| GET | `/api/v1/cart` | Ver carrinho |
| POST | `/api/v1/checkout` | Finalizar compra |
| POST | `/api/v1/coupons/validate` | Validar cupom |
| POST | `/api/v1/webhooks/payment` | Webhook pagamento Asaas |

### Lojista (JWT auth)
| Metodo | Rota | Descricao |
|--------|------|-----------|
| POST | `/api/v1/categories` | Criar categoria |
| GET | `/api/v1/categories` | Listar categorias |
| GET | `/api/v1/categories/:id` | Detalhe categoria |
| PUT | `/api/v1/categories/:id` | Atualizar categoria |
| DELETE | `/api/v1/categories/:id` | Remover categoria |
| POST | `/api/v1/products` | Criar produto |
| GET | `/api/v1/products` | Listar produtos |
| GET | `/api/v1/products/:id` | Detalhe produto |
| PUT | `/api/v1/products/:id` | Atualizar produto |
| DELETE | `/api/v1/products/:id` | Remover produto |
| PUT | `/api/v1/products/reorder` | Reordenar produtos |
| POST | `/api/v1/products/:id/photos` | Upload fotos |
| POST | `/api/v1/products/import` | Importar CSV |
| GET | `/api/v1/stock/alerts` | Alertas de estoque |
| GET | `/api/v1/orders` | Listar pedidos |
| GET | `/api/v1/orders/:id` | Detalhe pedido |

### Admin
| Metodo | Rota | Descricao |
|--------|------|-----------|
| GET | `/admin/customers` | Listar clientes |
| GET | `/admin/customers/:id` | Detalhe cliente |
| GET | `/admin/dashboard` | Dashboard de vendas |
| GET | `/admin/reports/sales` | Relatorio de vendas |
| GET | `/admin/reports/products` | Top produtos |
| GET | `/admin/reports/stock-alerts` | Alertas de estoque |
| GET | `/admin/reports/export` | Exportar vendas CSV |

## Modelo de Dados

```
lojas
├── categorias
│   └── produtos
│       ├── produto_fotos
│       └── produto_variacoes
├── cupons
├── pedidos (orders)
│   └── order_items
└── clientes (customers)
```

## Desenvolvimento Local

```bash
# Backend
cd backend
cp ../env.example .env
go run ./cmd/api

# Frontend Web (loja publica)
cd web
npm install
npm run dev

# Frontend Admin (painel lojista)
cd admin
npm install
npm run dev
```

## Variaveis de Ambiente

```env
DATABASE_URL=postgres://postgres:password@localhost:5432/storemake_db?sslmode=disable
JWT_SECRET=your-jwt-secret
WEBHOOK_SECRET=your-webhook-secret
PORT=3070
CORS_ORIGINS=http://localhost:5173,http://localhost:5174
```

## Clientes

- **Semi-joias** — catalogo de joias e bijuterias
- **Fabrica de Festas** — cardapio de salgados para eventos

## Ecossistema IIT

O StoreMake integra com:
- **checkout-service** — processamento de pagamento (Asaas)
- **cart-service** — carrinho de compras compartilhado
- **contract-service** — termos de uso e contratos
- **auth-service** — autenticacao OTP
- **notification-service** — emails transacionais
- **SocialMake** — marketing dos produtos nas redes sociais

## Deploy

```bash
# Docker
docker build -t storemake-backend -f backend/Dockerfile backend/

# K3s
kubectl apply -f k8s/
```

## Licenca

Propriedade do Instituto Itinerante de Tecnologia. Todos os direitos reservados.
