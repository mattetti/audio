// The chromagram package implements analysis tools to extract the time-chrome representation
// of a PCMBuffer.
// See https://en.wikipedia.org/wiki/Chroma_feature
package chromagram

import (
	"math"

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

// Process analyses the passed buffer and returns the value for each bin
func (cg *Chromagram) Process(buf *audio.PCMBuffer) ([]float64, error) {
	if cg.constantQ.Speckernel == nil {
		if err := cg.constantQ.NewSpeckernel(); err != nil {
			return nil, err
		}
	}
	octaves := int(math.Floor(float64(cg.constantQ.ConstantQBins)/float64(cg.constantQ.Config.BinsPerOctave)) - 1)

	// reset the results
	cg.Results = make([]float64, octaves*cg.constantQ.Config.BinsPerOctave)

	buf.SwitchPrimaryType(audio.Float)
	windowBuf := make([]float64, cg.FrameSize)
	if len(cg.Window) != cg.FrameSize {
		cg.Window = windows.Hamming(cg.FrameSize)
	}

	// apply the window into a buffer
	for i := 0; i < len(buf.Floats); i++ {
		windowBuf[i] = buf.Floats[i] * cg.Window[i]
	}

	// winBuf := fft.FFTReal(windowBuf)
	winBuf := audio.NewPCMFloatBuffer(windowBuf, buf.Format)

	// apply constant Q
	chromaData := cg.constantQ.Process(winBuf)
	for octave := 0; octave <= octaves; octave++ {
		firstBin := octave * cg.constantQ.Config.BinsPerOctave
		for i := 0; i < cg.constantQ.Config.BinsPerOctave; i++ {
			cg.Results[i] += chromaData[firstBin+i] // kabs( m_CQRe[ firstBin + i ], m_CQIm[ firstBin + i ])
		}
	}

	// TODO: normalize

	return cg.Results, nil
}
