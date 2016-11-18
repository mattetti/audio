package mp3_test

import (
	"os"
	"testing"

	"github.com/mattetti/audio/mp3"
)

func Test_SeemsValid(t *testing.T) {
	testCases := []struct {
		input   string
		isValid bool
	}{
		{"fixtures/frame.mp3", true},
		{"fixtures/HousyStab.mp3", true},
		{"../wav/fixtures/bass.wav", false},
	}

	for i, tc := range testCases {
		t.Logf("test case %d\n", i)
		f, err := os.Open(tc.input)
		if err != nil {
			panic(err)
		}
		if o := mp3.SeemsValid(f); o != tc.isValid {
			t.Fatalf("expected %t\ngot\n%t\n", tc.isValid, o)
		}
		f.Close()
	}
}
