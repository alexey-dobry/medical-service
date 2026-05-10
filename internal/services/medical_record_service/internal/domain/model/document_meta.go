package model

type DocumentMeta struct {
	ID       string
	Name     string
	MimeType string
	Size     int64
	RecordID string

	StorageKey string
}
