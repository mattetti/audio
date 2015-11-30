package filters

import "fmt"

type FIR struct {
	Sinc *Sinc
}

func (f *FIR) LowPass(input []float64) ([]float64, error) {
	return f.Convolve(input, f.Sinc.LowPassKernels()) // fir := &filters.FIR{Sinc: s}
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

var (
	lowPassCoefs = map[int][]float64{
		44100: []float64{0.000130,
			-0.000181,
			0.000221,
			-0.000237,
			0.000213,
			-0.000135,
			-0.000004,
			0.000207,
			-0.000466,
			0.000761,
			-0.001062,
			0.001329,
			-0.001514,
			0.001568,
			-0.001445,
			0.001111,
			-0.000551,
			-0.000229,
			0.001190,
			-0.002265,
			0.003356,
			-0.004343,
			0.005090,
			-0.005458,
			0.005324,
			-0.004592,
			0.003212,
			-0.001195,
			-0.001382,
			0.004364,
			-0.007527,
			0.010583,
			-0.013201,
			0.015024,
			-0.015701,
			0.014917,
			-0.012424,
			0.008067,
			-0.001807,
			-0.006260,
			0.015908,
			-0.026776,
			0.038394,
			-0.050205,
			0.061604,
			-0.071977,
			0.080742,
			-0.087399,
			0.091557,
			0.907029,
			0.091557,
			-0.087399,
			0.080742,
			-0.071977,
			0.061604,
			-0.050205,
			0.038394,
			-0.026776,
			0.015908,
			-0.006260,
			-0.001807,
			0.008067,
			-0.012424,
			0.014917,
			-0.015701,
			0.015024,
			-0.013201,
			0.010583,
			-0.007527,
			0.004364,
			-0.001382,
			-0.001195,
			0.003212,
			-0.004592,
			0.005324,
			-0.005458,
			0.005090,
			-0.004343,
			0.003356,
			-0.002265,
			0.001190,
			-0.000229,
			-0.000551,
			0.001111,
			-0.001445,
			0.001568,
			-0.001514,
			0.001329,
			-0.001062,
			0.000761,
			-0.000466,
			0.000207,
			-0.000004,
			-0.000135,
			0.000213,
			-0.000237,
			0.000221,
			-0.000181,
			0.000130},
	}
)
