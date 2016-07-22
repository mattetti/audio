package audio

import (
	"io"
	"math"
)

type SamplesInt []int

func (f SamplesInt) Get(channel, n int) int {
	return f[n*channel]
}

func (f SamplesInt) StereoToMono() SamplesInt {
	if f == nil {
		return nil
	}

	mono := make(SamplesInt, len(f)/2)
	var j int
	for i := 0; i+2 <= len(f); {
		mono[j] = AvgInt(f[i], f[i+1])
		i += 2
		j++
	}
	return mono
}

type SamplesFloat64 []float64

func (f SamplesFloat64) Get(channel, n int) float64 {
	return f[n*channel]
}

func (f SamplesFloat64) StereoToMono() SamplesFloat64 {
	if f == nil {
		return nil
	}

	mono := make(SamplesFloat64, len(f)/2)
	var j int
	for i := 0; i+2 <= len(f); {
		mono[j] = AvgFloat64(f[i], f[i+1])
		i += 2
		j++
	}
	return mono
}

type PCM interface {
	Ints(samples SamplesInt) (n int, err error)
	Float64s(samples SamplesFloat64) (n int, err error)
	NextInts(n int) (SamplesInt, error)
	NextFloat64s(n int) (SamplesFloat64, error)
	Read(buf []byte) (n int, err error)
	Offset() int64
	Seek(frameOffset int64, whence int) (offset int64, err error)
	Info() (numChannels, bitDepth int, sampleRate int64, err error)
	Size() int64
}

// FrameInfo represents the frame-level information.
type FrameInfo struct {
	// Channels represent the number of audio channels
	// (e.g. 1 for mono, 2 for stereo).
	Channels int
	// Bit depth is the number of bits used to represent
	// a single sample.
	BitDepth int

	// Sample rate is the number of samples to be played each second.
	SampleRate int64
}

type Chunk struct {
	ID   [4]byte
	Size int
	Pos  int
	R    io.Reader
}

// AvgInt averages the int values passed
func AvgInt(xs ...int) int {
	var output int
	for i := 0; i < len(xs); i++ {
		output += xs[i]
	}
	return output / len(xs)
}

// AvgFloat64 averages the passed float values
func AvgFloat64(xs ...float64) float64 {
	var output float64
	for i := 0; i < len(xs); i++ {
		output += xs[i]
	}
	return output / float64(len(xs))
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
