package utils

func MinFloat(first float64, second float64) float64 {
	if first < second {
		return first
	}

	return second
}
