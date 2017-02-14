package decoder

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-audio/aiff"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

var (
	ErrInvalidPath = errors.New("invalid path")
)

type Format string

var (
	Unknown Format = "unknown"
	Wav     Format = "wav"
	Aif     Format = "aiff"
)

type Decoder interface {
	FullPCMBuffer() (*audio.IntBuffer, error)
	PCMLen() int64
	FwdToPCM() error
	PCMBuffer(*audio.IntBuffer) (int, error)
	WasPCMAccessed() bool
	Format() *audio.Format
	Err() error
}

// FileFormat returns the known format of the passed path.
func FileFormat(path string) (Format, error) {
	if !fileExists(path) {
		return "", ErrInvalidPath
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	var triedWav bool
	var triedAif bool

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".wav", ".wave":
		triedWav = true
		d := wav.NewDecoder(f)
		if d.IsValidFile() {
			return Wav, nil
		}
	case ".aif", ".aiff":
		triedAif = true
		d := aiff.NewDecoder(f)
		if d.IsValidFile() {
			return Aif, nil
		}
	}
	// extension doesn't match, let's try again
	f.Seek(0, 0)
	if !triedWav {
		wd := wav.NewDecoder(f)
		if wd.IsValidFile() {
			return Wav, nil
		}
		f.Seek(0, 0)
	}
	if !triedAif {
		ad := aiff.NewDecoder(f)
		if ad.IsValidFile() {
			return Aif, nil
		}
	}
	return Unknown, nil
}

// helper checking if a file exists
func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
