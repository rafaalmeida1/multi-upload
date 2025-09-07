package models

import (
	"time"
)

type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
)

type Media struct {
	ID           int       `json:"id" db:"id"`
	UserID       int       `json:"user_id" db:"user_id"`
	Filename     string    `json:"filename" db:"filename"`
	OriginalName string    `json:"original_name" db:"original_name"`
	FilePath     string    `json:"file_path" db:"file_path"`
	FileSize     int64     `json:"file_size" db:"file_size"`
	MimeType     string    `json:"mime_type" db:"mime_type"`
	MediaType    MediaType `json:"media_type" db:"media_type"`
	SortOrder    int       `json:"sort_order" db:"sort_order"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type MediaListResponse struct {
	Data       []Media `json:"data"`
	Total      int     `json:"total"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
	TotalPages int     `json:"total_pages"`
}

type UploadResponse struct {
	Media   Media  `json:"media"`
	Message string `json:"message"`
}

type MediaUpdateRequest struct {
	SortOrder *int `json:"sort_order,omitempty"`
}

type SortOrderRequest struct {
	MediaIDs []int `json:"media_ids" binding:"required"`
}
