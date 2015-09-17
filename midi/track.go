package midi

// Track
// <Track Chunk> = <chunk type><length><MTrk event>
// <MTrk event> = <delta-time><event>
//
type Track struct {
	Size         uint32
	Events       []*Event
	ticksPerBeat uint16
}

// Add schedules the passed event after x beats
func (t *Track) Add(beatDelta float64, e *Event) {
	e.TimeDelta = uint32(beatDelta * float64(t.ticksPerBeat))
	t.Events = append(t.Events, e)
}
