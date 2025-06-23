package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"localMedia/internal/db"
	"localMedia/internal/indexer"
	routers "localMedia/internal/router"
	"localMedia/internal/streamer"
	"localMedia/internal/web"
	"log"
	"path/filepath"
	"time"
)

func main() {
	path := filepath.Join(".", "media")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := db.NewPostgres("postgres", "mysecretpassword", "localhost", "5432", "local_db")
	if err != nil {
		log.Fatalln("Ошибка запуска БД")
	}

	defer db.ClosePostgres(conn)

	migrator := &db.PostgresMigrator{Conn: conn}

	if err := migrator.CreateMedia(ctx); err != nil {
		log.Println(err)
	}

	repo := &db.PostgresRepository{Conn: conn}

	idx := &indexer.Indexer{Repo: repo}

	if err := idx.IndexFiles(path, ctx); err != nil {
		log.Println("Ошибка добавление в БД на старте")
	}

	ctxW := context.Background()

	go idx.WatchDir(ctxW, path)

	stream := streamer.NewStreamer(repo)

	home := &web.Home{Repo: repo}

	router := gin.Default()

	routers.Register(router, stream, home, ctx)

	router.LoadHTMLGlob(filepath.Join(".", "internal", "web", "templane", "*"))

	if err := router.Run(":8080"); err != nil {
		log.Fatalln("Ошибка запуска сервера")
	}
}
