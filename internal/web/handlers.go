package web

import (
	"context"
	"github.com/gin-gonic/gin"
	"localMedia/internal/db"
	"localMedia/internal/models"
	"log"
)

type Home struct {
	Repo *db.PostgresRepository
}

func (h *Home) GetFilms(ctx context.Context) ([]models.Media, error) {
	media, err := h.Repo.GetAllMedia(ctx)
	if err != nil {
		log.Println("GetFilms: Ошибка получения данных с БД:", err)
		return nil, err
	}
	return media, err
}

func (h *Home) HomePag(c *gin.Context) {
	media, err := h.GetFilms(c.Request.Context())
	if err != nil {
		log.Println("HomePag:", err)
		c.AbortWithStatusJSON(500, gin.H{
			"error": "Ошибка получения списка медиа",
		})
		return
	}

	c.HTML(200, "home.html", gin.H{"Film": media})
}
