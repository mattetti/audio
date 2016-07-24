package aiff

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/mattetti/audio"
)

// static check that PCM struct implements audio.PCM
var _ audio.PCM = (*PCM)(nil)

// PCM represents the PCM data contained in the aiff stream.
type PCM struct {
	r            io.ReadSeeker
	channels     int
	bitDepth     int
	sampleRate   int64
	sampleFrames int64
	readFrames   int64

	// decoder info
	offset     uint32
	blockSize  uint32
	offsetRead bool
}

// Offset returns the current frame offset
func (c *PCM) Offset() int64 {
	return c.readFrames
}

// Size returns the total number of frames available in this clip.
func (c *PCM) Size() int64 {
	if c == nil {
		return 0
	}
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

// Ints reads the PCM data and loads it into the passed frames.
// The number of frames read is returned so the caller can process
// only the populated frames.
func (c *PCM) Ints(frames audio.FramesInt) (n int, err error) {
	if c == nil || c.sampleFrames == 0 {
		return 0, nil
	}
	if err := c.readOffsetBlockSize(); err != nil {
		return 0, err
	}
	// TODO(mattetti): respect offset and block size

	decodeF, err := sampleDecodeFunc(c.bitDepth)
	if err != nil {
		return 0, fmt.Errorf("could not get sample decode func %v", err)
	}

	var v int
	for i := 0; i < len(frames); i++ {
		v, err = decodeF(c.r)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		frames[i] = v
		n++
	}
	return n, err
}

// NextInts returns the n next audio frames
func (c *PCM) NextInts(n int) (audio.FramesInt, error) {
	frames := make(audio.FramesInt, n)
	n, err := c.Ints(frames)
	return frames[:n], err
}

// Float64s reads the PCM data and loads it into the passed frames.
// The number of frames read is returned so the caller can process
// only the populated frames.
func (c *PCM) Float64s(frames audio.FramesFloat64) (n int, err error) {
	decodeF, err := sampleFloat64DecodeFunc(c.bitDepth)
	if err != nil {
		return 0, fmt.Errorf("could not get sample decode func %v", err)
	}

	var v float64
	for i := 0; i < len(frames); i++ {
		v, err = decodeF(c.r)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		frames[i] = v
		n++
	}
	return n, err
}

// NextFloat64s returns the n next audio frames
func (c *PCM) NextFloat64s(n int) (audio.FramesFloat64, error) {
	frames := make(audio.FramesFloat64, n)
	n, err := c.Float64s(frames)
	return frames[:n], err
}

// Next reads up to n frames from the clip.
// The frames as well as the number of full frames read are returned.
// This API is somewhat similar to https://golang.org/pkg/bytes/#Buffer.Next
func (c *PCM) Next(nFrames int) (frames audio.Frames, n int, err error) {
	if c == nil || c.sampleFrames == 0 {
		return nil, 0, nil
	}
	if err := c.readOffsetBlockSize(); err != nil {
		return nil, 0, err
	}
	// TODO(mattetti): respect offset and block size

	bytesPerSample := (c.bitDepth-1)/8 + 1
	sampleBufData := make([]byte, bytesPerSample)
	frames = make(audio.Frames, nFrames)
	for i := 0; i < c.channels; i++ {
		frames[i] = make([]int, c.channels)
	}

outter:
	for frameIDX := 0; frameIDX < nFrames; frameIDX++ {
		if frameIDX > len(frames) {
			break
		}

		frame := make([]int, c.channels)
		for j := 0; j < c.channels; j++ {
			_, err := c.r.Read(sampleBufData)
			if err != nil {
				if err == io.EOF {
					err = nil
				}
				break outter
			}

			sampleBuf := bytes.NewBuffer(sampleBufData)
			switch c.bitDepth {
			case 8:
				var v uint8
				binary.Read(sampleBuf, binary.BigEndian, &v)
				frame[j] = int(v)
			case 16:
				var v int16
				binary.Read(sampleBuf, binary.BigEndian, &v)
				frame[j] = int(v)
			case 24:
				// TODO: check if the conversion might not be inversed depending on
				// the encoding (BE vs LE)
				var output int32
				output |= int32(sampleBufData[2]) << 0
				output |= int32(sampleBufData[1]) << 8
				output |= int32(sampleBufData[0]) << 16
				frame[j] = int(output)
			case 32:
				var v int32
				binary.Read(sampleBuf, binary.BigEndian, &v)
				frame[j] = int(v)
			default:
				err = fmt.Errorf("%v bit depth not supported", c.bitDepth)
				break outter
			}
		}
		frames[frameIDX] = frame
		n++
	}

	return frames, n, err
}

// Read reads frames into the passed buffer and returns the number of full frames
// read.
func (c *PCM) Read(buf []byte) (n int, err error) {
	if c == nil || c.sampleFrames == 0 {
		return n, nil
	}
	if err := c.readOffsetBlockSize(); err != nil {
		return n, err
	}
	// TODO(mattetti): respect offset and block size

	bytesPerSample := (c.bitDepth-1)/8 + 1
	sampleBufData := make([]byte, bytesPerSample)

	frameSize := (bytesPerSample * c.channels)
	// track how many frames we previously read so we don't
	// read past the chunk
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

// Info returns the frame info for the PCM data
func (c *PCM) Info() (numChannels, bitDepth int, sampleRate int64, err error) {
	return c.channels, c.bitDepth, c.sampleRate, nil
}

func (c *PCM) readOffsetBlockSize() error {
	// reading the offset and blocksize should only happen once per chunk
	if c == nil || c.offsetRead == true {
		return nil
	}

	// TODO: endianness might depend on the encoding used to generate the aiff data.
	// check encSowt or encTwos

	if err := binary.Read(c.r, binary.BigEndian, &c.offset); err != nil {
		return err
	}
	if err := binary.Read(c.r, binary.BigEndian, &c.blockSize); err != nil {
		return err
	}

	c.offsetRead = true
	return nil
}
