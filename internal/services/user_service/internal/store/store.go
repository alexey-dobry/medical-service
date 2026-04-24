package store

type Store interface {
	User() UserRepository

	Meta() MetaRepository

	Search() SearchRepository

	Photos() PhotosRepository

	Close() error
}
