package windows

import "math"

// Hann generates a Hann window
// See https://en.wikipedia.org/wiki/Window_function#Hann_window
func Hann(L int) []float64 {
	r := make([]float64, L)
	arg := twoPi / float64(L-1)
	for i := 0; i < L; i++ {
		r[i] = 0.5 - 0.5*math.Cos(arg*float64(i))
	}

	return r
}
