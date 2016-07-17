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

func TestClip_Next(t *testing.T) {
	testCases := []struct {
		desc         string
		input        string
		framesToRead int
		output       audio.Frames
	}{
		{"mono 16 bit, 22.5khz",
			"fixtures/kick.aif",
			8,
			audio.Frames{
				[]int{76}, []int{76}, []int{75}, []int{75}, []int{72}, []int{71}, []int{72}, []int{69},
			}},
		{"stereo 16 bit, 44khz",
			"fixtures/bloop.aif",
			8,
			audio.Frames{
				[]int{-22, -22}, []int{-110, -110}, []int{-268, -268}, []int{-441, -441}, []int{-550, -550}, []int{-553, -553}, []int{-456, -456}, []int{-269, -269},
			}},
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
		clip := d.PCM()
		if d.Err() != nil {
			t.Fatal(d.Err())
		}
		frames, n, err := clip.Next(tc.framesToRead)
		if err != nil {
			t.Fatal(err)
		}
		if n != tc.framesToRead {
			t.Fatalf("expected to read %d frames but read %d", tc.framesToRead, n)
		}
		if len(frames) <= 0 {
			t.Fatal("unexpected empty frames")
		}
		for i := 0; i < len(frames); i++ {
			for j := 0; j < len(frames[i]); j++ {
				if frames[i][j] != tc.output[i][j] {
					t.Fatalf("unexpected frame - ch: %d, frame #: %d, got: %d, expected: %d",
						j, i, frames[i][j], tc.output[i][j])
				}
			}
		}
	}
}
