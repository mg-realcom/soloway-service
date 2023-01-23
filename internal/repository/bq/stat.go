package bq

import (
	"cloud.google.com/go/civil"
	"context"
	"time"
)

type PlacementStatDB struct {
	Clicks        int        `bigquery:"clicks"`
	Cost          int        `bigquery:"cost"`
	PlacementId   string     `bigquery:"placement_id"`
	PlacementName string     `bigquery:"placement_name"`
	Exposures     int        `bigquery:"exposures"`
	Date          civil.Date `bigquery:"date"`
	DateUpdate    time.Time  `bigquery:"date_update"`
}

func (db db) SendPlacementStatByDay(datasetId string, tableId string, stat []PlacementStatDB) (err error) {
	myDataset := db.client.Dataset(datasetId)
	table := myDataset.Table(tableId)
	u := table.Inserter()
	if err := u.Put(context.Background(), stat); err != nil {
		return err
	}
	return nil
}
