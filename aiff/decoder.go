package aiff

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/mattetti/audio/misc"
)

var (
	defaultChunkDecoderTimeout = 2 * time.Second
)

// Decoder is the wrapper structure for the AIFF container
type Decoder struct {
	r io.Reader
	// Chan is an Optional channel of chunks that is used to parse chunks
	Chan chan *Chunk
	// ChunkDecoderTimeout is the duration after which the main parser keeps going
	// if the dev hasn't reported the chunk parsing to be done.
	// By default: 2s
	ChunkDecoderTimeout time.Duration
	// The waitgroup is used to let the parser that it's ok to continue
	// after a chunk was passed to the optional parser channel.
	Wg sync.WaitGroup

	// ID is always 'FORM'. This indicates that this is a FORM chunk
	ID [4]byte
	// Size contains the size of data portion of the 'FORM' chunk.
	// Note that the data portion has been
	// broken into two parts, formType and chunks
	Size uint32
	// Format describes what's in the 'FORM' chunk. For Audio IFF files,
	// formType (aka Format) is always 'AIFF'.
	// This indicates that the chunks within the FORM pertain to sampled sound.
	Format [4]byte

	// Data coming from the COMM chunk
	commSize        uint32
	NumChans        uint16
	NumSampleFrames uint32
	SampleSize      uint16
	SampleRate      int

	// AIFC data
	Encoding     [4]byte
	EncodingName string
}

// NewDecoder creates a new reader reading the given reader and pushing audio data to the given channel.
// It is the caller's responsibility to call Close on the Decoder when done.
func NewDecoder(r io.Reader, c chan *Chunk) *Decoder {
	return &Decoder{r: r, Chan: c}
}

// Parse reads the aiff reader and populates the container structure with found information.
// The sound data or unknown chunks are passed to the optional channel if available.
func (p *Decoder) Parse() error {
	if err := binary.Read(p.r, binary.BigEndian, &p.ID); err != nil {
		return err
	}
	// Must start by a FORM header/ID
	if p.ID != formID {
		return fmt.Errorf("%s - %s", ErrFmtNotSupported, p.ID)
	}

	if err := binary.Read(p.r, binary.BigEndian, &p.Size); err != nil {
		return err
	}
	if err := binary.Read(p.r, binary.BigEndian, &p.Format); err != nil {
		return err
	}

	// Must be a AIFF or AIFC form type
	if p.Format != aiffID && p.Format != aifcID {
		return fmt.Errorf("%s - %s", ErrFmtNotSupported, p.Format)
	}

	var err error
	for err != io.EOF {
		id, size, err := p.IDnSize()
		if err != nil {
			break
		}
		switch id {
		case commID:
			p.parseCommChunk(size)
		default:
			p.dispatchToChan(id, size)
		}
	}

	if p.Chan != nil {
		close(p.Chan)
	}

	if err == io.EOF {
		return nil
	}
	return err
}

// Frames processes the reader and returns the basic data and LPCM audio frames.
// Very naive and inneficient approach loading the entire data set in memory.
func (r *Decoder) Frames() (info *Info, frames [][]int, err error) {
	ch := make(chan *Chunk)
	r.Chan = ch
	var sndDataFrames [][]int
	go func() {
		if err := r.Parse(); err != nil {
			panic(err)
		}
	}()

	for chunk := range ch {
		if sndDataFrames == nil {
			sndDataFrames = make([][]int, r.NumSampleFrames, r.NumSampleFrames)
		}
		id := string(chunk.ID[:])
		if id == "SSND" {
			var offset uint32
			var blockSize uint32
			// TODO: BE might depend on the encoding used to generate the aiff data.
			// check encSowt or encTwos
			chunk.ReadBE(&offset)
			chunk.ReadBE(&blockSize)

			// TODO: might want to use io.NewSectionDecoder
			bufData := make([]byte, chunk.Size-8)
			chunk.ReadBE(bufData)
			buf := bytes.NewReader(bufData)

			bytesPerSample := (r.SampleSize-1)/8 + 1
			frameCount := int(r.NumSampleFrames)

			if r.NumSampleFrames == 0 {
				chunk.Done()
				continue
			}

			for i := 0; i < frameCount; i++ {
				sampleBufData := make([]byte, bytesPerSample)
				frame := make([]int, r.NumChans)

				for j := uint16(0); j < r.NumChans; j++ {
					_, err := buf.Read(sampleBufData)
					if err != nil {
						if err == io.EOF {
							break
						}
						log.Println("error reading the buffer")
						log.Fatal(err)
					}

					sampleBuf := bytes.NewBuffer(sampleBufData)
					switch r.SampleSize {
					case 8:
						var v uint8
						binary.Read(sampleBuf, binary.BigEndian, &v)
						frame[j] = int(v)
					case 16:
						var v int16
						binary.Read(sampleBuf, binary.BigEndian, &v)
						frame[j] = int(v)
					case 24:
						// TODO: check if the conversion might not be inversed depending on
						// the encoding (BE vs LE)
						var output int32
						output |= int32(sampleBufData[2]) << 0
						output |= int32(sampleBufData[1]) << 8
						output |= int32(sampleBufData[0]) << 16
						frame[j] = int(output)
					case 32:
						var v int32
						binary.Read(sampleBuf, binary.BigEndian, &v)
						frame[j] = int(v)
					default:
						// TODO: nicer error instead of crashing
						log.Fatalf("%v bitrate not supported", r.SampleSize)
					}
				}
				sndDataFrames[i] = frame

			}
		}

		chunk.Done()
	}

	duration, err := r.Duration()
	if err != nil {
		return nil, sndDataFrames, err
	}

	info = &Info{
		NumChannels:   int(r.NumChans),
		SampleRate:    r.SampleRate,
		BitsPerSample: int(r.SampleSize),
		Duration:      duration,
	}

	return info, sndDataFrames, err
}

func (p *Decoder) parseCommChunk(size uint32) error {
	p.commSize = size

	if err := binary.Read(p.r, binary.BigEndian, &p.NumChans); err != nil {
		return fmt.Errorf("num of channels failed to parse - %s", err.Error())
	}
	if err := binary.Read(p.r, binary.BigEndian, &p.NumSampleFrames); err != nil {
		return fmt.Errorf("num of sample frames failed to parse - %s", err.Error())
	}
	if err := binary.Read(p.r, binary.BigEndian, &p.SampleSize); err != nil {
		return fmt.Errorf("sample size failed to parse - %s", err.Error())
	}
	var srBytes [10]byte
	if err := binary.Read(p.r, binary.BigEndian, &srBytes); err != nil {
		return fmt.Errorf("sample rate failed to parse - %s", err.Error())
	}
	p.SampleRate = misc.IeeeFloatToInt(srBytes)

	if p.Format == aifcID {
		if err := binary.Read(p.r, binary.BigEndian, &p.Encoding); err != nil {
			return fmt.Errorf("AIFC encoding failed to parse - %s", err)
		}
		// pascal style string with the description of the encoding
		var size uint8
		if err := binary.Read(p.r, binary.BigEndian, &size); err != nil {
			return fmt.Errorf("AIFC encoding failed to parse - %s", err)
		}

		desc := make([]byte, size)
		if err := binary.Read(p.r, binary.BigEndian, &desc); err != nil {
			return fmt.Errorf("AIFC encoding failed to parse - %s", err)
		}
		p.EncodingName = string(desc)
	}

	return nil

}

func (p *Decoder) dispatchToChan(id [4]byte, size uint32) error {
	if p.Chan == nil {
		if err := p.jumpTo(int(size)); err != nil {
			return err
		}
		return nil
	}
	okC := make(chan bool)
	p.Wg.Add(1)
	p.Chan <- &Chunk{ID: id, Size: int(size), R: p.r, okChan: okC, Wg: &p.Wg}
	p.Wg.Wait()
	// TODO: timeout
	return nil
}

// Duration returns the time duration for the current AIFF container
func (p *Decoder) Duration() (time.Duration, error) {
	if p == nil {
		return 0, errors.New("can't calculate the duration of a nil pointer")
	}
	duration := time.Duration(float64(p.NumSampleFrames) / float64(p.SampleRate) * float64(time.Second))
	return duration, nil
}

// String implements the Stringer interface.
func (c *Decoder) String() string {
	out := fmt.Sprintf("Format: %s - ", c.Format)
	if c.Format == aifcID {
		out += fmt.Sprintf("%s - ", c.EncodingName)
	}
	if c.SampleRate != 0 {
		out += fmt.Sprintf("%d channels @ %d / %d bits - ", c.NumChans, c.SampleRate, c.SampleSize)
		d, _ := c.Duration()
		out += fmt.Sprintf("Duration: %f seconds\n", d.Seconds())
	}
	return out
}

// IDnSize returns the next ID + block size
func (c *Decoder) IDnSize() ([4]byte, uint32, error) {
	var ID [4]byte
	var blockSize uint32
	if err := binary.Read(c.r, binary.BigEndian, &ID); err != nil {
		return ID, blockSize, err
	}
	if err := binary.Read(c.r, binary.BigEndian, &blockSize); err != err {
		return ID, blockSize, err
	}
	return ID, blockSize, nil
}

// jumpTo advances the reader to the amount of bytes provided
func (c *Decoder) jumpTo(bytesAhead int) error {
	var err error
	for bytesAhead > 0 {
		readSize := bytesAhead
		if readSize > 4000 {
			readSize = 4000
		}

		buf := make([]byte, readSize)
		err = binary.Read(c.r, binary.LittleEndian, &buf)
		if err != nil {
			return nil
		}
		bytesAhead -= readSize
	}
	return nil
}
