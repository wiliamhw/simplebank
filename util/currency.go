package util

const (
	USD = "USD"
	EUR = "EUR"
	CAD = "CAD"
)

var currencies = [...]string{EUR, USD, CAD}

func IsSupportedCurrency(currency string) bool {
	return InArray(currencies, currency)
}
