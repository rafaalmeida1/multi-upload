package middleware

import (
	"multi-upload-api/internal/auth"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware middleware para autenticação JWT
func AuthMiddleware(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token de autorização requerido",
			})
			c.Abort()
			return
		}

		// Verificar se o header tem o formato "Bearer <token>"
		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Formato de token inválido",
			})
			c.Abort()
			return
		}

		token := bearerToken[1]
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token inválido",
			})
			c.Abort()
			return
		}

		// Adicionar informações do usuário ao contexto
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

// GetUserID obtém o ID do usuário do contexto
func GetUserID(c *gin.Context) (int, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	return userID.(int), true
}
