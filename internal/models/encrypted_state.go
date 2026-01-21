package models

import (
	"time"

	"github.com/google/uuid"
)

type EncryptedState struct {
	ID uuid.UUID `json:"id"`
	AccountID uuid.UUID `json:"account_id"`
	DeviceID uuid.UUID `json:"device_id"`
	Key string `json:"key"`
	State []byte `json:"state"`
	Nonce []byte `json:"nonce"`
	Version int64 `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}