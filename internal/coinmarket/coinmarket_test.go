package coinmarket

import (
	"context"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

const CoinmarketResponse = `
{
  "status": {
    "timestamp": "2021-02-16T15:39:53.182Z",
    "error_code": 0,
    "error_message": null,
    "elapsed": 35,
    "credit_count": 1,
    "notice": null
  },
  "data": {
    "BTC": {
      "id": 1,
      "name": "Bitcoin",
      "symbol": "BTC",
      "slug": "bitcoin",
      "num_market_pairs": 9682,
      "date_added": "2013-04-28T00:00:00.000Z",
      "tags": [
        "mineable",
        "pow",
        "sha-256",
        "store-of-value",
        "state-channels",
        "coinbase-ventures-portfolio",
        "three-arrows-capital-portfolio",
        "polychain-capital-portfolio"
      ],
      "max_supply": 21000000,
      "circulating_supply": 18630306,
      "total_supply": 18630306,
      "is_active": 1,
      "platform": null,
      "cmc_rank": 1,
      "is_fiat": 0,
      "last_updated": "2021-02-16T15:38:02.000Z",
      "quote": {
        "USD": {
          "price": 49239.06561166671,
          "volume_24h": 79982713347.1132,
          "percent_change_1h": 0.35347662,
          "percent_change_24h": 3.04389851,
          "percent_change_7d": 6.31542234,
          "percent_change_30d": 40.10298659,
          "market_cap": 917338859499.428,
          "last_updated": "2021-02-16T15:38:02.000Z"
        }
      }
    }
  }
}`

func TestGetBTCPriceInUSD(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest",
		func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "1111-2222", req.Header.Get("X-CMC_PRO_API_KEY"))
			return httpmock.NewStringResponse(500, CoinmarketResponse), nil
		},
	)

	coinmarketService := NewCoinmarketService("1111-2222")
	btcPrice, err := coinmarketService.GetBTCPriceInUSD(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 49239.06561166671, btcPrice)
}
