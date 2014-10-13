package riff

import (
	"encoding/binary"
	"errors"
	"fmt"
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
	// Format name.
	// The representation of data in <wave-data>, and the content of the <format-specific-fields>
	// of the ‘fmt’ chunk, depend on the format category.
	// 0001h => Microsoft Pulse Code Modulation (PCM) format
	// 0050h => MPEG-1 Audio (audio only)
	Format [4]byte

	// WAV stuff
	// size of the wav specific fmt header
	wavHeaderSize uint32
	// A number indicating the WAVE format category of the file. The content of the
	// <format-specific-fields> portion of the ‘fmt’ chunk, and the interpretation of
	// the waveform data, depend on this value.
	// PCM = 1 (i.e. Linear quantization) Values other than 1 indicate some form of compression.
	WavAudioFormat uint16
	// The number of channels represented in the waveform data: 1 for mono or 2 for stereo.
	// Audio: Mono = 1, Stereo = 2, etc.
	// The EBU has defined the Multi-channel Broadcast Wave
	// Format [4] where more than two channels of audio are required.
	NumChannels uint16
	// The sampling rate (in sample per second) at which each channel should be played.
	// 8000, 44100, etc.
	SampleRate uint32
	// The average number of bytes per second at which the waveform data should be
	// transferred. Playback software can estimate the buffer size using this value.
	// SampleRate * NumChannels * BitsPerSample/8
	AvgBytesPerSec uint32
	// NumChannels * BitsPerSample/8 The number of bytes for one sample including
	// all channels.
	// The block alignment (in bytes) of the waveform data. Playback software needs
	// to process a multiple of <nBlockAlign> bytes of data at a time, so the value of
	// <BlockAlign> can be used for buffer alignment.
	BlockAlign uint16
	// 8, 16, 24...
	// Only available for PCM
	// The <nBitsPerSample> field specifies the number of bits of data used to represent each sample of
	// each channel. If there are multiple channels, the sample size is the same for each channel.
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
	if c.ID != riffID {
		return fmt.Errorf("%s - %s", c.ID, ErrFmtNotSupported)
	}
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

// String implements the Stringer interface.
func (c *Container) String() string {
	out := fmt.Sprintf("Format: %s - ", c.Format)
	if c.Format == wavFormatID {
		out += fmt.Sprintf("%d channels @ %d / %d bits - ", c.NumChannels, c.SampleRate, c.BitsPerSample)
		d, _ := c.Duration()
		out += fmt.Sprintf("Duration: %f seconds\n", d.Seconds())
	}
	return out
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

	for id != fmtID {
		// JUNK chunk should be skipped
		if id == junkID {
			if size%2 == 1 {
				size++
			}
		}
		// BFW: bext chunk described here
		// https://tech.ebu.ch/docs/tech/tech3285.pdf

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
	if err := binary.Read(c.r, binary.LittleEndian, &c.AvgBytesPerSec); err != nil {
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
	duration := time.Duration((float64(c.Size) / float64(c.AvgBytesPerSec)) * float64(time.Second))
	return duration, nil
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
