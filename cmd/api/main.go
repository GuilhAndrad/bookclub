package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/GuilhAndrad/bookclub/config"
	"github.com/GuilhAndrad/bookclub/internal/domain"
	"github.com/GuilhAndrad/bookclub/internal/server"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		// slog ainda não está configurado aqui, mas o processo deve encerrar.
		slog.Error("erro ao carregar configuração", "error", err)
		os.Exit(1)
	}

	setupLogger(cfg)

	db := mustConnectDB(cfg)
	mustMigrate(db)

	// ctx raiz — cancelado ao receber SIGTERM ou SIGINT.
	// Repassado ao server.New para encerrar goroutines internas limpo.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	r := server.New(ctx, db, cfg)

	srv := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           r,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
	}

	// Inicia o servidor em goroutine separada para não bloquear o select abaixo.
	go func() {
		slog.Info("servidor iniciado", "port", cfg.ServerPort, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("falha ao iniciar servidor", "error", err)
			os.Exit(1)
		}
	}()

	// Aguarda sinal de encerramento.
	<-ctx.Done()
	slog.Info("sinal de encerramento recebido, aguardando requisições em andamento...")

	// Graceful shutdown com timeout configurável.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("erro no graceful shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("servidor encerrado com sucesso")
}

// setupLogger configura o logger global slog.
// Em produção usa JSON (legível por ferramentas de log como Datadog, CloudWatch, Loki).
// Em desenvolvimento usa texto colorido legível no terminal.
func setupLogger(cfg *config.Config) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: cfg.IsProduction(), // adiciona arquivo:linha apenas em produção
	}

	if cfg.IsProduction() {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}

// mustConnectDB abre a conexão, configura o connection pool
// e encerra o processo em caso de falha.
func mustConnectDB(cfg *config.Config) *gorm.DB {
	gormLogger := logger.Default.LogMode(logger.Info)
	if cfg.IsProduction() {
		// Em produção o log do GORM é silencioso — queries aparecem via slog do app.
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		slog.Error("erro ao conectar ao banco", "error", err)
		os.Exit(1)
	}

	// Configura o pool de conexões do database/sql subjacente.
	// Sem isso o padrão permite conexões ilimitadas — problema real sob carga.
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("erro ao obter instância sql.DB", "error", err)
		os.Exit(1)
	}

	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.DBConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.DBConnMaxIdleTime)

	slog.Info("banco de dados conectado",
		"max_open", cfg.DBMaxOpenConns,
		"max_idle", cfg.DBMaxIdleConns,
	)

	return db
}

// mustMigrate executa as migrações automáticas e encerra o processo em caso de falha.
func mustMigrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&domain.User{},
		&domain.Book{},
		&domain.Review{},
		&domain.Like{},
		&domain.Comment{},
	)
	if err != nil {
		slog.Error("erro nas migrações", "error", err)
		os.Exit(1)
	}
	slog.Info("migrações executadas")
}