package aiff_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mattetti/audio"
	"github.com/mattetti/audio/aiff"
)

func TestClip_Read(t *testing.T) {
	expectations := []struct {
		input       string
		totalFrames int
	}{
		{"fixtures/kick.aif", 4484},
		{"fixtures/delivery.aiff", 17199},
	}

	for _, exp := range expectations {
		path, _ := filepath.Abs(exp.input)
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		d := aiff.NewDecoder(f)
		clip := d.PCM()
		totalFrames := int(clip.Size())
		if totalFrames != exp.totalFrames {
			t.Fatalf("Expected %d frames, got %d\n", exp.totalFrames, totalFrames)
		}
		readFrames := 0

		bufSize := 4096
		buf := make([]byte, bufSize)
		var n int
		for readFrames < totalFrames {
			n, err = clip.Read(buf)
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

func TestClip_NextInts(t *testing.T) {
	testCases := []struct {
		desc          string
		input         string
		samplesToRead int
		output        audio.FramesInt
	}{
		{"mono 16 bit, 22.5khz",
			"fixtures/kick.aif",
			8,
			audio.FramesInt{
				76, 76, 75, 75, 72, 71, 72, 69,
			}},
		{"stereo 16 bit, 44khz",
			"fixtures/bloop.aif",
			8,
			audio.FramesInt{
				-22, -22, -110, -110, -268, -268, -441, -441,
			},
		},
	}

	for i, tc := range testCases {
		t.Logf("test case %d - %s\n", i, tc.desc)
		path, _ := filepath.Abs(tc.input)
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		d := aiff.NewDecoder(f)
		pcm := d.PCM()
		if d.Err() != nil {
			t.Fatal(d.Err())
		}
		numChannels, _, _, _ := pcm.Info()

		frames, err := pcm.NextInts(tc.samplesToRead)
		if err != nil {
			t.Fatal(err)
		}
		if len(frames) != tc.samplesToRead {
			t.Fatalf("expected to read %d samples but read %d", tc.samplesToRead, len(frames))
		}
		if len(frames) <= 0 {
			t.Fatal("unexpected empty frames")
		}

		for i := 0; i+numChannels < len(frames); {
			for j := 0; j < numChannels; j++ {
				if frames[i] != tc.output[i] {
					t.Logf("%#v\n", frames)
					t.Logf("%#v\n", tc.output)
					t.Fatalf("frame value at position %d: %d didn't match expected: %d", i, frames[i], tc.output[i])
				}
				i++
			}
		}
	}
}
