package registry

import (
	"fmt"

	"github.com/google/uuid"
)

type NotFoundError struct {
	Kind string
	ID   uuid.UUID
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("Item [%s:%s] not found", e.Kind, e.ID.String())
}

func NewNotFoundError(kind string, id uuid.UUID) NotFoundError {
	return NotFoundError{
		Kind: kind,
		ID:   id,
	}
}
