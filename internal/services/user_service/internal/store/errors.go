package store

import "errors"

var (
	// ErrNotFound возвращается, когда запрашиваемая сущность не существует.
	ErrNotFound = errors.New("not found")
	// ErrInvalidField возвращается, когда обновление затрагивает поле, которое не разрешено.
	ErrInvalidField = errors.New("field is not allowed for update")
	// ErrAlreadyExists возвращается при нарушении ограничения уникальности.
	ErrAlreadyExists = errors.New("already exists")
)
