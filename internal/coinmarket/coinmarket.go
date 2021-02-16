package coinmarket

//go:generate mockery -all

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type CoinmarketService interface {
	GetBTCPriceInUSD(ctx context.Context) (float64, error)
}

type CoinmarketServiceImpl struct {
	ApiKey string
}

func NewCoinmarketService(apiKey string) CoinmarketService {
	return &CoinmarketServiceImpl{ApiKey: apiKey}
}

func (service *CoinmarketServiceImpl) GetBTCPriceInUSD(ctx context.Context) (float64, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest",
		http.NoBody,
	)
	if err != nil {
		return 0, err
	}

	q := url.Values{}
	q.Add("symbol", "BTC")

	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", service.ApiKey)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error sending request to server: %w", err)
	}
	defer resp.Body.Close()

	var jsonResponse interface{}
	if err = json.NewDecoder(resp.Body).Decode(&jsonResponse); err != nil {
		return 0, fmt.Errorf("invalid response from server: %w", err)
	}

	btcPrice, _ := getValueFromJson(jsonResponse, "data", "BTC", "quote", "USD", "price")
	btcPrice, ok := btcPrice.(float64)
	if !ok {
		return 0, fmt.Errorf("unexpected response from server: %w", err)
	}
	return btcPrice.(float64), nil
}

func getValueFromJson(jsonData interface{}, path ...string) (interface{}, error) {
	for idx, key := range path {
		jsonObject, ok := jsonData.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("missing json object under %v", path[0:idx])
		}
		jsonData, ok = jsonObject[key]
		if !ok {
			return nil, fmt.Errorf("missing key %q under %v", key, path[0:idx])
		}
	}

	return jsonData, nil
}
