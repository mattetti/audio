package riff

import (
	"encoding/binary"
	"errors"
	"io"
	"time"
)

// Container is a struct containing the overall container information.
type Container struct {
	r io.Reader
	// Must match RIFF
	ID [4]byte
	// This size is the size of the block
	// controlled by the RIFF header. Normally this equals the file size.
	Size uint32
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

// ParseHeaders reads the header of the passed container and populat the container with parsed info.
// Note that this code advances the container reader.
func (c *Container) ParseHeaders() error {
	id, size, err := c.IDnSize()
	if err != nil {
		return err
	}
	c.ID = id
	c.Size = size
	if err := binary.Read(c.r, binary.BigEndian, &c.Format); err != nil {
		return err
	}

	// Extra header parsing if the format is known
	switch c.Format {
	case wavFormatID:
		if err := c.parseWavHeaders(); err != nil {
			return err
		}
	}

	return nil
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
		return c.wavDuration()
	default:
		return 0, ErrFmtNotSupported
	}
}

// NextChunk returns a convenient structure to parse the next chunk.
// If the container is fully read, io.EOF is returned as an error.
func (c *Container) NextChunk() (*Chunk, error) {
	if c == nil {
		return nil, errors.New("can't calculate the duration of a nil pointer")
	}
	id, size, err := c.IDnSize()
	if err != nil {
		return nil, err
	}
	ch := &Chunk{
		ID:   id,
		Size: int(size),
		R:    c.r,
	}
	return ch, nil
}

// IDnSize returns the next ID + block size
func (c *Container) IDnSize() ([4]byte, uint32, error) {
	var ID [4]byte
	var blockSize uint32
	if err := binary.Read(c.r, binary.BigEndian, &ID); err != nil {
		return ID, blockSize, err
	}
	if err := binary.Read(c.r, binary.LittleEndian, &blockSize); err != err {
		return ID, blockSize, err
	}
	return ID, blockSize, nil
}

// parseWavHeaders parses the fmt chunk that comes right after the RIFF header
// This data is needed to calculate the duration of the file.
func (c *Container) parseWavHeaders() error {
	if c == nil {
		return errors.New("can't calculate the wav duration of a nil pointer")
	}
	if c.ID != riffID {
		return errors.New("headers not parsed, can't get the wav duration")
	}

	id, size, err := c.IDnSize()
	if err != nil {
		return nil
	}
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
func (c *Container) wavDuration() (time.Duration, error) {
	duration := time.Duration((float64(c.Size) / float64(c.ByteRate)) * float64(time.Second))
	return duration, nil
}
