package midi

import "bytes"

// Track
// <Track Chunk> = <chunk type><length><MTrk event>
// <MTrk event> = <delta-time><event>
//
type Track struct {
	Size         uint32
	Events       []*Event
	ticksPerBeat uint16
}

// Add schedules the passed event after x beats (relative to the previous event)
func (t *Track) Add(beatDelta float64, e *Event) {
	if t.ticksPerBeat == 0 {
		t.ticksPerBeat = 96
	}
	e.TimeDelta = uint32(beatDelta * float64(t.ticksPerBeat))
	t.Events = append(t.Events, e)
	t.Size += uint32(len(EncodeVarint(e.TimeDelta))) + e.Size()
}

// ChunkData converts the track and its events into a binary byte slice (chunk header included)
func (t *Track) ChunkData() ([]byte, error) {
	buff := bytes.NewBuffer(nil)
	// time signature
	// TODO: don't have 4/4 36, 8 hardcoded
	buff.Write([]byte{0x00, 0xFF, 0x58, 0x04, 0x04, 0x02, 0x24, 0x08})
	for _, e := range t.Events {
		if _, err := buff.Write(e.Encode()); err != nil {
			return nil, err
		}
	}
	// End of track meta event
	buff.Write([]byte{0x00, 0xFF, 0x2F, 0x00})
	return buff.Bytes(), nil
}
