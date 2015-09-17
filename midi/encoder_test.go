package midi

import (
	"bytes"
	"testing"
)

func TestNewEncoder(t *testing.T) {
	w := bytes.NewBuffer(nil)
	e := NewEncoder(w, SingleTrack, 96)
	tr := &Track{}
	// add a C3 at velocity 99, half a beat/quarter note
	tr.Add(0.5, NoteOn(1, KeyInt("C", 3), 99))
	t.Log(e)
}
