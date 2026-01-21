package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        string    `json:"id"`
	AccountID uuid.UUID `json:"account_id"`
	DeviceID  uuid.UUID `json:"device_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
