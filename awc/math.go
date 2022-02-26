package awc

func keepFloatInRange(value, min, max float32) float32 {
	if value <= min {
		return min
	}
	if value >= max {
		return max
	}
	return value
}
