package currency

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseBTCValid(t *testing.T) {
	testCases := []struct {
		btcString string
		expected  BTC
	}{
		{
			"12.01", BTC(1201000000),
		},
		{
			"12.", BTC(1200000000),
		},
		{
			"-12.01", BTC(-1201000000),
		},
		{
			"0.01", BTC(1000000),
		},
	}

	for i := range testCases {
		tc := testCases[i]
		btc, err := ParseBTC(tc.btcString)
		assert.Equal(t, tc.expected, btc)
		assert.NoError(t, err)
	}
}

func TestParseBTCInvalid(t *testing.T) {
	testCases := []struct {
		btcString string
	}{
		{
			"1-2.01",
		},
		{
			"",
		},
		{
			".",
		},
		{
			".1",
		},
	}
	for i := range testCases {
		tc := testCases[i]
		_, err := ParseBTC(tc.btcString)
		assert.Error(t, err)
	}
}

func TestBTCString(t *testing.T) {
	testCases := []struct {
		amount   BTC
		expected string
	}{
		{
			BTC(1201000000), "12.01000000",
		},
		{
			BTC(-1201000000), "-12.01000000",
		},
		{
			BTC(-1000000), "-0.01000000",
		},
	}

	for i := range testCases {
		tc := testCases[i]
		assert.Equal(t, tc.expected, tc.amount.String())
	}
}

func TestBTCFloat(t *testing.T) {
	testCases := []struct {
		amount   BTC
		expected float64
	}{
		{
			BTC(1201000000), 12.01,
		},
		{
			BTC(-1201000000), -12.01,
		},
		{
			BTC(-1000000), -0.01,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		assert.Equal(t, tc.expected, tc.amount.Float64())
	}
}

func TestBTCString2String(t *testing.T) {
	testCases := []struct {
		btcString string
		expected  string
	}{
		{
			"12.01", "12.01000000",
		},
		{
			"12.11111111111111", "12.11111111",
		},
		{
			"-12.01", "-12.01000000",
		},
		{
			"0.01", "0.01000000",
		},
		{
			"-0.01", "-0.01000000",
		},
	}

	for i := range testCases {
		tc := testCases[i]
		btc, err := ParseBTC(tc.btcString)
		assert.NoError(t, err)
		assert.Equal(t, tc.expected, btc.String())
	}
}
