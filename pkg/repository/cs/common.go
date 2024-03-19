package cs

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

func SendFile(ctx context.Context, bucket *storage.BucketHandle, filename string) (err error) {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("os.Open: %w", err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("error close file: %s", err)
		}
	}(f)

	ctxIn, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	items := strings.Split(filename, "/")
	o := bucket.Object(items[len(items)-1])
	o = o.If(storage.Conditions{DoesNotExist: true})
	wc := o.NewWriter(ctxIn)
	if _, err = io.Copy(wc, f); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %w", err)
	}

	return nil
}

func CreateBucket(ctx context.Context, bucket *storage.BucketHandle, bucketName string) (err error) {
	err = bucket.Create(ctx, bucketName,
		&storage.BucketAttrs{Lifecycle: storage.Lifecycle{Rules: []storage.LifecycleRule{
			{
				Action:    storage.LifecycleAction{Type: storage.DeleteAction},
				Condition: storage.LifecycleCondition{AgeInDays: 3},
			},
		}}})
	if err != nil {
		return fmt.Errorf("bucket creation: %w", err)
	}

	return nil
}
