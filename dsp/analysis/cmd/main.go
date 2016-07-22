package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattetti/audio"
	"github.com/mattetti/audio/aiff"
	"github.com/mattetti/audio/dsp/analysis"
	"github.com/mattetti/audio/wav"
)

func main() {
	path, _ := filepath.Abs("../../../decimator/beat.aiff")
	ext := filepath.Ext(path)
	var codec string
	switch strings.ToLower(ext) {
	case ".aif", ".aiff":
		codec = "aiff"
	case ".wav", ".wave":
		codec = "wav"
	default:
		fmt.Printf("files with extension %s not supported\n", ext)
	}

	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var monoFrames audio.SamplesInt
	var sampleRate int
	var sampleSize int
	switch codec {
	case "aiff":
		d := aiff.NewDecoder(f)
		frames, err := d.SamplesInt()
		if err != nil {
			panic(err)
		}
		sampleRate = d.SampleRate
		sampleSize = int(d.BitDepth)
		monoFrames = frames.StereoToMono()
	case "wav":
		d := wav.NewDecoder(f)
		frames, err := d.SamplesInt()
		if err != nil {
			panic(err)
		}
		sampleRate = int(d.SampleRate)
		sampleSize = int(d.BitDepth)
		monoFrames = frames.StereoToMono()
	}

	data := make([]float64, len(monoFrames))
	for i, f := range monoFrames {
		data[i] = float64(f)
	}
	dft := analysis.NewDFT(sampleRate, data)
	sndData := dft.IFFT()
	frames := make([]int, len(sndData))
	for i := 0; i < len(frames); i++ {
		frames[i] = int(sndData[i])
	}
	of, err := os.Create("roundtripped.aiff")
	if err != nil {
		panic(err)
	}
	defer of.Close()
	aiffe := aiff.NewEncoder(of, sampleRate, sampleSize, 1)
	if err := aiffe.Write(frames); err != nil {
		panic(err)
	}
	if err := aiffe.Close(); err != nil {
		panic(err)
	}
}
