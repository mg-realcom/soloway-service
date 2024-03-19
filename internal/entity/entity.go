package entity

import "time"

type StatPlacement struct {
	Clicks        int
	Cost          int
	PlacementID   string
	PlacementName string
	Exposures     int
	Date          time.Time
}

type User struct {
	Name     string
	Login    string
	Password string
}

type Placement struct {
	GUID string
	Name string
	ID   string
}
