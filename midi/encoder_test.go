package midi

import (
	"encoding/hex"
	"io/ioutil"
	"os"
	"testing"
)

func TestNewEncoder(t *testing.T) {
	w, err := tmpFile()
	if err != nil {
		t.Fatal(err)
	}
	//tmpFilePath, _ := ioutil.TempDir("", w.Name())
	defer func() {
		w.Close()
		os.Remove(w.Name())
	}()
	e := NewEncoder(w, SingleTrack, 96)
	tr := e.NewTrack()
	// add a C3 at velocity 99, half a beat/quarter note after the start
	tr.Add(0.5, NoteOn(1, KeyInt("C", 3), 99))
	// turn off the C3
	tr.Add(1, NoteOff(1, KeyInt("C", 3), 127))
	if err := e.Write(); err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v\n", e)
	w.Seek(0, 0)
	midiData, err := ioutil.ReadAll(w)
	if err != nil {
		t.Log(err)
	}
	t.Log(hex.Dump(midiData))
}

func tmpFile() (*os.File, error) {
	f, err := ioutil.TempFile("", "midi-test-")
	if err != nil {
		return nil, err
	}
	return f, nil
}
