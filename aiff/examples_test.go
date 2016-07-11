package aiff_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mattetti/audio/aiff"
	"github.com/mattetti/audio/misc"
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

func ExampleClip() {
	path, _ := filepath.Abs("fixtures/kick.aif")
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	d := aiff.NewDecoder(f)
	clip := d.Clip()
	totalFrames := int(clip.Size())
	buf := make([]byte, 4096)
	var (
		extractedFrames misc.AudioFrames
		readFrames      int
		n               int
	)

	for readFrames < totalFrames {
		n, err = clip.Read(buf)
		if err != nil || n == 0 {
			break
		}
		readFrames += n
		frames, err := d.DecodeFrames(buf)
		if err != nil {
			break
		}
		// It's very important to limit the number of frames we append
		// based on the number of frames contained in the buffer.
		// Otherwise if the buffer is bigger than the available frames,
		// we end up with blank/bad frames.
		extractedFrames = append(extractedFrames, frames[:n]...)
	}

	if err != nil {
		fmt.Printf("something went wrong %v", err)
		os.Exit(1)
	}

	fmt.Printf("%d PCM frames extracted, expected %d", len(extractedFrames), clip.Size())
	// Output:
	// 4484 PCM frames extracted, expected 4484
}
