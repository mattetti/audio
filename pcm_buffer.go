package audio

import (
	"bytes"
	"encoding/binary"
)

// TODO(mattetti): define a transform interface
// so we can easily implement RMS, channel reduction
// or other straightforward DSP operations.

// DataFormat is an enum type to indicate the underlying data format used.
type DataFormat int

const (
	// Unknown refers to an unknown format
	Unknown DataFormat = iota
	// Integer represents the int type.
	// it represents the native int format used in audio buffers.
	Integer
	// Float represents the float64 type.
	// It represents the native float format used in audio buffers.
	Float
	// Byte represents the byte type.
	Byte
)

// Format is a high level representation of the underlying data.
type Format struct {
	// NumChannels is the number of channels contained in the data
	NumChannels int
	// SampleRate is the sampling rate in Hz
	SampleRate int
	// BitDepth is the number of bits of data for each sample
	BitDepth int
	// Endianess indicate how the byte order of underlying bytes
	Endianness binary.ByteOrder
}

// PCMBuffer encapsulates uncompressed audio data
// and provides useful methods to read/manipulate this PCM data.
type PCMBuffer struct {
	// Format describes the format of the buffer data.
	Format *Format
	// Ints is a store for audio sample data as integers.
	Ints []int
	// Floats is a store for audio samples data as float64.
	Floats []float64
	// Bytes is a store for audio samples data as raw bytes.
	Bytes []byte
	// DataType indicates the primary format used for the underlying data.
	// The consumer of the buffer might want to look at this value to know what store
	// to use to optimaly retrieve data.
	DataType DataFormat
}

// NewPCMIntBuffer returns a new PCM buffer backed by the passed integer samples
func NewPCMIntBuffer(data []int, format *Format) *PCMBuffer {
	return &PCMBuffer{
		Format:   format,
		DataType: Integer,
		Ints:     data,
	}
}

// NewPCMFloatBuffer returns a new PCM buffer backed by the passed float samples
func NewPCMFloatBuffer(data []float64, format *Format) *PCMBuffer {
	return &PCMBuffer{
		Format:   format,
		DataType: Float,
		Floats:   data,
	}
}

// NewPCMByteBuffer returns a new PCM buffer backed by the passed float samples
func NewPCMByteBuffer(data []byte, format *Format) *PCMBuffer {
	return &PCMBuffer{
		Format:   format,
		DataType: Byte,
		Bytes:    data,
	}
}

// Len returns the length of the underlying data.
func (b *PCMBuffer) Len() int {
	if b == nil {
		return 0
	}

	switch b.DataType {
	case Integer:
		return len(b.Ints)
	case Float:
		return len(b.Floats)
	case Byte:
		return len(b.Bytes)
	default:
		return 0
	}
}

// Size returns the number of frames contained in the buffer.
func (b *PCMBuffer) Size() (numFrames int) {
	if b == nil || b.Format == nil {
		return 0
	}
	numChannels := b.Format.NumChannels
	if numChannels == 0 {
		numChannels = 1
	}
	switch b.DataType {
	case Integer:
		numFrames = len(b.Ints) / numChannels
	case Float:
		numFrames = len(b.Floats) / numChannels
	case Byte:
		sampleSize := int((b.Format.BitDepth-1)/8 + 1)
		numFrames = (len(b.Bytes) / sampleSize) / numChannels
	}
	return numFrames
}

// Int16s returns the buffer samples as int16 sample values.
func (b *PCMBuffer) Int16s() (out []int16) {
	if b == nil {
		return nil
	}
	switch b.DataType {
	case Integer, Float:
		out = make([]int16, len(b.Ints))
		for i := 0; i < len(b.Ints); i++ {
			out[i] = int16(b.Ints[i])
		}
	case Byte:
		// if the format isn't defined, we can't read the byte data
		if b.Format == nil || b.Format.Endianness == nil || b.Format.BitDepth == 0 {
			return out
		}
		bytesPerSample := int((b.Format.BitDepth-1)/8 + 1)
		buf := bytes.NewBuffer(b.Bytes)
		out := make([]int16, len(b.Bytes)/bytesPerSample)
		binary.Read(buf, b.Format.Endianness, &out)
	}
	return out
}

func (b *PCMBuffer) Int32s() []int32 {
	panic("not implemented")
}

func (b *PCMBuffer) Int64s() []int64 {
	panic("not implemented")
}

func (b *PCMBuffer) Float32s() []float32 {
	panic("not implemented")
}

func (b *PCMBuffer) Float64s() []float64 {
	panic("not implemented")
}
