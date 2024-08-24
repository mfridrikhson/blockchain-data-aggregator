package storage

import (
	"context"
	"log/slog"
	"rates/domain"
	"rates/logging"

	"cloud.google.com/go/storage"
)

var _ domain.StorageProvider = GoogleStorageProvider{}

type GoogleStorageProvider struct {
	context context.Context
	client *storage.Client
	bucket string
}

func NewGoogleStorageProvider(ctx context.Context, bucket string) (*GoogleStorageProvider, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		slog.Error("failed to initialize client", logging.ErrorAttr(err))
		return nil, err
	}

	provider := GoogleStorageProvider{
		context: ctx,
		client: client,
		bucket: bucket,
	}

	return &provider, nil
}

func (p GoogleStorageProvider) UploadToStorage(data []byte, path string) error {
	object := p.client.Bucket(p.bucket).Object(path)

	writer := object.NewWriter(p.context)
	defer func() {
		err := writer.Close()
		if err != nil {
			slog.Error("error on closing writer", logging.ErrorAttr(err))
			return
		}

		p.printObjectMetadata(object)
	}()

	bytesWritten, err := writer.Write(data)
	if err != nil {
		slog.Error("failed to write data", slog.Int("bytes_written", bytesWritten), logging.ErrorAttr(err))
	}

	return err
}

func (p GoogleStorageProvider) printObjectMetadata(object *storage.ObjectHandle) {
	attrs, err := object.Attrs(p.context)
	if err != nil {
		slog.Error("error reading object attrs", logging.ErrorAttr(err))
	} else {
		slog.Info("google cloud storage object created", slog.Time("created_at", attrs.Created))
	}
}

func (p GoogleStorageProvider) Close() {
	err := p.client.Close() 
	if err != nil {
		slog.Error("failed to close Google Storage client", logging.ErrorAttr(err))
	}
}