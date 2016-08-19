package audio

import "math"

// Frames are a representation of audio frames across multiple channels
// [] <- channels []int <- frame int values.
type Frames [][]int

// FloatFrames are a representation similar to Frames but containing float values
// [] <- channels []float64 <- frame float values.
type FloatFrames [][]float64

// ToFloatFrames converts the frame int values to values in the -1, 1 range.
func (f Frames) ToFloatFrames(srcBitDepth int) FloatFrames {
	out := make(FloatFrames, len(f))
	if len(f) < 1 {
		return out
	}
	max := IntMaxSignedValue(srcBitDepth)
	zeroF := float64(max / 2)

	nbrChannels := len(f[0])

	var i, j int
	var v float64
	for i = 0; i < len(f); i++ {
		out[i] = make([]float64, nbrChannels)
		for j = 0; j < nbrChannels; j++ {
			v = float64(f[i][j]) / zeroF
			if v > 1 {
				v = 1
			} else if v < -1 {
				v = -1
			}
			out[i][j] = v
		}
	}
	return out
}

// ToMonoFrames returns a new mono audio frame set.
func (f Frames) ToMonoFrames() Frames {
	return ToMonoFrames(f)
}

// MonoRMS is a representation of the audio frames (in mono)
// rms = sqrt ( (1/n) * (x12 + x22 + … + xn2) )
// multiplying by 1/n effectively assigns equal weights to all the terms, making it a rectangular window.
// Other window equations can be used instead which would favor terms in the middle of the window.
// This results in even greater accuracy of the RMS value since brand new samples (or old ones at
// the end of the window) have less influence over the signal’s power.)
func (fs FloatFrames) MonoRMS() []float64 {
	out := []float64{}
	if len(fs) == 0 {
		return out
	}
	buf := make([]float64, int(rmsWindowSize))

	processBuffer := func() {
		total := 0.0
		for i := 0; i < len(buf); i++ {
			total += buf[i]
		}
		out = append(out, math.Sqrt((1.0/rmsWindowSize)*total))
	}

	nbrChans := len(fs[0])
	i := 0
	for j, f := range fs {
		var v float64
		if nbrChans > 1 {
			v = (f[0] + f[1]) / 2
		} else {
			v = f[0]
		}
		buf[i] = v
		i++
		if i == 400 || j == (len(fs)-1) {
			i = 0
			processBuffer()
		}

	}
	return out
}

// ToMonoFrames converts stereo into mono frames by averaging each samples.
// Note that a stereo frame could have 2 samples in phase opposition which would lead
// to a zero value. This edge case isn't taken in consideration.
func ToMonoFrames(fs Frames) Frames {
	if fs == nil {
		return nil
	}

	mono := make(Frames, len(fs))
	for i, f := range fs {
		mono[i] = []int{AvgInt(f...)}
	}
	return mono
}

// SlicedValues converts AudioFrames into a 1 dimensional slice of ints.
func SlicedValues(fs Frames) []int {
	if fs == nil || len(fs) == 0 {
		return nil
	}
	out := make([]int, len(fs))
	for i := 0; i < len(fs); i++ {
		out[i] = fs[i][0]
	}
	return out
}

// SlicedFValues converts AudioFloatFrames into a 1 dimensional slice of floats.
func SlicedFValues(fs FloatFrames) []float64 {
	if fs == nil || len(fs) == 0 {
		return nil
	}
	out := make([]float64, len(fs))
	for i := 0; i < len(fs); i++ {
		out[i] = fs[i][0]
	}
	return out
}
