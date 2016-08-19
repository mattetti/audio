// filters implement easy to use audio filters.
package filters

import (
	"github.com/mattetti/audio"
	"github.com/mattetti/audio/dsp/filters"
	"github.com/mattetti/audio/dsp/windows"
)

// LowPass is a basic LowPass filter cutting off
// the audio buffer frequencies above the cutOff frequency.
func LowPass(buf *audio.PCMBuffer, cutOff float64) (err error) {
	s := &filters.Sinc{
		Taps:         62,
		SamplingFreq: buf.Format.SampleRate,
		CutOffFreq:   cutOff,
		Window:       windows.Blackman,
	}
	fir := &filters.FIR{Sinc: s}
	buf.Floats, err = fir.LowPass(buf.AsFloat64s())
	if buf.DataType != audio.Float {
		buf.DataType = audio.Float
		buf.Ints = nil
		buf.Bytes = nil
	}
	return err
}

// HighPass is a basic LowPass filter cutting off
// the audio buffer frequencies below the cutOff frequency.
func HighPass(buf *audio.PCMBuffer, cutOff float64) (err error) {
	s := &filters.Sinc{
		Taps:         62,
		SamplingFreq: buf.Format.SampleRate,
		CutOffFreq:   cutOff,
		Window:       windows.Blackman,
	}
	fir := &filters.FIR{Sinc: s}
	buf.Floats, err = fir.HighPass(buf.AsFloat64s())
	if buf.DataType != audio.Float {
		buf.DataType = audio.Float
		buf.Ints = nil
		buf.Bytes = nil
	}
	return err
}
