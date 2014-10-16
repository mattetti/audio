package midi

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestVarint(t *testing.T) {
	expecations := []struct {
		dec   uint32
		bytes []byte
	}{
		{0, []byte{0}},
		{42, []byte{0x2a}},
		{292, []byte{0xa4, 0x02}},
	}

	for _, exp := range expecations {
		conv := EncodeVarint(exp.dec)
		if bytes.Compare(conv, exp.bytes) != 0 {
			t.Fatalf("%d was converted to %#x didn't match %#x\n", exp.dec, conv, exp.bytes)
		}
	}

	for _, exp := range expecations {
		conv, _ := DecodeVarint(exp.bytes)
		if conv != exp.dec {
			t.Fatalf("%q was converted to %d didn't match %d\n", exp.bytes, conv, exp.dec)
		}
	}
}

func TestParsingFile(t *testing.T) {
	expectations := []struct {
		path                string
		format              uint16
		numTracks           uint16
		ticksPerQuarterNote uint16
		timeFormat          timeFormat
	}{
		{"fixtures/elise.mid", 1, 4, 960, MetricalTF},
		{"fixtures/elise1track.mid", 0, 1, 480, MetricalTF},
	}

	for _, exp := range expectations {
		t.Log(exp.path)
		path, _ := filepath.Abs(exp.path)
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		p := New(f)
		if err := p.Parse(); err != nil {
			t.Fatal(err)
		}

		if p.Format != exp.format {
			t.Fatalf("%s of %s didn't match %v, got %v", "format", exp.path, exp.format, p.Format)
		}
		if p.NumTracks != exp.numTracks {
			t.Fatalf("%s of %s didn't match %v, got %v", "numTracks", exp.path, exp.numTracks, p.NumTracks)
		}
		if p.TicksPerQuarterNote != exp.ticksPerQuarterNote {
			t.Fatalf("%s of %s didn't match %v, got %v", "ticksPerQuarterNote", exp.path, exp.ticksPerQuarterNote, p.TicksPerQuarterNote)
		}
		if p.TimeFormat != exp.timeFormat {
			t.Fatalf("%s of %s didn't match %v, got %v", "format", exp.path, exp.timeFormat, p.TimeFormat)
		}

		// check events
		//if i == 0 {
		if len(p.Tracks) == 0 {
			t.Fatal("Tracks not parsed")
		}
		t.Logf("%d tracks\n", len(p.Tracks))
		//if p.Events[0].Text == "foo" {
		//t.Fatalf("%#v\n", p.Events[0])
		//}
		//var found
		for _, tr := range p.Tracks {
			for _, ev := range tr.Events {
				t.Log(ev)
			}
		}
		//}

	}
}
