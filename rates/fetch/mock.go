package fetch

import (
	"log/slog"
	"os"
	"rates/domain"
	"rates/logging"
	"time"
)

var _ domain.DataFetcher = MockFetcher{}

type MockFetcher struct {}

func NewMockFetcher() MockFetcher {
	return MockFetcher{}
}

func (f MockFetcher) GetRatesData(_ time.Time) ([]byte, error) {
	data, err := os.ReadFile("fetch/mock_rates.json")
	if err != nil {
		slog.Error("failed reading file", logging.ErrorAttr(err))
	}

	return data, nil
}