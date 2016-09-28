package analysis

import (
	"fmt"
	"go-dsp/fft"
	"math"

	"github.com/mattetti/audio"
)

// ConstantQConfig is the config used to do a Constant Q transform
// it defines the geometrically-spaced frequency range to analyze.
type ConstantQConfig struct {
	// Fs is the sample rate
	Fs float64
	// MinFs is the minimum frequency (suggested: 55 or midi 36)
	MinFs float64
	// MaxFs is the maximum frequency (suggested: Fs/2 or midi 96)
	MaxFs float64
	// BinsPerOctave is the numbers of bins per octave (suggested: 12 to 24)
	BinsPerOctave int
	// Threshold is the sensitivity threshold (suggested: 0.0054)
	Threshold float64
}

// ConstantQ transform transforms a data series to the frequency domain.
// It is related to the Fourier transform, and very closely related to the
// complex Morlet wavelet transform.
// In general, the transform is well suited to musical data
// since it mirrors the human auditory system, whereby at lower frequencies spectral resolution is better,
// whereas temporal resolution improves at higher frequencies.
//
// As the output of the transform is effectively amplitude/phase
// against log frequency, fewer frequency bins are required to cover a given range
// effectively, and this proves useful where frequencies span several octaves.
// As the range of human hearing covers approximately ten octaves from 20 Hz to around 20 kHz,
// this reduction in output data is significant.
// See https://en.wikipedia.org/wiki/Constant_Q_transform
type ConstantQ struct {
	Config        *ConstantQConfig
	Data          []float64
	ConstantQBins int
	// FFTLen is the length of the FFT required for the filter bank
	FFTLen int
	// Q value of the filter bank
	Q          float64
	K          int
	HopSize    int
	Speckernel *Speckernel
}

// NewConstantQ configures and returns a constant Q transform ready to be used.
func NewConstantQ(config *ConstantQConfig) *ConstantQ {
	cq := &ConstantQ{Config: config}
	cq.ConstantQBins = int(
		math.Ceil(
			float64(config.BinsPerOctave) *
				(math.Log(config.MaxFs/config.MinFs) /
					math.Log(2.0))),
	)
	cq.Q = 1 / (math.Pow(2, (1/(float64(config.BinsPerOctave)))) - 1)
	cq.FFTLen = int(
		math.Pow(2,
			nextpow2(
				math.Ceil(cq.Q*config.Fs/config.MinFs),
			),
		))
	cq.HopSize = int(float64(cq.FFTLen) / 8)
	cq.K = int(
		math.Ceil(
			float64(cq.ConstantQBins) *
				math.Log(config.MaxFs/config.MinFs) /
				math.Log(2.0),
		))
	return cq
}

func (cq *ConstantQ) Process(buf *audio.PCMBuffer) []float64 {
	if cq == nil {
		return nil
	}

	fftData := fft.FFTReal(buf.AsFloat64s())

	if cq.Speckernel == nil {
		cq.NewSpeckernel()
	}

	fftbin := cq.Speckernel.IS
	cqbin := cq.Speckernel.JS
	// reals := cq.Speckernel.Real
	// imags := cq.Speckernel.Imag
	sparseCells := len(fftbin)
	cq.Data = make([]float64, sparseCells)

	var (
		row, col int
		// r1, r2, i1, i2 float64
		c1, c2 complex128
	)
	for i := 0; i < sparseCells; i++ {
		row = cqbin[i]
		col = fftbin[i]
		// r1 = reals[i]
		// i1 = imags[i]

		c1 = cq.Speckernel.Values[i]

		idx := cq.FFTLen - col - 1
		c2 = fftData[idx]

		// r2 = real(fftData[cq.FFTLen-2*col-2])
		// i2 = imag(fftData[cq.FFTLen-2*col-2+1])
		cq.Data[row] += real(c1 * c2)
		fmt.Println(cq.Data[row])

		// cq.Data[row] += (r1*r2 - i1*i2)
		// cq.Data[row+1] += (r1*r2 + i1*i2)
	}

	return cq.Data
}

// NewSpeckernel generates a new spectral kernel matrix
func (cq *ConstantQ) NewSpeckernel() error {
	if cq == nil {
		return nil
	}
	cq.Speckernel = &Speckernel{}

	hammingWindowRe := make([]float64, cq.FFTLen)
	hammingWindowIm := make([]float64, cq.FFTLen)

	// for each bin value K, calculate temporal kernel, take its fft to
	// calculate the spectral kernel then threshold it to make it sparse and
	// add it to the sparse kernels matrix
	squareThreshold := cq.Config.Threshold * cq.Config.Threshold

	for k := cq.K; k > 0; k-- {
		// Computing a hamming window
		hammingLength :=
			math.Ceil(
				cq.Q * cq.Config.Fs /
					(cq.Config.MinFs *
						math.Pow(2,
							(cq.Q)/float64(cq.ConstantQBins))),
			)

		origin := cq.FFTLen/2 - int(hammingLength)/2

		for i := 0; i < int(hammingLength); i++ {
			angle := 2 * math.Pi * cq.Q * float64(i) / float64(hammingLength)
			real := math.Cos(angle)
			imag := math.Sin(angle)
			absol := hamming(hammingLength, float64(i)) / hammingLength
			hammingWindowRe[origin+i] = absol * real
			hammingWindowIm[origin+i] = absol * imag
		}

		for i := 0; i < cq.FFTLen/2; i++ {
			temp := hammingWindowRe[i]
			hammingWindowRe[i] = hammingWindowRe[i+cq.FFTLen/2]
			hammingWindowRe[i+cq.FFTLen/2] = temp
			temp = hammingWindowIm[i]
			hammingWindowIm[i] = hammingWindowIm[i+cq.FFTLen/2]
			hammingWindowIm[i+cq.FFTLen/2] = temp
		}

		hammingWindow := make([]complex128, len(hammingWindowRe))
		for i := 0; i < len(hammingWindow); i++ {
			hammingWindow[i] = complex(hammingWindowRe[i], hammingWindowIm[i])
		}

		transfHammingWindow := fft.FFT(hammingWindow)

		for j := 0; j < cq.FFTLen; j++ {
			// perform thresholding
			squaredBin := transfHammingWindow[j] * transfHammingWindow[j]
			if real(squaredBin) <= squareThreshold {
				continue
			}

			// Insert non-zero position indexes, doubled because they are floats
			cq.Speckernel.IS = append(cq.Speckernel.IS, j)
			cq.Speckernel.JS = append(cq.Speckernel.JS, k)

			// take conjugate, normalise and add to array sparkernel
			cq.Speckernel.Values = append(cq.Speckernel.Values,
				transfHammingWindow[j]/complex(float64(cq.FFTLen), 0))
			cq.Speckernel.Real = append(cq.Speckernel.Real, real(transfHammingWindow[j])/float64(cq.FFTLen))
			cq.Speckernel.Imag = append(cq.Speckernel.Imag, -imag(transfHammingWindow[j])/float64(cq.FFTLen))
		}
	}

	return nil
}

// Speckernel is a spectral kernel matrix.
type Speckernel struct {
	IS     []int
	JS     []int
	Imag   []float64
	Real   []float64
	Values []complex128
}

func hamming(len, n float64) float64 {
	return 0.54 - 0.46*math.Cos(2*math.Pi*n/len)
}
