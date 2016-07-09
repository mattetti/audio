package aiff

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestClip(t *testing.T) {
	expectations := []struct {
		input           string
		id              [4]byte
		size            uint32
		format          [4]byte
		commSize        uint32
		numChans        uint16
		numSampleFrames uint32
		sampleSize      uint16
		sampleRate      int
		totalFrames     int64
	}{
		{"fixtures/kick.aif", formID, 9642, aiffID,
			18, 1, 4484, 16, 22050, 4484},
	}

	for _, exp := range expectations {
		path, _ := filepath.Abs(exp.input)
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		d := NewDecoder(f, nil)
		clip := d.Clip()
		if d.Err() != nil {
			t.Fatal(d.Err())
		}

		fi := clip.FrameInfo()
		if fi.BitDepth != int(exp.sampleSize) {
			t.Fatalf("%s of %s didn't match %d, got %d", "Clip bit depth", exp.input, exp.sampleSize, fi.BitDepth)
		}

		if fi.SampleRate != int64(exp.sampleRate) {
			t.Fatalf("%s of %s didn't match %d, got %d", "Clip sample rate", exp.input, exp.sampleRate, fi.SampleRate)
		}

		if fi.Channels != int(exp.numChans) {
			t.Fatalf("%s of %s didn't match %d, got %d", "Clip sample channels", exp.input, exp.numChans, fi.Channels)
		}

		if clip.Size() != exp.totalFrames {
			t.Fatalf("%s of %s didn't match %d, got %d", "Clip sample data size", exp.input, exp.totalFrames, clip.Size())
		}

		// decoder data, some will probably be deprecated

		if d.ID != exp.id {
			t.Fatalf("%s of %s didn't match %s, got %s", "ID", exp.input, exp.id, d.ID)
		}
		if d.Size != exp.size {
			t.Fatalf("%s of %s didn't match %d, got %d", "BlockSize", exp.input, exp.size, d.Size)
		}
		if d.Format != exp.format {
			t.Fatalf("%s of %s didn't match %q, got %q", "Format", exp.input, exp.format, d.Format)
		}
		// comm chunk
		if d.commSize != exp.commSize {
			t.Fatalf("%s of %s didn't match %d, got %d", "comm size", exp.input, exp.commSize, d.commSize)
		}
		if d.numChans != exp.numChans {
			t.Fatalf("%s of %s didn't match %d, got %d", "NumChans", exp.input, exp.numChans, d.numChans)
		}
		if d.numSampleFrames != exp.numSampleFrames {
			t.Fatalf("%s of %s didn't match %d, got %d", "NumSampleFrames", exp.input, exp.numSampleFrames, d.numSampleFrames)
		}
		if d.BitDepth != exp.sampleSize {
			t.Fatalf("%s of %s didn't match %d, got %d", "SampleSize", exp.input, exp.sampleSize, d.BitDepth)
		}
		if d.SampleRate != exp.sampleRate {
			t.Fatalf("%s of %s didn't match %d, got %d", "SampleRate", exp.input, exp.sampleRate, d.SampleRate)
		}
	}
}

func TestNewDecoder(t *testing.T) {
	path, _ := filepath.Abs("fixtures/kick.aif")
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	ch := make(chan *Chunk)
	c := NewDecoder(f, ch)
	go func() { c.Parse() }()

	for chunk := range ch {
		id := string(chunk.ID[:])
		t.Log(id, chunk.Size)
		if id != string(COMMID[:]) {
			buf := make([]byte, chunk.Size)
			chunk.ReadBE(buf)
			t.Log(hex.Dump(buf))
		}
		chunk.Done()
	}
	if c.Err() != nil {
		t.Fatal(c.Err())
	}
}

func TestReadFrames(t *testing.T) {
	path, _ := filepath.Abs("fixtures/kick.aif")
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	r := NewDecoder(f, nil)
	info, frames, err := r.Frames()
	if err != nil {
		t.Fatal(err)
	}
	if info.SampleRate != 22050 {
		t.Fatalf("unexpected sample rate: %d", info.SampleRate)
	}
	if info.BitDepth != 16 {
		t.Fatalf("unexpected sample size: %d", info.BitDepth)
	}
	if info.NumChannels != 1 {
		t.Fatalf("unexpected channel number: %d", info.NumChannels)
	}

	if totalFrames := len(frames); totalFrames != 4484 {
		t.Fatalf("unexpected total frames: %d", totalFrames)
	}
}

func TestClip_Read(t *testing.T) {
	expectations := []struct {
		input       string
		totalFrames int
	}{
		{"fixtures/kick.aif", 4484},
	}

	for _, exp := range expectations {
		path, _ := filepath.Abs(exp.input)
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		d := NewDecoder(f, nil)
		clip := d.Clip()
		totalFrames := int(clip.Size())
		if totalFrames != exp.totalFrames {
			t.Fatalf("Expected %d frames, got %d\n", exp.totalFrames, totalFrames)
		}
		readFrames := 0

		bufSize := 4096
		buf := make([]byte, bufSize)
		var n int
		for readFrames < totalFrames {
			n, err = clip.Read(buf)
			if err != nil || n == 0 {
				break
			}
			readFrames += n
		}
		if readFrames != totalFrames {
			t.Fatalf("file expected to have %d frames, only read %d, missing %d frames\n", totalFrames, readFrames, (totalFrames - readFrames))
		}

	}
}

func TestDuration(t *testing.T) {
	expectations := []struct {
		input    string
		duration time.Duration
	}{
		{"fixtures/kick.aif", time.Duration(203356009)},
	}

	for _, exp := range expectations {
		path, _ := filepath.Abs(exp.input)
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		c := NewDecoder(f, nil)
		err = c.Parse()
		if err != nil {
			t.Fatal(err)
		}
		d, err := c.Duration()
		if err != nil {
			t.Fatal(err)
		}
		if d != exp.duration {
			t.Fatalf("duration of %s didn't match %d milliseconds, got %d", exp.input, exp.duration, d)
		}
	}
}

func ExampleDecoder_Duration() {
	path, _ := filepath.Abs("fixtures/kick.aif")
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	c := NewDecoder(f, nil)
	if err = c.Parse(); err != nil {
		panic(err)
	}
	d, _ := c.Duration()
	fmt.Printf("kick.aif has a duration of %f seconds\n", d.Seconds())
	// Output:
	// kick.aif has a duration of 0.203356 seconds
}
