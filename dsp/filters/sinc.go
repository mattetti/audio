package filters

import (
	"math"

	"github.com/mattetti/audio/dsp/windows"
)

// Sinc represents a sinc function
// The sinc function also called the "sampling function," is a function that
// arises frequently in signal processing and the theory of Fourier transforms.
// The full name of the function is "sine cardinal," but it is commonly referred to by
// its abbreviation, "sinc."
// http://mathworld.wolfram.com/SincFunction.html
type Sinc struct {
	CutOffFreq    float64
	SamplingFreq  int
	Taps          int
	Window        windows.Function
	_lowPassCoefs []float64
}

// LowPassCoefs return the coeficients to create a low pass filter
func (s *Sinc) LowPassCoefs() []float64 {
	if s == nil {
		return nil
	}
	if s._lowPassCoefs != nil && len(s._lowPassCoefs) > 0 {
		return s._lowPassCoefs
	}
	size := s.Taps + 1
	b := (2 * math.Pi) * s.TransitionFreq()
	s._lowPassCoefs = make([]float64, size)
	winData := s.Window(size)

	for i := 0; i < (s.Taps / 2); i++ {
		c := float64(i) - float64(s.Taps)/2
		y := math.Sin(b*c) / (math.Pi * c)
		s._lowPassCoefs[i] = y * winData[i]
		s._lowPassCoefs[size-1-i] = s._lowPassCoefs[i]
	}

	s._lowPassCoefs[s.Taps/2] = 2 * s.TransitionFreq() * winData[s.Taps/2]
	return s._lowPassCoefs
}

func (s *Sinc) TransitionFreq() float64 {
	if s == nil {
		return 0
	}
	return s.CutOffFreq / float64(s.SamplingFreq)
}
