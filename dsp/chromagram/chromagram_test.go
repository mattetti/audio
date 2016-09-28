package chromagram_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mattetti/audio/dsp/analysis"
	"github.com/mattetti/audio/dsp/chromagram"
	"github.com/mattetti/audio/wav"
)

func Test_ChromagramProcess(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{"../../wav/fixtures/440hz.wav", ""},
	}

	config := &analysis.ConstantQConfig{
		Fs:            44100,
		MinFs:         55,
		MaxFs:         44100 / 2,
		BinsPerOctave: 12,
		Threshold:     0.0054,
	}
	chromaG := chromagram.New(config)
	for i, tc := range testCases {
		t.Logf("test case %d\n", i)
		path, err := filepath.Abs(tc.input)
		if err != nil {
			t.Fatal(err)
		}
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		d := wav.NewDecoder(f)
		buf, err := d.FullPCMBuffer()
		if err != nil {
			t.Fatal(err)
		}
		gram, err := chromaG.Process(buf)
		if err != nil {
			t.Error(err)
		}
		t.Logf("%+v\n", gram)
	}
}
