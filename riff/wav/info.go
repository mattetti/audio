package wav

import (
	"fmt"
	"time"
)

// Info represents the metadata of the wav file
type Info struct {
	// NumChannels is the number of channels represented in the waveform data:
	// 1 for mono or 2 for stereo.
	// Audio: Mono = 1, Stereo = 2, etc.
	// The EBU has defined the Multi-channel Broadcast Wave
	// Format [4] where more than two channels of audio are required.
	NumChannels uint16
	// SampleRate The sampling rate (in sample per second) at which each channel should be played.
	// 8000, 44100, etc.
	SampleRate uint32
	// AvgBytesPerSec The average number of bytes per second at which the waveform data should be
	// transferred. Playback software can estimate the buffer size using this value.
	// SampleRate * NumChannels * BitsPerSample/8
	AvgBytesPerSec uint32
	// BitsPerSample 8, 16, 24...
	// Only available for PCM
	// The <nBitsPerSample> field specifies the number of bits of data used to represent each sample of
	// each channel. If there are multiple channels, the sample size is the same for each channel.
	BitsPerSample uint16
	// Duration of the audio content
	Duration time.Duration
	// A number indicating the WAVE format category of the file. The content of the
	// <format-specific-fields> portion of the ‘fmt’ chunk, and the interpretation of
	// the waveform data, depend on this value.
	// PCM = 1 (i.e. Linear quantization) Values other than 1 indicate some form of compression.
	WavAudioFormat uint16
}

// String implements the Stringer interface
func (i *Info) String() string {
	return fmt.Sprintf("%d Hz @ %d bits, %d channel(s), %d avg bytes/sec, duration: %s", i.SampleRate, i.BitsPerSample, i.NumChannels, i.AvgBytesPerSec, i.Duration)
}

// TODO: create a generic file info interface and make this implement it.
