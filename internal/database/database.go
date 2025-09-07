package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// Connect conecta ao banco de dados PostgreSQL
func Connect(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir conexão: %w", err)
	}

	// Testar a conexão
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("erro ao conectar ao banco: %w", err)
	}

	// Configurar pool de conexões
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	return db, nil
}

// RunMigrations executa as migrations do banco de dados
func RunMigrations(databaseURL string) error {
	db, err := Connect(databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS media (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			filename VARCHAR(255) NOT NULL,
			original_name VARCHAR(255) NOT NULL,
			file_path VARCHAR(500) NOT NULL,
			file_size BIGINT NOT NULL,
			mime_type VARCHAR(100) NOT NULL,
			media_type VARCHAR(10) NOT NULL CHECK (media_type IN ('image', 'video')),
			sort_order INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_media_user_id ON media(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_media_sort_order ON media(sort_order)`,
		`CREATE INDEX IF NOT EXISTS idx_media_created_at ON media(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_media_media_type ON media(media_type)`,
		`CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ language 'plpgsql'`,
		`DROP TRIGGER IF EXISTS update_users_updated_at ON users`,
		`CREATE TRIGGER update_users_updated_at 
			BEFORE UPDATE ON users 
			FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,
		`DROP TRIGGER IF EXISTS update_media_updated_at ON media`,
		`CREATE TRIGGER update_media_updated_at 
			BEFORE UPDATE ON media 
			FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("erro na migration %d: %w", i+1, err)
		}
	}

	return nil
}
