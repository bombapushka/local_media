package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log"
	"time"
)

func NewPostgres(username, password, host, port, dbName string) (*pgx.Conn, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port, dbName)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	parse, err := pgx.ParseConfig(connStr)
	if err != nil {
		log.Println("Ошибка в получение строки для подключения к БД:", err)
		return nil, err
	}

	conn, err := pgx.ConnectConfig(ctx, parse)
	if err != nil {
		log.Println("Ошибка подключение к БД:", err)
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		conn.Close(ctx)
		log.Println("БД не пингуется:", err)
		return nil, err
	}

	return conn, nil
}

func ClosePostgres(conn *pgx.Conn) {
	conn.Close(context.Background())
}
