package currency

import "math"

type USD int64

const USDPrecision = 2
const USDBase = 100

func (usd USD) String() string {
	return convertIntToString(int64(usd), USDPrecision)
}

func (usd USD) Float64() float64 {
	return float64(usd) / USDBase
}

func (usd USD) Internal() int64 {
	return int64(usd)
}

func NewUSD(amount float64) USD {
	return USD(math.Round(amount * USDBase))
}

func ParseUSD(amountStr string) (USD, error) {
	amount, err := convertStringToAmount(amountStr, USDPrecision)
	return USD(amount), err
}
