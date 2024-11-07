package util

const (
	RMB = "RMB"
	USD = "USD"
	EUR = "EUR"
	CAD = "CAD"
)

func IsSupportedCurrency(currency string) bool {
	switch currency {
	case RMB, USD, EUR, CAD:
		return true
	}
	return false
}
