package midi

import "io"

const (
	SingleTrack uint16 = iota
	Syncronous
	Asyncronous
)

type Encoder struct {
	w io.Writer

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
}

func NewEncoder(w io.Writer, format uint16, ppqn uint16) *Encoder {
	return &Encoder{w: w, Format: format, TicksPerQuarterNote: ppqn}
}

func (e *Encoder) writeHeaders() {
	// chunk id [4] headerChunkID
	// size [uint32] 6
	// Format
	// numtracks
	// division [uint16] <-- contains the BPM/tempo
}

func (e *Encoder) Write() error {
	return nil
}

// NoteOn returns a pointer to a new event of type NoteOn (without the delta timing data)
func NoteOn(channel, key, vel int) *Event {
	return &Event{
		Channel:  uint8(channel),
		MsgType:  uint8(eventByteMap["NoteOn"]),
		Note:     uint8(key),
		Velocity: uint8(vel),
	}
}

// NoteOff return a pointer to a new event of type NoteOff (without the delta timing data)
func NoteOff(channel, key, vel int) *Event {
	return &Event{
		Channel:  uint8(channel),
		MsgType:  uint8(eventByteMap["NoteOff"]),
		Note:     uint8(key),
		Velocity: uint8(vel),
	}
}
