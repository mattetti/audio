package wav

import (
	"fmt"
	"io"
	"time"

	"github.com/mattetti/audio"
	"github.com/mattetti/audio/riff"
)

// Decoder handles the decoding of wav files.
type Decoder struct {
	r      io.Reader
	parser *riff.Parser
	info   *Info
}

// New creates a decoder for the passed wav reader.
// Note that the reader doesn't get rewinded as the container is processed.
func NewDecoder(r io.Reader, c chan *audio.Chunk) *Decoder {
	return &Decoder{
		r:      r,
		parser: riff.New(r),
	}
}

// Decode reads from a Read Seeker and converts the input to a PCM
// clip output.
func Decode(r io.ReadSeeker) (audio.Clip, error) {
	d := &Decoder{r: r, parser: riff.New(r)}
	nfo, err := d.Info()
	if err != nil {
		return nil, err
	}
	clip := &Clip{
		r:          r,
		byteSize:   int(nfo.Duration.Seconds()) * int(time.Second) * int(nfo.AvgBytesPerSec),
		channels:   int(nfo.NumChannels),
		bitDepth:   int(nfo.NumChannels),
		sampleRate: int64(nfo.SampleRate),
	}

	return clip, nil
}

// Parse reads the content of the file, populates the decoder fields
// and pass the chunks to the provided channel.
// Note that the channel consumer needs to call Done() on the chunk
// to release the wait group and deain the chunk if needed.
func (d *Decoder) Parse(ch chan *riff.Chunk) error {
	d.parser.Chan = ch

	if err := d.parser.Parse(); err != nil {
		return err
	}

	return nil
}

// Info returns the generic file information.
// Note that the information is cached can be called multiple times safely.
func (d *Decoder) Info() (*Info, error) {
	if d.info != nil {
		return d.info, nil
	}

	if d.parser.WavAudioFormat == 0 {
		if err := d.parser.Parse(); err != nil {
			return nil, err
		}
	}

	d.info = &Info{
		NumChannels:    d.parser.NumChannels,
		SampleRate:     d.parser.SampleRate,
		AvgBytesPerSec: d.parser.AvgBytesPerSec,
		BitsPerSample:  d.parser.BitsPerSample,
		WavAudioFormat: d.parser.WavAudioFormat,
	}
	var err error
	d.info.Duration, err = d.Duration()

	return d.info, err
}

// Duration returns the time duration of the decoded wav file.
func (d *Decoder) Duration() (time.Duration, error) {
	return d.parser.Duration()
}

// DecodeRawPCM converts a 'data' wav RAW PCM chunk into frames of samples.
// Each frame can contain one or more channels with their own value.
func (d *Decoder) DecodeRawPCM(chunk *riff.Chunk) ([][]int, error) {
	if chunk.ID != riff.DataFormatID {
		return nil, fmt.Errorf("can't decode chunk with ID %s as PCM data", string(chunk.ID[:]))
	}

	// Multi-channel digital audio samples are stored as interlaced wave data which simply means
	// that the audio samples of a multi-channel (such as stereo and surround)
	// wave file are stored by cycling through the audio samples for each channel
	// before advancing to the next sample time.
	// This is done so that the audio files can be played or streamed before the entire file can be read.
	// This is handy when playing a large file from disk (that may not completely fit into memory) or
	// streaming a file over the Internet.
	//
	// One point about sample data that may cause some confusion is that when samples are represented with 8-bits,
	// they are specified as unsigned values. All other sample bit-sizes are specified as signed values.
	// For example a 16-bit sample can range from -32,768 to +32,767 with a mid-point (silence) at 0.

	bytesPerSample := int(d.parser.BitsPerSample / 8)
	numSamples := chunk.Size / bytesPerSample
	numFrames := numSamples / int(d.parser.NumChannels)
	sndDataFrames := make([][]int, numFrames)

	decodeF, err := sampleDecodeFunc(d.parser.BitsPerSample)
	if err != nil {
		return nil, fmt.Errorf("could not get sample decode func %v", err)
	}
	sBuf := make([]byte, d.parser.BitsPerSample/8)

	for i := 0; i < numFrames; i++ {
		sndDataFrames[i] = make([]int, d.parser.NumChannels)
		for j := uint16(0); j < d.parser.NumChannels; j++ {

			if err := chunk.ReadLE(&sBuf); err != nil {
				return sndDataFrames, fmt.Errorf("failed to read sample %v", err)
			}
			sndDataFrames[i][j] = decodeF(sBuf)
		}
	}

	return sndDataFrames, nil
}

// ReadFrames decodes the file and returns its info and the audio frames
func (d *Decoder) ReadFrames() (info *Info, sndDataFrames [][]int, err error) {
	ch := make(chan *riff.Chunk)
	d.parser.Chan = ch

	go func() {
		if err := d.parser.Parse(); err != nil {
			panic(err)
		}
	}()

	for chunk := range ch {
		if chunk.ID == riff.DataFormatID {
			sndDataFrames, err = d.DecodeRawPCM(chunk)
		}
		chunk.Wg.Done()
	}

	info, err = d.Info()
	return info, sndDataFrames, err
}

func (d *Decoder) Frames() (info *Info, frames audio.Frames, err error) {
	var fs [][]int
	info, fs, err = d.ReadFrames()
	return info, audio.Frames(fs), err
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
