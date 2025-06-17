package repository

import (
	"github.com/google/uuid"
)

// validateUUID checks if the given string is a valid UUID
func validateUUID(id string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		return ErrInvalidUUID
	}
	return nil
}
