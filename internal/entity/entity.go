package entity

import "time"

type File struct {
	Name string
	Path string
	Date time.Time
}

type StatPlacement struct {
	Client        string
	Clicks        int
	Cost          int
	PlacementID   string
	PlacementName string
	Exposures     int
	Date          string
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
