package analysis

import "math"

// nextpow2 returns the smallest integer n (as a float) such that 2^n >= x
func nextpow2(x float64) float64 {
	y := math.Ceil(math.Log(x) / math.Log(2.0))
	return (float64(y))
}
