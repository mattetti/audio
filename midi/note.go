package midi

import (
	"strconv"
	"strings"
)

var Notes = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

var notesToInt = map[string]int{
	"C":  0,
	"C#": 1,
	"D":  2,
	"D#": 3,
	"E":  4,
	"F":  5,
	"F#": 6,
	"G":  7,
	"G#": 8,
	"A":  9,
	"A#": 10,
	"B":  11,
}

type Note struct {
	Channel  int
	Key      int
	Velocity int
}

// KeyInt cdonverts an A-G note notation to a midi note number value.
func KeyInt(n string, octave int) int {
	key := notesToInt[strings.ToUpper(n)]
	// octave starts at -2 but first note is at 0
	return key + (octave+2)*12
}

// MidiNoteToName converts a midi note value into its English name
func MidiNoteToName(note int) string {
	key := Notes[note%12]
	octave := ((note / 12) | 0) - 2 // The MIDI scale starts at octave = -2
	return key + strconv.Itoa(octave)
}
