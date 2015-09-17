package midi

import "strings"

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
	// octave starts at -1 but first note is at 0
	return key + (octave+1)*12
}
