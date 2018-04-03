package caf

// Metadata represent the amount of metadata one can store/retrieve from a caf file.
// See https://developer.apple.com/library/content/documentation/MusicAudio/Reference/CAFSpec/CAF_spec/CAF_spec.html
type Metadata struct {
	// instrument chunk

	// BaseNote The MIDI note number, and fractional pitch, for
	// the base note of the MIDI instrument. The integer portion of this field
	// indicates the base note, in the integer range 0 to 127, where a value of
	// 60 represents middle C and each integer is a step on a standard piano
	// keyboard (for example, 61 is C# above middle C). The fractional part of
	// the field specifies the fractional pitch; for example, 60.5 is a pitch
	// halfway between notes 60 and 61.
	BaseNote float32
	// MIDILowNote The lowest note for the region, in the integer range 0 to
	// 127, where a value of 60 represents middle C (following the MIDI
	// convention). This value represents the suggested lowest note on a
	// keyboard for playback of this instrument definition. The sound data
	// should be played if the instrument is requested to play a note between
	// MIDILowNote and MIDIHighNote, inclusive. The BaseNote value must be
	// within this range.
	MIDILowNote uint8
	// MIDIHighNote The highest note for the region when used as a MIDI
	// instrument, in the integer range 0 to 127, where a value of 60 represents
	// middle C. See the discussions of the BaseNote and MIDILowNote fields
	// for more information.
	MIDIHighNote uint8
	// MIDILowVelocity The lowest MIDI velocity for playing the region , in the integer range 0 to 127.
	MIDILowVelocity  uint8
	MIDIHighVelocity uint8
	DBGain           float32
	StartRegionID    uint32
	SustainRegionID  uint32
	ReleaseRegionID  uint32
	InstrumentID     uint32
	//
}
