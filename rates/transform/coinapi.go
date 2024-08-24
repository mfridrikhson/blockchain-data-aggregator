package transform

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"rates/domain"
	"rates/logging"
	"time"
)

type coinAPIResponse struct {
	Rates []coinAPIRate `json:"rates"`
}

type coinAPIRate struct {
	Time time.Time `json:"time"`
	AssetIdQuote string `json:"asset_id_quote"`
	Rate domain.Rate `json:"rate"`
}

var _ domain.DataTransformer = CoinAPITransformer{}

type CoinAPITransformer struct {}

func NewCoinAPITransformer() CoinAPITransformer {
	return CoinAPITransformer{}
}

func (t CoinAPITransformer) Transform(data []byte) ([]byte, error) {
	var apiResponse coinAPIResponse
	err := json.Unmarshal(data, &apiResponse)
	if err != nil {
		slog.Error("failed to parse input data", logging.ErrorAttr(err))
		return nil, err
	}

	var resultBuff bytes.Buffer

	for _, rate := range apiResponse.Rates {
		processedRow := domain.ProcessedDataRow{
			Symbol: rate.AssetIdQuote,
			Rate: rate.Rate,
			Date: rate.Time.Format(time.DateOnly),
		}

		marshalledRow, err := json.Marshal(processedRow)
		if err != nil {
			slog.Error("failed to marshal row", logging.ErrorAttr(err))
			return nil, err
		}

		resultBuff.Write(marshalledRow)
		resultBuff.WriteString("\n")
	}

	return resultBuff.Bytes(), nil
}

