package entity

import (
	"time"
)

type StatPlacement struct {
	Clicks        int
	Cost          int
	PlacementID   string
	PlacementName string
	Exposures     int
	Date          time.Time
}
