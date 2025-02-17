package renderutils

func Max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

// Clamp will return the value v as long as it is between low and high. If v is too low or high, it will
// return the value for low or high
func Clamp(v, low, high int) int {
	if high < low {
		high = low
	}
	if low > high {
		low = high
	}
	return Min(Max(v, low), high)
}
