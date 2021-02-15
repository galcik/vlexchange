package currency

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseUSDValid(t *testing.T) {
	testCases := []struct {
		usdString string
		expected  USD
	}{
		{
			"12.01", USD(1201),
		},
		{
			"12.", USD(1200),
		},
		{
			"-12.01", USD(-1201),
		},
		{
			"0.01", USD(1),
		},
	}

	for i := range testCases {
		tc := testCases[i]
		usd, err := ParseUSD(tc.usdString)
		assert.Equal(t, tc.expected, usd)
		assert.NoError(t, err)
	}
}

func TestParseUSDInvalid(t *testing.T) {
	testCases := []struct {
		usdString string
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
		_, err := ParseUSD(tc.usdString)
		assert.Error(t, err)
	}
}

func TestUSDString(t *testing.T) {
	testCases := []struct {
		amount   USD
		expected string
	}{
		{
			USD(1201), "12.01",
		},
		{
			USD(-1201), "-12.01",
		},
		{
			USD(-1), "-0.01",
		},
	}

	for i := range testCases {
		tc := testCases[i]
		assert.Equal(t, tc.expected, tc.amount.String())
	}
}

func TestUSDFloat(t *testing.T) {
	testCases := []struct {
		amount   USD
		expected float64
	}{
		{
			USD(1201), 12.01,
		},
		{
			USD(-1201), -12.01,
		},
		{
			USD(-1), -0.01,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		assert.Equal(t, tc.expected, tc.amount.Float64())
	}
}

func TestUSDString2String(t *testing.T) {
	testCases := []struct {
		usdString string
		expected  string
	}{
		{
			"12.01", "12.01",
		},
		{
			"12.11111111111111", "12.11",
		},
		{
			"-12.01", "-12.01",
		},
		{
			"0.01", "0.01",
		},
		{
			"-0.01", "-0.01",
		},
	}

	for i := range testCases {
		tc := testCases[i]
		usd, err := ParseUSD(tc.usdString)
		assert.NoError(t, err)
		assert.Equal(t, tc.expected, usd.String())
	}
}
