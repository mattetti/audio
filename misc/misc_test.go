package misc

import (
	"bytes"
	"reflect"
	"testing"
)

func TestToMonoFrames(t *testing.T) {
	input := AudioFrames([][]int{{2, 4}, {2, 2}, {-2, 2}, {10, 20}})
	output := ToMonoFrames(input)
	expected := AudioFrames([][]int{{3}, {2}, {0}, {15}})
	if !reflect.DeepEqual(output, expected) {
		t.Fatalf("expected:\t%q\n got:\t\t\t%q\n", expected, output)
	}
}

func TestIeeeFloat(t *testing.T) {
	testCases := []struct {
		ieeeFloat [10]byte
		intValue  int
	}{
		{
			[10]byte{0x40, 0x0D, 0xAC, 0x44, 00, 00, 00, 00, 00, 00},
			22050,
		},
		{
			[10]byte{0x40, 0x0E, 0xAC, 0x44, 00, 00, 00, 00, 00, 00},
			44100,
		},
		{
			[10]byte{0x40, 0x0E, 0xBB, 0x80, 00, 00, 00, 00, 00, 00},
			48000,
		},
		{
			[10]byte{0x40, 0x0F, 0xBB, 0x80, 00, 00, 00, 00, 00, 00},
			96000,
		},
	}

	for _, tc := range testCases {
		t.Logf("%d -> % X\n", tc.intValue, tc.ieeeFloat)
		if IeeeFloatToInt(tc.ieeeFloat) != tc.intValue {
			t.Logf("% X didn't convert to %d", tc.ieeeFloat, tc.intValue)
		}

		bs := IntToIeeeFloat(tc.intValue)
		if bytes.Compare(bs[:], tc.ieeeFloat[:]) != 0 {
			t.Fatalf("%d didn't convert to % X but to % X", tc.intValue, tc.ieeeFloat, bs)
		}
	}
}
