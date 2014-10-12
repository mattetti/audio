package riff

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseHeader(t *testing.T) {
	expectations := []struct {
		input     string
		id        [4]byte
		blockSize uint32
		format    [4]byte
	}{
		{"fixtures/sample.rmi", RIFF, 29632, RmiFormat},
		{"fixtures/sample.wav", RIFF, 53994, WavFormat},
		{"fixtures/sample.avi", RIFF, 230256, AviFormat},
	}

	for _, exp := range expectations {
		path, _ := filepath.Abs(exp.input)
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		c := NewContainer(f)
		err = c.ParseHeaders()
		if err != nil {
			t.Fatal(err)
		}
		if c.ID != exp.id {
			t.Fatalf("%s of %s didn't match %s, got %s", "ID", exp.input, exp.id, c.ID)
		}
		if c.BlockSize != exp.blockSize {
			t.Fatalf("%s of %s didn't match %d, got %d", "BlockSize", exp.input, exp.blockSize, c.BlockSize)
		}
		if c.Format != exp.format {
			t.Fatalf("%s of %s didn't match %q, got %q", "Format", exp.input, exp.format, c.Format)
		}
	}
}

func TestWavDuration(t *testing.T) {
	expectations := []struct {
		input string
		dur   time.Duration
	}{
		{"fixtures/sample.wav", time.Duration(612176870)},
	}

	for _, exp := range expectations {
		path, _ := filepath.Abs(exp.input)
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		d, err := Duration(f)
		if err != nil {
			t.Fatal(err)
		}
		if d != exp.dur {
			t.Fatalf("%s of %s didn't match %f, got %f", "Duration", exp.input, exp.dur.Seconds(), d.Seconds())
		}
	}

	for _, exp := range expectations {
		path, _ := filepath.Abs(exp.input)
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		c := NewContainer(f)
		d, err := c.Duration()
		if err != nil {
			t.Fatal(err)
		}
		if d != exp.dur {
			t.Fatalf("Container duration of %s didn't match %f, got %f", exp.input, exp.dur.Seconds(), d.Seconds())
		}
	}

}

func ExampleDuration() {
	path, _ := filepath.Abs("fixtures/sample.wav")
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	d, err := Duration(f)
	if err != nil {
		panic(err)
	}
	fmt.Printf("File with a duration of %f seconds", d.Seconds())
	// Output:
	// File with a duration of 0.612177 seconds
}
