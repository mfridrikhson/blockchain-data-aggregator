package main

import (
	"context"
	"log/slog"
	"time"

	"cloud.google.com/go/bigquery"
)

func Create(ctx context.Context, tableRef *bigquery.Table, schema bigquery.Schema) error {
	slog.Info("creating table", slog.String("dataset_id", tableRef.DatasetID), slog.String("table_id", tableRef.TableID))

	metadata := &bigquery.TableMetadata{
		Schema:         schema,
		ExpirationTime: time.Now().AddDate(0, 1, 0), // Table will be automatically deleted in 1 month.
	}
	if err := tableRef.Create(ctx, metadata); err != nil {
		slog.Error("failed to create table", slog.String("error", err.Error()))
		return err
	}
	
	return nil
}

func Populate(ctx context.Context, tableRef *bigquery.Table, sourceLocation string, sourceFormat bigquery.DataFormat, schema bigquery.Schema) error {
	slog.Info("loading data from source", slog.String("source", sourceLocation), slog.String("format", string(sourceFormat)))

	gcsReference := bigquery.NewGCSReference(sourceLocation)
	gcsReference.SourceFormat = sourceFormat
	if sourceFormat == bigquery.CSV {
		gcsReference.CSVOptions = bigquery.CSVOptions{
			// Skip header row
			SkipLeadingRows: 1,
		}
	}
	gcsReference.Schema = schema

	loader := tableRef.LoaderFrom(gcsReference)
	loader.WriteDisposition = bigquery.WriteEmpty

	job, err := loader.Run(ctx)
	if err != nil {
		return err
	}

	slog.Debug("running job", slog.String("id", job.ID()))
	
	status, err := job.Wait(ctx)
	if err != nil {
		return err
	}

	if err := status.Err(); err != nil {
		slog.Error("job completed with error", slog.String("error", err.Error()),  slog.String("id", job.ID()))
		return err
	}
	return nil
}