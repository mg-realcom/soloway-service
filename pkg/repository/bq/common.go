package bq

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/googleapi"
)

func CreateTable(ctx context.Context, schemaDTO any, table *bigquery.Table, fieldPartition *string,
	fieldClustering *[]string,
) error {
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
		return fmt.Errorf("can't create table: %w", err)
	}

	return nil
}

func SendFromCS(ctx context.Context, schemaDTO any, table *bigquery.Table, bucket string, object string) error {
	if table == nil {
		return errors.New("arg table can't be nil")
	}

	schema, err := bigquery.InferSchema(schemaDTO)
	if err != nil {
		return fmt.Errorf("bigquery.InferSchema: %w", err)
	}

	filePath := strings.Split(object, "/")
	gcsRef := bigquery.NewGCSReference(fmt.Sprintf("gs://%s/%s", bucket, filePath[len(filePath)-1]))

	gcsRef.SourceFormat = bigquery.JSON

	gcsRef.Schema = schema
	loader := table.LoaderFrom(gcsRef)
	loader.CreateDisposition = bigquery.CreateNever
	loader.WriteDisposition = bigquery.WriteAppend

	job, err := loader.Run(ctx)
	if err != nil {
		return fmt.Errorf("loader error: %w", err)
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
	if table == nil {
		return errors.New("arg table can't be nil")
	}

	if _, err := table.Metadata(ctx); err != nil {
		if e, ok := err.(*googleapi.Error); ok {
			if e.Code == http.StatusNotFound {
				return errors.New("dataset or table not found")
			}
		}
	}

	return nil
}

func TruncateTable(ctx context.Context, bqClient *bigquery.Client, table *bigquery.Table) error {
	q := bqClient.Query(`TRUNCATE TABLE` + fmt.Sprintf(" %s.%s", table.DatasetID, table.TableID))
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

func DeleteByArrayDateColumn(ctx context.Context, bqClient *bigquery.Client, table *bigquery.Table, dateColumn string, dates []time.Time) error {
	datesStr := timePrepare(dates)
	q := bqClient.Query(fmt.Sprintf("DELETE %s.%s ", table.DatasetID, table.TableID) + fmt.Sprintf("WHERE %s IN (%s)", dateColumn, datesStr))

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

func Query(ctx context.Context, bqClient *bigquery.Client, query string) error {
	q := bqClient.Query(query)

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

func timePrepare(t []time.Time) string {
	var result []string
	timeMap := make(map[string]bool)

	for _, v := range t {
		dateString := v.Format("2006-01-02")
		if _, ok := timeMap[dateString]; !ok {
			result = append(result, fmt.Sprintf("'%s'", dateString))
			timeMap[dateString] = true
		}
	}

	return strings.Join(result, ",")
}
