package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mattetti/audio/midi"
)

var (
	fileFlag = flag.String("file", "", "The path to the midi file to decode")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s \n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	f, err := os.Open(*fileFlag)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	decoder := midi.New(f)
	if err := decoder.Parse(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("format:", decoder.Format)
	fmt.Println(decoder.TicksPerQuarterNote, "ticks per quarter")
	for _, tr := range decoder.Tracks {
		for _, ev := range tr.Events {
			fmt.Println(ev)
		}
	}

}
