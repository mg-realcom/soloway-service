package cs

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/rs/zerolog"
	"google.golang.org/api/option"
)

type Client struct {
	client      *storage.Client
	storage     *storage.BucketHandle
	logger      *zerolog.Logger
	ctx         context.Context
	storageName string
}

func NewClient(ctx context.Context, logger *zerolog.Logger, storageName string, serviceKeyPath string) (*Client, error) {
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(serviceKeyPath))
	if err != nil {
		return nil, err
	}

	bucket := client.Bucket(storageName)

	return &Client{
		client:      client,
		storage:     bucket,
		logger:      logger,
		ctx:         ctx,
		storageName: storageName,
	}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}
