package db

import (
	"context"
	"github.com/jackc/pgx/v5"
	"log"
)

type PostgresMigrator struct {
	Conn *pgx.Conn
}

type Migrator interface {
	CreateMedia(cxt context.Context) error
}

func (pm *PostgresMigrator) CreateMedia(ctx context.Context) error {
	query := `CREATE TABLE IF NOT EXISTS media(
				id SERIAL PRIMARY KEY,
                name_media VARCHAR(100) NOT NULL,
    			file_path VARCHAR(100) NOT NULL,
    			type_file VARCHAR(15) NOT NULL,
    			duration INTEGER,
    			create_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)`

	_, err := pm.Conn.Exec(ctx, query)
	if err != nil {
		log.Printf("Ошибка создание таблицы media: %v", err)
		return err
	}

	return nil
}
