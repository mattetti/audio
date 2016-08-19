// generator example
package main

import (
	"flag"
	"os"

	"fmt"
	"io"
	"strings"

	"github.com/mattetti/audio"
	"github.com/mattetti/audio/generator"
	"github.com/mattetti/audio/wav"
)

var (
	freqFlag      = flag.Int("freq", 440, "frequency to generate")
	biteDepthFlag = flag.Int("biteDepth", 16, "bit size to use when generating the auid file")
	durationFlag  = flag.Int("duration", 4, "duration of the generated file")
	formatFlag    = flag.String("format", "wav", "the audio format of the output file")
)

func main() {
	flag.Parse()

	freq := *freqFlag
	fs := 44100
	biteDepth := *biteDepthFlag

	osc := generator.NewOsc(generator.WaveSine, float64(freq), fs)
	// our osc generates values from -1 to 1, we need to go back to PCM scale
	factor := float64(audio.IntMaxSignedValue(biteDepth))
	osc.Amplitude = factor
	data := make([]float64, fs**durationFlag)
	buf := audio.NewPCMFloatBuffer(data, audio.FormatMono4410016bBE)
	osc.Fill(buf)

	// generate the sound file
	var outName string
	var format string
	switch strings.ToLower(*formatFlag) {
	case "aif", "aiff":
		format = "aif"
		outName = "generated.aiff"
	default:
		format = "wav"
		outName = "generated.wav"
	}

	o, err := os.Create(outName)
	if err != nil {
		panic(err)
	}
	defer o.Close()
	if err := encode(format, buf, o); err != nil {
		panic(err)
	}
	fmt.Println(outName, "generated")
}

func encode(format string, buf *audio.PCMBuffer, w io.WriteSeeker) error {
	// switch format {
	// case "wav":
	e := wav.NewEncoder(w, buf.Format.SampleRate, buf.Format.BitDepth, buf.Format.NumChannels, 1)
	// }
	// e := aiff.NewEncoder(w, fs, bitDepth, 1)
	samples := buf.AsInts()
	if err := e.Write(samples); err != nil {
		return err
	}
	return e.Close()
}
