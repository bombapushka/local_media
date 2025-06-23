package streamer

import (
	"github.com/gin-gonic/gin"
	"localMedia/internal/db"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type Streamer struct {
	Repo    *db.PostgresRepository
	streams map[string]chan struct{}
	mu      sync.Mutex
}

func NewStreamer(repo *db.PostgresRepository) *Streamer {
	return &Streamer{
		Repo:    repo,
		streams: make(map[string]chan struct{}),
	}
}

func (s *Streamer) HandlersStreamer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	ctx := c.Request.Context()
	media, err := s.Repo.GetMediaByID(ctx, id)
	if err != nil {
		log.Println("HandlersStreamer: Ошибка получения данных из БД:", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Медиа не найдено"})
		return
	}
	
	c.HTML(http.StatusOK, "stream.html", gin.H{"Film": media})
}

func (s *Streamer) StreamHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	ctx := c.Request.Context()
	media, err := s.Repo.GetMediaByID(ctx, id)
	if err != nil {
		log.Println("StreamHandler: Ошибка получения данных из БД:", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Медиа не найдено"})
		return
	}

	ip := c.ClientIP()
	s.mu.Lock()
	if _, ok := s.streams[ip]; ok {
		s.mu.Unlock()
		log.Printf("StreamHandler: Стрим для IP %s уже активен", ip)
		c.JSON(http.StatusConflict, gin.H{"error": "Для этого IP уже есть активный стрим"})
		return
	}

	ch := make(chan struct{})
	s.streams[ip] = ch
	s.mu.Unlock()

	s.StreamMedia(media.FilePath, c, ip, ch)
}

func (s *Streamer) StreamMedia(filePath string, c *gin.Context, ip string, ch chan struct{}) {
	defer func() {
		s.mu.Lock()
		delete(s.streams, ip)
		s.mu.Unlock()
		close(ch)
		log.Printf("StreamMedia: Стрим для IP %s завершён", ip)
	}()

	filename := filepath.Base(filePath)
	if filename == "" {
		log.Println("StreamMedia: Ошибка получения имени файла")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("StreamMedia: Ошибка открытия файла %s: %v", filePath, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		log.Printf("StreamMedia: Ошибка получения информации о файле %s: %v", filePath, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Type", s.getMimeType(filename))
	c.Header("Access-Control-Allow-Origin", "*")

	select {
	case <-c.Request.Context().Done():
		log.Printf("StreamMedia: Клиент %s закрыл соединение", ip)
		return
	case <-ch:
		log.Printf("StreamMedia: Получен сигнал остановки для IP %s", ip)
		return
	default:
		http.ServeContent(c.Writer, c.Request, filename, stat.ModTime(), file)
	}
}

func (s *Streamer) getMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".avi":
		return "video/x-msvideo"
	case ".mov":
		return "video/quicktime"
	case ".mkv":
		return "video/x-matroska"
	default:
		return "video/" + strings.TrimPrefix(ext, ".")
	}
}
