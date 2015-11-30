package filters

import (
	"math"

	"github.com/mattetti/audio/dsp/windows"
)

type Sinc struct {
	CutOffFreq      float64
	SamplingFreq    int
	Order           int
	Window          windows.Function
	_lowPassKernels []float64
}

func (s *Sinc) TransitionFreq() float64 {
	if s == nil {
		return 0
	}
	return s.CutOffFreq / float64(s.SamplingFreq)
}

func (s *Sinc) LowPassKernels() []float64 {
	if s == nil {
		return nil
	}
	size := s.Order+1
	b := (2 * math.Pi) * s.TransitionFreq()
	s._lowPassKernels = make([]float64, size)
	winData := s.Window(size)
		
	for i := 0; i < (s.Order / 2); i++ {
		c := float64(i) - float64(s.Order)/2
		y := math.Sin(b*c) / (math.Pi * c)
		s._lowPassKernels[i] = y * winData[i]
		s._lowPassKernels[size-1-i] = s._lowPassKernels[i]
	}
	
	s._lowPassKernels[s.Order/2] = 2 * s.TransitionFreq() * winData[s.Order/2]
	return s._lowPassKernels
}
