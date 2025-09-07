package handlers

import (
	"database/sql"
	"multi-upload-api/internal/auth"
	"multi-upload-api/internal/models"
	"multi-upload-api/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userRepo   *repository.UserRepository
	jwtService *auth.JWTService
}

func NewAuthHandler(userRepo *repository.UserRepository, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

// Login autentica o usuário e retorna um token JWT
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Dados inválidos",
			"details": err.Error(),
		})
		return
	}

	// Buscar usuário
	user, err := h.userRepo.GetByUsername(req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Credenciais inválidas",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro interno do servidor",
		})
		return
	}

	// Verificar senha
	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Credenciais inválidas",
		})
		return
	}

	// Gerar token
	token, err := h.jwtService.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao gerar token",
		})
		return
	}

	response := models.LoginResponse{
		Token: token,
		User:  *user,
	}

	c.JSON(http.StatusOK, response)
}

// Me retorna informações do usuário autenticado
func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Usuário não autenticado",
		})
		return
	}

	user, err := h.userRepo.GetByID(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao buscar usuário",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}
