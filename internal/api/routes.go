package api

import (
	"database/sql"
	"multi-upload-api/internal/auth"
	"multi-upload-api/internal/config"
	"multi-upload-api/internal/handlers"
	"multi-upload-api/internal/middleware"
	"multi-upload-api/internal/repository"
	"multi-upload-api/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, db *sql.DB, cfg *config.Config) {
	// Inicializar serviços
	jwtService := auth.NewJWTService(cfg.JWTSecret)
	emailService := services.NewEmailService(cfg)

	// Inicializar repositórios
	userRepo := repository.NewUserRepository(db)
	mediaRepo := repository.NewMediaRepository(db)

	// Inicializar handlers
	authHandler := handlers.NewAuthHandler(userRepo, jwtService)
	mediaHandler := handlers.NewMediaHandler(mediaRepo, cfg.UploadPath)
	contactHandler := handlers.NewContactHandler(emailService)

	// Rotas públicas
	public := router.Group("/api/v1")
	{
		// Autenticação
		public.POST("/login", authHandler.Login)

		// Contato
		public.POST("/contact", contactHandler.SendContact)

		// Servir arquivos (público para visualização)
		public.GET("/files/*filepath", mediaHandler.Serve)

		// Galeria pública de mídias
		public.GET("/gallery", mediaHandler.ListPublic)
	}

	// Rotas protegidas
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(jwtService))
	{
		// Usuário
		protected.GET("/me", authHandler.Me)

		// Mídia
		media := protected.Group("/media")
		{
			media.POST("/upload", mediaHandler.Upload)
			media.GET("", mediaHandler.List)
			media.GET("/:id", mediaHandler.Get)
			media.PUT("/:id", mediaHandler.Update)
			media.PUT("/:id/replace", mediaHandler.Replace)
			media.DELETE("/:id", mediaHandler.Delete)
			media.POST("/sort", mediaHandler.UpdateSortOrder)
		}
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "API funcionando corretamente",
		})
	})
}
