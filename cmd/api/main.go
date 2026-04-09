package main

import (
	"log"

	"github.com/GuilhAndrad/bookclub/config"
	"github.com/GuilhAndrad/bookclub/internal/domain"
	"github.com/GuilhAndrad/bookclub/internal/server"
	"gorm.io/gorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm/logger"
)

func main() {
	// Carregar configuração (agora retorna *Config e error)
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("erro ao carregar configuração: %v", err)
	}

	db := mustConnectDB(cfg)
	mustMigrate(db)

	// Passar db e cfg para o server.New
	r := server.New(db, cfg)

	addr := ":" + cfg.ServerPort
	log.Printf("servidor rodando em http://localhost%s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("erro ao iniciar servidor: %v", err)
	}
}

// mustConnectDB agora recebe a configuração para obter o DSN.
func mustConnectDB(cfg *config.Config) *gorm.DB {
	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("erro ao conectar ao banco: %v", err)
	}
	log.Println("banco de dados conectado")
	return db
}

// mustMigrate executa as migrações automáticas.
func mustMigrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&domain.User{},
		&domain.Book{},
		&domain.Review{},
		&domain.Like{},
		&domain.Comment{},
		&domain.UserBook{},
	)
	if err != nil {
		log.Fatalf("erro nas migrações: %v", err)
	}
	log.Println("migrações executadas")
}
