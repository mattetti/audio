package filters

import "fmt"

// FIR represents a Finite Impulse Response filter taking a sinc.
// https://en.wikipedia.org/wiki/Finite_impulse_response
type FIR struct {
	Sinc *Sinc
}

// LowPass applies a low pass filter using the FIR
func (f *FIR) LowPass(input []float64) ([]float64, error) {
	return f.Convolve(input, f.Sinc.LowPassCoefs())
}

// Convolve "mixes" two signals together
func (f *FIR) Convolve(input, kernels []float64) ([]float64, error) {
	if f == nil {
		return nil, nil
	}
	if !(len(input) > len(kernels)) {
		return nil, fmt.Errorf("provided data set is not greater than the filter weights")
	}

	output := make([]float64, len(input))

	for i := 0; i < len(kernels); i++ {
		var sum float64

		for j := 0; j < i; j++ {
			sum += (input[j] * kernels[len(kernels)-(1+i-j)])
		}
		output[i] = sum
	}

	for i := len(kernels); i < len(input); i++ {
		var sum float64
		for j := 0; j < len(kernels); j++ {
			sum += (input[i-j] * kernels[j])
		}
		output[i] = sum
	}

	return output, nil
}
