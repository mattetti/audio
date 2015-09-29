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
