package audio

// DataFormat is an enum type to indicate the underlying data format used.
type DataFormat int

const (
	// Integer represents the int type
	Integer DataFormat = iota
	// Float represents the float64 type
	Float
	// Byte represents the byte type
	Byte
)

// Format is a high level representation of the underlying
type Format struct {
	// Channels is the number of channels contained in the data
	Channels int
	// SampleRate is the sampling rate in Hz
	SampleRate int
	// BitDepth is the number of bits of data for each sample
	BitDepth int
}

// PCMBuffer provides useful methods to read/manipulate audio buffers in PCM format
type PCMBuffer struct {
	Format *Format
	Ints   []int
	Floats []float64
	Bytes  []byte
	// DataType indicates the format used for the underlying data
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

// Size returns the number of frames contained in the buffer.
func (b *PCMBuffer) Size() (numFrames int) {
	if b == nil || b.Format == nil {
		return 0
	}
	numChannels := b.Format.Channels
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

func (b *PCMBuffer) Int16() []int16 {
	panic("not implemented")
}

func (b *PCMBuffer) Int32() []int32 {
	panic("not implemented")
}

func (b *PCMBuffer) Float32() []float32 {
	panic("not implemented")
}

func (b *PCMBuffer) Float64() []float64 {
	panic("not implemented")
}
