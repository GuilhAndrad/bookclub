package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config mantém todas as configurações da aplicação.
// Todos os campos são exportados para facilitar acesso nas camadas internas.
type Config struct {
	// Banco de dados
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBTimezone string
	DSN        string

	// Connection pool do banco — configurar evita esgotamento de conexões sob carga.
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBConnMaxIdleTime time.Duration

	// JWT
	JWTSecret          string
	JWTExpirationHours int

	// Servidor HTTP
	ServerPort        string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	ShutdownTimeout   time.Duration

	// Ambiente
	Env string
}

// Load carrega a configuração a partir do ambiente.
// Retorna erro se alguma configuração obrigatória estiver ausente ou inválida.
func Load() (*Config, error) {
	// .env é opcional em produção — variáveis do sistema têm precedência.
	_ = godotenv.Load()

	cfg := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "bookclub"),
		DBTimezone: getEnv("DB_TIMEZONE", "UTC"),

		DBMaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
		DBConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		DBConnMaxIdleTime: getEnvDuration("DB_CONN_MAX_IDLE_TIME", 2*time.Minute),

		JWTSecret: getEnv("JWT_SECRET", "change_me_in_production"),

		ServerPort:        getEnv("SERVER_PORT", "8080"),
		ReadTimeout:       getEnvDuration("SERVER_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:      getEnvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
		IdleTimeout:       getEnvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		ReadHeaderTimeout: getEnvDuration("SERVER_READ_HEADER_TIMEOUT", 5*time.Second),
		ShutdownTimeout:   getEnvDuration("SERVER_SHUTDOWN_TIMEOUT", 10*time.Second),

		Env: getEnv("ENV", "development"),
	}

	expHours, err := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "72"))
	if err != nil {
		return nil, fmt.Errorf("config: JWT_EXPIRATION_HOURS inválido: %w", err)
	}
	cfg.JWTExpirationHours = expHours

	cfg.DSN = cfg.buildDSN()

	return cfg, nil
}

// IsProduction retorna true quando o ambiente for produção.
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// buildDSN constrói a string de conexão com o banco de dados.
func (c *Config) buildDSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=%s",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort, c.DBTimezone,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}