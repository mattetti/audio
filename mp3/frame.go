package mp3

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"time"
)

type Frame struct {
	buf    []byte
	header FrameHeader
}

type (
	// FrameVersion is the MPEG version given in the frame header
	FrameVersion byte
	// FrameLayer is the MPEG layer given in the frame header
	FrameLayer byte
	// FrameEmphasis is the Emphasis value from the frame header
	FrameEmphasis byte
	// FrameChannelMode is the Channel mode from the frame header
	FrameChannelMode byte
	// FrameBitRate is the bit rate from the frame header
	FrameBitRate int
	// FrameSampleRate is the sample rate from teh frame header
	FrameSampleRate int
	// FrameSideInfo holds the SideInfo bytes from the frame
	FrameSideInfo []byte
)

// Header returns the header for this frame
func (f *Frame) Header() FrameHeader {
	if f.header == nil || len(f.header) == 0 {
		f.header = FrameHeader(f.buf[0:4])
	}
	return f.header
}

// SideInfoLength returns the expected side info length
//  the side information is 17 bytes in length for a single
// channel encoded file and 32 bytes for dual channel mode. Information in
// side information allows decoding the main data correctly.
func (f *Frame) SideInfoLength() int {
	switch f.Header().Version() {
	case MPEG1:
		switch f.Header().ChannelMode() {
		case SingleChannel:
			return 17
		case Stereo, JointStereo, DualChannel:
			return 32
		default:
			panic("bad channel mode")
		}
	case MPEG2, MPEG25:
		switch f.Header().ChannelMode() {
		case SingleChannel:
			return 9
		case Stereo, JointStereo, DualChannel:
			return 17
		default:
			panic("bad channel mode")
		}
	default:
		log.Println(f.Header())
		panic("unknown mpeg version")
	}
}

// Size calculates the expected size of this frame in bytes based on the header
// information
func (f *Frame) Size() int {
	bps := float64(f.Samples()) / 8
	fsize := (bps * float64(f.Header().BitRate())) / float64(f.Header().SampleRate())
	if f.Header().Pad() {
		fsize += float64(slotSize[f.Header().Layer()])
	}
	return int(fsize)
}

// Samples determines the number of samples based on the MPEG version and Layer from the header
func (f *Frame) Samples() int {
	return samplesPerFrame[f.Header().Version()][f.Header().Layer()]
}

// Duration calculates the time duration of this frame based on the samplerate and number of samples
func (f *Frame) Duration() time.Duration {
	ms := (1000 / float64(f.Header().SampleRate())) * float64(f.Samples())
	return time.Duration(int(float64(time.Millisecond) * ms))
}

// CRC returns the CRC word stored in this frame
func (f *Frame) CRC() uint16 {
	var crc uint16
	if !f.Header().Protection() {
		return 0
	}
	crcdata := bytes.NewReader(f.buf[4:6])
	//log.Println(f.buf[4:6])
	// TODO: check error (add it to decoder)
	binary.Read(crcdata, binary.BigEndian, &crc)
	//log.Println(err)
	return crc
}

// SideInfo returns the  side info for this frame
func (f *Frame) SideInfo() FrameSideInfo {
	if f.Header().Protection() {
		return FrameSideInfo(f.buf[6:])
	} else {
		return FrameSideInfo(f.buf[4:])
	}
}

// Frame returns a string describing this frame, header and side info
func (f *Frame) String() string {
	str := ""
	str += fmt.Sprintf("Header: \n%s", f.Header())
	str += fmt.Sprintf("SideInfo: \n%s", f.SideInfo())
	str += fmt.Sprintf("CRC: %x\n", f.CRC())
	str += fmt.Sprintf("Samples: %v\n", f.Samples())
	str += fmt.Sprintf("Size: %v\n", f.Size())
	str += fmt.Sprintf("Duration: %v\n", f.Duration())
	return str
}

// NDataBegin is the number of bytes before the frame header at which the sample data begins
// 0 indicates that the data begins after the side channel information. This data is the
// data from the "bit resevoir" and can be up to 511 bytes
func (i FrameSideInfo) NDataBegin() uint16 {
	return (uint16(i[0]) << 1 & (uint16(i[1]) >> 7))
}
