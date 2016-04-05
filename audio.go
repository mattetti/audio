package audio

import (
	"io"
	"sync"
)

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
	ID     [4]byte
	Size   int
	Pos    int
	R      io.Reader
	okChan chan bool
	Wg     *sync.WaitGroup
}

// Clipper represents a linear PCM formatted audio io.ReadSeeker.
// Clipper can seek and read from a section and allow users to
// consume a small section of the underlying audio data.
//
// FrameInfo returns the basic frame-level information about the clip audio.
//
// Size returns the total number of bytes of the underlying audio data.
// TODO(jbd): Support cases where size is unknown?
type Clipper interface {
	io.ReadSeeker
	FrameInfo() FrameInfo
	Size() int64
}

type Clip struct {
	R          io.ReadSeeker
	DataSize   int64
	Channels   int
	BitDepth   int
	SampleRate int64
}

func (c *Clip) Read(p []byte) (n int, err error) {
	return c.R.Read(p)
}

func (c *Clip) Seek(offset int64, whence int) (int64, error) {
	return c.R.Seek(offset, whence)
}

func (c *Clip) FrameInfo() FrameInfo {
	return FrameInfo{
		Channels:   c.Channels,
		BitDepth:   c.BitDepth,
		SampleRate: c.SampleRate,
	}
}

func (c *Clip) Size() int64 {
	return c.DataSize
}
