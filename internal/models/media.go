package models

import "time"

type Media struct {
	ID        int       `json:"id"`
	NameMedia string    `json:"name_media"`
	FilePath  string    `json:"file_path"`
	TypeFile  string    `json:"type_file"`
	Duration  int       `json:"duration"`
	CreateAt  time.Time `json:"create_at"`
}
