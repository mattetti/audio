package transforms

import (
	"errors"

	"github.com/mattetti/audio"
)

// PCMScale converts a buffer with audio content from -1 to 1 into
// the PCM scale based on the buffer's bitdepth.
func PCMScale(buf *audio.PCMBuffer) error {
	if buf == nil || buf.Format == nil {
		return errors.New("nil buffer")
	}
	factor := float64(audio.IntMaxSignedValue(buf.Format.BitDepth))
	data := buf.AsFloat64s()
	for i := 0; i < len(data); i++ {
		data[i] *= factor
	}
	buf.SwitchPrimaryType(audio.Float)

	return nil
}
