package wav

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/mattetti/audio"
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

	err     error
	pcmClip *PCM
}

// New creates a decoder for the passed wav reader.
// Note that the reader doesn't get rewinded as the container is processed.
func NewDecoder(r io.ReadSeeker) *Decoder {
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

// ReadInfo reads the underlying reader until the comm header is parsed.
// This method is safe to call multiple times.
func (d *Decoder) ReadInfo() {
	d.err = d.readHeaders()
}

// PCM returns an audio.PCM compatible value to consume the PCM data
// contained in the underlying wav data.
func (d *Decoder) PCM() *PCM {
	if d.pcmClip != nil {
		return d.pcmClip
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

	d.pcmClip = &PCM{
		r:          d.r,
		byteSize:   chunk.Size,
		channels:   int(d.NumChans),
		bitDepth:   int(d.BitDepth),
		sampleRate: int64(d.SampleRate),
		blockSize:  chunk.Size,
	}

	return d.pcmClip
}

// NextChunk returns the next available chunk
func (d *Decoder) NextChunk() (*riff.Chunk, error) {
	if d.err = d.readHeaders(); d.err != nil {
		d.err = fmt.Errorf("failed to read header - %v", d.err)
		return nil, d.err
	}

	var (
		id   [4]byte
		size uint32
	)

	id, size, d.err = d.parser.IDnSize()
	if d.err != nil {
		d.err = fmt.Errorf("error reading chunk header - %v", d.err)
		return nil, d.err
	}

	c := &riff.Chunk{
		ID:   id,
		Size: int(size),
		R:    io.LimitReader(d.r, int64(size)),
	}
	return c, d.err
}

// FramesInt returns the audio frames contained in reader.
// Notes that this method allocates a lot of memory (depending on the duration of the underlying file).
// Consider using the decoder clip and reading/decoding using a buffer.
func (d *Decoder) FramesInt() (frames audio.FramesInt, err error) {
	pcm := d.PCM()
	if pcm == nil {
		return nil, fmt.Errorf("no PCM data available")
	}
	totalFrames := int(pcm.Size()) * int(d.NumChans)
	frames = make(audio.FramesInt, totalFrames)
	n, err := pcm.Ints(frames)
	return frames[:n], err
}

// DecodeFrames decodes PCM bytes into audio frames based on the decoder context.
// This function is usually used in conjunction with Clip.Read which returns the amount
// of frames read into the buffer. It's highly recommended to slice the returned frames
// of this function by the amount of total frames reads into the buffer.
// The reason being that if the buffer didn't match the exact size of the frames,
// some of the data might be garbage but will still be converted into frames.
func (d *Decoder) DecodeFrames(data []byte) (frames audio.Frames, err error) {
	numChannels := int(d.NumChans)
	r := bytes.NewBuffer(data)

	bytesPerSample := int((d.BitDepth-1)/8 + 1)
	sampleBufData := make([]byte, bytesPerSample)
	decodeF, err := sampleDecodeFunc(int(d.BitDepth))
	if err != nil {
		return nil, fmt.Errorf("could not get sample decode func %v", err)
	}

	frames = make(audio.Frames, len(data)/bytesPerSample)
	for j := 0; j < int(numChannels); j++ {
		frames[j] = make([]int, numChannels)
	}
	n := 0

outter:
	for i := 0; (i + (bytesPerSample * numChannels)) <= len(data); {
		frame := make([]int, numChannels)
		for j := 0; j < numChannels; j++ {
			_, err = r.Read(sampleBufData)
			if err != nil {
				break outter
			}
			frame[j] = decodeF(sampleBufData)
			i += bytesPerSample
		}
		frames[n] = frame
		n++
	}

	return frames, err
}

// Duration returns the time duration for the current audio container
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
func sampleDecodeFunc(bitsPerSample int) (func([]byte) int, error) {
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
		return nil, fmt.Errorf("unhandled byte depth:%d", bitsPerSample)
	}
}

// sampleDecodeFloat64Func returns a function that can be used to convert
// a byte range into a float64 value based on the amount of bits used per sample.
func sampleFloat64DecodeFunc(bitsPerSample int) (func([]byte) float64, error) {
	bytesPerSample := bitsPerSample / 8
	switch bytesPerSample {
	case 1:
		// 8bit values are unsigned
		return func(s []byte) float64 {
			return float64(uint8(s[0]))
		}, nil
	case 2:
		return func(s []byte) float64 {
			return float64(int(s[0]) + int(s[1])<<8)
		}, nil
	case 3:
		return func(s []byte) float64 {
			var output int32
			output |= int32(s[2]) << 0
			output |= int32(s[1]) << 8
			output |= int32(s[0]) << 16
			return float64(output)
		}, nil
	case 4:
		// TODO: fix the float64 conversion (current int implementation)
		return func(s []byte) float64 {
			return float64(int(s[0]) + int(s[1])<<8 + int(s[2])<<16 + int(s[3])<<24)
		}, nil
	default:
		return nil, fmt.Errorf("unhandled byte depth:%d", bitsPerSample)
	}
}
