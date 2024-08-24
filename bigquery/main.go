package main

import (
	"context"
	"fmt"
	"log/slog"

	"cloud.google.com/go/bigquery"
)

func main() {
	container := NewContainer()
	defer container.Client.Close()

	slog.Info("creating dataset and tables in BigQuery")

	err := createDataset(container.Ctx, container.Dataset, *container.Config.Location)
	if err != nil {
		panic(err)
	}

	slog.Info("successfully created dataset", slog.String("dataset_id", container.Dataset.DatasetID))

	err = createEventsTable(container.Ctx, container.Dataset, *container.Config.EventsTableId, *container.Config.SourceBucketId)
	if err != nil {
		panic(err)
	}
	err = createRatesTable(container.Ctx, container.Dataset, *container.Config.RatesTableId, *container.Config.SourceBucketId)
	if err != nil {
		panic(err)
	}
	err = createResultTable(container.Ctx, container.Dataset, *container.Config.ResultTableId)
	if err != nil {
		panic(err)
	}
	
	slog.Info("successfully created tables")
	
	err = joinAndAggregateData(
		container.Ctx, 
		container.Client, 
		*container.Config.DatasetId,
		*container.Config.EventsTableId,
		*container.Config.RatesTableId,
		*container.Config.ResultTableId,
		*container.Config.Location,
	)
	if err != nil {
		panic(err)
	}

	slog.Info("successfully aggregated data")
}

func createDataset(ctx context.Context, datasetRef *bigquery.Dataset, location string) error {
	metadata := &bigquery.DatasetMetadata{Location: location}
	if err := datasetRef.Create(ctx, metadata); err != nil {
		slog.Error("failed to create dataset", slog.String("error", err.Error()))
		return err 
	}
	return nil
}

func createEventsTable(ctx context.Context, dataset *bigquery.Dataset, tableName, sourceBucketId string) error {
	eventsTable := dataset.Table(tableName)
	eventsSchema := bigquery.Schema{
		{Name: "app", Type: bigquery.StringFieldType},
		{Name: "ts", Type: bigquery.TimestampFieldType},
		{Name: "event", Type: bigquery.StringFieldType},
		{Name: "project_id", Type: bigquery.IntegerFieldType},
		{Name: "source", Type: bigquery.StringFieldType},
		{Name: "ident", Type: bigquery.IntegerFieldType},
		{Name: "user_id", Type: bigquery.StringFieldType},
		{Name: "session_id", Type: bigquery.StringFieldType},
		{Name: "country", Type: bigquery.StringFieldType},
		{Name: "device_type", Type: bigquery.StringFieldType},
		{Name: "device_os", Type: bigquery.StringFieldType},
		{Name: "device_os_ver", Type: bigquery.StringFieldType},
		{Name: "device_browser", Type: bigquery.StringFieldType},
		{Name: "device_browser_ver", Type: bigquery.StringFieldType},
		{Name: "props", Type: bigquery.JSONFieldType},
		{Name: "nums", Type: bigquery.JSONFieldType},
	}
	eventsSourceLocation := getSourceLocation("events/*.csv", sourceBucketId)
	eventsSourceFormat := bigquery.CSV

	err := Create(ctx, eventsTable, eventsSchema)
	if err != nil {
		return err
	}
	
	return Populate(ctx, eventsTable, eventsSourceLocation, eventsSourceFormat, eventsSchema)
}

func createRatesTable(ctx context.Context, dataset *bigquery.Dataset, tableName, sourceBucketId string) error {
	ratesTable := dataset.Table(tableName)
	ratesSchema := bigquery.Schema{
		{Name: "date", Type: bigquery.DateFieldType},
		{Name: "rate", Type: bigquery.FloatFieldType},
		{Name: "symbol", Type: bigquery.StringFieldType},
	}
	ratesSourceLocation := getSourceLocation("exchange-rates/*.json", sourceBucketId)
	ratesSourceFormat := bigquery.JSON

	err := Create(ctx, ratesTable, ratesSchema)
	if err != nil {
		return err
	}
	return Populate(ctx, ratesTable, ratesSourceLocation, ratesSourceFormat, ratesSchema)
}

func getSourceLocation(filesLocation, sourceBucketId string) string {
	return fmt.Sprintf("gs://%s/raw/%s", sourceBucketId, filesLocation)
}

func createResultTable(ctx context.Context, dataset *bigquery.Dataset, tableName string) error {
	resultTable := dataset.Table(tableName)
	resultSchema := bigquery.Schema{
		{Name: "date", Type: bigquery.DateFieldType},
		{Name: "project_id", Type: bigquery.IntegerFieldType},
		{Name: "num_transactions", Type: bigquery.IntegerFieldType},
		{Name: "total_volume_usd", Type: bigquery.FloatFieldType},
	}
	
	return Create(ctx, resultTable, resultSchema)
}

func joinAndAggregateData(ctx context.Context, client *bigquery.Client, datasetId, eventsTableId, ratesTableId, resultTableId, location string) error {
	/*
		Putting the query here just for the sake of test assignment.
		In production scenario I would only store it in the BigQuery and 
		schedule it to run on a specific cadence.

		In my opinion Go is not the language to handle operations on
		data - at least not in case of BigQuery data. 
		BigQuery has libraries for Python to operate on data in a 
		more developer-friendly Pandas-like manner.
	*/
	query := client.Query(fmt.Sprintf(`
	with events as (
		select 
			date(ts) as date, 
			project_id,
			count(*) as num_transactions,
			case json_value(props.currencySymbol)
			-- Normalize USDC.E symbol value
			when 'USDC.E' then 'USDCE'
			else json_value(props.currencySymbol)
			end as currency_symbol, 
			-- MATIC decimal values seem to be the same as raw ones
			-- Empirically I was able to deduct the coefficient of conversion to decimal,
			-- so while it may be incorrect, it yields far more realistic results.
			sum(
				case json_value(props.currencySymbol)
				when 'MATIC' then cast(json_value(nums.currencyValueDecimal) as decimal) / 1000000000000000000
				else cast(json_value(nums.currencyValueDecimal) as decimal)
				end
			) as currency_value
		from %s.%s 
		group by date(ts), project_id, currency_symbol
	)
	select e.date, project_id, sum(e.num_transactions) as num_transactions, sum(currency_value / rate) as total_volume_usd
	-- Some currencies were not available in exchange rates data for April,
	-- so we are unable to calculate USD values for them.
	from events e left join %s.%s r 
	on e.date = r.date and e.currency_symbol = r.symbol
	group by e.date, project_id
	order by e.date, project_id;
	`, datasetId, eventsTableId, datasetId, ratesTableId))

	query.Location = location
	query.QueryConfig.Dst = client.Dataset(datasetId).Table(resultTableId)

	slog.Info("aggregating events and rates data")

	job, err := query.Run(ctx)
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