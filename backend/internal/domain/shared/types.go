package shared

import (
	"fmt"

	"github.com/google/uuid"
)

// ID is a strongly-typed identifier backed by UUID v4.
type ID string

// NewID generates a new random UUID-based ID.
func NewID() ID {
	return ID(uuid.New().String())
}

// String returns the underlying string value of the ID.
func (id ID) String() string {
	return string(id)
}

// IsEmpty reports whether the ID is the zero value.
func (id ID) IsEmpty() bool {
	return string(id) == ""
}

// ParseID validates that s is a well-formed UUID string and returns it as an ID.
func ParseID(s string) (ID, error) {
	if _, err := uuid.Parse(s); err != nil {
		return "", fmt.Errorf("invalid ID format: %w", err)
	}
	return ID(s), nil
}
