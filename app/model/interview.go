package model

import "time"

type Interview struct {
	ID            int       `json:"id"`
	ApplicationID int       `json:"application_id"`
	ScheduledAt   time.Time `json:"scheduled_at"`
	Location      string    `json:"location"`
	Notes         string    `json:"notes"`
}
