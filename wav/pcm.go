package wav

import (
	"fmt"
	"io"

	"github.com/mattetti/audio"
)

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

// Ints reads the PCM data and loads it into the passed frames.
// The number of frames read is returned so the caller can process
// only the populated frames.
func (c *PCM) Ints(frames audio.FramesInt) (n int, err error) {
	bytesPerSample := (c.bitDepth-1)/8 + 1
	sampleBufData := make([]byte, bytesPerSample)
	decodeF, err := sampleDecodeFunc(c.bitDepth)
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

// NextInts returns the n next audio frames
func (c *PCM) NextInts(n int) (audio.FramesInt, error) {
	frames := make(audio.FramesInt, n)
	n, err := c.Ints(frames)
	return frames[:n], err
}

// Float32s reads the PCM data and loads it into the passed frames.
// The number of frames read is returned so the caller can process
// only the populated frames.
func (c *PCM) Float32s(frames audio.FramesFloat32) (n int, err error) {
	bytesPerSample := (c.bitDepth-1)/8 + 1
	sampleBufData := make([]byte, bytesPerSample)
	decodeF, err := sampleFloat32DecodeFunc(c.bitDepth)
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

// Info returns the frame info for the PCM data
func (c *PCM) Info() (numChannels, bitDepth int, sampleRate int64, err error) {
	return c.channels, c.bitDepth, c.sampleRate, nil
}

// NextFloat32s returns the n next audio frames
func (c *PCM) NextFloat32s(n int) (audio.FramesFloat32, error) {
	frames := make(audio.FramesFloat32, n)
	n, err := c.Float32s(frames)
	return frames[:n], err
}

// Read reads frames into the passed buffer and returns the number of full frames
// read.
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

// Size returns the total number of frames available in this PCM.
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

// Seek seeks to the frame offset
func (c *PCM) Seek(offset int64, whence int) (int64, error) {
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
