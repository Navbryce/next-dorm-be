package services

import (
	"cloud.google.com/go/storage"
	"context"
	firebase "firebase.google.com/go/v4"
)

type StorageBucket struct {
	*storage.BucketHandle
}

func NewStorageBucket(ctx context.Context, app *firebase.App, bucketName string) (*StorageBucket, error) {
	client, err := app.Storage(ctx)
	if err != nil {
		return nil, err
	}
	bucketHandle, err := client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}

	return &StorageBucket{
		bucketHandle,
	}, nil
}

func (sb *StorageBucket) Exists(ctx context.Context, blobName string) (bool, error) {
	if len(blobName) == 0 {
		return false, nil
	}
	handle := sb.Object(blobName)
	if _, err := handle.Attrs(ctx); err != nil {
		if err == storage.ErrObjectNotExist {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
