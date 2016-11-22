package mp3

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/mattetti/audio/mp3/id3v1"
	"github.com/mattetti/audio/mp3/id3v2"
)

// Decoder operates on a reader and extracts important information
// See http://www.mp3-converter.com/mp3codec/mp3_anatomy.htm
type Decoder struct {
	r         io.Reader
	NbrFrames int

	ID3v2tag *id3v2.Tag
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
	if bytes.Compare(buf, id3v2.HeaderTagID) == 0 {
		return true
	}
	// MPEG-1 Layer 3 file without an ID3 tag or with an ID3v1 tag (which's appended at the end of the file)
	if bytes.Compare(buf[:2], ID31HBytes) == 0 {
		return true
	}
	return false
}

// Duration returns the time duration for the current mp3 file
// The entire reader will be consumed, the consumer might want to rewind the reader
// if they want to read more from the feed.
func (d *Decoder) Duration() (time.Duration, error) {
	if d == nil {
		return 0, errors.New("can't calculate the duration of a nil pointer")
	}
	fr := &Frame{}
	var duration time.Duration
	var err error
	for {
		err = d.Next(fr)
		if err != nil {
			// bad headers can be ignored and hopefully skipped
			if err == ErrInvalidHeader {
				continue
			}
			break
		}
		duration += fr.Duration()
		d.NbrFrames++
	}
	if err == io.EOF || err == io.ErrUnexpectedEOF || err == io.ErrShortBuffer {
		err = nil
	}

	return duration, err
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

	_, err := io.ReadAtLeast(d.r, f.buf, hLen)
	if err != nil {
		return err
	}

	// ID3v1 tag at the beggining
	if bytes.Compare(f.buf[:3], id3v1.HeaderTagID) == 0 {
		// the ID3v1 tag is always 128 bytes long, we already read 4 bytes
		// so we need to read the rest.
		buf := make([]byte, 124)
		// TODO: parse the actual header
		n, err := io.ReadFull(d.r, buf)
		if err != nil || n != 124 {
			return ErrInvalidHeader
		}
		buf = append(f.buf, buf...)
		// that wasn't a frame
		f = &Frame{}
		return nil
	}

	// ID3v2 tag
	if bytes.Compare(f.buf[:3], id3v2.HeaderTagID) == 0 {
		d.ID3v2tag = &id3v2.Tag{}
		// we already read 4 bytes, an id3v2 tag header is of zie 10, read the rest
		// and append it to what we already have.
		buf := make([]byte, 6)
		n, err := d.r.Read(buf)
		if err != nil || n != 6 {
			return ErrInvalidHeader
		}
		buf = append(f.buf, buf...)

		th := id3v2.TagHeader{}
		copy(th[:], buf)
		if err = d.ID3v2tag.ReadHeader(th); err != nil {
			return err
		}
		// TODO: parse the actual tag
		// Skip the tag for now
		bytesToSkip := int64(d.ID3v2tag.Header.Size)
		var cn int64
		if cn, err = io.CopyN(ioutil.Discard, d.r, bytesToSkip); cn != bytesToSkip {
			return ErrInvalidHeader
		}
		f = &Frame{}
		return err
	}

	f.Header = FrameHeader(f.buf)

	dataSize := f.Header.Size() - 4
	f.buf = append(f.buf, make([]byte, dataSize)...)
	_, err = io.ReadAtLeast(d.r, f.buf[4:], int(dataSize))
	return err
}
