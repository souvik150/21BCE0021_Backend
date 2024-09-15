package schemas

import "github.com/google/uuid"

type FileCache struct {
	ID       uuid.UUID `json:"id"`
	URL      string    `json:"url"`
	FileName string    `json:"file_name"`
	Size     int64     `json:"size"`
	FileType string    `json:"file_type"`
}
