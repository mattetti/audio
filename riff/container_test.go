package riff

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseHeader(t *testing.T) {
	expectations := []struct {
		input  string
		id     [4]byte
		size   uint32
		format [4]byte
	}{
		{"fixtures/sample.rmi", riffID, 29632, rmiFormatID},
		{"fixtures/sample.wav", riffID, 53994, wavFormatID},
		{"fixtures/sample.avi", riffID, 230256, aviFormatID},
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
		if c.Size != exp.size {
			t.Fatalf("%s of %s didn't match %d, got %d", "BlockSize", exp.input, exp.size, c.Size)
		}
		if c.Format != exp.format {
			t.Fatalf("%s of %s didn't match %q, got %q", "Format", exp.input, exp.format, c.Format)
		}
	}
}

func TestParseWavHeaders(t *testing.T) {
	expectations := []struct {
		input         string
		headerSize    uint32
		format        uint16
		numChans      uint16
		sampleRate    uint32
		byteRate      uint32
		blockAlign    uint16
		bitsPerSample uint16
	}{
		{"fixtures/sample.wav", 16, 1, 1, 44100, 88200, 2, 16},
	}

	for _, exp := range expectations {
		path, _ := filepath.Abs(exp.input)
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		c := NewContainer(f)
		if err := c.ParseHeaders(); err != nil {
			t.Fatal(err)
		}
		if c.wavHeaderSize != exp.headerSize {
			t.Fatalf("%s didn't match %d, got %d", "header size", exp.headerSize, c.wavHeaderSize)
		}
		if c.WavAudioFormat != exp.format {
			t.Fatalf("%s didn't match %d, got %d", "audio format", exp.format, c.WavAudioFormat)
		}
		if c.NumChannels != exp.numChans {
			t.Fatalf("%s didn't match %d, got %d", "# of channels", exp.numChans, c.NumChannels)
		}
		if c.SampleRate != exp.sampleRate {
			t.Fatalf("%s didn't match %d, got %d", "SampleRate", exp.sampleRate, c.SampleRate)
		}
		if c.ByteRate != exp.byteRate {
			t.Fatalf("%s didn't match %d, got %d", "ByteRate", exp.byteRate, c.ByteRate)
		}
		if c.BlockAlign != exp.blockAlign {
			t.Fatalf("%s didn't match %d, got %d", "BlockAlign", exp.blockAlign, c.BlockAlign)
		}
		if c.BitsPerSample != exp.bitsPerSample {
			t.Fatalf("%s didn't match %d, got %d", "BitsPerSample", exp.bitsPerSample, c.BitsPerSample)
		}
	}

}

func TestContainerDuration(t *testing.T) {
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

func TestWavNextChunk(t *testing.T) {
	path, _ := filepath.Abs("fixtures/sample.wav")
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	c := NewContainer(f)
	if err := c.ParseHeaders(); err != nil {
		t.Fatal(err)
	}
	ch, err := c.NextChunk()
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
	c := NewContainer(f)
	if err := c.ParseHeaders(); err != nil {
		t.Fatal(err)
	}
	ch, err := c.NextChunk()
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
