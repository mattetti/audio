package transforms

import (
	"github.com/mattetti/audio"
	"github.com/mattetti/audio/dsp/analysis"
)

// BitCrush reduces the resolution of the sample to the target bit depth
// Note that bit crusher effects are usually made of this feature + a decimator
func BitCrush(buf *audio.PCMBuffer, bitDepth int) {
	buf.SwitchPrimaryType(audio.Float)
	min, max := analysis.MinMaxFloat(buf)
	if min >= -1 && max <= 1 {
		PCMScale(buf)
	}
	buf.SwitchPrimaryType(audio.Integer)
	for i := 0; i < len(buf.Ints); i++ {
		buf.Ints[i] = dropBit(buf.Ints[i], bitDepth)
	}
}

func dropBit(input, bitsToKeep int) int {
	if bitsToKeep > 16 {
		return input
	}
	return input & (-1 << (16 - uint8(bitsToKeep)))
}
