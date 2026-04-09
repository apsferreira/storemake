package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Env                    string
	Port                   string
	DatabaseURL            string
	JWTSecret              string
	CORSOrigins            string
	WebhookSecret          string
	WAPhoneNumberID        string
	WAAccessToken          string
	WAWebhookVerifyToken   string
	WAAppSecret            string // VUL-003: HMAC-SHA256 para validar X-Hub-Signature-256
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		Env:                  getEnv("ENV", "development"),
		Port:                 getEnv("PORT", "3080"),
		DatabaseURL:          getEnv("DATABASE_URL", ""),
		JWTSecret:            requireEnv("JWT_SECRET"),    // BKL-114: sem fallback — segredo crítico
		CORSOrigins:          requireEnv("CORS_ORIGINS"),  // BKL-107: sem wildcard default — segurança
		WebhookSecret:        getEnv("WEBHOOK_SECRET", ""),
		WAPhoneNumberID:      getEnv("WHATSAPP_PHONE_NUMBER_ID", ""),    // BKL-144: opcional
		WAAccessToken:        getEnv("WHATSAPP_ACCESS_TOKEN", ""),       // BKL-144: opcional
		WAWebhookVerifyToken: getEnv("WHATSAPP_WEBHOOK_VERIFY_TOKEN", ""),
		WAAppSecret:          getEnv("WHATSAPP_APP_SECRET", ""),         // VUL-003: validação HMAC
	}
}

// requireEnv retorna o valor da variável ou encerra o processo com log.Fatal.
// BKL-107: usar para variáveis críticas de segurança sem default inseguro.
func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("variável de ambiente obrigatória não definida: %s", key)
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
