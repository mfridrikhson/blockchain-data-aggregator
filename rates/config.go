package main

import (
	"context"
	"flag"
	"log/slog"
	"rates/domain"
	"rates/fetch"
	"rates/storage"
	"rates/transform"
	"time"
)

type AppConfig struct {
	Date *string
	BucketName *string
	ApiKey *string
}

func DefineAppConfig() AppConfig {
	date := flag.String("date", "", "Date in YYYY-MM-DD format to fetch exchange rates for")
	bucketName := flag.String("bucketName", "", "Name of the target Google Storage bucket")
	apiKey := flag.String("apiKey", "", "Exchange rates API access key")

	return AppConfig{
		Date: date,
		BucketName: bucketName,
		ApiKey: apiKey,
	}
}

type Container struct {
	Context context.Context
	Date time.Time
	DataFetcher domain.DataFetcher
	DataTransformer domain.DataTransformer
	StorageProvider domain.StorageProvider
}

func NewContainer() Container {
	ctx := context.Background()
	
	config := DefineAppConfig()
	flag.Parse()
	
	date, err := parseDate(*config.Date)
	if err != nil {
		panic(err)
	}
	
	
	fetcher := fetch.NewCoinAPIFetcher(ctx, *config.ApiKey)

	dataTransformer := transform.NewCoinAPITransformer()

	storageProvider, err := storage.NewGoogleStorageProvider(ctx, *config.BucketName)
	if err != nil {
		panic(err)
	}

	return Container{
		Context: ctx,
		Date: date,
		DataFetcher: fetcher,
		DataTransformer: dataTransformer,
		StorageProvider: storageProvider,
	}
}

func parseDate(dateStr string) (time.Time, error) {
	date, err := time.Parse(time.DateOnly, dateStr)
	if err != nil {
		slog.Error("invalid date", slog.String("date", dateStr))
		return time.Time{}, err
	}

	return date, nil
}