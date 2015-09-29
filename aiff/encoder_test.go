package aiff

import (
	"os"
	"testing"
)

func TestEncoderRoundTrip(t *testing.T) {
	os.Mkdir("testOutput", 0777)
	testCases := []struct {
		in  string
		out string
	}{
		// 22050, 16bit, mono
		{"fixtures/kick.aif", "testOutput/kick.aiff"},
		// 44100, 16bit, mono
		{"fixtures/subsynth.aif", "testOutput/subsynth.aif"},
		// 44100, 16bit, stereo
		{"fixtures/bloop.aif", "testOutput/bloop.aif"},
	}

	for i, tc := range testCases {
		t.Logf("%d - in: %s, out: %s", i, tc.in, tc.out)
		in, err := os.Open(tc.in)
		if err != nil {
			t.Fatalf("couldn't open %s %v", tc.in, err)
		}
		sampleRate, sampleSize, numChans, frames := ReadFrames(in)
		in.Close()

		out, err := os.Create(tc.out)
		if err != nil {
			t.Fatalf("couldn't create %s %v", tc.out, err)
		}
		defer out.Close()

		e := NewEncoder(out, sampleRate, sampleSize, numChans)
		e.Frames = frames
		if err := e.Write(); err != nil {
			t.Fatal(err)
		}
		out.Close()

		// TODO compare frames
		nf, err := os.Open(tc.out)
		if err != nil {
			t.Fatal(err)
		}

		nsampleRate, nsampleSize, nnumChans, nframes := ReadFrames(nf)
		nf.Close()
		if nsampleRate != sampleRate {
			t.Fatalf("sample rate didn't support roundtripping exp: %d, got: %d", sampleRate, nsampleRate)
		}
		if nsampleSize != sampleSize {
			t.Fatalf("sample size didn't support roundtripping exp: %d, got: %d", sampleSize, nsampleSize)
		}
		if nnumChans != numChans {
			t.Fatalf("the number of channels didn't support roundtripping exp: %d, got: %d", numChans, nnumChans)
		}

		if len(frames) != len(nframes) {
			t.Fatalf("the number of frames didn't support roundtripping, exp: %d, got: %d", len(frames), len(nframes))
		}
		for i := range frames {
			for j := 0; j < e.NumChans; j++ {
				if frames[i][j] != nframes[i][j] {
					t.Fatalf("frames[%d][%d]: %d didn't match nframes[%d][%d]: %d", i, j, frames[i][j], i, j, nframes[i][j])
				}
			}
		}
		os.Remove(nf.Name())
	}
}
