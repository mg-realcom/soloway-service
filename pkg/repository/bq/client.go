package bq

import (
	"context"

	"cloud.google.com/go/bigquery"
	"github.com/rs/zerolog"
	"google.golang.org/api/option"
)

type Client struct {
	db     *bigquery.Client
	logger *zerolog.Logger
	ctx    context.Context
}

func NewClient(ctx context.Context, logger *zerolog.Logger, projectID, serviceKeyPath string) (*Client, error) {
	client, err := bigquery.NewClient(ctx, projectID, option.WithCredentialsFile(serviceKeyPath))
	if err != nil {
		return nil, err
	}

	return &Client{
		db:     client,
		logger: logger,
		ctx:    ctx,
	}, nil
}

func (c *Client) Close() error {
	return c.db.Close()
}
