package wav_test

import (
	"fmt"
	"log"
	"os"

	"github.com/mattetti/audio/wav"
)

func ExampleDecoder_Duration() {
	f, err := os.Open("fixtures/kick.wav")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	dur, err := wav.NewDecoder(f).Duration()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s duration: %s\n", f.Name(), dur)
	// Output: fixtures/kick.wav duration: 204.172335ms
}

func ExampleEncoder_Write() {
	f, err := os.Open("fixtures/kick.wav")
	if err != nil {
		panic(fmt.Sprintf("couldn't open audio file - %v", err))
	}

	// Decode the original audio file
	// and collect audio content and information.
	d := wav.NewDecoder(f)
	pcm := d.PCM()
	numChannels, bitDepth, sampleRate, err := pcm.Info()
	if err != nil {
		panic(err)
	}
	frames, err := d.SamplesInt()
	if err != nil {
		panic(err)
	}
	f.Close()
	fmt.Println("Old file ->", d)

	// Destination file
	out, err := os.Create("testOutput/kick.wav")
	if err != nil {
		panic(fmt.Sprintf("couldn't create output file - %v", err))
	}

	// setup the encoder and write all the frames
	e := wav.NewEncoder(out,
		int(sampleRate),
		bitDepth,
		numChannels,
		int(d.WavAudioFormat))
	if err := e.Write(frames); err != nil {
		panic(err)
	}
	// close the encoder to make sure the headers are properly
	// set and the data is flushed.
	if err := e.Close(); err != nil {
		panic(err)
	}
	out.Close()

	// reopen to confirm things worked well
	out, err = os.Open("testOutput/kick.wav")
	if err != nil {
		panic(err)
	}
	d2 := wav.NewDecoder(out)
	d2.ReadInfo()
	fmt.Println("New file ->", d2)
	out.Close()
	os.Remove(out.Name())

	// Output:
	// Old file -> Format: WAVE - 1 channels @ 22050 / 16 bits - Duration: 0.204172 seconds
	// New file -> Format: WAVE - 1 channels @ 22050 / 16 bits - Duration: 0.204172 seconds
}
