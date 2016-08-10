package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mattetti/audio"
	"github.com/mattetti/audio/generator"
	"github.com/mattetti/audio/midi"
)

func main() {
	f, err := os.Open("daFunk.mid")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	decoder := midi.NewDecoder(f)
	if err := decoder.Parse(); err != nil {
		log.Fatal(err)
	}
	tr := decoder.Tracks[0]

	bpm := 110.0
	freq := 440.0
	fs := 44100
	biteDepth := 16
	osc := generator.NewOsc(generator.WaveSine, float64(freq), fs)
	// our osc generates values from -1 to 1, we need to go back to PCM scale
	factor := float64(audio.IntMaxSignedValue(biteDepth))
	osc.Amplitude = factor

	// ticker for the definition of a tick duration
	// for each tick, check what notes are on and add samples.
	msPerTick := ((bpm * float64(decoder.TicksPerQuarterNote)) / 60)
	tickDur := time.Duration(int(msPerTick)*int(time.Millisecond)) / 2
	idx := 0
	maxIDX := len(tr.Events)

	// real time player
	ticker := time.NewTicker(tickDur)
	var evT time.Time
	var ev *midi.Event
	ticksPerBeat := float64(decoder.TicksPerQuarterNote)
	start := time.Now()

	for t := range ticker.C {
		if idx+1 == maxIDX {
			break
		}
		if evT.IsZero() {
			ev = tr.Events[idx]
			idx++
			evType := midi.EventMap[ev.MsgType]
			if evType == "NoteOn" || evType == "NoteOff" {
				evT = time.Now().Add(durForEvent(ev, bpm, ticksPerBeat))
				// fmt.Println(idx, maxIDX, ev, evT)
			}
		} else if t.Equal(evT) || t.After(evT) {
			evType := midi.EventMap[ev.MsgType]
			switch evType {
			case "NoteOn":
				fmt.Printf("%s at %s", midi.MidiNoteToName(int(ev.Note)), t.Sub(start))
			case "NoteOff":
				fmt.Printf("\r")
			}

			// fmt.Println(t, "act on", ev)
			evT = time.Time{}
		}

	}
	fmt.Println()
	ticker.Stop()

	// // Play back the events
	// for _, ev := range tr.Events {
	// 	switch midi.EventMap[ev.MsgType] {
	// 	case "NoteOn", "NoteOff":
	// 		beats := float64(ev.TimeDelta) / float64(decoder.TicksPerQuarterNote)
	// 		secs := (beats / bpm) * 60
	// 		fmt.Printf("After %fs, secs, %s : %s\n",
	// 			secs,
	// 			midi.MidiNoteToName(int(ev.Note)),
	// 			midi.EventMap[ev.MsgType])
	// 	}
	// }

}

func durForEvent(ev *midi.Event, bpm, ticksPerBeat float64) time.Duration {
	beats := float64(ev.TimeDelta) / ticksPerBeat
	secs := int((beats / bpm) * 60)
	return time.Duration(secs * int(time.Second))
}
