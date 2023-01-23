package services

import (
	"solowayStat/internal/repository"
	"solowayStat/pkg/solowaysdk"
	"time"
)

type Info interface {
	Login() (err error)
	Whoami() (err error)
	GetPlacements() (placements solowaysdk.PlacementsInfo, err error)
}

type Stat interface {
	Login() (err error)
	GetPlacementsStat(placements solowaysdk.PlacementsInfo, startDate time.Time, stopDate time.Time, withArchived bool) (err error)
	GetPlacementStatByDay(placementGuid string, startDate time.Time, stopDate time.Time) (stat solowaysdk.PlacementsStatByDay, err error)
	SendOnePlacementStatByDay(datasetId string, tableId string, stat solowaysdk.PlacementsStatByDay, placements solowaysdk.PlacementsInfo) (err error)
	SendAnyPlacementStatByDay(datasetId string, tableId string, stat []solowaysdk.PlacementsStatByDay, placements solowaysdk.PlacementsInfo) (err error)
}

type Service struct {
	Info
	Stat
}

func NewService(solowayClient solowaysdk.Client, repo repository.Repository) *Service {
	return &Service{
		Info: NewInfoService(solowayClient, repo),
		Stat: NewStatService(solowayClient, repo),
	}
}
