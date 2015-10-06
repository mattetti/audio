package wav

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/mattetti/audio/misc"
	"github.com/mattetti/audio/riff"
)

type Encoder struct {
	w             io.WriteSeeker
	SampleRate    int
	BitsPerSample int
	NumChannels   int
	Frames        [][]int

	WrittenBytes int
}

// NewEncoder creates a new encoder to create a new aiff file.
// Don't forget to add Frames to the encoder before writing.
func NewEncoder(w io.WriteSeeker, sampleRate, sampleSize, numChans int) *Encoder {
	return &Encoder{
		w:             w,
		SampleRate:    sampleRate,
		BitsPerSample: sampleSize,
		NumChannels:   numChans,
	}
}

// Add serializes and adds the passed value using little endian
func (e *Encoder) Add(src interface{}) error {
	e.WrittenBytes += binary.Size(src)
	return binary.Write(e.w, binary.LittleEndian, src)
}

// AddBE serializes and adds the passed value using big endian
func (e *Encoder) AddBE(src interface{}) error {
	e.WrittenBytes += binary.Size(src)
	return binary.Write(e.w, binary.BigEndian, src)
}

func (e *Encoder) Write() error {
	if e == nil {
		return fmt.Errorf("can't write a nil encoder")
	}
	if e.w == nil {
		return fmt.Errorf("can't write to a nil writer")
	}

	// HEADERS
	// riff ID
	if err := e.Add(riff.RiffID); err != nil {
		return err
	}
	// file size uint32 in BE, to update later on.
	if err := e.Add(uint32(42)); err != nil {
		return err
	}

	// wave headers
	// form
	if err := e.Add(riff.FmtID); err != nil {
		return err
	}
	// chunk size
	if err := e.Add(uint32(16)); err != nil {
		return err
	}
	// wave format
	if err := e.Add(uint16(1)); err != nil {
		return err
	}
	// num channels
	if err := e.Add(e.NumChannels); err != nil {
		return err
	}
	// samplerate
	if err := e.Add(e.SampleRate); err != nil {
		return err
	}
	// avg bytes per sec
	if err := e.Add(uint32(e.SampleRate * e.NumChannels * e.BitsPerSample / 8)); err != nil {
		return err
	}
	// block align
	if err := e.Add(uint16(0)); err != nil {
		return err
	}
	// bits per sample
	if err := e.Add(e.BitsPerSample); err != nil {
		return err
	}

	// TODO add frames

	return nil
}

func (e *Encoder) addFrame(frame []int) error {
	for i := 0; i < e.NumChannels; i++ {
		switch e.BitsPerSample {
		case 8:
			if err := e.Add(uint8(frame[i])); err != nil {
				return err
			}
		case 16:
			if err := e.Add(uint16(frame[i])); err != nil {
				return err
			}
		case 24:
			if err := e.Add(misc.Uint32toUint24Bytes(uint32(frame[i]))); err != nil {
				return err
			}
		case 32:
			if err := e.Add(uint32(frame[i])); err != nil {
				return err
			}
		default:
			return fmt.Errorf("can't add frames of bit size %d", e.BitsPerSample)
		}
	}

	return nil
}
