package caf

import (
	"errors"
	"io"
)

var (
	fileHeaderID = [4]byte{'c', 'a', 'f', 'f'}

	// Chunk IDS
	StreamDescriptionChunkID = "desc"
	AudioDataChunkID         = "data"
	ChannelLayoutChunkID     = "chan"
	FillerChunkID            = "free"
	MarkerChunkID            = "mark"
	RegionChunkID            = "regn"
	InstrumentChunkID        = "inst"
	MagicCookieID            = "kuki"
	InfoStringsChunkID       = "info"
	EditCommentsChunkID      = "edct"
	PacketTableChunkID       = "pakt"
	StringsChunkID           = "strg"
	UUIDChunkID              = "uuid"
	PeakChunkID              = "peak"
	OverviewChunkID          = "ovvw"
	MIDIChunkID              = "midi"
	UMIDChunkID              = "umid"
	FormatListID             = "ldsc"
	iXMLChunkID              = "iXML"

	// ErrFmtNotSupported is a generic error reporting an unknown format.
	ErrFmtNotSupported = errors.New("format not supported")
	// ErrUnexpectedData is a generic error reporting that the parser encountered unexpected data.
	ErrUnexpectedData = errors.New("unexpected data content")
)

func New(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

//func NewParser(r io.Reader, ch chan *TBD) *Decoder {
//return &Decoder{r: r, Ch: ch}
//}
