package aiff_test

import (
	"bytes"
	"encoding/hex"
	"os"
	"testing"

	"github.com/mattetti/audio/aiff"
)

// TODO(mattetti): switch to using github.com/mattetti/filebuffer

func TestEncoderRoundTrip(t *testing.T) {
	os.Mkdir("testOutput", 0777)
	testCases := []struct {
		in  string
		out string
		// a round trip decoding/encoding doesn't mean we get a perfect match
		// in this test, we do drop extra chunks
		perfectMatch bool
	}{
		// 22050, 8bit, mono
		{"fixtures/kick8b.aiff", "testOutput/kick8b.aiff", true},
		// 22050, 16bit, mono (extra chunk)
		{"fixtures/kick.aif", "testOutput/kick.aif", false},
		// 22050, 16bit, mono
		{"fixtures/kick32b.aiff", "testOutput/kick32b.aiff", true},
		// 44100, 16bit, mono
		{"fixtures/subsynth.aif", "testOutput/subsynth.aif", true},
		// 44100, 16bit, stereo
		{"fixtures/bloop.aif", "testOutput/bloop.aif", true},
		// 48000, 16bit, stereo
		{"fixtures/zipper.aiff", "testOutput/zipper.aiff", true},
		// 48000, 24bit, stereo
		{"fixtures/zipper24b.aiff", "testOutput/zipper24b.aiff", true},
	}

	for i, tc := range testCases {
		t.Logf("%d - in: %s, out: %s", i, tc.in, tc.out)
		in, err := os.Open(tc.in)
		if err != nil {
			t.Fatalf("couldn't open %s %v", tc.in, err)
		}
		d := aiff.NewDecoder(in)
		frames, err := d.Frames()
		if err != nil {
			t.Fatal(err)
		}
		defer in.Close()

		out, err := os.Create(tc.out)
		if err != nil {
			t.Fatalf("couldn't create %s %v", tc.out, err)
		}

		e := aiff.NewEncoder(out, int(d.SampleRate), int(d.BitDepth), int(d.NumChans))
		e.Frames = frames
		if err := e.Write(); err != nil {
			t.Fatal(err)
		}

		nf, err := os.Open(tc.out)
		if err != nil {
			t.Fatal(err)
		}

		d2 := aiff.NewDecoder(nf)
		d2.PCM()
		// TODO(mattetti): using d2.Duration() messes the later Frames() call
		info, err := nf.Stat()
		if err != nil {
			t.Fatal(err)
		}
		expectedHeaderSize := info.Size() - 8
		if d.Size != d2.Size {
			t.Logf("the encoded size didn't match the original, expected: %d, got %d", d.Size, d2.Size)
		}
		if expectedHeaderSize != int64(d2.Size) {
			t.Fatalf("wrong header size data, expected %d, got %d", expectedHeaderSize, d2.Size)
		}
		nframes, err := d2.Frames()
		if err != nil {
			t.Fatal(err)
		}
		if d2.SampleRate != d.SampleRate {
			t.Fatalf("sample rate didn't support roundtripping exp: %d, got: %d", d.SampleRate, d2.SampleRate)
		}
		if d2.BitDepth != d.BitDepth {
			t.Fatalf("sample size didn't support roundtripping exp: %d, got: %d", d.BitDepth, d2.BitDepth)
		}
		if d2.NumChans != d.NumChans {
			t.Fatalf("the number of channels didn't support roundtripping exp: %d, got: %d", d.NumChans, d2.NumChans)
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

		if tc.perfectMatch {
			// binary comparison
			in.Seek(0, 0)
			nf.Seek(0, 0)
			buf1 := make([]byte, 32)
			buf2 := make([]byte, 32)

			var err1, err2 error
			var n int
			readBytes := 0
			for err1 == nil && err2 == nil {
				n, err1 = in.Read(buf1)
				_, err2 = nf.Read(buf2)
				readBytes += n
				if bytes.Compare(buf1, buf2) != 0 {
					t.Fatalf("round trip failed, data differed after %d bytes\n%s\n%s", readBytes, hex.Dump(buf1), hex.Dump(buf2))
				}
			}
		}

		nf.Close()
		os.Remove(nf.Name())
	}
}
