package main

import (
	"log"
	"os"

	"github.com/mattetti/audio/midi"
)

func main() {
	f, err := os.Create("midi.mid")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		f.Close()
	}()
	e := midi.NewEncoder(f, midi.SingleTrack, 96)
	tr := e.NewTrack()
	/*
		// add a C3 at velocity 100, a beat/quarter note after the start
		tr.Add(0, midi.NoteOn(0, midi.KeyInt("C", 3), 100))
		// turn off the C3
		tr.Add(1, midi.NoteOff(0, midi.KeyInt("C", 3)))
		if err := e.Write(); err != nil {
			log.Fatal(err)
		}
	*/

	// 1 beat with 1 note for nothing
	tr.Add(1, midi.NoteOff(0, 60))

	vel := 90
	//C3 to B3
	var j float64
	for i := 60; i < 72; i++ {
		tr.Add(j, midi.NoteOn(0, i, vel))
		tr.Add(1, midi.NoteOff(0, i))
		j = 1
	}
	// empty note at the end
	tr.Add(1, midi.NoteOff(0, 60))

	if err := e.Write(); err != nil {
		log.Fatal(err)
	}

}
