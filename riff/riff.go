package riff

import (
	"encoding/binary"
	"errors"
	"io"
	"time"
)

var (
	riffID      = [4]byte{'R', 'I', 'F', 'F'}
	fmtID       = [4]byte{'f', 'm', 't', ' '}
	wavFormatID = [4]byte{'W', 'A', 'V', 'E'}
	rmiFormatID = [4]byte{'R', 'M', 'I', 'D'}
	aviFormatID = [4]byte{'A', 'V', 'I', ' '}
	// ErrFmtNotSupported is a generic error reporting an unknown format.
	ErrFmtNotSupported = errors.New("format not supported")
	// ErrUnexpectedData is a generic error reporting that the parser encountered unexpected data.
	ErrUnexpectedData = errors.New("unexpected data content")
)

// NewContainer creates a container wrapper for a reader.
// Note that the reader doesn't get rewinded as the container is processed.
func NewContainer(r io.Reader) *Container {
	return &Container{r: r}
}

// ParseHeaders reads the header of the passed container and populat the container with parsed info.
// Note that this code advances the container reader.
func (c *Container) ParseHeaders() error {
	id, size := c.IDnSize()
	c.ID = id
	c.BlockSize = size
	if err := binary.Read(c.r, binary.BigEndian, &c.Format); err != nil {
		return err
	}

	// Extra header parsing if the format is known
	switch c.Format {
	case wavFormatID:
		if err := c.ParseWavHeaders(); err != nil {
			return err
		}
	}

	return nil
}

// Duration returns the time duration of the passed reader if the sub format is supported.
func Duration(r io.Reader) (time.Duration, error) {
	c := NewContainer(r)
	if err := c.ParseHeaders(); err != nil {
		return 0, err
	}
	return c.Duration()
}

// Container is a struct containing the overall container information.
type Container struct {
	r io.Reader
	// Must match RIFF
	ID [4]byte
	// This size is the size of the block
	// controlled by the RIFF header. Normally this equals the file size.
	BlockSize uint32
	// Format name. This is the format name of the RIFF
	Format [4]byte

	// WAV stuff
	// size of the wav specific fmt header
	wavHeaderSize uint32
	// PCM = 1 (i.e. Linear quantization) Values other than 1 indicate some form of compression.
	WavAudioFormat uint16
	// Audio: Mono = 1, Stereo = 2, etc.
	NumChannels uint16
	// 8000, 44100, etc.
	SampleRate uint32
	// SampleRate * NumChannels * BitsPerSample/8
	ByteRate uint32
	// NumChannels * BitsPerSample/8 The number of bytes for one sample including
	// all channels
	BlockAlign uint16
	// 8, 16, 24...
	BitsPerSample uint16
}

// IDnSize returns the next ID + block size
func (c *Container) IDnSize() ([4]byte, uint32) {
	var ID [4]byte
	var blockSize uint32
	binary.Read(c.r, binary.BigEndian, &ID)
	binary.Read(c.r, binary.LittleEndian, &blockSize)
	return ID, blockSize
}

// Duration returns the time duration for the current RIFF container
// based on the sub format (wav etc...)
func (c *Container) Duration() (time.Duration, error) {
	if c == nil {
		return 0, errors.New("can't calculate the duration of a nil pointer")
	}
	if c.ID == [4]byte{} {
		err := c.ParseHeaders()
		if err != nil {
			return 0, nil
		}
	}
	switch c.Format {
	case wavFormatID:
		return c.WavDuration()
	default:
		return 0, ErrFmtNotSupported
	}
}

// ParseWavHeaders parses the fmt chunk that comes right after the RIFF header
// This data is needed to calculate the duration of the file.
func (c *Container) ParseWavHeaders() error {
	if c == nil {
		return errors.New("can't calculate the wav duration of a nil pointer")
	}
	if c.ID != riffID {
		return errors.New("headers not parsed, can't get the wav duration")
	}

	id, size := c.IDnSize()
	if id != fmtID {
		return ErrUnexpectedData
	}

	c.wavHeaderSize = size
	if err := binary.Read(c.r, binary.LittleEndian, &c.WavAudioFormat); err != nil {
		return err
	}
	if err := binary.Read(c.r, binary.LittleEndian, &c.NumChannels); err != nil {
		return err
	}
	if err := binary.Read(c.r, binary.LittleEndian, &c.SampleRate); err != nil {
		return err
	}
	if err := binary.Read(c.r, binary.LittleEndian, &c.ByteRate); err != nil {
		return err
	}
	if err := binary.Read(c.r, binary.LittleEndian, &c.BlockAlign); err != nil {
		return err
	}
	if err := binary.Read(c.r, binary.LittleEndian, &c.BitsPerSample); err != nil {
		return err
	}

	// if we aren't dealing with a PCM file, we advance to reader to the
	// end of the chunck.
	if size > 16 {
		extra := make([]byte, size-16)
		binary.Read(c.r, binary.LittleEndian, &extra)
	}
	return nil
}

// WavDuration returns the time duration of a wav container.
func (c *Container) WavDuration() (time.Duration, error) {
	duration := time.Duration((float64(c.BlockSize) / float64(c.ByteRate)) * float64(time.Second))
	return duration, nil
}
