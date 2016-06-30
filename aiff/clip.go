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
	// TODO(mattetti): should this return raw bytes or PCM data
	// TODO(mattetti): the underlying reader might pass the size limit, we probably
	// need to use some sort of limitreader.
	return
}

// Seek seeks into the clip
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

func (c *Clip) Size() int64 {
	return c.size
}
