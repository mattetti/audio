// given a PCM audio file, convert it to mono and decimates it.
// Because of nyquist law, we can't simply average or drop samples otherwise we will have alisasing.
// A proper decimation is needed http://dspguru.com/dsp/faqs/multirate/decimation
// Decimation is useful to reduce the amount of data to process.
// Max hearing frequency would be around 20kHz, so we need to low pass to remove anything above 20kHz so
// we don't get any aliasing.
package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattetti/audio/aiff"
	"github.com/mattetti/audio/dsp/filters"
	"github.com/mattetti/audio/dsp/windows"
	"github.com/mattetti/audio/generator"
	"github.com/mattetti/audio/misc"
	"github.com/mattetti/audio/riff/wav"
)

var (
	fileFlag   = flag.String("file", "", "file to downsample (copy will be done)")
	factorFlag = flag.Int("factor", 2, "The decimator factor divides the sampling rate")
	outputFlag = flag.String("format", "aiff", "output format, aiff or wav")
)

func main() {
	flag.Parse()

	if *fileFlag == "" {
		freq := 440
		fs := 44100
		bitSize := 16
		fmt.Printf("Target fs: %d\n", fs / *factorFlag)

		// generate a wave sine
		osc := generator.NewOsc(generator.WaveSine, float64(freq), fs)
		data := osc.Signal(fs * 4)

		// sinc function to run a low pass filter
		s := &filters.Sinc{
			Taps:         62,
			SamplingFreq: fs,
			CutOffFreq:   20000,
			Window:       windows.Blackman,
		}
		fir := &filters.FIR{Sinc: s}
		filtered, err := fir.LowPass(data)
		if err != nil {
			panic(err)
		}

		// our osc generates values from -1 to 1, we need to go back to PCM scale
		factor := float64(intMaxSignedValue(bitSize))
		// build the audio frames
		frames := make([][]int, len(data) / *factorFlag)
		for i := 0; i < len(frames); i++ {
			frames[i] = []int{int(filtered[i**factorFlag] * factor)}
		}

		// generate the sound file
		o, err := os.Create("resampled.aiff")
		if err != nil {
			panic(err)
		}
		defer o.Close()
		e := aiff.NewEncoder(o, fs / *factorFlag, 16, 1)
		e.Frames = frames
		if err := e.Write(); err != nil {
			panic(err)
		}
		return
	}

	ext := filepath.Ext(*fileFlag)
	var codec string
	switch strings.ToLower(ext) {
	case ".aif", ".aiff":
		codec = "aiff"
	case ".wav", ".wave":
		codec = "wav"
	default:
		fmt.Printf("files with extension %s not supported\n", ext)
	}

	f, err := os.Open(*fileFlag)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var monoFrames misc.AudioFrames
	var sampleRate int
	var sampleSize int
	switch codec {
	case "aiff":
		d := aiff.NewDecoder(f)
		frames, err := d.Frames()
		if err != nil {
			panic(err)
		}
		sampleRate = d.SampleRate
		sampleSize = int(d.BitDepth)
		monoFrames = misc.ToMonoFrames(frames)

	case "wav":
		info, frames, err := wav.NewDecoder(f, nil).ReadFrames()
		if err != nil {
			panic(err)
		}
		sampleRate = int(info.SampleRate)
		sampleSize = int(info.BitsPerSample)
		monoFrames = misc.ToMonoFrames(frames)
	}

	fmt.Printf("undersampling -> %s file at %dHz to %d samples (%d)\n", codec, sampleRate, sampleRate / *factorFlag, sampleSize)

	switch sampleRate {
	case 44100:
	case 48000:
	default:
		log.Fatalf("input sample rate of %dHz not supported", sampleRate)
	}

	amplitudesF := make([]float64, len(monoFrames))
	for i, f := range monoFrames {
		amplitudesF[i] = float64(f[0])
	}

	// low pass filter before we drop some samples to avoid aliasing
	s := &filters.Sinc{Taps: 62, SamplingFreq: sampleRate, CutOffFreq: float64(sampleRate / 2), Window: windows.Blackman}
	fir := &filters.FIR{Sinc: s}
	filtered, err := fir.LowPass(amplitudesF)
	if err != nil {
		panic(err)
	}
	frames := make([][]int, len(amplitudesF) / *factorFlag)
	for i := 0; i < len(frames); i++ {
		frames[i] = []int{int(filtered[i**factorFlag])}
	}

	of, err := os.Create("resampled.aiff")
	if err != nil {
		panic(err)
	}
	defer of.Close()
	aiffe := aiff.NewEncoder(of, sampleRate / *factorFlag, sampleSize, 1)
	aiffe.Frames = frames
	if err := aiffe.Write(); err != nil {
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

//This function calculates the zeroth order Bessel function
func Ino(x float64) float64 {
	var (
		d  float64 = 0
		ds float64 = 1
		s  float64 = 1
	)
	for ds > float64(s*1e-6) {
		d += 2
		ds *= x * x / (d * d)
		s += ds
	}
	return s
}

type filterParams struct {
	Fs, Fa, Fb, Att float64
}

func (f *filterParams) Window(M int) []float64 {
	var (
		Np    = int(float64(M-1) / 2)
		A     = make([]float64, int(Np))
		Alpha float64
		//pi = math.PI
		//Inoalpha
		H   = make([]float64, M)
		Fs  = f.Fs
		Fa  = f.Fa
		Fb  = f.Fb
		Att = f.Att
	)

	// Calculate the impulse response of the ideal filter
	A[0] = 2 * (Fb - Fa) / Fs
	for j := 1; j < int(Np); j++ {
		A[j] = (math.Sin(2 * float64(j) * math.Pi * Fb / Fs)) - (math.Sin(2*float64(j)*math.Pi*Fa/Fs) / (float64(j) * math.Pi))
	}

	// Calculate the desired shape factor for the Kaiser-Bessel window
	if Att < 21 {
		Alpha = 0
	} else if Att > 50 {
		Alpha = 0.1102 * (Att - 8.7)
	} else {
		Alpha = 0.5842*math.Pow((Att-21), 0.4) + 0.07886*(Att-21)
	}

	// Window the ideal response with the Kaiser-Bessel window
	Inoalpha := Ino(Alpha)
	for j := 0; j < Np; j++ {
		H[Np+j] = A[j] * Ino(Alpha*math.Sqrt(1-(float64(j*j)/float64(Np*Np)))) / Inoalpha
	}
	for j := 0; j < Np; j++ {
		H[j] = H[M-1-j]
	}

	return H
}

/*
 * This function calculates Kaiser windowed
 * FIR filter coefficients for a single passband
 * based on
 * "DIGITAL SIGNAL PROCESSING, II" IEEE Press pp 123-126.
 *
 * Fs=Sampling frequency
 * Fa=Low freq ideal cut off (0=low pass)
 * Fb=High freq ideal cut off (Fs/2=high pass)
 * Att=Minimum stop band attenuation (>21dB)
 * M=Number of points in filter (ODD number)
 * H[] holds the output coefficients (they are symetric only half generated)
 */
func calcFilter(Fs, Fa, Fb float64, M int, Att float64) []float64 {
	var (
		Np    = int(float64(M-1) / 2)
		A     = make([]float64, int(Np))
		Alpha float64
		//pi = math.PI
		//Inoalpha
		H = make([]float64, M)
	)

	// Calculate the impulse response of the ideal filter
	A[0] = 2 * (Fb - Fa) / Fs
	for j := 1; j < int(Np); j++ {
		A[j] = (math.Sin(2 * float64(j) * math.Pi * Fb / Fs)) - (math.Sin(2*float64(j)*math.Pi*Fa/Fs) / (float64(j) * math.Pi))
	}

	// Calculate the desired shape factor for the Kaiser-Bessel window
	if Att < 21 {
		Alpha = 0
	} else if Att > 50 {
		Alpha = 0.1102 * (Att - 8.7)
	} else {
		Alpha = 0.5842*math.Pow((Att-21), 0.4) + 0.07886*(Att-21)
	}

	// Window the ideal response with the Kaiser-Bessel window
	Inoalpha := Ino(Alpha)
	for j := 0; j < Np; j++ {
		H[Np+j] = A[j] * Ino(Alpha*math.Sqrt(1-(float64(j*j)/float64(Np*Np)))) / Inoalpha
	}
	for j := 0; j < Np; j++ {
		H[j] = H[M-1-j]
	}

	return H
}
