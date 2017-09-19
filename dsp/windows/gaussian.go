package windows

import "math"

// Gaussian generates a gaussian window with a set standard deviation (0.4 recommended)
// See https://en.wikipedia.org/wiki/Window_function#Gaussian_window
func Gaussian(L int, stdDev float64) []float64 {
	r := make([]float64, L)
	var coef float64
	for i := 0; i < L; i++ {
		coef = (float64(i) - float64(L-1)/2.0) / (stdDev * float64(L-1) / 2.0)
		r[i] = math.Exp(-0.5 * coef * coef)
	}

	return r
}
