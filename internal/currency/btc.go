package currency

import "math"

type BTC int64

const BTCPrecision = 8
const BTCBase = 100_000_000

func (btc BTC) String() string {
	return convertIntToString(int64(btc), BTCPrecision)
}

func (btc BTC) Float64() float64 {
	return float64(btc) / BTCBase
}

func (btc BTC) Internal() int64 {
	return int64(btc)
}

func NewBTC(amount float64) BTC {
	return BTC(math.Round(amount * BTCBase))
}

func ParseBTC(amountStr string) (BTC, error) {
	amount, err := convertStringToAmount(amountStr, BTCPrecision)
	return BTC(amount), err
}

func (btc BTC) USD(btcPrice float64) USD {
	return USD(math.Round(btc.Float64() * btcPrice * 100))
}
