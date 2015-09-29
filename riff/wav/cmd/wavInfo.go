package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mattetti/audio/riff/wav"
)

func main() {
	f, err := os.Open("../fixtures/kick.wav")
	if err != nil {
		log.Fatal(err)
	}
	d := wav.NewDecoder(f)
	info, frames, err := d.ReadFrames()
	fmt.Println(info, frames, err)
}
