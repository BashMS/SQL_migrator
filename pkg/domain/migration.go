package domain

import (
	"time"
)

//Migration.
type Migration struct {
	Version   uint64    `json:"version"`
	Name      string    `json:"name"`
	IsApplied bool      `json:"is_applied"`
	UpdateAt  time.Time `json:"update_at"`
}
