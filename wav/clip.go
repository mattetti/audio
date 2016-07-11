package wav

import (
	"io"

	"github.com/mattetti/audio"
)

type Clip struct {
	r          io.ReadSeeker
	channels   int
	bitDepth   int
	sampleRate int64
	byteSize   int

	sampleFrames int64
	readFrames   int
	blockSize    int
}

// Read reads frames into the passed buffer and returns the number of full frames
// read.
func (c *Clip) Read(buf []byte) (n int, err error) {
	if c == nil || c.Size() == 0 {
		return n, nil
	}

	bytesPerSample := (c.bitDepth-1)/8 + 1
	sampleBufData := make([]byte, bytesPerSample)
	frameSize := (bytesPerSample * c.channels)

	startingAtFrame := c.readFrames
	if startingAtFrame >= int(c.sampleFrames) {
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
		if c.readFrames >= int(c.sampleFrames) {
			break
		}
	}

	n = c.readFrames - startingAtFrame
	return n, err
}

// Size returns the total number of frames available in this clip.
func (c *Clip) Size() int64 {
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

// Seek seeks into the clip
// TODO(mattetti): Seek offset should be in frames, not bytes
func (c *Clip) Seek(offset int64, whence int) (int64, error) {
	if c == nil {
		return 0, nil
	}

	return c.r.Seek(offset, whence)
}

func (c *Clip) FrameInfo() audio.FrameInfo {
	return audio.FrameInfo{
		Channels:   c.channels,
		BitDepth:   c.bitDepth,
		SampleRate: c.sampleRate,
	}
}
