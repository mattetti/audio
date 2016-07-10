package wav

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/mattetti/audio/misc"
)

// Decoder is the wrapper structure for the WAV container
type Decoder struct {
	r io.ReadSeeker

	err      error
	clipInfo *Clip
}

// NewDecoder creates a new reader reading the given reader and pushing audio data to the given channel.
// It is the caller's responsibility to call Close on the Decoder when done.
func NewDecoder(r io.ReadSeeker) *Decoder {
	return &Decoder{r: r}
}

// Err returns the first non-EOF error that was encountered by the Decoder.
func (d *Decoder) Err() error {
	if d.err == io.EOF {
		return nil
	}
	return d.err
}

// Clip returns the audio Clip information including a reader to reads its content.
// This method is safe to be called multiple times but the reader might need to be rewinded
// if previously read.
// This is the recommended, default way to consume an AIFF file.
func (d *Decoder) Clip() *Clip {
	if d.clipInfo != nil {
		return d.clipInfo
	}
	if d.err = d.readHeaders(); d.err != nil {
		d.err = fmt.Errorf("failed to read header - %v", d.err)
		return nil
	}

	d.clipInfo = &Clip{}

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
	panic("not implemented")
}

// Duration returns the time duration for the current AIFF container
func (d *Decoder) Duration() (time.Duration, error) {
	if d == nil {
		return 0, errors.New("can't calculate the duration of a nil pointer")
	}
	panic("not implemented")
}

// String implements the Stringer interface.
func (d *Decoder) String() string {
	panic("not implemented")
}

// readHeaders is safe to call multiple times
func (d *Decoder) readHeaders() error {
	panic("not implemented")
}
