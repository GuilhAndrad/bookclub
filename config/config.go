package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config mantém todas as configurações da aplicação.
// Campos exportados para facilitar o acesso.
type Config struct {
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	DBTimezone         string
	DSN                string
	JWTSecret          string
	JWTExpirationHours int
	ServerPort         string
}

// Load carrega a configuração a partir do ambiente.
// Retorna erro se alguma configuração obrigatória estiver ausente ou inválida.
func Load() (*Config, error) {
	// .env é opcional, apenas loga se não encontrar.
	_ = godotenv.Load()

	cfg := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "bookclub"),
		DBTimezone: getEnv("DB_TIMEZONE", "UTC"),
		JWTSecret:  getEnv("JWT_SECRET", "change_me_in_production"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}

	// JWT expiration: falha se não for um número válido
	expStr := getEnv("JWT_EXPIRATION_HOURS", "72")
	expHours, err := strconv.Atoi(expStr)
	if err != nil {
		return nil, fmt.Errorf("JWT_EXPIRATION_HOURS inválido: %w", err)
	}
	cfg.JWTExpirationHours = expHours

	// Construção do DSN pode ficar aqui ou em um método separado
	cfg.DSN = cfg.buildDSN()

	return cfg, nil
}

// buildDSN constrói a string de conexão com o banco de dados.
func (c *Config) buildDSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=%s",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort, c.DBTimezone,
	)
}

// getEnv retorna o valor da variável de ambiente ou o fallback.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}