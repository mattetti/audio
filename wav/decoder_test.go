package wav_test

import (
	"os"
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

func TestDecoder_Clip(t *testing.T) {
	testCases := []struct {
		in   string
		size int64
		info audio.FrameInfo
	}{
		{in: "fixtures/kick.wav",
			size: 4484,
			info: audio.FrameInfo{
				Channels:   2,
				SampleRate: 22050,
				BitDepth:   16,
			},
		},
	}

	for _, tc := range testCases {
		f, err := os.Open(tc.in)
		if err != nil {
			t.Fatal(err)
		}
		d := wav.NewDecoder(f)
		clip := d.Clip()
		f.Close()
		if clip.Size() != tc.size {
			t.Fatalf("expected the clip to report containing %d frames but it has %d", tc.size, clip.Size())
		}
		info := clip.FrameInfo()
		if info.SampleRate != tc.info.SampleRate {
			t.Fatalf("expected info to have a sample rate of %d but it has %d", tc.info.SampleRate, info.SampleRate)
		}
		if info.BitDepth != tc.info.BitDepth {
			t.Fatalf("expected info to have %d bits per sample but it has %d", tc.info.BitDepth, info.BitDepth)
		}
	}
}
