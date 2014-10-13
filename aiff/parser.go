package aiff

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"
)

var (
	defaultChunkParserTimeout = 2 * time.Second
)

// Parser is the wrapper structure for the AIFF container
type Parser struct {
	r io.Reader
	// c is an Optional channel of chunks that is used to parse chunks
	Chan chan *Chunk
	// ChunkParserTimeout is the duration after which the main parser keeps going
	// if the dev hasn't reported the chunk parsing to be done.
	// By default: 2s
	ChunkParserTimeout time.Duration
	// The ok channel is used to let the parser that it's ok to continue
	// after a chunk was passed to the optional parser channel.
	okChan chan bool

	// ID is always 'FORM'. This indicates that this is a FORM chunk
	ID [4]byte
	// Size contains the size of data portion of the 'FORM' chunk.
	// Note that the data portion has been
	// broken into two parts, formType and chunks
	Size uint32
	// describes what's in the 'FORM' chunk. For Audio IFF files,
	// formType (aka Format) is always 'AIFF'.
	// This indicates that the chunks within the FORM pertain to sampled sound.
	Format [4]byte

	// Data coming from the COMM chunk
	commSize        uint32
	NumChans        uint16
	NumSampleFrames uint32
	SampleSize      uint16
	SampleRate      int

	// AIFC data
	Encoding     [4]byte
	EncodingName string
}

// Parse reads the aiff reader and populates the container structure with found information.
// The sound data or unknown chunks are passed to the optional channel if available.
func (c *Parser) Parse() error {
	if err := binary.Read(c.r, binary.BigEndian, &c.ID); err != nil {
		return err
	}
	// Must start by a FORM header/ID
	if c.ID != formID {
		return fmt.Errorf("%s - %s", ErrFmtNotSupported, c.ID)
	}

	if err := binary.Read(c.r, binary.BigEndian, &c.Size); err != nil {
		return err
	}
	if err := binary.Read(c.r, binary.BigEndian, &c.Format); err != nil {
		return err
	}

	// Must be a AIFF or AIFC form type
	if c.Format != aiffID && c.Format != aifcID {
		return fmt.Errorf("%s - %s", ErrFmtNotSupported, c.Format)
	}

	id, size, err := c.IDnSize()
	if err != nil {
		return err
	}
	for id != commID {

		if c.Chan != nil {
			okC := make(chan bool)
			c.Chan <- &Chunk{ID: id, Size: int(size), R: c.r, okChan: okC}
			timeout := c.ChunkParserTimeout
			if timeout == 0 {
				timeout = defaultChunkParserTimeout
			}
			for {
				select {
				case <-okC:
					break
				case <-time.After(timeout):
					fmt.Printf(".")
				}
			}
		} else {
			// we don't support other chunks ATM, skip them all
			// TODO: push data to an optional channel
			if err := c.jumpTo(int(size)); err != nil {
				return err
			}
		}
		id, size, err = c.IDnSize()
		if err != nil {
			return err
		}
	}

	c.commSize = size

	if err := binary.Read(c.r, binary.BigEndian, &c.NumChans); err != nil {
		return fmt.Errorf("num of channels failed to parse - %s", err.Error())
	}
	if err := binary.Read(c.r, binary.BigEndian, &c.NumSampleFrames); err != nil {
		return fmt.Errorf("num of sample frames failed to parse - %s", err.Error())
	}
	if err := binary.Read(c.r, binary.BigEndian, &c.SampleSize); err != nil {
		return fmt.Errorf("sample size failed to parse - %s", err.Error())
	}
	var srBytes [10]byte
	if err := binary.Read(c.r, binary.BigEndian, &srBytes); err != nil {
		return fmt.Errorf("sample rate failed to parse - %s", err.Error())
	}
	c.SampleRate = IeeeFloatToInt(srBytes)

	if c.Format == aifcID {
		if err := binary.Read(c.r, binary.BigEndian, &c.Encoding); err != nil {
			return fmt.Errorf("AIFC encoding failed to parse - %s", err)
		}
		// pascal style string with the description of the encoding
		var size uint8
		if err := binary.Read(c.r, binary.BigEndian, &size); err != nil {
			return fmt.Errorf("AIFC encoding failed to parse - %s", err)
		}

		desc := make([]byte, size)
		if err := binary.Read(c.r, binary.BigEndian, &desc); err != nil {
			return fmt.Errorf("AIFC encoding failed to parse - %s", err)
		}
		c.EncodingName = string(desc)
	}

	return nil
}

// Duration returns the time duration for the current AIFF container
func (p *Parser) Duration() (time.Duration, error) {
	if p == nil {
		return 0, errors.New("can't calculate the duration of a nil pointer")
	}
	duration := time.Duration(float64(p.NumSampleFrames) / float64(p.SampleRate) * float64(time.Second))
	return duration, nil
}

// String implements the Stringer interface.
func (c *Parser) String() string {
	out := fmt.Sprintf("Format: %s - ", c.Format)
	if c.Format == aifcID {
		out += fmt.Sprintf("%s - ", c.EncodingName)
	}
	if c.SampleRate != 0 {
		out += fmt.Sprintf("%d channels @ %d / %d bits - ", c.NumChans, c.SampleRate, c.SampleSize)
		d, _ := c.Duration()
		out += fmt.Sprintf("Duration: %f seconds\n", d.Seconds())
	}
	return out
}

// IDnSize returns the next ID + block size
func (c *Parser) IDnSize() ([4]byte, uint32, error) {
	var ID [4]byte
	var blockSize uint32
	if err := binary.Read(c.r, binary.BigEndian, &ID); err != nil {
		return ID, blockSize, err
	}
	if err := binary.Read(c.r, binary.BigEndian, &blockSize); err != err {
		return ID, blockSize, err
	}
	return ID, blockSize, nil
}

// jumpTo advances the reader to the amount of bytes provided
func (c *Parser) jumpTo(bytesAhead int) error {
	var err error
	for bytesAhead > 0 {
		readSize := bytesAhead
		if readSize > 4000 {
			readSize = 4000
		}

		buf := make([]byte, readSize)
		err = binary.Read(c.r, binary.LittleEndian, &buf)
		if err != nil {
			return nil
		}
		bytesAhead -= readSize
	}
	return nil
}
