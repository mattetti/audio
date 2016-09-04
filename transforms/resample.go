package transforms

import "github.com/mattetti/audio"

// TODO
func Resample(buf *audio.PCMBuffer, fs float64) error {
	if buf == nil || buf.Format == nil {
		return audio.ErrInvalidBuffer
	}
	// check the target fs
	// if < than the buffer, then decimate (with anti aliasing)
	// otherwise oversample
	panic("not implemented")
}
