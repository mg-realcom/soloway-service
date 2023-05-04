package bq

import (
	"cloud.google.com/go/bigquery"
	"context"
	"errors"
	"fmt"
	"google.golang.org/api/googleapi"
	"net/http"
	"time"
)

func CreateTable(ctx context.Context, schemaDTO any, table *bigquery.Table, fieldPartition *string,
	fieldClustering *[]string) error {
	schema, err := bigquery.InferSchema(schemaDTO)
	if err != nil {
		return fmt.Errorf("bigquery.InferSchema: %w", err)
	}

	metadata := &bigquery.TableMetadata{
		Schema: schema,
	}

	if fieldPartition != nil {
		partition := bigquery.TimePartitioning{
			Type:  bigquery.DayPartitioningType,
			Field: *fieldPartition,
		}
		metadata.TimePartitioning = &partition
	}

	if fieldClustering != nil {
		clustering := bigquery.Clustering{
			Fields: *fieldClustering,
		}
		metadata.Clustering = &clustering
	}

	if err := table.Create(ctx, metadata); err != nil {
		return err
	}

	return nil
}

func DeleteByDateColumn(ctx context.Context, bqClient *bigquery.Client, table *bigquery.Table, client string, dateColumn string, dateStart time.Time, dateFinish time.Time) error {
	q := bqClient.Query(fmt.Sprintf("DELETE %s.%s ", table.DatasetID, table.TableID) + fmt.Sprintf("WHERE %s >= '%s' AND %s <= '%s' AND client_name = '%s'", dateColumn, dateStart.Format(time.DateOnly),
		dateColumn, dateFinish.Format(time.DateOnly), client))

	job, err := q.Run(ctx)
	if err != nil {
		return err
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return err
	}

	if err := status.Err(); err != nil {
		return err
	}

	return nil
}

func TableExists(ctx context.Context, table *bigquery.Table) error {
	if _, err := table.Metadata(ctx); err != nil {
		if e, ok := err.(*googleapi.Error); ok {
			if e.Code == http.StatusNotFound {
				return errors.New("dataset or table not found")
			}
		}
	}

	return nil
}
