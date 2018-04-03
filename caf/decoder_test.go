package caf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mattetti/filebuffer"
)

func TestBadFileHeaderData(t *testing.T) {
	r := filebuffer.New([]byte{'m', 'a', 't', 't', 0, 0, 0})
	d := NewDecoder(r)
	if err := d.ReadInfo(); err == nil {
		t.Fatalf("Expected bad data to return %s", ErrFmtNotSupported)
	}

	r = filebuffer.New([]byte{'c', 'a', 'f', 'f', 2, 0, 0})
	d = NewDecoder(r)
	if err := d.ReadInfo(); err == nil {
		t.Fatalf("Expected bad data to return %s", ErrFmtNotSupported)
	}
}

func TestParsingFile(t *testing.T) {
	expectations := []struct {
		path    string
		format  [4]byte
		version uint16
		flags   uint16
	}{
		{"fixtures/ring.caf", fileHeaderID, 1, 0},
		{"fixtures/bass.caf", fileHeaderID, 1, 0},
	}

	for _, exp := range expectations {
		t.Run(exp.path, func(t *testing.T) {
			path, _ := filepath.Abs(exp.path)
			f, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			d := NewDecoder(f)
			if err := d.ReadInfo(); err != nil {
				t.Fatal(err)
			}

			if d.Format != exp.format {
				t.Fatalf("%s of %s didn't match %v, got %v", "format", exp.path, exp.format, d.Format)
			}
			if d.Version != exp.version {
				t.Fatalf("%s of %s didn't match %d, got %v", "version", exp.path, exp.version, d.Version)
			}
			if d.Flags != exp.flags {
				t.Fatalf("%s of %s didn't match %d, got %v", "flags", exp.path, exp.flags, d.Flags)
			}

		})
	}
}
