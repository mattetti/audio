package midi

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
)

const (
	SingleTrack uint16 = iota
	Syncronous
	Asyncronous
)

type Encoder struct {
	// we need a write seeker because we will update the size at the enc
	// and need to back to the beginning of the file.
	w io.WriteSeeker

	/*
	   Format describes the tracks format

	   0	-	single-track
	   Format 0 file has a header chunk followed by one track chunk. It
	   is the most interchangable representation of data. It is very useful
	   for a simple single-track player in a program which needs to make
	   synthesizers make sounds, but which is primarily concerened with
	   something else such as mixers or sound effect boxes. It is very
	   desirable to be able to produce such a format, even if your program
	   is track-based, in order to work with these simple programs. On the
	   other hand, perhaps someone will write a format conversion from
	   format 1 to format 0 which might be so easy to use in some setting
	   that it would save you the trouble of putting it into your program.


	   Synchronous multiple tracks means that the tracks will all be vertically synchronous, or in other words,
	    they all start at the same time, and so can represent different parts in one song.
	    1	-	multiple tracks, synchronous
	    Asynchronous multiple tracks do not necessarily start at the same time, and can be completely asynchronous.
	    2	-	multiple tracks, asynchronous
	*/
	Format uint16

	// NumTracks represents the number of tracks in the midi file
	NumTracks uint16

	// resolution for delta timing
	TicksPerQuarterNote uint16

	TimeFormat timeFormat
	Tracks     []*Track

	size int
}

func NewEncoder(w io.WriteSeeker, format uint16, ppqn uint16) *Encoder {
	return &Encoder{w: w, Format: format, TicksPerQuarterNote: ppqn}
}

// NewTrack adds and return a new track (not thread safe)
func (e *Encoder) NewTrack() *Track {
	t := &Track{ticksPerBeat: e.TicksPerQuarterNote}
	e.Tracks = append(e.Tracks, t)
	return t
}

// NoteOn returns a pointer to a new event of type NoteOn (without the delta timing data)
func NoteOn(channel, key, vel int) *Event {
	return &Event{
		MsgChan:  uint8(channel),
		MsgType:  uint8(eventByteMap["NoteOn"]),
		Note:     uint8(key),
		Velocity: uint8(vel),
	}
}

// NoteOff returns a pointer to a new event of type NoteOff (without the delta timing data)
func NoteOff(channel, key int) *Event {
	return &Event{
		MsgChan:  uint8(channel),
		MsgType:  uint8(eventByteMap["NoteOff"]),
		Note:     uint8(key),
		Velocity: 64,
	}
}

// AfterTouch returns a pointer to a new aftertouch event
func Aftertouch(channel, key, vel int) *Event {
	return &Event{
		MsgChan:  uint8(channel),
		MsgType:  uint8(eventByteMap["AfterTouch"]),
		Note:     uint8(key),
		Velocity: uint8(vel),
	}
}

// ControlChange sets a new value for a given controller
// The controller number is between 0-119.
// The new controller value is between 0-127.
func ControlChange(channel, controller, newVal int) *Event {
	return &Event{
		MsgChan:    uint8(channel),
		MsgType:    uint8(eventByteMap["ControlChange"]),
		Controller: uint8(controller),
		NewValue:   uint8(newVal),
	}
}

// ProgramChange sets a new value the same way as ControlChange
// but implements Mode control and special message by using reserved controller numbers 120-127.
func ProgramChange(channel, controller, newVal int) *Event {
	return &Event{
		MsgChan:    uint8(channel),
		MsgType:    uint8(eventByteMap["ProgramChange"]),
		Controller: uint8(controller),
		NewValue:   uint8(newVal),
	}
}

// ChannelAfterTouch is a global aftertouch with a value from 0 to 127
func ChannelAfterTouch(channel, vel int) *Event {
	return &Event{
		MsgChan:  uint8(channel),
		MsgType:  uint8(eventByteMap["ChannelAfterTouch"]),
		Pressure: uint8(vel),
	}
}

// PitchWheelChange is sent to indicate a change in the pitch bender.
// The possible value goes between 0 and 16383 where 8192 is the center.
func PitchWheelChange(channel, int, val int) *Event {
	return &Event{
		MsgChan:      uint8(channel),
		MsgType:      uint8(eventByteMap["PitchWheelChange"]),
		AbsPitchBend: uint16(val),
	}
}

// TODO
func Meta(channel int) *Event {
	return nil
}

// Write writes the binary representation to the writer
func (e *Encoder) Write() error {
	if e == nil {
		return errors.New("Can't write a nil encoder")
	}
	e.writeHeaders()
	for _, t := range e.Tracks {
		if err := e.encodeTrack(t); err != nil {
			return err
		}
	}
	// go back and update body size in header
	return nil
}

func (e *Encoder) writeHeaders() error {
	// chunk id [4] headerChunkID
	if _, err := e.w.Write(headerChunkID[:]); err != nil {
		return err
	}
	// header size
	if err := binary.Write(e.w, binary.BigEndian, uint32(6)); err != nil {
		return err
	}
	// Format
	if err := binary.Write(e.w, binary.BigEndian, e.Format); err != nil {
		return err
	}
	// numtracks (not trusting the field value, but checking the actual amount of tracks
	if err := binary.Write(e.w, binary.BigEndian, uint16(len(e.Tracks))); err != nil {
		return err
	}
	// division [uint16] <-- contains precision
	if err := binary.Write(e.w, binary.BigEndian, e.TicksPerQuarterNote); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) encodeTrack(t *Track) error {
	// chunk id [4]
	if _, err := e.w.Write(trackChunkID[:]); err != nil {
		return err
	}
	data, err := t.ChunkData()
	if err != nil {
		return err
	}
	// chunk size
	if err := binary.Write(e.w, binary.BigEndian, uint32(len(data))); err != nil {
		log.Fatalf("106", err)

		return err
	}
	// chunk data
	if _, err := e.w.Write(data); err != nil {
		return err
	}

	return nil
}
