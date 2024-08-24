package domain

import "time"

type DataFetcher interface {
	GetRatesData(date time.Time) ([]byte, error)
}