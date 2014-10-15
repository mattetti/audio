package midi

import (
	"errors"
	"io"
)

var (
	headerChunkID = [4]byte{0x4D, 0x54, 0x68, 0x64}
	trackChunkID  = [4]byte{0x4D, 0x54, 0x72, 0x6B}

	// ErrFmtNotSupported is a generic error reporting an unknown format.
	ErrFmtNotSupported = errors.New("format not supported")
	// ErrUnexpectedData is a generic error reporting that the parser encountered unexpected data.
	ErrUnexpectedData = errors.New("unexpected data content")
)

func New(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

func NewParser(r io.Reader, ch chan *Track) *Decoder {
	return &Decoder{r: r, Ch: ch}
}
