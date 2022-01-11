package utils

import "math"

func PriceDiff(last float64, prev float64) (float64, float64) {
	var diffAmt = 0.0
	var diffPerc = 0.0
	if last < 1 {
		diffAmt = ToFixed(last-prev, 4)
		if prev != 0 {
			diffPerc = RoundUp(diffAmt / prev * 100)
		}
	} else {
		diffAmt = RoundUp(last - prev)
		if prev != 0 {
			diffPerc = RoundUp(diffAmt / prev * 100)
		}
	}
	return diffAmt, diffPerc
}

//ToFixed rounds a float to precision
func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

//RoundUp rounds the float value to 2 decimal places
func RoundUp(x float64) float64 {
	return math.Ceil(x*100) / 100
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}
