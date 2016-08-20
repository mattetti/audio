package transforms

import (
	"math"

	"github.com/mattetti/audio"
)

// FullWaveRectifier to make all signal positive
// See https://en.wikipedia.org/wiki/Rectifier#Full-wave_rectification
func FullWaveRectifier(buf *audio.PCMBuffer) error {
	if buf == nil {
		return nil
	}
	samples := buf.AsFloat64s()
	buf.Floats = make([]float64, len(samples))
	for i := 0; i < len(samples); i++ {
		buf.Floats[i] = math.Abs(samples[i])
	}
	buf.DataType = audio.Float
	buf.Ints = nil
	buf.Bytes = nil

	return nil
}

// MonoRMS converts the buffer to mono and apply an RMS treatment.
// rms = sqrt ( (1/n) * (x12 + x22 + … + xn2) )
// multiplying by 1/n effectively assigns equal weights to all the terms, making it a rectangular window.
// Other window equations can be used instead which would favor terms in the middle of the window.
// This results in even greater accuracy of the RMS value since brand new samples (or old ones at
// the end of the window) have less influence over the signal’s power.)
func MonoRMS(b *audio.PCMBuffer, windowSize int) error {
	if b == nil || b.Len() == 0 {
		return nil
	}
	out := []float64{}
	winBuf := make([]float64, windowSize)
	windowSizeF := float64(windowSize)

	processWindow := func(idx int) {
		total := 0.0
		for i := 0; i < len(winBuf); i++ {
			total += winBuf[idx] * winBuf[idx]
		}
		v := math.Sqrt((1.0 / windowSizeF) * total)
		out = append(out, v)
	}

	nbrChans := b.Format.NumChannels
	samples := b.AsFloat64s()

	var windowIDX int
	// process each frame, convert it to mono and them RMS it
	for i := 0; i < len(samples); i++ {
		v := samples[i]
		if nbrChans > 1 {
			for j := 1; j < nbrChans; j++ {
				i++
				v += samples[i]
			}
			v /= float64(nbrChans)
		}
		winBuf[windowIDX] = v
		windowIDX++
		if windowIDX == windowSize || i == (len(samples)-1) {
			windowIDX = 0
			processWindow(windowIDX)
		}
	}

	b.Format.NumChannels = 1
	b.Format.SampleRate /= windowSize
	b.DataType = audio.Float
	b.Floats = out
	return nil
}
