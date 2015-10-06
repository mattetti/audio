package aiff

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/mattetti/audio/misc"
)

// Encoder encodes LPCM data into an aiff content.
type Encoder struct {
	w          io.WriteSeeker
	SampleRate int
	SampleSize int
	NumChans   int
	Frames     [][]int

	WrittenBytes int
}

// NewEncoder creates a new encoder to create a new aiff file.
// Don't forget to add Frames to the encoder before writing.
func NewEncoder(w io.WriteSeeker, sampleRate, sampleSize, numChans int) *Encoder {
	return &Encoder{
		w:          w,
		SampleRate: sampleRate,
		SampleSize: sampleSize,
		NumChans:   numChans,
	}
}

// Add serializes and adds the passed value using big endian
func (e *Encoder) Add(src interface{}) error {
	e.WrittenBytes += binary.Size(src)
	return binary.Write(e.w, binary.BigEndian, src)
}

// AddLE serializes and adds the passed value using little endian
func (e *Encoder) AddLE(src interface{}) error {
	e.WrittenBytes += binary.Size(src)
	return binary.Write(e.w, binary.LittleEndian, src)
}

func (e *Encoder) addFrame(frame []int) error {
	for i := 0; i < e.NumChans; i++ {
		switch e.SampleSize {
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
			return fmt.Errorf("can't add frames of bit size %d", e.SampleSize)
		}
	}

	return nil
}

func (e *Encoder) numSampleFrames() int {
	return len(e.Frames)
}

func (e *Encoder) Write() error {
	if e == nil {
		return fmt.Errorf("can't write a nil encoder")
	}
	if e.w == nil {
		return fmt.Errorf("can't write to a nil writer")
	}

	// ID
	if err := e.Add(formID); err != nil {
		return fmt.Errorf("%v when writing FORM header", err)
	}
	// size, will need to be updated later on (total size - 8)
	if err := e.Add(uint32(0)); err != nil {
		return fmt.Errorf("%v when writing size header", err)
	}
	// Format
	if err := e.Add(aiffID); err != nil {
		return fmt.Errorf("%v when writing format header", err)
	}

	// comm chunk
	if err := e.Add(commID); err != nil {
		return fmt.Errorf("%v when writing comm chunk ID header", err)
	}
	// blocksize uint32
	if err := e.Add(uint32(18)); err != nil {
		return fmt.Errorf("%v when writing comm chunk size header", err)
	}
	if err := e.Add(uint16(e.NumChans)); err != nil {
		return fmt.Errorf("%v when writing comm chan numbers", err)
	}
	if err := e.Add(uint32(e.numSampleFrames())); err != nil {
		return fmt.Errorf("%v when writing comm num sample frames", err)
	}
	if err := e.Add(uint16(e.SampleSize)); err != nil {
		return fmt.Errorf("%v when writing comm chan numbers", err)
	}
	// sample rate in IeeeFloat (10 bytes)
	if err := e.Add(misc.IntToIeeeFloat(int(e.SampleRate))); err != nil {
		return fmt.Errorf("%v when writing comm sample rate", err)
	}

	// other chunks
	// audio frames
	if err := e.Add([]byte("SSND")); err != nil {
		return fmt.Errorf("%v when writing SSND chunk ID header", err)
	}

	// blocksize uint32
	chunksize := uint32((int(e.SampleSize)/8)*int(e.NumChans)*len(e.Frames) + 8)
	if err := e.Add(uint32(chunksize)); err != nil {
		return fmt.Errorf("%v when writing SSND chunk size header", err)
	}

	if err := e.Add(uint32(0)); err != nil {
		return fmt.Errorf("%v when writing SSND offset", err)
	}
	if err := e.Add(uint32(0)); err != nil {
		return fmt.Errorf("%v when writing SSND block size", err)
	}

	for i, frame := range e.Frames {
		if err := e.addFrame(frame); err != nil {
			return fmt.Errorf("%v when writing frame %d", err, i)
		}
	}

	// go back and write total size
	e.w.Seek(4, 0)
	if err := e.Add(uint32(e.WrittenBytes) - 8); err != nil {
		return fmt.Errorf("%v when writing the total written bytes")
	}
	// jump to the end of the file.
	e.w.Seek(0, 2)
	switch e.w.(type) {
	case *os.File:
		e.w.(*os.File).Sync()
	}
	return nil
}
