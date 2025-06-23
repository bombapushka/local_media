package indexer

import (
	"context"
	"github.com/fsnotify/fsnotify"
	"localMedia/internal/db"
	"localMedia/internal/models"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Indexer struct {
	Repo *db.PostgresRepository
}

func (i *Indexer) WatchDir(ctx context.Context, path string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("WatchDir: Ошибка создания нового события:", err)
		return err
	}

	defer watcher.Close()

	if err := watcher.Add(path); err != nil {
		log.Println("WatchDir: Не удалось добавить путь для прослушивания события:", err)
		return err
	}

	var wg sync.WaitGroup

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Println("WatchDir: Канал event закрыт")
				return nil
			}

			if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
				continue
			}

			if event.Op&fsnotify.Create == fsnotify.Create {
				log.Println("Файл создан:", event.Name)
				wg.Add(1)
				go i.ProcessFile(&wg, event.Name, ctx)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				log.Println("WatchDir: Канал error закрыт")
				return nil
			}
			log.Println("Warcher error: %v", err)
		case <-ctx.Done():
			log.Println("WatchDir: context закрыт")
			return ctx.Err()
		}
	}
}

func (i *Indexer) IndexFiles(path string, ctx context.Context) error {

	var wg sync.WaitGroup

	dir, err := os.Open(path)
	if err != nil {
		log.Println("IndexFiles: Ошибка открытия файла:", err)
		return err
	}
	defer dir.Close()

	files, err := dir.ReadDir(-1)
	if err != nil {
		log.Println("IndexFiles: Ошибка просмотра каталога:", err)
		return err
	}

	for _, file := range files {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if !file.IsDir() {
				pathFile := filepath.Join(path, file.Name())
				wg.Add(1)
				go func() {
					if err := i.ProcessFile(&wg, pathFile, ctx); err != nil {
						log.Println("Ошибка в Горутине:", err)
					}
				}()
			}
		}
	}
	wg.Wait()

	return nil
}

func (i *Indexer) ProcessFile(wg *sync.WaitGroup, pathFile string, ctx context.Context) error {
	defer wg.Done()

	var media models.Media

	fileInfo, err := os.Stat(pathFile)
	if err != nil {
		log.Println("ProcessFile: Ошибка получения данных файла:", err)
		return err
	}

	fileType := strings.Split(fileInfo.Name(), ".")

	if len(fileType) < 2 {
		log.Println("ProcessFile: Ошибка неверный тип файла:", err)
		return err
	}

	media.NameMedia = fileInfo.Name()
	media.FilePath = pathFile
	media.TypeFile = fileType[1]
	media.Duration = 0

	ctxTime, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	media.CreateAt = time.Now()

	if err := i.Repo.CreateMedia(ctxTime, media); err != nil {
		log.Println("ProcessFile: Ошибка в отправки данных для CreateMedia:", err)
		return err
	}

	return nil
}
