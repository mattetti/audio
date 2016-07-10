package wav

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/mattetti/audio"
	"github.com/mattetti/audio/misc"
	"github.com/mattetti/audio/riff"
)

// Decoder handles the decoding of wav files.
type Decoder struct {
	r      io.ReadSeeker
	parser *riff.Parser

	NumChans   uint16
	BitDepth   uint16
	SampleRate uint32

	AvgBytesPerSec uint32
	WavAudioFormat uint16

	err      error
	clipInfo *Clip
}

// New creates a decoder for the passed wav reader.
// Note that the reader doesn't get rewinded as the container is processed.
func NewDecoder(r io.ReadSeeker, c chan *audio.Chunk) *Decoder {
	return &Decoder{
		r:      r,
		parser: riff.New(r),
	}
}

// Err returns the first non-EOF error that was encountered by the Decoder.
func (d *Decoder) Err() error {
	if d.err == io.EOF {
		return nil
	}
	return d.err
}

// EOF returns positively if the underlying reader reached the end of file.
func (d *Decoder) EOF() bool {
	if d == nil || d.err == io.EOF {
		return true
	}
	return false
}

// Clip returns the audio Clip information including a reader to reads its content.
// This method is safe to be called multiple times but the reader might need to be rewinded
// if previously read.
// This is the recommended, default way to consume an AIFF file.
func (d *Decoder) Clip() *Clip {
	if d.clipInfo != nil {
		return d.clipInfo
	}
	d.err = d.readHeaders()
	if d.err != nil {
		return nil
	}

	var chunk *riff.Chunk
	for d.err == nil {
		chunk, d.err = d.parser.NextChunk()
		if d.err != nil {
			break
		}
		if chunk.ID == riff.DataFormatID {
			break
		}
	}
	if chunk == nil {
		return nil
	}

	d.clipInfo = &Clip{
		r:          d.r,
		byteSize:   chunk.Size,
		channels:   int(d.NumChans),
		bitDepth:   int(d.BitDepth),
		sampleRate: int64(d.SampleRate),
		blockSize:  chunk.Size,
	}

	return d.clipInfo
}

// Frames returns the audio frames contained in reader.
// Notes that this method allocates a lot of memory (depending on the duration of the underlying file).
// Consider using the decoder clip and reading/decoding using a buffer.
func (d *Decoder) Frames() (frames misc.AudioFrames, err error) {
	panic("not implemented")
}

// DecodeFrames decodes PCM bytes into audio frames based on the decoder context
func (d *Decoder) DecodeFrames(data []byte) (frames misc.AudioFrames, err error) {
	numChannels := int(d.NumChans)
	// r := bytes.NewBuffer(data)

	bytesPerSample := int((d.BitDepth-1)/8 + 1)
	// sampleBufData := make([]byte, bytesPerSample)

	frames = make(misc.AudioFrames, len(data)/bytesPerSample)
	for j := 0; j < int(numChannels); j++ {
		frames[j] = make([]int, numChannels)
	}
	n := 0

	// outter:
	for i := 0; (i + (bytesPerSample * numChannels)) <= len(data); {
		frame := make([]int, numChannels)
		for j := 0; j < numChannels; j++ {
			// TODO
			panic("not implemented")
		}
		frames[n] = frame
		n++
	}

	return frames, err
}

// Duration returns the time duration for the current AIFF container
func (d *Decoder) Duration() (time.Duration, error) {
	if d == nil || d.parser == nil {
		return 0, errors.New("can't calculate the duration of a nil pointer")
	}
	return d.parser.Duration()
}

// String implements the Stringer interface.
func (d *Decoder) String() string {
	return d.parser.String()
}

// readHeaders is safe to call multiple times
func (d *Decoder) readHeaders() error {
	if d == nil || d.NumChans > 0 {
		return nil
	}

	id, size, err := d.parser.IDnSize()
	if err != nil {
		return err
	}
	d.parser.ID = id
	if d.parser.ID != riff.RiffID {
		return fmt.Errorf("%s - %s", d.parser.ID, riff.ErrFmtNotSupported)
	}
	d.parser.Size = size
	if err := binary.Read(d.r, binary.BigEndian, &d.parser.Format); err != nil {
		return err
	}

	var chunk *riff.Chunk
	var rewindBytes int64

	for err == nil {
		chunk, err = d.parser.NextChunk()
		if err != nil {
			break
		}

		if chunk.ID == riff.FmtID {
			chunk.DecodeWavHeader(d.parser)
			d.NumChans = d.parser.NumChannels
			d.BitDepth = d.parser.BitsPerSample
			d.SampleRate = d.parser.SampleRate
			d.WavAudioFormat = d.parser.WavAudioFormat
			d.AvgBytesPerSec = d.parser.AvgBytesPerSec

			if rewindBytes > 0 {
				d.r.Seek(-(rewindBytes + int64(chunk.Size)), 1)
			}
			break
		} else {
			// unexpected chunk order
			rewindBytes += int64(chunk.Size)
		}

	}

	return d.err
}

// sampleDecodeFunc returns a function that can be used to convert
// a byte range into an int value based on the amount of bits used per sample.
// Note that 8bit samples are unsigned, all other values are signed.
func sampleDecodeFunc(bitsPerSample uint16) (func([]byte) int, error) {
	bytesPerSample := bitsPerSample / 8
	switch bytesPerSample {
	case 1:
		// 8bit values are unsigned
		return func(s []byte) int {
			return int(uint8(s[0]))
		}, nil
	case 2:
		return func(s []byte) int {
			return int(s[0]) + int(s[1])<<8
		}, nil
	case 3:
		return func(s []byte) int {
			var output int32
			output |= int32(s[2]) << 0
			output |= int32(s[1]) << 8
			output |= int32(s[0]) << 16
			return int(output)
		}, nil
	case 4:
		return func(s []byte) int {
			return int(s[0]) + int(s[1])<<8 + int(s[2])<<16 + int(s[3])<<24
		}, nil
	default:
		return nil, fmt.Errorf("unhandled bytesPerSample! b:%d", bytesPerSample)
	}
}
