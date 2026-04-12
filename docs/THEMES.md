# StoreMake — Biblioteca de Temas Iniciais

> BKL-421 | Versão 1.0 | Abril 2026

Três temas prontos para uso imediato ao criar uma loja no StoreMake. Cada tema é um ponto de partida — todos os elementos são editáveis pelo painel.

---

## Tema 1: Minimalista

**Tagline:** Produto em foco. Tudo o mais, fora do caminho.

### Paleta de Cores

| Token | Hex | Uso |
|-------|-----|-----|
| `--color-bg` | `#FFFFFF` | Fundo principal |
| `--color-surface` | `#F8F8F8` | Cards de produto |
| `--color-text-primary` | `#0A0A0A` | Títulos, preços |
| `--color-text-secondary` | `#6B6B6B` | Descrições, labels |
| `--color-accent` | `#000000` | Botão CTA, links |
| `--color-accent-hover` | `#222222` | Hover nos CTAs |
| `--color-border` | `#E5E5E5` | Divisores, bordas de card |
| `--color-success` | `#16A34A` | Badges "Em estoque", confirmações |

### Tipografia

| Elemento | Fonte | Peso | Tamanho |
|----------|-------|------|---------|
| Logotipo / Hero título | `Inter` | 800 (ExtraBold) | 48–64px |
| Título de produto | `Inter` | 700 (Bold) | 24–28px |
| Preço | `Inter` | 700 (Bold) | 22px |
| Descrição | `Inter` | 400 (Regular) | 16px |
| Botão | `Inter` | 600 (SemiBold) | 15px |
| Labels / badges | `Inter` | 500 (Medium) | 12px |

Alternativa Google Fonts gratuita: `Plus Jakarta Sans`.

### Layout de Componentes

```
┌─────────────────────────────────────────────┐
│  LOGO                         [Cart (0)]    │
├─────────────────────────────────────────────┤
│                                             │
│  [Imagem produto — ocupa 55% da largura]    │
│                                             │
│  Nome do Produto                            │
│  R$ 299,00                                  │
│                                             │
│  ──────────────────────────────────         │
│  Descrição curta, máximo 2 linhas.          │
│  ──────────────────────────────────         │
│                                             │
│  [COMPRAR AGORA]  [+ FAVORITAR]             │
│                                             │
├─────────────────────────────────────────────┤
│  Você também pode gostar                    │
│  [P1]  [P2]  [P3]  [P4]                    │
└─────────────────────────────────────────────┘
```

**Regras de layout:**
- Grid de produtos: 2 colunas mobile, 4 colunas desktop
- Espaçamento generoso: padding mínimo de 24px entre elementos
- Sem banners de promoção no hero — produto ocupa o espaço inteiro
- Imagens quadradas (1:1) ou retrato (4:5), fundo branco obrigatório
- Carrinho fixo no topo ao rolar

### Casos de Uso Ideais

- Produtos únicos ou coleção pequena (até 50 SKUs)
- Moda minimalista, lifestyle, papelaria premium
- Produtos onde a fotografia é o argumento de venda
- Marcas que querem transmitir sofisticação discreta
- Venda direta de produtos digitais (cursos, ebooks, templates)

---

## Tema 2: Colorido

**Tagline:** Energia, variedade e prova social na primeira dobra.

### Paleta de Cores

| Token | Hex | Uso |
|-------|-----|-----|
| `--color-bg` | `#FAFAFA` | Fundo principal |
| `--color-primary` | `#FF5733` | CTAs principais, badges de oferta |
| `--color-primary-hover` | `#E04520` | Hover nos CTAs |
| `--color-secondary` | `#6C3DE0` | Acentos, categorias em destaque |
| `--color-tertiary` | `#00C2A8` | Tags "Novo", confirmações |
| `--color-warning` | `#FFAA00` | Estrelas de avaliação, urgência |
| `--color-text-primary` | `#1A1A2E` | Títulos |
| `--color-text-secondary` | `#555577` | Descrições |
| `--color-surface` | `#FFFFFF` | Cards |
| `--color-border` | `#E0E0EF` | Bordas |

### Tipografia

| Elemento | Fonte | Peso | Tamanho |
|----------|-------|------|---------|
| Logotipo / Hero título | `Nunito` | 900 (Black) | 40–56px |
| Título de produto | `Nunito` | 700 (Bold) | 18–20px |
| Preço cheio | `Nunito` | 800 (ExtraBold) | 20px |
| Preço riscado | `Nunito` | 400 | 14px, `line-through` |
| Descrição | `Nunito` | 400 (Regular) | 14px |
| Botão | `Nunito` | 700 (Bold) | 14px, uppercase |
| Badges / contadores | `Nunito` | 700 | 11px |

Alternativa: `Poppins` (igualmente expressiva, mais neutra).

### Layout de Componentes

```
┌────────────────────────────────────────────────────────┐
│  LOGO          [Categorias v]    [Busca]    [Cart (3)] │
├────────────────────────────────────────────────────────┤
│  🔥 OFERTA DA SEMANA — até 40% OFF   [ver tudo →]      │
│  [Banner principal — imagem colorida, full width]       │
├────────────────────────────────────────────────────────┤
│  MAIS VENDIDOS                                         │
│  ┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐           │
│  │    │ │    │ │    │ │    │ │    │ │    │           │
│  │ P1 │ │ P2 │ │ P3 │ │ P4 │ │ P5 │ │ P6 │           │
│  │    │ │    │ │    │ │    │ │    │ │    │           │
│  └────┘ └────┘ └────┘ └────┘ └────┘ └────┘           │
│  R$XX  R$XX  R$XX  R$XX  R$XX  R$XX                   │
├────────────────────────────────────────────────────────┤
│  ⭐⭐⭐⭐⭐  "Recebi em 2 dias, produto incrível!"        │
│  ⭐⭐⭐⭐⭐  "Exatamente como na foto, recomendo!"         │
│  ⭐⭐⭐⭐⭐  "Já é a terceira vez que compro, perfeito!"   │
└────────────────────────────────────────────────────────┘
```

**Regras de layout:**
- Grid de produtos: 2 colunas mobile, 3-4-6 colunas desktop (responsivo)
- Banner rotativo (carousel) com 3-5 slides no hero
- Seção de avaliações sempre visível acima do dobramento em mobile
- Badges visíveis em todos os cards: "Mais vendido", "Novo", "% OFF"
- Contador regressivo em ofertas com prazo (aumenta urgência)
- Filtro de categorias fixo ao rolar em mobile (bottom bar)

### Casos de Uso Ideais

- Lojas com catálogo amplo (50+ SKUs)
- E-commerce de moda jovem, acessórios, itens de decoração coloridos
- Produtos de consumo recorrente com prova social forte
- Mercados onde o preço é argumento principal
- Lojas que dependem de tráfego pago (Meta Ads, TikTok) — alta conversão visual

---

## Tema 3: Premium

**Tagline:** Luxo discreto. A fotografia faz o trabalho.

### Paleta de Cores

| Token | Hex | Uso |
|-------|-----|-----|
| `--color-bg` | `#0E0E0E` | Fundo principal (dark) |
| `--color-surface` | `#1A1A1A` | Cards, painéis |
| `--color-surface-elevated` | `#242424` | Modais, dropdowns |
| `--color-gold` | `#C9A84C` | Acentos, badges, ícones premium |
| `--color-gold-light` | `#E8C76A` | Hover nos acentos |
| `--color-text-primary` | `#F5F0E8` | Títulos, preços |
| `--color-text-secondary` | `#9A9580` | Descrições, metadata |
| `--color-text-muted` | `#5C5850` | Datas, labels secundários |
| `--color-border` | `#2E2E2E` | Bordas sutis |
| `--color-cta` | `#C9A84C` | Botão principal |
| `--color-cta-text` | `#0E0E0E` | Texto no botão (contraste) |

### Tipografia

| Elemento | Fonte | Peso | Tamanho |
|----------|-------|------|---------|
| Logotipo | `Cormorant Garamond` | 300 (Light) | 28–36px, letter-spacing: 0.15em |
| Hero título | `Cormorant Garamond` | 600 (SemiBold) | 52–72px |
| Título de produto | `Cormorant Garamond` | 500 | 28px |
| Preço | `Inter` | 300 (Light) | 24px |
| Descrição | `Inter` | 300 (Light) | 16px, line-height: 1.8 |
| Botão | `Inter` | 500 (Medium) | 13px, letter-spacing: 0.1em, uppercase |
| Subtítulos / seção | `Cormorant Garamond` | 400, italic | 18px |

Alternativa para logotipo/títulos: `Playfair Display` (mais disponível, igualmente elegante).

### Layout de Componentes

```
┌──────────────────────────────────────────────────────┐
│  [LOGO — discreto, centralizado]                     │
│  Início  |  Coleção  |  Sobre  |  Contato            │
├──────────────────────────────────────────────────────┤
│                                                      │
│  [Foto de produto — full viewport height]            │
│  Imagem escura, produto em destaque com luz lateral  │
│                                                      │
│                    Nome do Produto                   │
│                    Uma linha descritiva              │
│                                                      │
│                   [ADICIONAR AO CARRINHO]            │
│                                                      │
├──────────────────────────────────────────────────────┤
│  A Peça                                              │
│  Texto descritivo em 3-4 parágrafos. Material,       │
│  processo de fabricação, diferenciais. Sem bullets.  │
│                                                      │
│  [Foto detalhe 1]  [Foto detalhe 2]  [Foto detalhe 3]│
├──────────────────────────────────────────────────────┤
│  Outras Peças da Coleção                             │
│  [P1 — foto grande]    [P2 — foto grande]            │
└──────────────────────────────────────────────────────┘
```

**Regras de layout:**
- Hero sempre full-screen (100vh) — fotografia domina
- Máximo 4 produtos por página de coleção em desktop
- Sem badges de desconto ou urgência — contradiz o posicionamento
- Animações suaves (fade-in ao rolar, 300ms ease)
- Cursor customizado (dot pequeno, dourado) para reforçar experiência
- Checkout simplificado: menos campos, mais confiança

### Casos de Uso Ideais

- Joias, relógios, artigos de couro, perfumaria artesanal
- Moda de alto valor (peça acima de R$ 500)
- Arte original, fotografia impressa, escultura
- Produtos de edição limitada ou sob encomenda
- Marcas pessoais de estilistas ou artesãos que priorizam percepção de valor

---

## Decisão de Tema: Guia Rápido

| Quero... | Use |
|----------|-----|
| Produto único, foto impecável | Minimalista |
| Catálogo grande, tráfego pago | Colorido |
| Preço alto, percepção de valor | Premium |
| Testar sem saber ainda | Minimalista (mais versátil) |

## Próximos Temas (Backlog)

- **Marketplace Local** — cards horizontais, filtro por bairro, foco em food/serviços
- **Eventos** — timeline, contagem regressiva, integração MeuIngresso
- **Digital Downloads** — mockup de produto digital, licença, download imediato
