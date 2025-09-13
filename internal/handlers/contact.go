package handlers

import (
	"multi-upload-api/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ContactHandler struct {
	emailService *services.EmailService
}

func NewContactHandler(emailService *services.EmailService) *ContactHandler {
	return &ContactHandler{
		emailService: emailService,
	}
}

func (h *ContactHandler) SendContact(c *gin.Context) {
	var req services.ContactRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Dados inválidos",
			"details": err.Error(),
		})
		return
	}

	// Enviar email
	if err := h.emailService.SendContactEmail(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Erro ao enviar email",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Mensagem enviada com sucesso!",
	})
}
