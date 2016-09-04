package presenters

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/mattetti/audio"
)

// GnuplotBin exports the buffer content as a binary gnuplot file.
func GnuplotBin(buf *audio.PCMBuffer, path string) error {
	if buf == nil || buf.Format == nil {
		return audio.ErrInvalidBuffer
	}
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	for _, s := range buf.AsFloat32s() {
		if err = binary.Write(out, binary.BigEndian, s); err != nil {
			break
		}
	}
	return err
}

func GnuplotText(buf *audio.PCMBuffer, path string) error {
	if buf == nil || buf.Format == nil {
		return audio.ErrInvalidBuffer
	}
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err = out.WriteString("#"); err != nil {
		return err
	}
	for i := 0; i < buf.Format.NumChannels; i++ {
		if _, err = out.WriteString(fmt.Sprintf("%d\t", i+1)); err != nil {
			return err
		}
	}
	if _, err = out.WriteString("\n"); err != nil {
		return err
	}

	buf.SwitchPrimaryType(audio.Float)
	for i := 0; i < buf.Size(); i++ {
		for j := 0; j < buf.Format.NumChannels; j++ {
			out.WriteString(fmt.Sprintf("%f\t", buf.Floats[i*buf.Format.NumChannels+j]))
		}
		out.WriteString("\n")
	}
	return err
}
