package currency

import (
	"fmt"
	"strconv"
	"strings"
)

func convertIntToString(amount int64, decimalPlaces int) string {
	isNegative := amount < 0
	if isNegative {
		amount = -amount
	}

	baseString := strconv.FormatInt(amount, 10)
	paddingLength := decimalPlaces + 1 - len(baseString)
	if paddingLength > 0 {
		baseString = strings.Repeat("0", paddingLength) + baseString
	}

	dotPos := len(baseString) - decimalPlaces
	baseString = baseString[0:dotPos] + "." + baseString[dotPos:]
	if isNegative {
		return "-" + baseString
	}
	return baseString
}

func convertStringToAmount(amountStr string, decimalPlaces int) (int64, error) {
	amountStr = strings.TrimSpace(amountStr)

	var isNegative bool
	if strings.HasPrefix(amountStr, "-") {
		amountStr = strings.TrimSpace(amountStr[1:])
		isNegative = true
	}

	if len(amountStr) == 0 {
		return 0, fmt.Errorf("unparsable amount %q", amountStr)
	}

	parsedStr := strings.SplitN(amountStr, ".", 2)
	intPartStr := parsedStr[0]
	var fracPartStr string
	if len(parsedStr) == 2 {
		fracPartStr = parsedStr[1]
	}

	// no rounding
	fracPartStr = fracPartStr + strings.Repeat("0", decimalPlaces)
	fracPartStr = fracPartStr[0:decimalPlaces]

	intPart, intErr := parseInt(intPartStr)
	fracPart, fracErr := parseInt(fracPartStr)

	if intErr != nil || fracErr != nil || fracPart < 0 {
		return 0, fmt.Errorf("unparsable amount %q", amountStr)
	}

	for i := 0; i < decimalPlaces; i++ {
		intPart *= 10
	}

	amount := intPart + fracPart
	if isNegative {
		return -amount, nil
	}
	return amount, nil
}

func parseInt(s string) (int64, error) {
	if len(s) == 0 {
		return 0, fmt.Errorf("empty string")
	}
	val := int64(0)
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("only digits are allowed")
		}

		val = val*10 + int64(r-'0')
	}

	return val, nil
}
