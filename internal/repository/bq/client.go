package bq

import (
	"cloud.google.com/go/bigquery"
)

type db struct {
	client bigquery.Client
}

func NewBigQueryDB(client bigquery.Client) *db {
	return &db{
		client: client,
	}
}
