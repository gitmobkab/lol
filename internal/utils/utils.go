package utils

func Warp(value int, min int, max int) int {
	if value < min {
		return max
	}
	if value > max {
		return min
	}
	return value
}