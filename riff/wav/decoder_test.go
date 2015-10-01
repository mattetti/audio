package wav

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"
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
		dur, err := NewDecoder(f).Duration()
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
		if dur != tc.duration {
			t.Fatalf("expected duration to be: %s but was %s", tc.duration, dur)
		}
	}

}

func ExampleDecoder_Duration() {
	f, err := os.Open("fixtures/kick.wav")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	dur, err := NewDecoder(f).Duration()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s duration: %s\n", f.Name(), dur)
	// Output: fixtures/kick.wav duration: 204.172335ms
}

func TestDecoder_Info(t *testing.T) {
	testCases := []struct {
		in   string
		info *Info
	}{
		{"fixtures/kick.wav",
			&Info{
				NumChannels:    1,
				SampleRate:     22050,
				AvgBytesPerSec: 44100,
				BitsPerSample:  16,
			},
		},
	}

	for _, tc := range testCases {
		f, err := os.Open(tc.in)
		if err != nil {
			t.Fatal(err)
		}
		info, err := NewDecoder(f).Info()
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
		if info.NumChannels != tc.info.NumChannels {
			t.Fatalf("expected info to have %d channels but it has %s", tc.info.NumChannels, info.NumChannels)
		}
		if info.SampleRate != tc.info.SampleRate {
			t.Fatalf("expected info to have a sample rate of %d but it has %s", tc.info.SampleRate, info.SampleRate)
		}
		if info.AvgBytesPerSec != tc.info.AvgBytesPerSec {
			t.Fatalf("expected info to have %d avg bytes per sec but it has\n%s", tc.info.AvgBytesPerSec, info.AvgBytesPerSec)
		}
		if info.BitsPerSample != tc.info.BitsPerSample {
			t.Fatalf("expected info to have %d bits per sample but it has\n%s", tc.info.BitsPerSample, info.BitsPerSample)
		}
	}

}

func ExampleDecoder_Info() {
	f, err := os.Open("fixtures/kick.wav")
	if err != nil {
		log.Fatal(err)
	}
	info, err := NewDecoder(f).Info()
	if err != nil {
		log.Fatal(err)
	}
	f.Close()
	fmt.Println(info)
	// Output: 22050 Hz @ 16 bits, 1 channel(s), 44100 avg bytes/sec
}
