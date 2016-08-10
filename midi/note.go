package midi

import (
	"math"
	"strconv"
	"strings"
)

var Notes = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

var notesToInt = map[string]int{
	"C":  0,
	"C#": 1,
	"DB": 1,
	"D":  2,
	"D#": 3,
	"EB": 3,
	"E":  4,
	"F":  5,
	"F#": 6,
	"GB": 6,
	"G":  7,
	"G#": 8,
	"AB": 8,
	"A":  9,
	"A#": 10,
	"BB": 10,
	"B":  11,
}

type Note struct {
	Channel  int
	Key      int
	Velocity int
}

// KeyInt converts an A-G note notation to a midi note number value.
func KeyInt(n string, octave int) int {
	key := notesToInt[strings.ToUpper(n)]
	// octave starts at -2 but first note is at 0
	return key + (octave+2)*12
}

// KeyFreq returns the frequency for the given key/octave combo
// https://en.wikipedia.org/wiki/MIDI_Tuning_Standard#Frequency_values
func KeyFreq(n string, octave int) float64 {
	return 440.0 * math.Pow(2, (float64(KeyInt(n, octave)-69)/12))
}

// MidiNoteToName converts a midi note value into its English name
func MidiNoteToName(note int) string {
	key := Notes[note%12]
	octave := ((note / 12) | 0) - 2 // The MIDI scale starts at octave = -2
	return key + strconv.Itoa(octave)
}
