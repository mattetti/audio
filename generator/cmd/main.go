// generator example
package main

import (
	"flag"
	"os"

	"github.com/mattetti/audio"
	"github.com/mattetti/audio/aiff"
	"github.com/mattetti/audio/generator"
)

var (
	freqFlag     = flag.Int("freq", 440, "frequency to generate")
	bitSizeFlag  = flag.Int("bitsize", 16, "bit size to use when generating the auid file")
	durationFlag = flag.Int("duration", 4, "duration of the generated file")
	// TODO: support waveform types
)

func main() {
	flag.Parse()

	freq := *freqFlag
	fs := 44100
	bitSize := *bitSizeFlag

	osc := generator.NewOsc(generator.WaveSine, float64(freq), fs)
	// our osc generates values from -1 to 1, we need to go back to PCM scale
	factor := float64(audio.IntMaxSignedValue(bitSize))
	osc.Amplitude = factor
	// xs of sound
	data := osc.Signal(fs * *durationFlag)
	// build the audio frames
	frames := make([][]int, len(data))
	for i := 0; i < len(frames); i++ {
		frames[i] = []int{int(data[i])}
	}

	// generate the sound file
	o, err := os.Create("generated.aiff")
	if err != nil {
		panic(err)
	}
	defer o.Close()
	e := aiff.NewEncoder(o, fs, 16, 1)
	e.Frames = frames
	if err := e.Write(); err != nil {
		panic(err)
	}
}

func intMaxSignedValue(b int) int {
	switch b {
	case 8:
		return 255 / 2
	case 16:
		return 65535 / 2
	case 24:
		return 16777215 / 2
	case 32:
		return 4294967295 / 2
	default:
		return 0
	}
}
