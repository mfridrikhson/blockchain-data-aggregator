package fetch

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"rates/domain"
	"rates/logging"
	"time"
)

const baseApiEndpoint = "https://rest.coinapi.io/v1"

var _ domain.DataFetcher = CoinAPIFetcher{}

type CoinAPIFetcher struct {
	ctx context.Context
	apiKey string
}

func NewCoinAPIFetcher(ctx context.Context, apiKey string) CoinAPIFetcher {
	return CoinAPIFetcher{
		ctx: ctx, 
		apiKey: apiKey,
	}
}

func (f CoinAPIFetcher) GetRatesData(date time.Time) ([]byte, error) {
	dateString := date.Format(time.DateOnly)
	
	endpoint := getRatesDataEndpoint(dateString)
	request, err := http.NewRequestWithContext(f.ctx, "GET", endpoint, nil)
	if err != nil {
		slog.Error("failed to create request", logging.ErrorAttr(err))
		return nil, err
	}
	request.Header.Add("X-CoinAPI-Key", f.apiKey)

	slog.Debug("requesting rates", slog.String("date", dateString))

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		slog.Error("failed to do request", logging.ErrorAttr(err))
		return nil, err
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		slog.Error("failed to read response body", logging.ErrorAttr(err))
		return nil, err
	}

	return body, nil
}

func getRatesDataEndpoint(date string) string {
	assetIdBase := "USD"
	values := url.Values{}
	values.Set("time", date)
	return fmt.Sprintf("%s/exchangerate/%s?%s", baseApiEndpoint, assetIdBase, values.Encode())
}