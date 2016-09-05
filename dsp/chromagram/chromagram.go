// The chromagram package implements analysis tools to extract the time-chrome representation
// of a PCMBuffer.
// See https://en.wikipedia.org/wiki/Chroma_feature
package chromagram

import (
	"github.com/mattetti/audio"
	"github.com/mattetti/audio/dsp/analysis"
	"github.com/mattetti/audio/dsp/windows"
)

type Chromagram struct {
	constantQ *analysis.ConstantQ
	Results   []float64
	FrameSize int
	HopSize   int
	FFTData   []complex128
	Window    []float64
}

func New(config *analysis.ConstantQConfig) *Chromagram {
	c := &Chromagram{constantQ: analysis.NewConstantQ(config)}
	c.Results = make([]float64, config.BinsPerOctave)
	c.FrameSize = c.constantQ.FFTLen
	c.HopSize = c.constantQ.HopSize

	return c
}

func (cg *Chromagram) Process(buf *audio.PCMBuffer) ([]float64, error) {
	if cg.constantQ.Speckernel == nil {
		if err := cg.constantQ.NewSpeckernel(); err != nil {
			return nil, err
		}
	}

	buf.SwitchPrimaryType(audio.Float)
	windowBuf := make([]float64, cg.FrameSize)
	if len(cg.Window) != cg.FrameSize {
		cg.Window = windows.Hamming(cg.FrameSize)
	}

	// apply the window into a buffer
	for i := 0; i < cg.FrameSize; i++ {
		windowBuf[i] = buf.Floats[i] * cg.Window[i]
	}

	//data := fft.FFTReal(windowBuf)

	// TODO: apply constant Q
	return nil, nil
}
