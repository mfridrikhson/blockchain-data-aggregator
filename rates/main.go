package main

import (
	"fmt"
	"log/slog"
	"time"
)

func main() {
	container := NewContainer()
	defer container.StorageProvider.Close()

	slog.Info("running exchange rates ingestion", slog.Time("date", container.Date))

	rawData, err := container.DataFetcher.GetRatesData(container.Date)
	if err != nil {
		panic(err)
	}

	transformedData, err := container.DataTransformer.Transform(rawData)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(transformedData))

	fileName := getFileName(container.Date, "json")
	err = container.StorageProvider.UploadToStorage(transformedData, fileName)
	if err != nil {
		panic(err)
	}
}

func getFileName(date time.Time, format string) string {
	dateString := date.Format(time.DateOnly)

	directory := "raw/exchange-rates"

	return fmt.Sprintf("%s/%s.%s", directory, dateString, format)
}