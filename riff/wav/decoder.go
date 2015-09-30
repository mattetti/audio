package wav

import (
	"io"
	"time"

	"github.com/mattetti/audio/riff"
)

// Decoder handles the decoding of wav files.
type Decoder struct {
	r      io.Reader
	parser *riff.Parser
	Info   *Info
}

// New creates a parser for the passed wav reader.
// Note that the reader doesn't get rewinded as the container is processed.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r:      r,
		parser: riff.New(r),
	}
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

	d.Info = &Info{
		NumChannels:    d.parser.NumChannels,
		SampleRate:     d.parser.SampleRate,
		AvgBytesPerSec: d.parser.AvgBytesPerSec,
		BitsPerSample:  d.parser.BitsPerSample,
	}

	return nil
}

// Duration returns the time duration of the decoded wav file.
func (d *Decoder) Duration() (time.Duration, error) {
	return d.parser.Duration()
}

// ReadFrames decodes the file and returns its info and the audio frames
func (d *Decoder) ReadFrames() (*Info, [][]int, error) {
	ch := make(chan *riff.Chunk)
	d.parser.Chan = ch

	go func() {
		if err := d.parser.Parse(); err != nil {
			panic(err)
		}
	}()

	var sndDataFrames [][]int
	for chunk := range ch {
		id := string(chunk.ID[:])
		if id == "data" {
			// decode LPCM data
		}
		chunk.Wg.Done()
	}

	d.Info = &Info{
		NumChannels:    d.parser.NumChannels,
		SampleRate:     d.parser.SampleRate,
		AvgBytesPerSec: d.parser.AvgBytesPerSec,
		BitsPerSample:  d.parser.BitsPerSample,
	}

	return d.Info, sndDataFrames, nil
}
