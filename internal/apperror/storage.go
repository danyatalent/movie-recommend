package apperror

import (
	"errors"
)

var (
	ErrEntityNotFound       = errors.New("entity not found")
	ErrEntityExists         = errors.New("entity exists")
	ErrConstraintUniqueCode = "23505"
)
