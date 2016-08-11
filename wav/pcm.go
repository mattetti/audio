package wav

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/mattetti/audio"
)

// PCM is a data structure representing the underlying PCM data.
type PCM struct {
	r          io.ReadSeeker
	channels   int
	bitDepth   int
	sampleRate int64
	byteSize   int

	sampleFrames int64
	readFrames   int64
	blockSize    int
}

// Offset returns the current frame offset
func (c *PCM) Offset() int64 {
	return c.readFrames
}

// FullBuffer is an inneficient way to access all the PCM data contained in the
// audio container. The entire PCM data is held in memory.
// Consider using Buffer() instead.
func (c *PCM) FullBuffer() (*audio.PCMBuffer, error) {
	format := &audio.Format{
		Channels:   c.channels,
		SampleRate: int(c.sampleRate),
		BitDepth:   int(c.bitDepth),
		Endianness: binary.BigEndian,
	}

	buf := audio.NewPCMIntBuffer(make([]int, 4096), format)

	bytesPerSample := (c.bitDepth-1)/8 + 1
	sampleBufData := make([]byte, bytesPerSample)
	decodeF, err := sampleDecodeFunc(c.bitDepth)
	if err != nil {
		return nil, fmt.Errorf("could not get sample decode func %v", err)
	}

	i := 0
	for err == nil {
		_, err = c.r.Read(sampleBufData)
		if err != nil {
			break
		}
		buf.Ints[i] = decodeF(sampleBufData)
		i++
		// grow the underlying slice if needed
		if i == len(buf.Ints) {
			buf.Ints = append(buf.Ints, make([]int, 4096)...)
		}
	}
	buf.Ints = buf.Ints[:i]

	if err == io.EOF {
		err = nil
	}

	return buf, err
}

// Buffer populates the passed PCM buffer
func (c *PCM) Buffer(buf *audio.PCMBuffer) error {
	if buf == nil {
		return nil
	}
	// TODO: avoid a potentially unecessary allocation
	format := &audio.Format{
		Channels:   c.channels,
		SampleRate: int(c.sampleRate),
		BitDepth:   int(c.bitDepth),
		Endianness: binary.BigEndian,
	}

	bytesPerSample := (c.bitDepth-1)/8 + 1
	sampleBufData := make([]byte, bytesPerSample)
	decodeF, err := sampleDecodeFunc(c.bitDepth)
	if err != nil {
		return fmt.Errorf("could not get sample decode func %v", err)
	}

	// Note that we populate the buffer even if the
	// size of the buffer doesn't fit an even number of frames.
	for i := 0; i < len(buf.Ints); i++ {
		_, err = c.r.Read(sampleBufData)
		if err != nil {
			break
		}
		buf.Ints[i] = decodeF(sampleBufData)
	}
	if err == io.EOF {
		err = nil
	}
	buf.Format = format
	if buf.DataType != audio.Integer {
		buf.DataType = audio.Integer
	}

	return err
}

// Ints reads the PCM data and loads it into the passed frames.
// The number of frames read is returned so the caller can process
// only the populated frames.
// DEPRECATED
func (c *PCM) Ints(frames audio.FramesInt) (n int, err error) {
	bytesPerSample := (c.bitDepth-1)/8 + 1
	sampleBufData := make([]byte, bytesPerSample)
	decodeF, err := sampleDecodeFunc(c.bitDepth)
	if err != nil {
		return 0, fmt.Errorf("could not get sample decode func %v", err)
	}

outter:
	for i := 0; i+c.channels <= len(frames); {
		for j := 0; j < c.channels; j++ {
			_, err = c.r.Read(sampleBufData)
			if err != nil {
				break outter
			}
			frames[i] = decodeF(sampleBufData)
			i++
		}
		n++
	}
	if err == io.EOF {
		err = nil
	}
	return n, err
}

// NextInts returns the n next audio frames
// DEPRECATED
func (c *PCM) NextInts(n int) (audio.FramesInt, error) {
	frames := make(audio.FramesInt, n)
	n, err := c.Ints(frames)
	return frames[:n], err
}

// Float64s reads the PCM data and loads it into the passed frames.
// The number of frames read is returned so the caller can process
// only the populated frames.
// DEPRECATED
func (c *PCM) Float64s(frames audio.FramesFloat64) (n int, err error) {
	bytesPerSample := (c.bitDepth-1)/8 + 1
	sampleBufData := make([]byte, bytesPerSample)
	decodeF, err := sampleFloat64DecodeFunc(c.bitDepth)
	if err != nil {
		return 0, fmt.Errorf("could not get sample decode func %v", err)
	}

	for i := 0; i < len(frames); i++ {
		_, err = c.r.Read(sampleBufData)
		if err != nil {
			break
		}
		frames[i] = decodeF(sampleBufData)
		n++
	}
	if err == io.EOF {
		err = nil
	}
	return n, err
}

// Info returns the frame info for the PCM data.
func (c *PCM) Info() (numChannels, bitDepth int, sampleRate int64, err error) {
	return c.channels, c.bitDepth, c.sampleRate, nil
}

// NextFloat64s returns the n next audio frames.
// DEPRECATED
func (c *PCM) NextFloat64s(n int) (audio.FramesFloat64, error) {
	frames := make(audio.FramesFloat64, n)
	n, err := c.Float64s(frames)
	return frames[:n], err
}

// Read reads frames into the passed buffer and returns the number of full frames
// read.
// DEPRECATED
func (c *PCM) Read(buf []byte) (n int, err error) {
	if c == nil || c.Size() == 0 {
		return n, nil
	}

	bytesPerSample := (c.bitDepth-1)/8 + 1
	sampleBufData := make([]byte, bytesPerSample)
	frameSize := (bytesPerSample * c.channels)

	startingAtFrame := c.readFrames
	if startingAtFrame >= c.sampleFrames {
		return 0, nil
	}

outter:
	for i := 0; i+frameSize < len(buf); {
		for j := 0; j < c.channels; j++ {
			_, err := c.r.Read(sampleBufData)
			if err != nil {
				if err == io.EOF {
					err = nil
				}
				break outter
			}
			for _, b := range sampleBufData {
				buf[i] = b
				i++
			}
		}
		c.readFrames++
		if c.readFrames >= c.sampleFrames {
			break
		}
	}

	n = int(c.readFrames - startingAtFrame)
	return n, err
}

// Size returns the total number of frames available in the PCM data.
func (c *PCM) Size() int64 {
	if c == nil {
		return 0
	}
	if c.sampleFrames != 0 {
		return c.sampleFrames
	}
	bytesPerSample := (c.bitDepth-1)/8 + 1
	frameSize := (bytesPerSample * c.channels)
	c.sampleFrames = int64(c.blockSize / frameSize)
	return c.sampleFrames
}

// SeekFrames seeks to the frame offset
func (c *PCM) SeekFrames(offset int64, whence int) (int64, error) {
	if c == nil {
		return 0, nil
	}

	bytesPerSample := (c.bitDepth-1)/8 + 1
	frameSize := int64(bytesPerSample * c.channels)
	switch whence {
	case 0:
		c.readFrames = offset
	case 1:
		c.readFrames += offset
	case 2:
		c.readFrames = c.Size() - offset
	}
	return c.r.Seek(offset*frameSize, whence)
}
