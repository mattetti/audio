package aiff_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mattetti/audio"
	"github.com/mattetti/audio/aiff"
)

func ExampleDecoder_Duration() {
	path, _ := filepath.Abs("fixtures/kick.aif")
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	c := aiff.NewDecoder(f)
	d, _ := c.Duration()
	fmt.Printf("kick.aif has a duration of %f seconds\n", d.Seconds())
	// Output:
	// kick.aif has a duration of 0.203356 seconds
}

func ExamplePCM() {
	path, _ := filepath.Abs("fixtures/kick.aif")
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	d := aiff.NewDecoder(f)
	pcm := d.PCM()
	totalFrames := int(pcm.Size())
	buf := make(audio.SamplesInt, 2048)
	var (
		extractedFrames audio.SamplesInt
		readFrames      int
		n               int
	)

	for readFrames < totalFrames {
		n, err = pcm.Ints(buf)
		if err != nil || n == 0 {
			break
		}
		readFrames += n
		// It's very important to limit the number of frames we append
		// based on the number of frames contained in the buffer.
		// Otherwise if the buffer is bigger than the available frames,
		// we end up with blank/bad frames.
		// We could have also used pcm.NextInts(2048) if we didn't care to reuse
		// a buffer.
		extractedFrames = append(extractedFrames, buf[:n]...)
	}

	if err != nil {
		fmt.Printf("something went wrong %v", err)
		os.Exit(1)
	}

	fmt.Printf("%d PCM frames extracted", len(extractedFrames))
	// Output:
	// 4484 PCM frames extracted
}
