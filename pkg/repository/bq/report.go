package bq

import (
	"context"
	"fmt"
	"time"

	repository "soloway/internal/repository/report"
)

func (c *Client) TableExists(ctx context.Context, destination repository.Destination) error {
	dataset := c.db.Dataset(destination.DatasetID)
	table := dataset.Table(destination.TableID)

	err := TableExists(ctx, table)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) CreateTable(ctx context.Context, destination repository.Destination, schema any) error {
	dataset := c.db.Dataset(destination.DatasetID)
	table := dataset.Table(destination.TableID)

	err := CreateTable(ctx, schema, table, nil, nil)
	if err != nil {
		return fmt.Errorf("createTable: %w", err)
	}

	return nil
}

func (c *Client) TruncateTable(ctx context.Context, destination repository.Destination) error {
	dataset := c.db.Dataset(destination.DatasetID)
	table := dataset.Table(destination.TableID)

	err := TruncateTable(ctx, c.db, table)
	if err != nil {
		return fmt.Errorf("TruncateTable: %w", err)
	}

	return nil
}

func (c *Client) ImportFromCS(ctx context.Context, destination repository.Destination, bucket string, object string, schema any) error {
	dataset := c.db.Dataset(destination.DatasetID)
	table := dataset.Table(destination.TableID)

	err := SendFromCS(ctx, schema, table, bucket, object)
	if err != nil {
		return fmt.Errorf("SendFromCS: %w", err)
	}

	return nil
}

func (c *Client) DeleteByDateColumn(ctx context.Context, destination repository.Destination, client string, dateColumn string, dateStart, dateFinish time.Time) error {
	query := fmt.Sprintf("DELETE `%s.%s` ", destination.DatasetID, destination.TableID) + fmt.Sprintf("WHERE %s >= '%s' AND %s <= '%s' AND client_name = '%s'", dateColumn, dateStart.Format(time.DateOnly), dateColumn, dateFinish.Format(time.DateOnly), client)

	err := Query(ctx, c.db, query)
	if err != nil {
		return fmt.Errorf("DeleteByDateColumn: %w", err)
	}

	return nil
}
