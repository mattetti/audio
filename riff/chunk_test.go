package riff

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestWavNextChunk(t *testing.T) {
	path, _ := filepath.Abs("fixtures/sample.wav")
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	c := New(f)
	if err := c.ParseHeaders(); err != nil {
		t.Fatal(err)
	}
	// fmt
	ch, err := c.NextChunk()
	if err != nil {
		t.Fatal(err)
	}
	if ch.ID != fmtID {
		t.Fatalf("Expected the next chunk to have an ID of %q but got %q", fmtID, ch.ID)
	}
	if ch.Size != 16 {
		t.Fatalf("Expected the next chunk to have a size of %d but got %d", 16, ch.Size)
	}
	ch.Done()
	//
	ch, err = c.NextChunk()
	if err != nil {
		t.Fatal(err)
	}
	if ch.ID != dataFormatID {
		t.Fatalf("Expected the next chunk to have an ID of %q but got %q", dataFormatID, ch.ID)
	}
	if ch.Size != 53958 {
		t.Fatalf("Expected the next chunk to have a size of %d but got %d", 53958, ch.Size)
	}
	if int(c.Size) != (ch.Size + 36) {
		t.Fatal("Looks like we have some extra data in this wav file?")
	}
}

func TestNextChunk(t *testing.T) {
	path, _ := filepath.Abs("fixtures/sample.wav")
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	c := New(f)
	if err := c.ParseHeaders(); err != nil {
		t.Fatal(err)
	}
	ch, err := c.NextChunk()
	if err != nil {
		t.Fatal(err)
	}
	ch.DecodeWavHeader(c)

	ch, err = c.NextChunk()
	if err != nil {
		t.Fatal(err)
	}

	nextSample := func() []byte {
		var s = make([]byte, c.BlockAlign)
		if err := ch.ReadLE(s); err != nil {
			t.Fatal(err)
		}
		return s
	}
	firstSample := nextSample()
	if ch.Pos != int(c.BlockAlign) {
		t.Fatal("Chunk position wasn't moved as expected")
	}
	expectedSample := []byte{0, 0}
	if bytes.Compare(firstSample, expectedSample) != 0 {
		t.Fatalf("First sample doesn't seem right, got %q, expected %q", firstSample, expectedSample)
	}

	desideredPos := 1541
	bytePos := desideredPos * 2
	for ch.Pos < bytePos {
		nextSample()
	}
	s := nextSample()
	expectedSample = []byte{0xfe, 0xff}
	if bytes.Compare(s, expectedSample) != 0 {
		t.Fatalf("1542nd sample doesn't seem right, got %q, expected %q", s, expectedSample)
	}

}

func ExampleParser_NextChunk() {
	// Example showing how to access the sound data
	path, _ := filepath.Abs("fixtures/sample.wav")
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	c := New(f)
	if err := c.ParseHeaders(); err != nil {
		panic(err)
	}
	soundData, err := c.NextChunk()
	if err != nil {
		panic(err)
	}

	nextSample := func() []byte {
		var s = make([]byte, c.BlockAlign)
		if err := soundData.ReadLE(s); err != nil {
			panic(err)
		}
		return s
	}

	// jump to a specific sample since first samples are blank
	desideredPos := 1541
	bytePos := desideredPos * 2
	for soundData.Pos < bytePos {
		nextSample()
	}

	sample := nextSample()
	fmt.Printf("1542nd sample: %#X %#X\n", sample[0], sample[1])
	// Output:
	// 1542nd sample: 0XFE 0XFF
}
