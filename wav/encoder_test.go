package wav_test

import (
	"os"
	"testing"

	"github.com/mattetti/audio/wav"
)

func TestEncoderRoundTrip(t *testing.T) {
	os.Mkdir("testOutput", 0777)
	testCases := []struct {
		in   string
		out  string
		desc string
	}{
		{"fixtures/kick.wav", "testOutput/kick.wav", "22050 Hz @ 16 bits, 1 channel(s), 44100 avg bytes/sec, duration: 204.172335ms"},
		{"fixtures/kick-16b441k.wav", "testOutput/kick-16b441k.wav", "2 ch,  44100 Hz, 'lpcm' 16-bit little-endian signed integer"},
		{"fixtures/bass.wav", "testOutput/bass.wav", "44100 Hz @ 24 bits, 2 channel(s), 264600 avg bytes/sec, duration: 543.378684ms"},
	}

	for i, tc := range testCases {
		t.Logf("%d - in: %s, out: %s\n%s", i, tc.in, tc.out, tc.desc)
		in, err := os.Open(tc.in)
		if err != nil {
			t.Fatalf("couldn't open %s %v", tc.in, err)
		}
		d := wav.NewDecoder(in)
		pcm := d.PCM()
		numChannels, bitDepth, sampleRate, err := pcm.Info()
		if err != nil {
			t.Fatal(err)
		}
		totalFrames := pcm.Size()
		frames, err := d.FramesInt()
		if err != nil {
			t.Fatal(err)
		}
		in.Close()
		t.Logf("%s - total frames %d - total samples %d", tc.in, totalFrames, len(frames))

		out, err := os.Create(tc.out)
		if err != nil {
			t.Fatalf("couldn't create %s %v", tc.out, err)
		}

		e := wav.NewEncoder(out,
			int(sampleRate),
			bitDepth,
			numChannels,
			int(d.WavAudioFormat))
		if err := e.Write(frames); err != nil {
			t.Fatal(err)
		}
		if err := e.Close(); err != nil {
			t.Fatal(err)
		}
		out.Close()

		nf, err := os.Open(tc.out)
		if err != nil {
			t.Fatal(err)
		}

		nd := wav.NewDecoder(nf)
		nPCM := nd.PCM()
		if nPCM == nil {
			t.Fatalf("couldn't extract the PCM from %s - %v", nf.Name(), d.Err())
		}
		nNumChannels, nBitDepth, nSampleRate, err := nPCM.Info()
		nTotalFrames := nPCM.Size()
		nframes, err := nd.FramesInt()
		if err != nil {
			t.Fatal(err)
		}

		nf.Close()
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := os.Remove(nf.Name()); err != nil {
				panic(err)
			}
		}()

		if nSampleRate != sampleRate {
			t.Fatalf("sample rate didn't support roundtripping exp: %d, got: %d", sampleRate, nSampleRate)
		}
		if nBitDepth != bitDepth {
			t.Fatalf("sample size didn't support roundtripping exp: %d, got: %d", bitDepth, nBitDepth)
		}
		if nNumChannels != numChannels {
			t.Fatalf("the number of channels didn't support roundtripping exp: %d, got: %d", numChannels, nNumChannels)
		}
		if totalFrames != nTotalFrames {
			t.Fatalf("the reported number of frames didn't support roundtripping, exp: %d, got: %d", totalFrames, nTotalFrames)
		}
		if len(frames) != len(nframes) {
			t.Fatalf("the number of frames didn't support roundtripping, exp: %d, got: %d", len(frames), len(nframes))
		}
		for i := 0; i < len(frames); i++ {
			if frames[i] != nframes[i] {
				t.Fatalf("frame value at position %d: %d didn't match nframes position %d: %d", i, frames[i], i, nframes[i])
			}
		}

	}
}
