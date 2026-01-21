package models

import (
	"time"

	"github.com/google/uuid"
)

type Presence struct {
	AccountID uuid.UUID `json:"account_id"`
	DeviceID  uuid.UUID `json:"device_id"`
	Status    string    `json:"status"`
	LastSeen  time.Time `json:"last_seen"`
}

type PresenceStatus string

const (
	StatusOnline  PresenceStatus = "online"
	StatusOffline PresenceStatus = "offline"
	StatusAway    PresenceStatus = "away"
)
