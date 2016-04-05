package misc

import "math"

var (
	RMSWindowSize = 400.0
)

type AudioFrames [][]int
type AudioFloatFrames [][]float64

// ToFloatFrames converts the frame int values to values in the -1, 1 range.
func (f AudioFrames) ToFloatFrames(srcBitDepth int) AudioFloatFrames {
	out := make(AudioFloatFrames, len(f))
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
func (f AudioFrames) ToMonoFrames() AudioFrames {
	return ToMonoFrames(f)
}

// RMS representation og the audio frames (in mono)
// rms = sqrt ( (1/n) * (x12 + x22 + … + xn2) )
// multiplying by 1/n effectively assigns equal weights to all the terms, making it a rectangular window.
// Other window equations can be used instead which would favor terms in the middle of the window.
// This results in even greater accuracy of the RMS value since brand new samples (or old ones at
// the end of the window) have less influence over the signal’s power.)
func (fs AudioFloatFrames) MonoRMS() []float64 {
	out := []float64{}
	if len(fs) == 0 {
		return out
	}
	buf := make([]float64, int(RMSWindowSize))

	processBuffer := func() {
		total := 0.0
		for i := 0; i < len(buf); i++ {
			total += buf[i]
		}
		out = append(out, math.Sqrt((1.0/RMSWindowSize)*total))
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
func ToMonoFrames(fs AudioFrames) AudioFrames {
	if fs == nil {
		return nil
	}

	mono := make(AudioFrames, len(fs))
	for i, f := range fs {
		mono[i] = []int{AvgInt(f...)}
	}
	return mono
}

// SlicedValues converts AudioFrames into a 1 dimensional slice of ints.
func SlicedValues(fs AudioFrames) []int {
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
func SlicedFValues(fs AudioFloatFrames) []float64 {
	if fs == nil || len(fs) == 0 {
		return nil
	}
	out := make([]float64, len(fs))
	for i := 0; i < len(fs); i++ {
		out[i] = fs[i][0]
	}
	return out
}

// AvgInt averages the int values passed
func AvgInt(xs ...int) int {
	var output int
	for i := 0; i < len(xs); i++ {
		output += xs[i]
	}
	return output / len(xs)
}

// IntMaxSignedValue returns the max value of an integer
// based on its memory size
func IntMaxSignedValue(b int) int {
	switch b {
	case 8:
		return 255 / 2
	case 16:
		return 65535 / 2
	case 24:
		return 16777215 / 2
	case 32:
		return 4294967295 / 2
	default:
		return 0
	}
}

// IeeeFloatToInt converts a 10 byte IEEE float into an int.
func IeeeFloatToInt(b [10]byte) int {
	var i uint32
	// Negative number
	if (b[0] & 0x80) == 1 {
		return 0
	}

	// Less than 1
	if b[0] <= 0x3F {
		return 1
	}

	// Too big
	if b[0] > 0x40 {
		return 67108864
	}

	// Still too big
	if b[0] == 0x40 && b[1] > 0x1C {
		return 800000000
	}

	i = (uint32(b[2]) << 23) | (uint32(b[3]) << 15) | (uint32(b[4]) << 7) | (uint32(b[5]) >> 1)
	i >>= (29 - uint32(b[1]))

	return int(i)
}

// IntToIeeeFloat converts an int into a 10 byte IEEE float.
func IntToIeeeFloat(i int) [10]byte {
	b := [10]byte{}
	num := float64(i)

	var sign int
	var expon int
	var fMant, fsMant float64
	var hiMant, loMant uint

	if num < 0 {
		sign = 0x8000
	} else {
		sign = 0
	}

	if num == 0 {
		expon = 0
		hiMant = 0
		loMant = 0
	} else {
		fMant, expon = math.Frexp(num)
		if (expon > 16384) || !(fMant < 1) { /* Infinity or NaN */
			expon = sign | 0x7FFF
			hiMant = 0
			loMant = 0 /* infinity */
		} else { /* Finite */
			expon += 16382
			if expon < 0 { /* denormalized */
				fMant = math.Ldexp(fMant, expon)
				expon = 0
			}
			expon |= sign
			fMant = math.Ldexp(fMant, 32)
			fsMant = math.Floor(fMant)
			hiMant = uint(fsMant)
			fMant = math.Ldexp(fMant-fsMant, 32)
			fsMant = math.Floor(fMant)
			loMant = uint(fsMant)
		}
	}

	b[0] = byte(expon >> 8)
	b[1] = byte(expon)
	b[2] = byte(hiMant >> 24)
	b[3] = byte(hiMant >> 16)
	b[4] = byte(hiMant >> 8)
	b[5] = byte(hiMant)
	b[6] = byte(loMant >> 24)
	b[7] = byte(loMant >> 16)
	b[8] = byte(loMant >> 8)
	b[9] = byte(loMant)

	return b
}

// Uint24to32 converts a 3 byte uint23 into a uint32
func Uint24to32(bytes []byte) uint32 {
	var output uint32
	output |= uint32(bytes[2]) << 0
	output |= uint32(bytes[1]) << 8
	output |= uint32(bytes[0]) << 16

	return output
}

// Uint32toUint24Bytes converts a uint32 into a 3 byte uint24 representation
func Uint32toUint24Bytes(n uint32) []byte {
	bytes := make([]byte, 3)
	bytes[0] = byte(n >> 16)
	bytes[1] = byte(n >> 8)
	bytes[2] = byte(n >> 0)

	return bytes
}
