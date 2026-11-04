package repository

import "errors"

var (
	// ErrNotFound возвращается, когда запрошенная запись не существует.
	ErrNotFound = errors.New("entry not found")

	// ErrAlreadyExists возвращается при нарушении ограничения уникальности.
	ErrAlreadyExists = errors.New("entry already exists")
)
