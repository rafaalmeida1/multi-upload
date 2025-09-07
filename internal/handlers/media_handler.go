package handlers

import (
	"io"
	"math"
	"multi-upload-api/internal/middleware"
	"multi-upload-api/internal/models"
	"multi-upload-api/internal/repository"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MediaHandler struct {
	mediaRepo  *repository.MediaRepository
	uploadPath string
}

func NewMediaHandler(mediaRepo *repository.MediaRepository, uploadPath string) *MediaHandler {
	return &MediaHandler{
		mediaRepo:  mediaRepo,
		uploadPath: uploadPath,
	}
}

// Upload faz upload de arquivos de mídia
func (h *MediaHandler) Upload(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	// Parse multipart form
	err := c.Request.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao processar formulário"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arquivo não encontrado"})
		return
	}
	defer file.Close()

	// Validar tipo de arquivo
	contentType := header.Header.Get("Content-Type")
	mediaType := h.getMediaType(contentType)
	if mediaType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tipo de arquivo não suportado. Apenas imagens e vídeos são permitidos",
		})
		return
	}

	// Validar tamanho do arquivo (100MB max)
	const maxSize = 100 * 1024 * 1024
	if header.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Arquivo muito grande. Tamanho máximo: 100MB",
		})
		return
	}

	// Gerar nome único para o arquivo
	fileExt := filepath.Ext(header.Filename)
	fileName := uuid.New().String() + fileExt

	// Criar diretório por data
	now := time.Now()
	dateDir := now.Format("2006/01/02")
	fullDir := filepath.Join(h.uploadPath, dateDir)

	if err := os.MkdirAll(fullDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar diretório"})
		return
	}

	// Caminho completo do arquivo
	filePath := filepath.Join(fullDir, fileName)

	// Salvar arquivo
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar arquivo"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar arquivo"})
		return
	}

	// Salvar no banco de dados
	media := &models.Media{
		UserID:       userID,
		Filename:     fileName,
		OriginalName: header.Filename,
		FilePath:     filepath.Join(dateDir, fileName),
		FileSize:     header.Size,
		MimeType:     contentType,
		MediaType:    models.MediaType(mediaType),
	}

	if err := h.mediaRepo.Create(media); err != nil {
		// Remover arquivo se falhar ao salvar no banco
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar no banco de dados"})
		return
	}

	c.JSON(http.StatusCreated, models.UploadResponse{
		Media:   *media,
		Message: "Arquivo enviado com sucesso",
	})
}

// List lista arquivos com paginação
func (h *MediaHandler) List(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	// Parâmetros de paginação
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Validar parâmetros
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Filtros
	mediaType := c.Query("type")
	orderBy := c.DefaultQuery("order_by", "sort_order")

	// Buscar dados
	medias, total, err := h.mediaRepo.List(userID, page, pageSize, mediaType, orderBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar arquivos"})
		return
	}

	// Calcular total de páginas
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	response := models.MediaListResponse{
		Data:       medias,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

// ListPublic lista arquivos publicamente para galeria (sem autenticação)
func (h *MediaHandler) ListPublic(c *gin.Context) {
	// Parâmetros de paginação
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Validar parâmetros
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Filtros
	mediaType := c.Query("type")
	orderBy := c.DefaultQuery("order_by", "sort_order")

	// Buscar dados publicamente
	medias, total, err := h.mediaRepo.ListPublic(page, pageSize, mediaType, orderBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar arquivos"})
		return
	}

	// Calcular total de páginas
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	response := models.MediaListResponse{
		Data:       medias,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

// Get busca um arquivo específico
func (h *MediaHandler) Get(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	media, err := h.mediaRepo.GetByID(id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Arquivo não encontrado"})
		return
	}

	c.JSON(http.StatusOK, media)
}

// Update atualiza um arquivo
func (h *MediaHandler) Update(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var req models.MediaUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	// Buscar mídia atual
	media, err := h.mediaRepo.GetByID(id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Arquivo não encontrado"})
		return
	}

	// Atualizar campos
	if req.SortOrder != nil {
		media.SortOrder = *req.SortOrder
	}

	if err := h.mediaRepo.Update(media); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar arquivo"})
		return
	}

	c.JSON(http.StatusOK, media)
}

// Replace substitui um arquivo existente
func (h *MediaHandler) Replace(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	// Buscar mídia atual
	oldMedia, err := h.mediaRepo.GetByID(id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Arquivo não encontrado"})
		return
	}

	// Parse multipart form
	err = c.Request.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao processar formulário"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arquivo não encontrado"})
		return
	}
	defer file.Close()

	// Validar tipo de arquivo
	contentType := header.Header.Get("Content-Type")
	mediaType := h.getMediaType(contentType)
	if mediaType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tipo de arquivo não suportado",
		})
		return
	}

	// Remover arquivo antigo
	oldFilePath := filepath.Join(h.uploadPath, oldMedia.FilePath)
	os.Remove(oldFilePath)

	// Salvar novo arquivo
	fileExt := filepath.Ext(header.Filename)
	fileName := uuid.New().String() + fileExt

	now := time.Now()
	dateDir := now.Format("2006/01/02")
	fullDir := filepath.Join(h.uploadPath, dateDir)

	if err := os.MkdirAll(fullDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar diretório"})
		return
	}

	filePath := filepath.Join(fullDir, fileName)
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar arquivo"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar arquivo"})
		return
	}

	// Atualizar no banco
	oldMedia.Filename = fileName
	oldMedia.OriginalName = header.Filename
	oldMedia.FilePath = filepath.Join(dateDir, fileName)
	oldMedia.FileSize = header.Size
	oldMedia.MimeType = contentType
	oldMedia.MediaType = models.MediaType(mediaType)

	if err := h.mediaRepo.Update(oldMedia); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar banco de dados"})
		return
	}

	c.JSON(http.StatusOK, models.UploadResponse{
		Media:   *oldMedia,
		Message: "Arquivo substituído com sucesso",
	})
}

// Delete exclui um arquivo
func (h *MediaHandler) Delete(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	// Buscar mídia
	media, err := h.mediaRepo.GetByID(id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Arquivo não encontrado"})
		return
	}

	// Remover arquivo físico
	filePath := filepath.Join(h.uploadPath, media.FilePath)
	os.Remove(filePath)

	// Remover do banco
	if err := h.mediaRepo.Delete(id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir arquivo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Arquivo excluído com sucesso"})
}

// UpdateSortOrder atualiza a ordem dos arquivos
func (h *MediaHandler) UpdateSortOrder(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	var req models.SortOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	if len(req.MediaIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lista de IDs não pode estar vazia"})
		return
	}

	if err := h.mediaRepo.UpdateSortOrders(userID, req.MediaIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar ordem"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ordem atualizada com sucesso"})
}

// Serve serve arquivos estáticos
func (h *MediaHandler) Serve(c *gin.Context) {
	filePath := c.Param("filepath")
	fullPath := filepath.Join(h.uploadPath, filePath)

	// Verificar se arquivo existe
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Arquivo não encontrado"})
		return
	}

	c.File(fullPath)
}

// getMediaType determina o tipo de mídia baseado no content-type
func (h *MediaHandler) getMediaType(contentType string) string {
	if strings.HasPrefix(contentType, "image/") {
		return "image"
	}
	if strings.HasPrefix(contentType, "video/") {
		return "video"
	}
	return ""
}
