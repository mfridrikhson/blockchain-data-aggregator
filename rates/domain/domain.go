package domain

import "time"

type Symbol = string
type Rate = float64

type ProcessedDataRow struct {
	Symbol Symbol `json:"symbol"`
	Rate Rate `json:"rate"`
	Date string `json:"date"`
}

type DataFetcher interface {
	GetRatesData(date time.Time) ([]byte, error)
}

type DataTransformer interface {
	Transform(data []byte) ([]byte, error)
}

type StorageProvider interface {
	UploadToStorage(data []byte, path string) error
	Close()
}