package services

import (
	"cloud.google.com/go/civil"
	"solowayStat/internal/repository"
	"solowayStat/internal/repository/bq"
	"solowayStat/pkg/solowaysdk"
	"time"
)

type StatService struct {
	client solowaysdk.Client
	repo   repository.Stat
}

func NewStatService(client solowaysdk.Client, repo repository.Repository) *StatService {
	return &StatService{
		client: client,
		repo:   repo,
	}
}

func (s *StatService) Login() (err error) {
	err = s.client.Login()
	if err != nil {
		return err
	}
	return nil
}
func (s *StatService) GetPlacementsStat(placements solowaysdk.PlacementsInfo, startDate time.Time, stopDate time.Time, withArchived bool) (err error) {
	placementIds := make([]string, 0, len(placements.List))
	for _, place := range placements.List {
		placementIds = append(placementIds, place.Id)
	}
	err = s.client.GetPlacementsStat(placementIds, startDate, stopDate, withArchived)
	if err != nil {
		return err
	}
	return nil
}
func (s *StatService) GetPlacementStatByDay(placementGuid string, startDate time.Time, stopDate time.Time) (stat solowaysdk.PlacementsStatByDay, err error) {
	stat, err = s.client.GetPlacementStatByDay(placementGuid, startDate, stopDate)
	if err != nil {
		return stat, err
	}
	return stat, nil
}
func (s *StatService) SendOnePlacementStatByDay(datasetId string, tableId string, stat solowaysdk.PlacementsStatByDay, placements solowaysdk.PlacementsInfo) (err error) {
	var resultStat []bq.PlacementStatDB
	places := placements.ToMap()
	dateUpdate := time.Now()
	for _, place := range stat.List {
		date, err := civil.ParseDate(place.Date)
		if err != nil {
			return err
		}
		item := bq.PlacementStatDB{
			Clicks:        place.Clicks,
			Cost:          place.Cost,
			PlacementId:   place.PlacementId,
			PlacementName: places[place.PlacementId].Doc.Name,
			Exposures:     place.Exposures,
			Date:          date,
			DateUpdate:    dateUpdate,
		}
		resultStat = append(resultStat, item)
	}

	err = s.repo.SendPlacementStatByDay(datasetId, tableId, resultStat)
	if err != nil {
		return err
	}
	return nil
}
func (s *StatService) SendAnyPlacementStatByDay(datasetId string, tableId string, stat []solowaysdk.PlacementsStatByDay, placements solowaysdk.PlacementsInfo) (err error) {
	var resultStat []bq.PlacementStatDB
	places := placements.ToMap()
	dateUpdate := time.Now()
	for _, place := range stat {
		for _, row := range place.List {
			date, err := civil.ParseDate(row.Date)
			if err != nil {
				return err
			}
			item := bq.PlacementStatDB{
				Clicks:        row.Clicks,
				Cost:          row.Cost,
				PlacementId:   row.PlacementId,
				PlacementName: places[row.PlacementId].Doc.Name,
				Exposures:     row.Exposures,
				Date:          date,
				DateUpdate:    dateUpdate,
			}
			resultStat = append(resultStat, item)
		}
	}

	err = s.repo.SendPlacementStatByDay(datasetId, tableId, resultStat)
	if err != nil {
		return err
	}
	return nil
}
