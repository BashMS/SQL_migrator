package domain

import (
	"time"
)

// Migration.
type Migration struct {
	Version   uint64    `json:"version"`
	Name      string    `json:"name"`
	IsApplied bool      `json:"isApplied"`
	UpdateAt  time.Time `json:"updateAt"`
}
