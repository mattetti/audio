package wav_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mattetti/audio"
	"github.com/mattetti/audio/wav"
)

func TestDecoder_Duration(t *testing.T) {
	testCases := []struct {
		in       string
		duration time.Duration
	}{
		{"fixtures/kick.wav", time.Duration(204172335 * time.Nanosecond)},
	}

	for _, tc := range testCases {
		f, err := os.Open(tc.in)
		if err != nil {
			t.Fatal(err)
		}
		dur, err := wav.NewDecoder(f).Duration()
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
		if dur != tc.duration {
			t.Fatalf("expected duration to be: %s but was %s", tc.duration, dur)
		}
	}

}

func TestDecoder_Attributes(t *testing.T) {
	testCases := []struct {
		in             string
		numChannels    int
		sampleRate     int
		avgBytesPerSec int
		bitDepth       int
	}{
		{in: "fixtures/kick.wav",
			numChannels:    1,
			sampleRate:     22050,
			avgBytesPerSec: 44100,
			bitDepth:       16,
		},
	}

	for _, tc := range testCases {
		f, err := os.Open(tc.in)
		if err != nil {
			t.Fatal(err)
		}
		d := wav.NewDecoder(f)
		d.ReadInfo()
		f.Close()
		if int(d.NumChans) != tc.numChannels {
			t.Fatalf("expected info to have %d channels but it has %d", tc.numChannels, d.NumChans)
		}
		if int(d.SampleRate) != tc.sampleRate {
			t.Fatalf("expected info to have a sample rate of %d but it has %d", tc.sampleRate, d.SampleRate)
		}
		if int(d.AvgBytesPerSec) != tc.avgBytesPerSec {
			t.Fatalf("expected info to have %d avg bytes per sec but it has %d", tc.avgBytesPerSec, d.AvgBytesPerSec)
		}
		if int(d.BitDepth) != tc.bitDepth {
			t.Fatalf("expected info to have %d bits per sample but it has %d", tc.bitDepth, d.BitDepth)
		}
	}
}

func TestDecoder_Buffer(t *testing.T) {
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

		intBuf := make([]int, len(tc.samples))
		buf := audio.NewPCMIntBuffer(intBuf, nil)
		err = d.Buffer(buf)
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

// DEPRECATED
func TestDecoder_Clip(t *testing.T) {
	testCases := []struct {
		in          string
		size        int64
		numChannels int
		sampleRate  int64
		bitDepth    int
	}{
		{in: "fixtures/kick.wav",
			size:        4484,
			numChannels: 1,
			sampleRate:  22050,
			bitDepth:    16,
		},
	}

	for _, tc := range testCases {
		f, err := os.Open(tc.in)
		if err != nil {
			t.Fatal(err)
		}
		d := wav.NewDecoder(f)
		pcm := d.PCM()
		f.Close()
		if pcm.Size() != tc.size {
			t.Fatalf("expected the pcm to report containing %d frames but it has %d", tc.size, pcm.Size())
		}
		numChannels, bitDepth, sampleRate, err := pcm.Info()
		if numChannels != tc.numChannels {
			t.Fatalf("expected info to have %d channels but it has %d", tc.numChannels, numChannels)
		}
		if sampleRate != tc.sampleRate {
			t.Fatalf("expected info to have a sample rate of %d but it has %d", tc.sampleRate, sampleRate)
		}
		if bitDepth != tc.bitDepth {
			t.Fatalf("expected info to have %d bits per sample but it has %d", tc.bitDepth, bitDepth)
		}
	}
}
