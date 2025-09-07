package main

import (
	"log"
	"multi-upload-api/internal/api"
	"multi-upload-api/internal/config"
	"multi-upload-api/internal/database"
	"multi-upload-api/internal/middleware"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Carregar variáveis de ambiente
	if err := godotenv.Load("config.env"); err != nil {
		log.Printf("Aviso: arquivo config.env não encontrado: %v", err)
	}

	// Configurar aplicação
	cfg := config.Load()

	// Conectar ao banco de dados
	db, err := database.Connect(cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("Erro ao conectar com o banco de dados: %v", err)
	}
	defer db.Close()

	// Executar migrations
	if err := database.RunMigrations(cfg.DatabaseURL()); err != nil {
		log.Fatalf("Erro ao executar migrations: %v", err)
	}

	// Criar diretório de uploads se não existir
	if err := os.MkdirAll(cfg.UploadPath, 0755); err != nil {
		log.Fatalf("Erro ao criar diretório de uploads: %v", err)
	}

	// Configurar Gin
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Middleware global
	router.Use(middleware.CORS())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.ErrorHandler())

	// Configurar rotas
	api.SetupRoutes(router, db, cfg)

	// Iniciar servidor
	log.Printf("Servidor iniciando na porta %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
