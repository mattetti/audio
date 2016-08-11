package wav_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mattetti/audio"
	"github.com/mattetti/audio/wav"
)

func TestPCM_Ints(t *testing.T) {
	expectations := []struct {
		input       string
		totalFrames int
	}{
		{"fixtures/kick.wav", 4484},
		{"fixtures/dirty-kick-24b441k.wav", 21340},
	}

	for i, exp := range expectations {
		t.Logf("%d - %s\n", i, exp.input)
		path, _ := filepath.Abs(exp.input)
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		d := wav.NewDecoder(f)
		pcm := d.PCM()
		totalFrames := int(pcm.Size())
		if totalFrames != exp.totalFrames {
			t.Fatalf("Expected %d frames, got %d\n", exp.totalFrames, totalFrames)
		}
		readFrames := 0

		bufSize := 4096
		buf := make(audio.FramesInt, bufSize/int(d.NumChans))
		var n int
		for readFrames < totalFrames {
			n, err = pcm.Ints(buf)
			if err != nil || n == 0 {
				break
			}
			readFrames += n
		}
		if readFrames != totalFrames {
			t.Fatalf("file expected to have %d frames, only read %d, off by %d frames\n", totalFrames, readFrames, (totalFrames - readFrames))
		}

	}
}

func TestPCM_Buffer(t *testing.T) {
	testCases := []struct {
		input   string
		desc    string
		samples []int
	}{
		{"fixtures/bass.wav",
			"2 ch,  44100 Hz, 'lpcm' 24-bit little-endian signed integer",
			[]int{0, 0, 28160, 26368, 16128, 14848, -746240, -705536, 596480, 565504, 2161408, 2050304, -306944, -271872, -607488, -537856, -1624064, -1477376, -4554752, -4233472, -16599808, -15644160, -21150208, -20034560, -6344192, -6146816, 28578048, 26699520, 60357888, 56626176, 70529280},
		},
	}

	for i, tc := range testCases {
		t.Logf("%d - %s\n", i, tc.input)
		path, _ := filepath.Abs(tc.input)
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		d := wav.NewDecoder(f)
		pcm := d.PCM()

		intBuf := make([]int, len(tc.samples))
		buf := audio.NewPCMIntBuffer(intBuf, nil)
		err = pcm.Buffer(buf)
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
