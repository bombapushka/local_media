package routers

import (
	"context"
	"github.com/gin-gonic/gin"
	"localMedia/internal/streamer"
	"localMedia/internal/web"
)

func Register(c *gin.Engine, s *streamer.Streamer, h *web.Home, ctx context.Context) {

	c.GET("/film", h.HomePag)
	c.GET("/film/:id", s.HandlersStreamer)
	c.GET("/film/:id/stream", s.StreamHandler)
}
