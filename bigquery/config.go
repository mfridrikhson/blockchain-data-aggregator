package main

import (
	"context"
	"flag"
	"log/slog"

	"cloud.google.com/go/bigquery"
)

type AppConfig struct {
	ProjectId *string
	Location *string
	SourceBucketId *string
	DatasetId *string
	EventsTableId *string
	RatesTableId *string
	ResultTableId *string
}

func DefineAppConfig() AppConfig {
	projectId := flag.String("projectId", "", "Google Cloud project ID to create tables in")
	location := flag.String("location", "us-west1", "GCP data and compute location of created entities")
	sourceBucketId := flag.String("sourceBucketId", "", "Google Storage bucket ID to populate tables from")
	datasetId := flag.String("datasetId", "", "Name of the BigQuery dataset to be created")
	eventsTableId := flag.String("eventsTableId", "", "Name of the event data BigQuery table to be created")
	ratesTableId := flag.String("ratesTableId", "", "Name of the exchange rates data BigQuery table to be created")
	resultTableId := flag.String("resultTableId", "", "Name of the BigQuery table with aggregated data to be created")

	return AppConfig{
		ProjectId: projectId,
		Location: location,
		SourceBucketId: sourceBucketId,
		DatasetId: datasetId,
		EventsTableId: eventsTableId,
		RatesTableId: ratesTableId,
		ResultTableId: resultTableId,
	}
}

type Container struct {
	Ctx context.Context
	Config AppConfig
	Client *bigquery.Client
	Dataset *bigquery.Dataset
}

func NewContainer() Container {
	ctx := context.Background()
	config := DefineAppConfig()
	flag.Parse()

	slog.Info("initializing program to create dataset and tables in BigQuery",
		slog.String("project_id", *config.ProjectId),
		slog.String("dataset_id", *config.DatasetId),
		slog.String("events_table_id", *config.EventsTableId),
		slog.String("rates_table_id", *config.RatesTableId),
		slog.String("result_table_id", *config.ResultTableId),
	)

	client, err := bigquery.NewClient(ctx, *config.ProjectId)
	if err != nil {
		slog.Error("failed to create BigQuery client", slog.String("error", err.Error()))
		panic(err)
	}

	datasetRef := client.Dataset(*config.DatasetId)

	return Container{
		Ctx: ctx,
		Config: config,
		Client: client,
		Dataset: datasetRef,
	}
}