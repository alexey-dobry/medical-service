package model

import "time"

type Photo struct {
	ID       string
	Name     string
	MimeType string
	Size     int64
	UserID   string

	StorageKey string
}

type StorageObjectInfo struct {
	Size         int64
	ContentType  string
	LastModified time.Time
}
