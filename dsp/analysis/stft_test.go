package analysis

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-audio/transforms"
	"github.com/go-audio/wav"
	"github.com/mattetti/audio/dsp/windows"
)

func loadFixtureData(t *testing.T, fixturePath string) []float64 {
	path, _ := filepath.Abs(fixturePath)
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d := wav.NewDecoder(f)

	intBuf, err := d.FullPCMBuffer()
	if err != nil {
		t.Fatal(err)
	}
	fBuf := intBuf.AsFloatBuffer()
	transforms.MonoDownmix(fBuf)
	return fBuf.Data
}

func TestRoundtripping_STFT_ISTFT(t *testing.T) {
	var (
		testData            = loadFixtureData(t, "../../wav/fixtures/r9y916k.wav")
		testFrameLen        = []int{4096, 2048, 1024, 512}
		testFrameShiftDenom = []int{2, 3, 4, 5, 6, 7, 8} // 50% overlap 75% ...
		errTolerance        = 1.2
	)

	for _, frameLen := range testFrameLen {
		windowFunctions := prepareWindowFunctions(frameLen)
		for winI, win := range windowFunctions {
			for _, denom := range testFrameShiftDenom {
				testName := fmt.Sprintf("win size %d, windowFn %d, - hop size: %d", frameLen, winI, frameLen/denom)
				t.Run(testName, func(t *testing.T) {
					s := &STFT{
						FFTSize:    frameLen,
						WindowSize: frameLen,
						HopSize:    frameLen / denom,
						Window:     win,
					}

					reconstructed := s.ISTFT(s.STFT(testData))
					if containNAN(reconstructed) {
						t.Errorf("NAN contained, want non NAN contained.")
					}

					err := absErr(reconstructed, testData)
					if err > errTolerance {
						t.Errorf("[Frame length %d, hop size %d] %f error, want less than %f", frameLen, s.HopSize, err, errTolerance)
					}
				})
			}
		}
	}
}

func prepareWindowFunctions(frameLen int) [][]float64 {
	windowFunctions := make([][]float64, 4)
	windowFunctions[0] = windows.Hann(frameLen)
	windowFunctions[1] = windows.Hamming(frameLen)
	windowFunctions[2] = windows.Blackman(frameLen)
	windowFunctions[3] = windows.Gaussian(frameLen, 0.4)
	return windowFunctions
}

func containNAN(a []float64) bool {
	for _, val := range a {
		if math.IsNaN(val) {
			return true
		}
	}
	return false
}

func absErr(a, b []float64) float64 {
	length := len(a)
	if len(b) < length {
		length = len(b)
	}
	err := 0.0
	for i := 0; i < length; i++ {
		err += math.Abs(a[i] - b[i])
	}
	return err / float64(length)
}
