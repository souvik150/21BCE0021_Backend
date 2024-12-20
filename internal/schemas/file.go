package schemas

import (
	"time"

	"github.com/google/uuid"
)

type UploadedFileResponse struct {
	FileNames []string `json:"file_names"`
}

type FileResponse struct {
	ID            uuid.UUID `json:"id"`
	FileName      string    `json:"file_name"`
	Size          int64     `json:"size"`
	FileType      string    `json:"file_type"`
	CreatedAt     string    `json:"created_at"`
	UpdatedAt     string    `json:"updated_at"`
	AccessedAt    string    `json:"accessed_at"`
	DeletedStatus bool      `json:"deleted_status"`
}

type FilesResponse struct {
	Files []FileResponse `json:"files"`
}


type RenameFileResponse struct {
	ID        string    `json:"id"`
	FileName  string    `json:"fileName"`
	FileType	string    `json:"fileType"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
