package analysis

import "math"

// ConstantQConfig is the config used to do a Constant Q transform
// it defines the geometrically-spaced frequency range to analyze.
type ConstantQConfig struct {
	// Fs is the sample rate
	Fs float64
	// MinFs is the minimum frequency (suggested: 55 or midi 36)
	MinFs float64
	// MaxFs is the maximum frequency (suggested: Fs/2 or midi 96)
	MaxFs float64
	// BinsPerOctave is the numbers of bins per octave (suggested: 12 to 24)
	BinsPerOctave int
	// Threshold is the sensitivity threshold (suggested: 0.0054)
	Threshold float64
}

// Speckernel is a spectral kernel matrix.
type Speckernel struct {
	IS   []int
	JS   []int
	Imag []float64
	Real []float64
}

// ConstantQ transform transforms a data series to the frequency domain.
// It is related to the Fourier transform, and very closely related to the
// complex Morlet wavelet transform.
// In general, the transform is well suited to musical data
// since it mirrors the human auditory system, whereby at lower frequencies spectral resolution is better,
// whereas temporal resolution improves at higher frequencies.
//
// As the output of the transform is effectively amplitude/phase
// against log frequency, fewer frequency bins are required to cover a given range
// effectively, and this proves useful where frequencies span several octaves.
// As the range of human hearing covers approximately ten octaves from 20 Hz to around 20 kHz,
// this reduction in output data is significant.
// See https://en.wikipedia.org/wiki/Constant_Q_transform
type ConstantQ struct {
	Config        *ConstantQConfig
	Data          []float64
	ConstantQBins int
	// FFTLen is the length of the FFT required for the filter bank
	FFTLen int
	// Q value of the filter bank
	Q       float64
	HopSize int
}

// NewConstantQ configures and returns a constant Q transform ready to be used.
func NewConstantQ(config *ConstantQConfig) *ConstantQ {
	cq := &ConstantQ{Config: config}
	cq.ConstantQBins = int(
		math.Ceil(
			float64(config.BinsPerOctave) *
				(math.Log(config.MaxFs/config.MinFs) /
					math.Log(2.0))),
	)
	cq.Q = 1 / (math.Pow(2, (1/(float64(config.BinsPerOctave)))) - 1)
	cq.FFTLen = int(
		math.Pow(2,
			nextpow2(
				math.Ceil(cq.Q*config.Fs/config.MinFs),
			),
		))
	cq.HopSize = int(float64(cq.FFTLen) / 8)
	cq.Data = make([]float64, 2*cq.ConstantQBins)
	return cq
}
