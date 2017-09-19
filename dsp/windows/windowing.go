package windows

// Windowing returns a windowed signal for given input and window signal.
func Windowing(input, window []float64) []float64 {
	result := make([]float64, len(input))

	for i := range input {
		result[i] = input[i] * window[i]
	}

	return result
}
