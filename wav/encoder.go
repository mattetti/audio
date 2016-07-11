package wav

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/mattetti/audio/misc"
	"github.com/mattetti/audio/riff"
)

// Encoder encodes LPCM data into a wav containter.
type Encoder struct {
	w              io.WriteSeeker
	SampleRate     int
	BitDepth       int
	NumChans       int
	Frames         [][]int
	WavAudioFormat int

	WrittenBytes int
}

// NewEncoder creates a new encoder to create a new wav file.
// Don't forget to add Frames to the encoder before writing.
func NewEncoder(w io.WriteSeeker, sampleRate, bitDepth, numChans, audioFormat int) *Encoder {
	return &Encoder{
		w:              w,
		SampleRate:     sampleRate,
		BitDepth:       bitDepth,
		NumChans:       numChans,
		WavAudioFormat: audioFormat,
	}
}

// Add serializes and adds the passed value using little endian
func (e *Encoder) Add(src interface{}) error {
	e.WrittenBytes += binary.Size(src)
	return binary.Write(e.w, binary.LittleEndian, src)
}

// AddLE serializes and adds the passed value using little endian
func (e *Encoder) AddLE(src interface{}) error {
	return e.Add(src)
}

// AddBE serializes and adds the passed value using big endian
func (e *Encoder) AddBE(src interface{}) error {
	e.WrittenBytes += binary.Size(src)
	return binary.Write(e.w, binary.BigEndian, src)
}

func (e *Encoder) addFrame(frame []int) error {
	if frame == nil {
		return fmt.Errorf("can't add a nil frame")
	}
	for i := 0; i < e.NumChans; i++ {
		switch e.BitDepth {
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
			return fmt.Errorf("can't add frames of bit size %d", e.BitDepth)
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

	// HEADERS
	// riff ID
	if err := e.Add(riff.RiffID); err != nil {
		return err
	}
	// file size uint32, to update later on.
	if err := e.Add(uint32(42)); err != nil {
		return err
	}

	// wave headers
	if err := e.Add(riff.WavFormatID); err != nil {
		return err
	}
	// form
	if err := e.Add(riff.FmtID); err != nil {
		return err
	}
	// chunk size
	if err := e.AddLE(uint32(16)); err != nil {
		return err
	}
	// wave format
	if err := e.Add(uint16(e.WavAudioFormat)); err != nil {
		return err
	}
	// num channels
	if err := e.Add(uint16(e.NumChans)); err != nil {
		return fmt.Errorf("error encoding the number of channels - %v", err)
	}
	// samplerate
	if err := e.Add(uint32(e.SampleRate)); err != nil {
		return fmt.Errorf("error encoding the sample rate - %v", err)
	}
	// avg bytes per sec
	if err := e.Add(uint32(e.SampleRate * e.NumChans * e.BitDepth / 8)); err != nil {
		return fmt.Errorf("error encoding the avg bytes per sec - %v", err)
	}
	// block align
	if err := e.Add(uint16(2)); err != nil {
		return err
	}
	// bits per sample
	if err := e.Add(uint16(e.BitDepth)); err != nil {
		return fmt.Errorf("error encoding bits per sample - %v", err)
	}

	// sound header
	if err := e.Add(riff.DataFormatID); err != nil {
		return fmt.Errorf("error encoding sound header %v", err)
	}

	chunksize := uint32((int(e.BitDepth) / 8) * int(e.NumChans) * len(e.Frames))
	if err := e.Add(uint32(chunksize)); err != nil {
		return fmt.Errorf("%v when writing wav data chunk size header", err)
	}

	for i, frame := range e.Frames {
		if err := e.addFrame(frame); err != nil {
			return fmt.Errorf("%v when writing frame %d", err, i)
		}
	}

	// go back and write total size
	e.w.Seek(4, 0)
	if err := e.AddLE(uint32(e.WrittenBytes) - 8); err != nil {
		return fmt.Errorf("%v when writing the total written bytes", err)
	}
	// jump to the end of the file.
	e.w.Seek(0, 2)
	switch e.w.(type) {
	case *os.File:
		e.w.(*os.File).Sync()
	}
	return nil
}
