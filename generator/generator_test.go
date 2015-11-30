package generator

import (
	"math"
	"testing"
)

func TestSine(t *testing.T) {
	testCases := []struct {
		in  float64
		out float64
	}{
		{-math.Pi, 0},
		{0.007, 0.006909727339533104},
		{-0.5, -0.47932893655759223},
		{0.1, 0.09895415534087945},
		{1.5862234, 0.9998818440160414},
		{2.0, 0.909795856141705},
		{3.0, 0.14008939955174454},
		{math.Pi, 0},
	}

	for i, tc := range testCases {
		if out := Sine(tc.in); out != tc.out {
			t.Logf("[%d] sine(%f) => %f != %f", i, tc.in, out, tc.out)
			t.Fail()
		}
	}
}
