package user

import (
	"time"

	"github.com/liang21/heka/internal/domain/shared"
)

type User struct {
	ID           shared.ID
	Name         string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}
