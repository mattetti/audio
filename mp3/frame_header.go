package mp3

import "fmt"

type (
	// FrameHeader represents the entire header of a frame
	FrameHeader []byte
)

// Version returns the MPEG version from the header
func (h FrameHeader) Version() FrameVersion {
	return FrameVersion((h[1] >> 3) & 0x03)
}

// Layer returns the MPEG layer from the header
func (h FrameHeader) Layer() FrameLayer {
	return FrameLayer((h[1] >> 1) & 0x03)
}

// Protection indicates if there is a CRC present after the header (before the side data)
func (h FrameHeader) Protection() bool {
	return (h[1] & 0x01) != 0x01
}

// BitRate returns the calculated bit rate from the header
func (h FrameHeader) BitRate() FrameBitRate {
	bitrateIdx := (h[2] >> 4) & 0x0F
	if bitrateIdx == 0x0F {
		return ErrInvalidBitrate
	}
	br := bitrates[h.Version()][h.Layer()][bitrateIdx] * 1000
	if br == 0 {
		return ErrInvalidBitrate
	}
	return FrameBitRate(br)
}

// SampleRate returns the samplerate from the header
func (h FrameHeader) SampleRate() FrameSampleRate {
	sri := (h[2] >> 2) & 0x03
	if sri == 0x03 {
		return ErrInvalidSampleRate
	}
	return FrameSampleRate(sampleRates[h.Version()][sri])
}

// Pad returns the pad bit, indicating if there are extra samples
// in this frame to make up the correct bitrate
func (h FrameHeader) Pad() bool {
	return ((h[2] >> 1) & 0x01) == 0x01
}

// Private retrusn the Private bit from the header
func (h FrameHeader) Private() bool {
	return (h[2] & 0x01) == 0x01
}

// ChannelMode returns the channel mode from the header
func (h FrameHeader) ChannelMode() FrameChannelMode {
	return FrameChannelMode((h[3] >> 6) & 0x03)
}

// CopyRight returns the CopyRight bit from the header
func (h FrameHeader) CopyRight() bool {
	return (h[3]>>3)&0x01 == 0x01
}

// Original returns the "original content" bit from the header
func (h FrameHeader) Original() bool {
	return (h[3]>>2)&0x01 == 0x01
}

// Emphasis returns the Emphasis from the header
func (h FrameHeader) Emphasis() FrameEmphasis {
	return FrameEmphasis((h[3] & 0x03))
}

// String dumps the frame header as a string for display purposes
func (h FrameHeader) String() string {
	str := ""
	str += fmt.Sprintf(" Layer: %v\n", h.Layer())
	str += fmt.Sprintf(" Version: %v\n", h.Version())
	str += fmt.Sprintf(" Protection: %v\n", h.Protection())
	str += fmt.Sprintf(" BitRate: %v\n", h.BitRate())
	str += fmt.Sprintf(" SampleRate: %v\n", h.SampleRate())
	str += fmt.Sprintf(" Pad: %v\n", h.Pad())
	str += fmt.Sprintf(" Private: %v\n", h.Private())
	str += fmt.Sprintf(" ChannelMode: %v\n", h.ChannelMode())
	str += fmt.Sprintf(" CopyRight: %v\n", h.CopyRight())
	str += fmt.Sprintf(" Original: %v\n", h.Original())
	str += fmt.Sprintf(" Emphasis: %v\n", h.Emphasis())
	return str
}
