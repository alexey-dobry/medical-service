package store

type Store interface {
	Entry() RecordRepository

	Document() DocumentRepository

	Close() error
}
