package midi

// Track
// <Track Chunk> = <chunk type><length><MTrk event>
// <MTrk event> = <delta-time><event>
//
type Track struct {
	Size   uint32
	Events []*Event
}
