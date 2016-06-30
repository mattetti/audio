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

	"github.com/mattetti/audio"
	"github.com/mattetti/audio/misc"
)

var (
	defaultChunkDecoderTimeout = 2 * time.Second
)

// Decoder is the wrapper structure for the AIFF container
type Decoder struct {
	r io.ReadSeeker
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

	err      error
	dataClip *audio.Clip
}

// NewDecoder creates a new reader reading the given reader and pushing audio data to the given channel.
// It is the caller's responsibility to call Close on the Decoder when done.
func NewDecoder(r io.ReadSeeker, c chan *Chunk) *Decoder {
	return &Decoder{r: r, Chan: c}
}

// Decode reads from a Read Seeker and converts the input to a PCM
// clip output.
func Decode(r io.ReadSeeker) (audio.Clipper, error) {
	d := &Decoder{r: r}
	if d.err = d.readHeaders(); d.err != nil {
		return nil, d.err
	}

	// read the file information to setup the audio clip
	// find the beginning of the SSND chunk and set the clip reader to it.
	clip := &audio.Clip{}

	var err error
	var rewindBytes int64
	for err != io.EOF {
		id, size, err := d.IDnSize()
		if err != nil {
			d.err = err
			break
		}
		switch id {
		case COMMID:
			d.parseCommChunk(size)
			clip.Channels = int(d.NumChans)
			clip.BitDepth = int(d.SampleSize)
			clip.SampleRate = int64(d.SampleRate)
			// if we found the sound data before the COMM,
			// we need to rewind the reader so we can properly
			// set the clip reader.
			if rewindBytes > 0 {
				r.Seek(-rewindBytes, 1)
				break
			}
		case SSNDID:
			clip.DataSize = int64(size)
			// if we didn't read the COMM, we are going to need to come back
			if clip.SampleRate == 0 {
				rewindBytes += int64(size)
				d.dispatchToChan(id, size)
			} else {
				break
			}
		default:
			// if we read SSN but didn't read the COMM, we need to track location
			if clip.DataSize != 0 {
				rewindBytes += int64(size)
			}
			d.dispatchToChan(id, size)
		}
	}
	clip.R = r
	return clip, nil
}

// Err returns the first non-EOF error that was encountered by the Decoder.
func (d *Decoder) Err() error {
	if d.err == io.EOF {
		return nil
	}
	return d.err
}

// Clip returns the audio Clip information including a reader to reads its content.
// This method is safe to be called multiple times but the reader might need to be rewinded
// if previously read.
// This is the recommended, default way to consume an AIFF file.
func (d *Decoder) Clip() *audio.Clip {
	if d.dataClip != nil {
		return d.dataClip
	}
	if d.err = d.readHeaders(); d.err != nil {
		d.err = fmt.Errorf("failed to read header - %v", d.err)
		return nil
	}

	d.dataClip = &audio.Clip{}

	// read the file information to setup the audio clip
	// find the beginning of the SSND chunk and set the clip reader to it.
	var (
		id          [4]byte
		size        uint32
		rewindBytes int64
	)
	for d.err != io.EOF {
		id, size, d.err = d.IDnSize()
		if d.err != nil {
			d.err = fmt.Errorf("error reading chunk header - %v", d.err)
			break
		}
		switch id {
		case COMMID:
			d.parseCommChunk(size)
			d.dataClip.Channels = int(d.NumChans)
			d.dataClip.BitDepth = int(d.SampleSize)
			d.dataClip.SampleRate = int64(d.SampleRate)
			// if we found the sound data before the COMM,
			// we need to rewind the reader so we can properly
			// set the clip reader.
			if rewindBytes > 0 {
				d.r.Seek(-rewindBytes, 1)
				break
			}
		case SSNDID:
			d.dataClip.DataSize = int64(size)
			// if we didn't read the COMM, we are going to need to come back
			if d.dataClip.SampleRate == 0 {
				rewindBytes += int64(size)
				if d.err = d.jumpTo(int(size)); d.err != nil {
					return nil
				}
			}
			d.dataClip.R = d.r
			return d.dataClip

		default:
			// if we read SSN but didn't read the COMM, we need to track location
			if d.dataClip.DataSize == 0 {
				rewindBytes += int64(size)
			}
			if d.err = d.jumpTo(int(size)); d.err != nil {
				return nil
			}
		}
	}

	return d.dataClip
}

func (d *Decoder) readHeaders() error {
	// prevent the headers to be re-read
	if d.Size > 0 {
		return nil
	}
	if d.err = binary.Read(d.r, binary.BigEndian, &d.ID); d.err != nil {
		return d.err
	}
	// Must start by a FORM header/ID
	if d.ID != formID {
		d.err = fmt.Errorf("%s - %s", ErrFmtNotSupported, d.ID)
		return d.err
	}

	if d.err = binary.Read(d.r, binary.BigEndian, &d.Size); d.err != nil {
		return d.err
	}
	if d.err = binary.Read(d.r, binary.BigEndian, &d.Format); d.err != nil {
		return d.err
	}

	// Must be a AIFF or AIFC form type
	if d.Format != aiffID && d.Format != aifcID {
		d.err = fmt.Errorf("%s - %s", ErrFmtNotSupported, d.Format)
		return d.err
	}

	return nil
}

// Parse reads the aiff reader and populates the container structure with found information.
// The sound data or unknown chunks are passed to the optional channel if available.
// Instead of checking the returned error, it's recommended to read d.Err()
func (d *Decoder) Parse() error {
	if d.err = d.readHeaders(); d.err != nil {
		return d.err
	}

	var id [4]byte
	var size uint32
	for d.err != io.EOF {
		id, size, d.err = d.IDnSize()
		if d.err != nil {
			break
		}
		switch id {
		case COMMID:
			d.parseCommChunk(size)
		default:
			d.dispatchToChan(id, size)
		}
	}

	if d.Chan != nil {
		close(d.Chan)
	}

	return d.Err()
}

// Frames processes the reader and returns the basic data and LPCM audio frames.
// Very naive and inneficient approach loading the entire data set in memory.
func (d *Decoder) Frames() (info *Info, frames [][]int, err error) {
	ch := make(chan *Chunk)
	d.Chan = ch
	var sndDataFrames [][]int
	go func() {
		if err := d.Parse(); err != nil {
			panic(err)
		}
	}()

	for chunk := range ch {
		if sndDataFrames == nil {
			sndDataFrames = make([][]int, d.NumSampleFrames, d.NumSampleFrames)
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

			bytesPerSample := (d.SampleSize-1)/8 + 1
			frameCount := int(d.NumSampleFrames)

			if d.NumSampleFrames == 0 {
				chunk.Done()
				continue
			}

			for i := 0; i < frameCount; i++ {
				sampleBufData := make([]byte, bytesPerSample)
				frame := make([]int, d.NumChans)

				for j := uint16(0); j < d.NumChans; j++ {
					_, err := buf.Read(sampleBufData)
					if err != nil {
						if err == io.EOF {
							break
						}
						log.Println("error reading the buffer")
						log.Fatal(err)
					}

					sampleBuf := bytes.NewBuffer(sampleBufData)
					switch d.SampleSize {
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
						log.Fatalf("%v bitrate not supported", d.SampleSize)
					}
				}
				sndDataFrames[i] = frame

			}
		}

		chunk.Done()
	}

	duration, err := d.Duration()
	if err != nil {
		return nil, sndDataFrames, err
	}

	info = &Info{
		NumChannels: int(d.NumChans),
		SampleRate:  d.SampleRate,
		BitDepth:    int(d.SampleSize),
		Duration:    duration,
	}

	return info, sndDataFrames, err
}

func (d *Decoder) parseCommChunk(size uint32) error {
	d.commSize = size
	// don't re-parse the comm chunk
	if d.NumChans > 0 {
		return nil
	}

	if d.err = binary.Read(d.r, binary.BigEndian, &d.NumChans); d.err != nil {
		d.err = fmt.Errorf("num of channels failed to parse - %s", d.err)
		return d.err
	}
	if d.err = binary.Read(d.r, binary.BigEndian, &d.NumSampleFrames); d.err != nil {
		d.err = fmt.Errorf("num of sample frames failed to parse - %s", d.err)
		return d.err
	}
	if d.err = binary.Read(d.r, binary.BigEndian, &d.SampleSize); d.err != nil {
		d.err = fmt.Errorf("sample size failed to parse - %s", d.err)
		return d.err
	}
	var srBytes [10]byte
	if d.err = binary.Read(d.r, binary.BigEndian, &srBytes); d.err != nil {
		d.err = fmt.Errorf("sample rate failed to parse - %s", d.err)
		return d.err
	}
	d.SampleRate = misc.IeeeFloatToInt(srBytes)

	if d.Format == aifcID {
		if d.err = binary.Read(d.r, binary.BigEndian, &d.Encoding); d.err != nil {
			d.err = fmt.Errorf("AIFC encoding failed to parse - %s", d.err)
			return d.err
		}
		// pascal style string with the description of the encoding
		var size uint8
		if d.err = binary.Read(d.r, binary.BigEndian, &size); d.err != nil {
			d.err = fmt.Errorf("AIFC encoding failed to parse - %s", d.err)
			return d.err
		}

		desc := make([]byte, size)
		if d.err = binary.Read(d.r, binary.BigEndian, &desc); d.err != nil {
			d.err = fmt.Errorf("AIFC encoding failed to parse - %s", d.err)
			return d.err
		}
		d.EncodingName = string(desc)
	}

	return nil

}

func (d *Decoder) dispatchToChan(id [4]byte, size uint32) error {
	if d.Chan == nil {
		if d.err = d.jumpTo(int(size)); d.err != nil {
			return d.err
		}
		return nil
	}
	okC := make(chan bool)
	bSize := int(size)
	d.Wg.Add(1)
	d.Chan <- &Chunk{
		ID:     id,
		Size:   bSize,
		R:      io.LimitReader(d.r, int64(bSize)),
		okChan: okC,
		Wg:     &d.Wg,
	}
	d.Wg.Wait()
	// TODO: timeout
	return nil
}

// Duration returns the time duration for the current AIFF container
func (d *Decoder) Duration() (time.Duration, error) {
	if d == nil {
		return 0, errors.New("can't calculate the duration of a nil pointer")
	}
	duration := time.Duration(float64(d.NumSampleFrames) / float64(d.SampleRate) * float64(time.Second))
	return duration, nil
}

// String implements the Stringer interface.
func (d *Decoder) String() string {
	out := fmt.Sprintf("Format: %s - ", d.Format)
	if d.Format == aifcID {
		out += fmt.Sprintf("%s - ", d.EncodingName)
	}
	if d.SampleRate != 0 {
		out += fmt.Sprintf("%d channels @ %d / %d bits - ", d.NumChans, d.SampleRate, d.SampleSize)
		dur, _ := d.Duration()
		out += fmt.Sprintf("Duration: %f seconds\n", dur.Seconds())
	}
	return out
}

// IDnSize returns the next ID + block size
func (d *Decoder) IDnSize() ([4]byte, uint32, error) {
	var ID [4]byte
	var blockSize uint32
	if d.err = binary.Read(d.r, binary.BigEndian, &ID); d.err != nil {
		return ID, blockSize, d.err
	}
	if d.err = binary.Read(d.r, binary.BigEndian, &blockSize); d.err != nil {
		return ID, blockSize, d.err
	}
	return ID, blockSize, nil
}

// jumpTo advances the reader to the amount of bytes provided
func (d *Decoder) jumpTo(bytesAhead int) error {
	for bytesAhead > 0 {
		readSize := bytesAhead
		if readSize > 4000 {
			readSize = 4000
		}

		buf := make([]byte, readSize)
		d.err = binary.Read(d.r, binary.LittleEndian, &buf)
		if d.err != nil {
			return d.err
		}
		bytesAhead -= readSize
	}
	return nil
}
