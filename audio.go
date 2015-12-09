package audio

import "io"

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

// Clip represents a linear PCM formatted audio io.ReadSeeker.
// Clip can seek and read from a section and allow users to
// consume a small section of the underlying audio data.
//
// FrameInfo returns the basic frame-level information about the clip audio.
//
// Size returns the total number of bytes of the underlying audio data.
// TODO(jbd): Support cases where size is unknown?
type Clip interface {
	io.ReadSeeker
	FrameInfo() FrameInfo
	Size() int64
}
