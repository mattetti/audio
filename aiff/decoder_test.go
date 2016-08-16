package aiff

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mattetti/audio"
)

// TODO(mattetti): benchmark allocations

func TestContainerAttributes(t *testing.T) {
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
		d := NewDecoder(f)
		clip := d.PCM()
		if d.Err() != nil {
			t.Fatal(d.Err())
		}

		numChannels, bitDepth, sampleRate, err := clip.Info()
		if err != nil {
			t.Fatal(err)
		}
		if bitDepth != int(exp.sampleSize) {
			t.Fatalf("%s of %s didn't match %d, got %d", "Clip bit depth", exp.input, exp.sampleSize, bitDepth)
		}

		if sampleRate != int64(exp.sampleRate) {
			t.Fatalf("%s of %s didn't match %d, got %d", "Clip sample rate", exp.input, exp.sampleRate, sampleRate)
		}

		if numChannels != int(exp.numChans) {
			t.Fatalf("%s of %s didn't match %d, got %d", "Clip sample channels", exp.input, exp.numChans, numChannels)
		}

		if clip.Size() != exp.totalFrames {
			t.Fatalf("%s of %s didn't match %d, got %d", "Clip sample data size", exp.input, exp.totalFrames, clip.Size())
		}

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
		if d.NumChans != exp.numChans {
			t.Fatalf("%s of %s didn't match %d, got %d", "NumChans", exp.input, exp.numChans, d.NumChans)
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

func Test_Frames(t *testing.T) {
	testCases := []struct {
		input string
	}{
		// 22050, 8bit, mono
		{"fixtures/kick8b.aiff"},
		// 22050, 16bit, mono
		{"fixtures/kick.aif"},
		// 22050, 16bit, mono
		{"fixtures/kick32b.aiff"},
		// 44100, 16bit, mono
		{"fixtures/subsynth.aif"},
		// 44100, 16bit, stereo
		{"fixtures/bloop.aif"},
		// 48000, 16bit, stereo
		{"fixtures/zipper.aiff"},
		// 48000, 24bit, stereo
		{"fixtures/zipper24b.aiff"},
	}

	for i, tc := range testCases {
		t.Logf("test case %d\n", i)
		in, err := os.Open(tc.input)
		if err != nil {
			t.Fatalf("couldn't open %s %v", tc.input, err)
		}
		d := NewDecoder(in)
		clip := d.PCM()
		frames, err := d.Frames()
		if err != nil {
			t.Fatal(err)
		}
		if int(clip.Size()) != len(frames) {
			t.Fatalf("expected %d frames, got %d", clip.Size(), len(frames))
		}
	}
}

func TestDecoder_Duration(t *testing.T) {
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
		c := NewDecoder(f)
		d, err := c.Duration()
		if err != nil {
			t.Fatal(err)
		}
		if d != exp.duration {
			t.Fatalf("duration of %s didn't match %d milliseconds, got %d", exp.input, exp.duration, d)
		}
	}
}

func TestDecoderPCMBuffer(t *testing.T) {
	testCases := []struct {
		input   string
		desc    string
		samples []int
	}{
		{"fixtures/kick.aif",
			"1 ch,  22050 Hz, 'lpcm' (0x0000000E) 16-bit big-endian signed integer",
			[]int{76, 76, 75, 75, 72, 71, 72, 69, 70, 68, 65, 73, 529, 1425, 2245, 2941, 3514, 3952, 4258, 4436, 4486, 4413, 4218, 3903, 3474, 2938, 2295, 1553, 711, -214, -1230, -2321, -3489, -4721, -6007, -7352, -8738, -10172, -11631, -13127, -14642, -16029, -17322, -18528, -19710, -20877},
		},
	}

	for i, tc := range testCases {
		t.Logf("%d - %s - %s\n", i, tc.input, tc.desc)
		path, _ := filepath.Abs(tc.input)
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		d := NewDecoder(f)

		intBuf := make([]int, len(tc.samples))
		buf := audio.NewPCMIntBuffer(intBuf, nil)
		err = d.PCMBuffer(buf)
		if err != nil {
			t.Fatal(err)
		}
		if len(buf.Ints) != len(tc.samples) {
			t.Fatalf("the length of the buffer (%d) didn't match what we expected (%d)", len(buf.Ints), len(tc.samples))
		}
		for i := 0; i < len(buf.Ints); i++ {
			if buf.Ints[i] != tc.samples[i] {
				t.Fatalf("Expected %d at position %d, but got %d", tc.samples[i], i, buf.Ints[i])
			}
		}
	}
}
