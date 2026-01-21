package models

import (
	"time"
	"github.com/google/uuid"
)

type Account struct {
	ID uuid.UUID `json:"id"`
	Email string `json:"email"`
	PasswordHash string `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`
}
