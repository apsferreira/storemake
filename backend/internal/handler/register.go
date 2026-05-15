package handler

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/apsferreira/storemaker/internal/pkg/database"
	"github.com/apsferreira/storemaker/internal/repository"
)

var slugRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,48}[a-z0-9]$`)

// RegisterRequest é o payload para criar uma nova loja.
// O campo auth_token deve ser o access_token retornado pelo auth-service após verificação do OTP.
type RegisterRequest struct {
	AuthToken string `json:"auth_token"`
	StoreName string `json:"store_name"`
	Slug      string `json:"slug"`
	PlanSlug  string `json:"plan_slug"` // "free", "starter", "pro"
}

// RegisterResponse retorna o JWT do storemake e os dados da loja criada.
type RegisterResponse struct {
	Token   string `json:"token"`
	StoreID string `json:"store_id"`
	Slug    string `json:"slug"`
}

// Register cria uma nova loja para um lojista autenticado via OTP.
// POST /api/v1/public/register
//
// Fluxo:
//  1. Valida o auth_token do auth-service (mesmo JWT_SECRET compartilhado)
//  2. Extrai o user_id (sub) do token
//  3. Verifica se o lojista já tem uma loja
//  4. Valida e normaliza slug + nome
//  5. Resolve o plano pelo slug
//  6. Cria a loja no banco
//  7. Emite JWT próprio do storemake com store_id
func Register(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req RegisterRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo da requisição inválido"})
		}

		// 1. Validar campos obrigatórios
		req.AuthToken = strings.TrimSpace(req.AuthToken)
		req.StoreName = strings.TrimSpace(req.StoreName)
		req.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
		req.PlanSlug = strings.ToLower(strings.TrimSpace(req.PlanSlug))

		if req.AuthToken == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "auth_token é obrigatório"})
		}
		if req.StoreName == "" || len(req.StoreName) > 100 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "store_name deve ter entre 1 e 100 caracteres"})
		}
		if !slugRe.MatchString(req.Slug) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "slug inválido: use apenas letras minúsculas, números e hífens (3-50 caracteres, sem começar ou terminar com hífen)",
			})
		}
		validPlans := map[string]bool{"free": true, "starter": true, "pro": true}
		if !validPlans[req.PlanSlug] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "plan_slug deve ser free, starter ou pro"})
		}

		// 2. Validar o JWT do auth-service
		token, err := jwt.Parse(req.AuthToken, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "método de assinatura inválido")
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "token de autenticação inválido ou expirado"})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "token malformado"})
		}

		ownerIDStr, _ := claims["sub"].(string)
		if ownerIDStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "token sem identificador de usuário"})
		}

		ownerID, err := uuid.Parse(ownerIDStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "identificador de usuário inválido"})
		}

		ctx := context.Background()

		// 3. Verificar se o owner já tem loja
		existing, err := getLojaByOwnerID(ownerID.String())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao verificar conta existente"})
		}
		if existing != nil {
			// Lojista já tem loja — emite novo JWT e redireciona
			storeToken, tokenErr := mintStoreJWT(jwtSecret, ownerID.String(), existing.ID)
			if tokenErr != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao gerar token"})
			}
			return c.Status(fiber.StatusOK).JSON(RegisterResponse{
				Token:   storeToken,
				StoreID: existing.ID,
				Slug:    existing.Slug,
			})
		}

		// 4. Resolver plano
		planoRepo := repository.NewPlanoRepository(database.DB)
		plano, err := planoRepo.GetBySlug(ctx, req.PlanSlug)
		if err != nil || plano == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "plano não encontrado"})
		}

		// 5. Criar loja
		lojaID := uuid.New()
		_, err = database.DB.Exec(
			`INSERT INTO lojas (id, owner_id, name, slug, template, plano_id, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, 'generico', $5, NOW(), NOW())`,
			lojaID.String(), ownerID.String(), req.StoreName, req.Slug, plano.ID.String(),
		)
		if err != nil {
			if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "esse slug já está em uso, escolha outro"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao criar loja"})
		}

		// Ativar módulos padrão para a loja recém-criada
		modules := []string{"storefront", "inventory", "crm", "whatsapp"}
		for _, mod := range modules {
			_, _ = database.DB.Exec(
				`INSERT INTO tenant_modules (tenant_id, module, enabled, config, updated_at)
				 VALUES ($1, $2, true, '{}', NOW())
				 ON CONFLICT (tenant_id, module) DO NOTHING`,
				lojaID.String(), mod,
			)
		}

		// 6. Emitir JWT do storemake
		storeToken, err := mintStoreJWT(jwtSecret, ownerID.String(), lojaID.String())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "loja criada, mas erro ao gerar token — faça login"})
		}

		// BKL-1250: Pulse tracking — fire-and-forget.
		if pulseClient != nil {
			pulseClient.Track(context.Background(), "store.created", ownerID.String(), map[string]string{
				"store_id": lojaID.String(),
				"plan":     req.PlanSlug,
			})
		}

		return c.Status(fiber.StatusCreated).JSON(RegisterResponse{
			Token:   storeToken,
			StoreID: lojaID.String(),
			Slug:    req.Slug,
		})
	}
}

// SuggestSlug gera sugestões de slug a partir do nome da loja.
// GET /api/v1/public/suggest-slug?name=...
func SuggestSlug(c *fiber.Ctx) error {
	name := strings.TrimSpace(c.Query("name"))
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name é obrigatório"})
	}

	base := toSlug(name)
	if len(base) < 3 {
		base = base + "-loja"
	}

	// Verifica disponibilidade e retorna até 3 sugestões
	suggestions := make([]string, 0, 3)
	candidates := []string{base, base + "-shop", base + "-store"}

	for _, candidate := range candidates {
		if !slugRe.MatchString(candidate) {
			continue
		}
		var count int
		_ = database.DB.QueryRow(`SELECT COUNT(*) FROM lojas WHERE slug = $1`, candidate).Scan(&count)
		suggestions = append(suggestions, candidate)
		if count == 0 && len(suggestions) == 1 {
			// Slug base disponível — retorna só ele como primary
			break
		}
	}

	return c.JSON(fiber.Map{"suggestions": suggestions})
}

// CheckSlug verifica se um slug está disponível.
// GET /api/v1/public/check-slug?slug=...
func CheckSlug(c *fiber.Ctx) error {
	slug := strings.ToLower(strings.TrimSpace(c.Query("slug")))
	if !slugRe.MatchString(slug) {
		return c.JSON(fiber.Map{"available": false, "reason": "slug inválido"})
	}

	var count int
	_ = database.DB.QueryRow(`SELECT COUNT(*) FROM lojas WHERE slug = $1`, slug).Scan(&count)

	return c.JSON(fiber.Map{"available": count == 0})
}

// getLojaByOwnerID busca a loja de um owner pelo owner_id.
func getLojaByOwnerID(ownerID string) (*struct{ ID, Slug string }, error) {
	row := database.DB.QueryRow(`SELECT id, slug FROM lojas WHERE owner_id = $1 LIMIT 1`, ownerID)
	var id, slug string
	err := row.Scan(&id, &slug)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &struct{ ID, Slug string }{ID: id, Slug: slug}, nil
}

// mintStoreJWT emite um JWT do storemake com store_id e role owner.
func mintStoreJWT(secret, userID, storeID string) (string, error) {
	claims := jwt.MapClaims{
		"sub":      userID,
		"store_id": storeID,
		"role":     "owner",
		"exp":      time.Now().Add(30 * 24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

// toSlug converte um nome para formato slug.
func toSlug(name string) string {
	// Normalizar: remover acentos ingenuamente (substituição simples)
	replacer := strings.NewReplacer(
		"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"í", "i", "ì", "i", "î", "i", "ï", "i",
		"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
		"ú", "u", "ù", "u", "û", "u", "ü", "u",
		"ç", "c", "ñ", "n",
		"Á", "a", "À", "a", "Â", "a", "Ã", "a",
		"É", "e", "Ê", "e",
		"Í", "i", "Î", "i",
		"Ó", "o", "Ô", "o", "Õ", "o",
		"Ú", "u", "Û", "u",
		"Ç", "c",
	)
	s := replacer.Replace(name)

	var sb strings.Builder
	prevHyphen := false
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(r)
			prevHyphen = false
		} else if !prevHyphen && sb.Len() > 0 {
			sb.WriteRune('-')
			prevHyphen = true
		}
	}

	result := strings.TrimRight(sb.String(), "-")
	if len(result) > 48 {
		result = result[:48]
		result = strings.TrimRight(result, "-")
	}
	return result
}

