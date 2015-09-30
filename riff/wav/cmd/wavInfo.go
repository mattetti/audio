package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mattetti/audio/riff"
	"github.com/mattetti/audio/riff/wav"
)

var (
	inputFlag = flag.String("path", "../fixtures/kick.wav", "path to the file to parse")
)

func main() {
	flag.Parse()

	f, err := os.Open(*inputFlag)
	if err != nil {
		log.Fatal(err)
	}
	d := wav.NewDecoder(f)
	ch := make(chan *riff.Chunk)

	go func() {
		if err := d.Parse(ch); err != nil {
			log.Fatal(err)
		}
	}()

	for chunk := range ch {
		fmt.Println(string(chunk.ID[:]))
	}

	//info, frames, err := d.ReadFrames()
	//fmt.Println(info, frames, err)
}
