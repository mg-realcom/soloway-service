package repository

import (
	"cloud.google.com/go/bigquery"
	"solowayStat/internal/repository/bq"
)

type Stat interface {
	SendPlacementStatByDay(datasetId string, tableId string, stat []bq.PlacementStatDB) (err error)
}

type Repository struct {
	Stat
}

func NewRepository(clientBQ bigquery.Client) *Repository {
	return &Repository{
		Stat: bq.NewBigQueryDB(clientBQ)}
}
