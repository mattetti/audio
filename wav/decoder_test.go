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
		{"fixtures/kick-16b441k.wav",
			"2 ch,  44100 Hz, 'lpcm' (0x0000000C) 16-bit little-endian signed integer",
			[]int{0, 0, 0, 0, 0, 0, 3, 3, 28, 28, 130, 130, 436, 436, 1103, 1103, 2140, 2140, 3073, 3073, 2884, 2884, 760, 760, -2755, -2755, -5182, -5182, -3860, -3860, 1048, 1048, 5303, 5303, 3885, 3885, -3378, -3378, -9971, -9971, -8119, -8119, 2616, 2616, 13344, 13344, 13297, 13297, 553, 553, -15013, -15013, -20341, -20341, -10692, -10692, 6553, 6553, 18819, 18819, 18824, 18824, 8617, 8617, -4253, -4253, -13305, -13305, -16289, -16289, -13913, -13913, -7552, -7552, 1334, 1334, 10383, 10383, 16409, 16409, 16928, 16928, 11771, 11771, 3121, 3121, -5908, -5908, -12829, -12829, -16321, -16321, -15990, -15990, -12025, -12025, -5273, -5273, 2732, 2732, 10094, 10094, 15172, 15172, 17038, 17038, 15563, 15563, 11232, 11232, 4973, 4971, -2044, -2044, -8602, -8602, -13659, -13659, -16458, -16458, -16574, -16575, -14012, -14012, -9294, -9294, -3352, -3352, 2823, 2823, 8485, 8485, 13125, 13125, 16228, 16228, 17214, 17214, 15766, 15766, 12188, 12188, 7355, 7355, 2152, 2152, -2973, -2973, -7929, -7929, -12446, -12446, -15806, -15806, -17161, -17161, -16200, -16200, -13407, -13407, -9681, -9681, -5659, -5659, -1418, -1418, 3212, 3212, 8092, 8092, 12567, 12567, 15766, 15766, 17123, 17123, 16665, 16665, 14863, 14863, 12262, 12262, 9171, 9171, 5644, 5644, 1636, 1636, -2768, -2768, -7262, -7262, -11344, -11344, -14486, -14486, -16310, -16310, -16710, -16710, -15861, -15861, -14093, -14093, -11737, -11737, -8974, -8974, -5840, -5840, -2309, -2309, 1577, 1577, 5631, 5631, 9510, 9510, 12821, 12821, 15218, 15218, 16500, 16500, 16663, 16663, 15861, 15861, 14338, 14338, 12322, 12322, 9960, 9960},
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
