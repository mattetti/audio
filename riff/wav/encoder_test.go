package wav

import (
	"os"
	"testing"
)

func TestEncoderRoundTrip(t *testing.T) {
	os.Mkdir("testOutput", 0777)
	testCases := []struct {
		in   string
		out  string
		desc string
	}{
		// 22050, 8bit, mono
		{"fixtures/kick.wav", "testOutput/kick.wav", "22050 Hz @ 16 bits, 1 channel(s), 44100 avg bytes/sec, duration: 204.172335ms"},
		{"fixtures/bass.wav", "testOutput/bass.wav", "44100 Hz @ 24 bits, 2 channel(s), 264600 avg bytes/sec, duration: 543.378684ms"},
	}

	for i, tc := range testCases {
		t.Logf("%d - in: %s, out: %s", i, tc.in, tc.out)
		in, err := os.Open(tc.in)
		if err != nil {
			t.Fatalf("couldn't open %s %v", tc.in, err)
		}
		d := NewDecoder(in)
		info, frames, err := d.ReadFrames()
		if info.String() != tc.desc {
			t.Fatalf("expected the parser to report something else than %s", info)
		}
		if err != nil {
			t.Fatal(err)
		}
		in.Close()

		out, err := os.Create(tc.out)
		if err != nil {
			t.Fatalf("couldn't create %s %v", tc.out, err)
		}
		defer out.Close()

		e := NewEncoder(out, info)
		e.Frames = frames
		if err := e.Write(); err != nil {
			t.Fatal(err)
		}
		out.Close()

		nf, err := os.Open(tc.out)
		if err != nil {
			t.Fatal(err)
		}

		nd := NewDecoder(nf)
		ninfo, nframes, err := nd.ReadFrames()
		nf.Close()
		if err != nil {
			t.Fatal(err)
		}
		if ninfo.SampleRate != info.SampleRate {
			t.Fatalf("sample rate didn't support roundtripping exp: %d, got: %d", info.SampleRate, ninfo.SampleRate)
		}
		if ninfo.BitsPerSample != info.BitsPerSample {
			t.Fatalf("sample size didn't support roundtripping exp: %d, got: %d", info.BitsPerSample, ninfo.BitsPerSample)
		}
		if ninfo.NumChannels != info.NumChannels {
			t.Fatalf("the number of channels didn't support roundtripping exp: %d, got: %d", info.NumChannels, ninfo.NumChannels)
		}

		if len(frames) != len(nframes) {
			t.Fatalf("the number of frames didn't support roundtripping, exp: %d, got: %d", len(frames), len(nframes))
		}
		for i := range frames {
			for j := 0; j < e.NumChannels; j++ {
				if frames[i][j] != nframes[i][j] {
					t.Fatalf("frames[%d][%d]: %d didn't match nframes[%d][%d]: %d", i, j, frames[i][j], i, j, nframes[i][j])
				}
			}
		}
		os.Remove(nf.Name())
	}
}
