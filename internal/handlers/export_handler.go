package handlers

import (
	"crypto/subtle"
	"fmt"
	"multi-upload-api/internal/repository"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ExportHandler expõe endpoints para extrair dados do sistema antigo
// (manifest JSON + arquivos), pensado para migração em massa.
type ExportHandler struct {
	mediaRepo  *repository.MediaRepository
	uploadPath string
	token      string
}

func NewExportHandler(mediaRepo *repository.MediaRepository, uploadPath, token string) *ExportHandler {
	return &ExportHandler{
		mediaRepo:  mediaRepo,
		uploadPath: uploadPath,
		token:      token,
	}
}

// authorize garante que, se EXPORT_TOKEN foi configurado, a requisição
// traga o mesmo valor via header X-Export-Token ou query string `token`.
// Se nenhum token estiver configurado, o endpoint fica público (útil pra LAN).
func (h *ExportHandler) authorize(c *gin.Context) bool {
	if h.token == "" {
		return true
	}
	provided := c.GetHeader("X-Export-Token")
	if provided == "" {
		provided = c.Query("token")
	}
	if subtle.ConstantTimeCompare([]byte(provided), []byte(h.token)) == 1 {
		return true
	}
	c.JSON(http.StatusUnauthorized, gin.H{"error": "token de export inválido"})
	return false
}

// Manifest devolve TODOS os registros de media em uma única resposta JSON.
// Sem paginação — usado para snapshot completo do banco para migração.
func (h *ExportHandler) Manifest(c *gin.Context) {
	if !h.authorize(c) {
		return
	}

	rows, err := h.mediaRepo.ListAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "erro ao buscar manifest",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"version":     1,
		"total":       len(rows),
		"upload_path": h.uploadPath,
		"items":       rows,
	})
}

// Stats devolve estatísticas rápidas (contagem e tamanho total) para o
// usuário ter ideia de quantos arquivos / quantos GBs vai migrar.
func (h *ExportHandler) Stats(c *gin.Context) {
	if !h.authorize(c) {
		return
	}

	count, totalBytes, byType, err := h.mediaRepo.Stats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "erro ao calcular estatísticas",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_files":     count,
		"total_bytes":     totalBytes,
		"total_human":     humanBytes(totalBytes),
		"by_type":         byType,
	})
}

// File serve um arquivo individual pelo ID (atalho seguro para o
// migrator, que prefere buscar por ID a montar paths manualmente).
func (h *ExportHandler) File(c *gin.Context) {
	if !h.authorize(c) {
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	media, err := h.mediaRepo.GetByIDPublic(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "arquivo não encontrado"})
		return
	}

	fullPath := filepath.Join(h.uploadPath, media.FilePath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "arquivo físico ausente"})
		return
	}

	c.Header("Content-Type", media.MimeType)
	c.Header("X-Original-Name", media.OriginalName)
	c.Header("X-File-Path", media.FilePath)
	c.File(fullPath)
}

func humanBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
