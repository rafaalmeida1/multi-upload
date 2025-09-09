package repository

import (
	"database/sql"
	"fmt"
	"multi-upload-api/internal/models"
	"strings"
)

type MediaRepository struct {
	db *sql.DB
}

func NewMediaRepository(db *sql.DB) *MediaRepository {
	return &MediaRepository{db: db}
}

// Create cria um novo registro de mídia
func (r *MediaRepository) Create(media *models.Media) error {
	query := `INSERT INTO media (user_id, filename, original_name, file_path, file_size, mime_type, media_type, sort_order)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, COALESCE((SELECT MAX(sort_order) + 1 FROM media WHERE user_id = $1), 1))
			  RETURNING id, sort_order, created_at, updated_at`

	return r.db.QueryRow(query, media.UserID, media.Filename, media.OriginalName,
		media.FilePath, media.FileSize, media.MimeType, media.MediaType).Scan(
		&media.ID, &media.SortOrder, &media.CreatedAt, &media.UpdatedAt,
	)
}

// GetByID busca mídia por ID
func (r *MediaRepository) GetByID(id int, userID int) (*models.Media, error) {
	query := `SELECT id, user_id, filename, original_name, file_path, file_size, 
			  mime_type, media_type, sort_order, created_at, updated_at
			  FROM media WHERE id = $1 AND user_id = $2`

	media := &models.Media{}
	err := r.db.QueryRow(query, id, userID).Scan(
		&media.ID, &media.UserID, &media.Filename, &media.OriginalName,
		&media.FilePath, &media.FileSize, &media.MimeType, &media.MediaType,
		&media.SortOrder, &media.CreatedAt, &media.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return media, nil
}

// List lista mídias com paginação e filtros
func (r *MediaRepository) List(userID int, page, pageSize int, mediaType string, orderBy string) ([]models.Media, int, error) {
	offset := (page - 1) * pageSize

	// Construir query base
	baseQuery := `FROM media WHERE user_id = $1`
	args := []interface{}{userID}
	argCount := 1

	// Adicionar filtro de tipo de mídia se especificado
	if mediaType != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND media_type = $%d", argCount)
		args = append(args, mediaType)
	}

	// Construir ORDER BY (padrão: mais novos primeiro)
	orderClause := " ORDER BY created_at DESC"
	switch strings.ToLower(orderBy) {
	case "sort_order":
		orderClause = " ORDER BY sort_order ASC"
	case "created_at_desc":
		orderClause = " ORDER BY created_at DESC"
	case "created_at_asc":
		orderClause = " ORDER BY created_at ASC"
	case "filename_asc":
		orderClause = " ORDER BY filename ASC"
	case "filename_desc":
		orderClause = " ORDER BY filename DESC"
	case "size_asc":
		orderClause = " ORDER BY file_size ASC"
	case "size_desc":
		orderClause = " ORDER BY file_size DESC"
	}

	// Query para contar total
	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Query para buscar dados
	dataQuery := `SELECT id, user_id, filename, original_name, file_path, file_size,
				  mime_type, media_type, sort_order, created_at, updated_at ` +
		baseQuery + orderClause + fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount+1, argCount+2)

	args = append(args, pageSize, offset)

	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var medias []models.Media
	for rows.Next() {
		var media models.Media
		if err := rows.Scan(
			&media.ID, &media.UserID, &media.Filename, &media.OriginalName,
			&media.FilePath, &media.FileSize, &media.MimeType, &media.MediaType,
			&media.SortOrder, &media.CreatedAt, &media.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		medias = append(medias, media)
	}

	return medias, total, nil
}

// ListPublic lista todas as mídias publicamente (para galeria)
func (r *MediaRepository) ListPublic(page, pageSize int, mediaType string, orderBy string) ([]models.Media, int, error) {
	offset := (page - 1) * pageSize

	// Construir query base (sem filtro de usuário)
	baseQuery := `FROM media WHERE 1=1`
	args := []interface{}{}
	argCount := 0

	// Adicionar filtro de tipo de mídia se especificado
	if mediaType != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND media_type = $%d", argCount)
		args = append(args, mediaType)
	}

	// Construir ORDER BY (padrão: mais novos primeiro)
	orderClause := " ORDER BY created_at DESC"
	switch strings.ToLower(orderBy) {
	case "sort_order":
		orderClause = " ORDER BY sort_order ASC"
	case "created_at_desc":
		orderClause = " ORDER BY created_at DESC"
	case "created_at_asc":
		orderClause = " ORDER BY created_at ASC"
	case "filename_asc":
		orderClause = " ORDER BY filename ASC"
	case "filename_desc":
		orderClause = " ORDER BY filename DESC"
	case "size_asc":
		orderClause = " ORDER BY file_size ASC"
	case "size_desc":
		orderClause = " ORDER BY file_size DESC"
	}

	// Query para contar total
	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Query para buscar dados
	dataQuery := `SELECT id, user_id, filename, original_name, file_path, file_size,
				  mime_type, media_type, sort_order, created_at, updated_at ` +
		baseQuery + orderClause + fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount+1, argCount+2)

	args = append(args, pageSize, offset)

	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var medias []models.Media
	for rows.Next() {
		var media models.Media
		if err := rows.Scan(
			&media.ID, &media.UserID, &media.Filename, &media.OriginalName,
			&media.FilePath, &media.FileSize, &media.MimeType, &media.MediaType,
			&media.SortOrder, &media.CreatedAt, &media.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		medias = append(medias, media)
	}

	return medias, total, nil
}

// Update atualiza uma mídia
func (r *MediaRepository) Update(media *models.Media) error {
	query := `UPDATE media SET sort_order = $1, updated_at = CURRENT_TIMESTAMP
			  WHERE id = $2 AND user_id = $3`

	_, err := r.db.Exec(query, media.SortOrder, media.ID, media.UserID)
	return err
}

// Delete exclui uma mídia
func (r *MediaRepository) Delete(id int, userID int) error {
	query := `DELETE FROM media WHERE id = $1 AND user_id = $2`
	_, err := r.db.Exec(query, id, userID)
	return err
}

// UpdateSortOrders atualiza a ordem de múltiplas mídias
func (r *MediaRepository) UpdateSortOrders(userID int, mediaIDs []int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `UPDATE media SET sort_order = $1, updated_at = CURRENT_TIMESTAMP 
			  WHERE id = $2 AND user_id = $3`

	for i, mediaID := range mediaIDs {
		if _, err := tx.Exec(query, i+1, mediaID, userID); err != nil {
			return err
		}
	}

	return tx.Commit()
}
