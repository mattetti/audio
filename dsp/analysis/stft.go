package analysis

import (
	"go-dsp/fft"

	"github.com/mattetti/audio/dsp/windows"
)

// STFT is the Short Time Fourier Transform representation of a signal
// https://en.wikipedia.org/wiki/Short-time_Fourier_transform
// See https://github.com/r9y9/gossp/ where most of this code comes from.
type STFT struct {
	// FFTSize is the FFT window size
	FFTSize int
	// WindowSize is the size to use by the window function and will be padded
	// with zeros if shorter than FFTSize
	WindowSize int
	// HopSize number audio of frames between STFT columns
	HopSize int
	// Window is the window created byt a windowing function such as
	// windows.Hann and passing it the window size.
	Window []float64
}

func NewSTFT(fftSize int) *STFT {
	return &STFT{
		FFTSize:    fftSize,
		WindowSize: fftSize,
		HopSize:    fftSize / 4,
		Window:     windows.Hann(fftSize),
	}
}

// NumFrames returns the number of frames that will be analyzed in STFT.
func (s *STFT) numFrames(input []float64) int {
	return int(float64(len(input)-s.WindowSize)/float64(s.HopSize)) + 1
}

// DivideFrames returns overlapping divided frames for STFT.
func (s *STFT) DivideFrames(input []float64) [][]float64 {
	numFrames := s.numFrames(input)
	frames := make([][]float64, numFrames)
	for i := 0; i < numFrames; i++ {
		frames[i] = s.FrameAt(input, i)
	}
	return frames
}

// FrameAt returns frame at specified index given an input signal.
// Note that it doesn't make copy of input.
func (s *STFT) FrameAt(input []float64, index int) []float64 {
	return input[index*s.HopSize : index*s.HopSize+s.WindowSize]
}

// STFT returns complex spectrogram given an input signal.
func (s *STFT) STFT(input []float64) [][]complex128 {
	numFrames := s.numFrames(input)
	spectrogram := make([][]complex128, numFrames)

	frames := s.DivideFrames(input)
	for i, frame := range frames {
		// Windowing
		windowed := windows.Windowing(frame, s.Window)
		// Complex Spectrum
		spectrogram[i] = fft.FFTReal(windowed)
	}

	return spectrogram
}

// ISTFT performs invere STFT signal reconstruction and returns reconstructed
// signal.
func (s *STFT) ISTFT(spectrogram [][]complex128) []float64 {
	WindowSize := len(spectrogram[0])
	numFrames := len(spectrogram)
	reconstructedSignal := make([]float64, WindowSize+numFrames*s.HopSize)

	// Griffin's method
	windowSum := make([]float64, len(reconstructedSignal))
	for i := 0; i < numFrames; i++ {
		buf := fft.IFFT(spectrogram[i])
		index := 0
		for t := i * s.HopSize; t < i*s.HopSize+WindowSize; t++ {
			reconstructedSignal[t] += real(buf[index]) * s.Window[index]
			windowSum[t] += s.Window[index] * s.Window[index]
			index++
		}
	}

	// Normalize by window
	for n := range reconstructedSignal {
		if windowSum[n] > 1.0e-21 {
			reconstructedSignal[n] /= windowSum[n]
		}
	}

	return reconstructedSignal
}
