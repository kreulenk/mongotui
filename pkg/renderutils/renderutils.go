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

func Clamp(v, low, high int) int {
	return Min(Max(v, low), high)
}
