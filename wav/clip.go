package wav

import (
	"io"

	"github.com/mattetti/audio"
	"github.com/mattetti/audio/misc"
)

type Clip struct {
	r          io.ReadSeeker
	channels   int
	bitDepth   int
	sampleRate int64

	sampleFrames int
	readFrames   int
}

// ReadPCM reads up to n frames from the clip.
// The frames as well as the number of frames/items read are returned.
// TODO(mattetti): misc.AudioFrames is a temporary solution that needs to be improved.
func (c *Clip) ReadPCM(nFrames int) (frames misc.AudioFrames, n int, err error) {
	if c == nil || c.sampleFrames == 0 {
		return nil, 0, nil
	}
	panic("not implemented")
}

// Read reads frames into the passed buffer and returns the number of full frames
// read.
func (c *Clip) Read(buf []byte) (n int, err error) {
	if c == nil || c.sampleFrames == 0 {
		return n, nil
	}
	panic("not implemented")
}

// Size returns the total number of frames available in this clip.
func (c *Clip) Size() int64 {
	if c == nil {
		return 0
	}
	panic("not implemented")
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
