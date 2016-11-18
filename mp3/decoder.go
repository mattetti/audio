package mp3

import (
	"bytes"
	"fmt"
	"io"
	"log"

	"github.com/mattetti/audio/mp3/id3v2"
)

// Decoder operates on a reader and extracts important information
// See http://www.mp3-converter.com/mp3codec/mp3_anatomy.htm
type Decoder struct {
	r   io.Reader
	err error

	id3v2tag *id3v2.Tag
}

// NewDecoder creates a new reader reading the given reader and parsing its data.
// It is the caller's responsibility to call Close on the reader when done.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// SeemsValid checks if the mp3 file looks like a valid mp3 file by looking at the first few bytes.
// The data can be corrupt but at least the header seems alright.
// It is the caller's responsibility to rewind/close the reader when done.
func SeemsValid(r io.Reader) bool {
	buf := make([]byte, 3)
	n, err := r.Read(buf)
	if err != nil {
		return false
	}
	if n != 3 {
		return false
	}
	// MP3 file with an ID3v2 container
	if bytes.Compare(buf, []byte{0x49, 0x44, 0x33}) == 0 {
		return true
	}
	// MPEG-1 Layer 3 file without an ID3 tag or with an ID3v1 tag (which's appended at the end of the file)
	if bytes.Compare(buf[:2], []byte{0xFF, 0xFB}) == 0 {
		return true
	}
	return false
}

// Next decodes the next frame into the provided frame structure.
func (d *Decoder) Next(f *Frame) error {
	if f == nil {
		return fmt.Errorf("can't decode to a nil Frame")
	}

	hLen := 4
	if f.buf == nil {
		f.buf = make([]byte, hLen)
	} else {
		f.buf = f.buf[:hLen]
	}

	n, err := io.ReadFull(d.r, f.buf)
	switch {
	case err != nil:
		return err
	case n != 4:
		return ErrPrematureEOF
	}

	skipped := 0
	// read first 2 bytes
	for {
		if f.buf[0] == 0xFF &&
			(f.buf[1]&0xE0 == 0xE0) &&
			f.Header().Emphasis() != EmphReserved &&
			f.Header().Layer() != LayerReserved &&
			f.Header().Version() != MPEGReserved &&
			f.Header().SampleRate() != -1 &&
			f.Header().BitRate() != -1 {
			break
		}
		switch {
		// skip first byte
		case f.buf[1] == 0xFF:
			f.buf[0] = f.buf[1]
			skipped++
			_, err = io.ReadFull(d.r, f.buf[1:])
		// discard both bytes
		default:
			skipped += 2
			_, err = io.ReadFull(d.r, f.buf)
		}
		if err != nil {
			if skipped != 0 {
				log.Printf("Skipped %v bytes\n", skipped)
			}
			return err
		}
	}

	// CRC check
	crcLen := 0
	if f.Header().Protection() {
		crcLen = 2
		f.buf = append(f.buf, make([]byte, crcLen)...)
		n, err = io.ReadFull(d.r, f.buf[hLen:hLen+crcLen])
		if n != crcLen {
			return ErrPrematureEOF
		}
	}

	sideLen := f.SideInfoLength()
	f.buf = append(f.buf, make([]byte, sideLen)...)

	n, err = io.ReadFull(d.r, f.buf[hLen+crcLen:hLen+crcLen+sideLen])
	if n != sideLen {
		return ErrPrematureEOF
	}

	dataLen := f.Size()
	f.buf = append(f.buf, make([]byte, dataLen-len(f.buf))...)
	f.buf = f.buf[0:dataLen]

	n, err = io.ReadFull(d.r, f.buf[hLen+crcLen+sideLen:dataLen])
	if n != dataLen-hLen-sideLen-crcLen {
		return ErrPrematureEOF
	}

	return nil
}
