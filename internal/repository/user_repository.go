package repository

import (
	"database/sql"
	"multi-upload-api/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByUsername busca usuário por username
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	query := `SELECT id, username, password, created_at, updated_at 
			  FROM users WHERE username = $1`

	user := &models.User{}
	err := r.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Password,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByID busca usuário por ID
func (r *UserRepository) GetByID(id int) (*models.User, error) {
	query := `SELECT id, username, password, created_at, updated_at 
			  FROM users WHERE id = $1`

	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Password,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// Create cria um novo usuário
func (r *UserRepository) Create(user *models.User) error {
	query := `INSERT INTO users (username, password) 
			  VALUES ($1, $2) 
			  RETURNING id, created_at, updated_at`

	return r.db.QueryRow(query, user.Username, user.Password).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt,
	)
}
