package models

import (
	"time"

	"github.com/google/uuid"
)

type SyncEvent struct {
	ID uuid.UUID `json:"id"`
	AccountID uuid.UUID `json:"account_id"`
	DeviceID uuid.UUID `json:"device_id"`
	EventType string `json:"event_type"`
	StateKey string `json:"state_key"`
	SequenceNumber int64 `json:"sequence_number"`
	Payload []byte `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}