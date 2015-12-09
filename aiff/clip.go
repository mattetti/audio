package aiff

import (
	"io"

	"github.com/mattetti/audio"
)

type Clip struct {
	r          io.ReadSeeker
	size       int64
	channels   int
	bitDepth   int
	sampleRate int64
}

func (c *Clip) Read(p []byte) (n int, err error) {
	return
}

func (c *Clip) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (c *Clip) FrameInfo() audio.FrameInfo {
	return audio.FrameInfo{
		Channels:   c.channels,
		BitDepth:   c.bitDepth,
		SampleRate: c.sampleRate,
	}
}

func (c *Clip) Size() int64 {
	return c.size
}
