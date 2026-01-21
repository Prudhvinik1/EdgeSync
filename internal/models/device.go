package models

import (
	"time"

	"github.com/google/uuid"
)

type Device struct {
	ID uuid.UUID `json:"id"`
	AccountID uuid.UUID `json:"account_id"`
	Name string `json:"name"`
	DeviceType string `json:"device_type"`
	PublicKey *string `json:"-"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

