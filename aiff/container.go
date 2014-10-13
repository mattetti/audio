package aiff

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"
)

// Container is the wrapper structure for the AIFF container
type Container struct {
	r io.Reader
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

// ParseHeaders reads the header of the passed container and populat the container with parsed info.
// Note that this code advances the container reader.
func (c *Container) ParseHeaders() error {
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
		// we don't support other chunks ATM, skip them all
		// TODO: push data to an optional channel
		if err := c.jumpTo(int(size)); err != nil {
			return err
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
func (c *Container) Duration() (time.Duration, error) {
	if c == nil {
		return 0, errors.New("can't calculate the duration of a nil pointer")
	}
	duration := time.Duration(float64(c.NumSampleFrames) / float64(c.SampleRate) * float64(time.Second))
	return duration, nil
}

// String implements the Stringer interface.
func (c *Container) String() string {
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
func (c *Container) IDnSize() ([4]byte, uint32, error) {
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
func (c *Container) jumpTo(bytesAhead int) error {
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
