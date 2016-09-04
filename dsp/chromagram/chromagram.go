// The chromagram package implements analysis tools to extract the time-chrome representation
// of a PCMBuffer.
// See https://en.wikipedia.org/wiki/Chroma_feature
package chromagram

import "github.com/mattetti/audio/dsp/analysis"

type Chromagram struct {
	constantQ *analysis.ConstantQ
	Results   []float64
	FrameSize int
	HopSize   int
	//FFT
}

func New(config *analysis.ConstantQConfig) *Chromagram {
	c := &Chromagram{constantQ: analysis.NewConstantQ(config)}
	c.Results = make([]float64, config.BinsPerOctave)
	c.FrameSize = c.constantQ.FFTLen
	c.HopSize = c.constantQ.HopSize
	// c.FFT of the framesize
	return c
}
