package db

import (
	"context"
	"github.com/jackc/pgx/v5"
	"localMedia/internal/models"
	"log"
)

type PostgresRepository struct {
	Conn *pgx.Conn
}

type MediaRepository interface {
	CreateMedia(ctx context.Context, media models.Media) error
	GetMediaByID(ctx context.Context, id string) (models.Media, error)
	GetAllMedia(ctx context.Context) ([]models.Media, error)
}

func (pp *PostgresRepository) CreateMedia(ctx context.Context, media models.Media) error {
	query := "INSERT INTO media (name_media, file_path, type_file, duration, create_at) VALUES ($1, $2, $3, $4, $5)"

	_, err := pp.Conn.Exec(ctx, query, media.NameMedia, media.FilePath, media.TypeFile, media.Duration, media.CreateAt)
	if err != nil {
		log.Println("CreateMedia: Ошибка добавления данных в таблицу:", err)
		return err
	}

	return nil
}

func (pp *PostgresRepository) GetMediaByID(ctx context.Context, id int) (models.Media, error) {
	query := "SELECT * FROM media WHERE id = $1"

	rows := pp.Conn.QueryRow(ctx, query, id)

	var media models.Media

	if err := rows.Scan(&media.ID, &media.NameMedia, &media.FilePath, &media.TypeFile, &media.Duration, &media.CreateAt); err != nil {
		if err == pgx.ErrNoRows {
			log.Println("GetMediaByID: Фильм с ID", id, "не найден")
			return models.Media{}, pgx.ErrNoRows
		}
		log.Println("GetMediaByID: Ошибка получения данных с БД:", err, pgx.ErrNoRows)
		return models.Media{}, err
	}

	return media, nil
}

func (pp *PostgresRepository) GetAllMedia(ctx context.Context) ([]models.Media, error) {
	query := "SELECT * FROM media"

	rows, err := pp.Conn.Query(ctx, query)
	if err != nil {
		log.Println("GetAllMedia: Ошибка получение таблицы:", err)
		return nil, err
	}
	defer rows.Close()

	var medias []models.Media

	for rows.Next() {
		var media models.Media
		if err := rows.Scan(&media.ID, &media.NameMedia, &media.FilePath, &media.TypeFile, &media.Duration, &media.CreateAt); err != nil {
			log.Println("GetAllMedia: Ошибка получения данных с БД:", err)
			return nil, err
		}
		medias = append(medias, media)
	}

	return medias, nil
}
