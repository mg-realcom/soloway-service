package cs

import (
	"context"
	"fmt"
)

func (c *Client) SendFile(ctx context.Context, filename string) (err error) {
	_, err = c.storage.Attrs(ctx)
	if err != nil {
		err = CreateBucket(ctx, c.storage, c.storageName)
		if err != nil {
			return fmt.Errorf("bucket creation: %w", err)
		}
	}

	err = SendFile(ctx, c.storage, filename)
	if err != nil {
		return fmt.Errorf("SendFile: %w", err)
	}

	return nil
}
