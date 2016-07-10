package aiff

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/mattetti/audio/misc"
)

// Decoder is the wrapper structure for the AIFF container
type Decoder struct {
	r io.ReadSeeker

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
	numSampleFrames uint32
	BitDepth        uint16
	SampleRate      int

	// AIFC data
	Encoding     [4]byte
	EncodingName string

	err      error
	clipInfo *Clip
}

// NewDecoder creates a new reader reading the given reader and pushing audio data to the given channel.
// It is the caller's responsibility to call Close on the Decoder when done.
func NewDecoder(r io.ReadSeeker) *Decoder {
	return &Decoder{r: r}
}

// Err returns the first non-EOF error that was encountered by the Decoder.
func (d *Decoder) Err() error {
	if d.err == io.EOF {
		return nil
	}
	return d.err
}

// EOF returns positively if the underlying reader reached the end of file.
func (d *Decoder) EOF() bool {
	if d == nil || d.err == io.EOF {
		return true
	}
	return false
}

// Clip returns the audio Clip information including a reader to reads its content.
// This method is safe to be called multiple times but the reader might need to be rewinded
// if previously read.
// This is the recommended, default way to consume an AIFF file.
func (d *Decoder) Clip() *Clip {
	if d.clipInfo != nil {
		return d.clipInfo
	}
	if d.err = d.readHeaders(); d.err != nil {
		d.err = fmt.Errorf("failed to read header - %v", d.err)
		return nil
	}

	d.clipInfo = &Clip{}

	// read the file information to setup the audio clip
	// find the beginning of the SSND chunk and set the clip reader to it.
	var (
		id          [4]byte
		size        uint32
		rewindBytes int64
	)
	for d.err != io.EOF {
		id, size, d.err = d.iDnSize()
		if d.err != nil {
			d.err = fmt.Errorf("error reading chunk header - %v", d.err)
			break
		}
		switch id {
		case COMMID:
			d.parseCommChunk(size)
			d.clipInfo.channels = int(d.NumChans)
			d.clipInfo.bitDepth = int(d.BitDepth)
			d.clipInfo.sampleRate = int64(d.SampleRate)
			d.clipInfo.sampleFrames = int(d.numSampleFrames)
			d.clipInfo.blockSize = size
			// if we found the sound data before the COMM,
			// we need to rewind the reader so we can properly
			// set the clip reader.
			if rewindBytes > 0 {
				d.r.Seek(-rewindBytes, 1)
				break
			}
		case SSNDID:
			d.clipInfo.blockSize = size
			// if we didn't read the COMM, we are going to need to come back
			if d.clipInfo.sampleRate == 0 {
				rewindBytes += int64(size)
				if d.err = d.jumpTo(int(size)); d.err != nil {
					return nil
				}
			}
			d.clipInfo.r = d.r
			return d.clipInfo

		default:
			// if we read SSN but didn't read the COMM, we need to track location
			if d.clipInfo.sampleRate == 0 {
				rewindBytes += int64(size)
			}
			if d.err = d.jumpTo(int(size)); d.err != nil {
				return nil
			}
		}
	}

	return d.clipInfo
}

// NextChunk returns the next available chunk
func (d *Decoder) NextChunk() (*Chunk, error) {
	if d.err = d.readHeaders(); d.err != nil {
		d.err = fmt.Errorf("failed to read header - %v", d.err)
		return nil, d.err
	}

	var (
		id   [4]byte
		size uint32
	)

	id, size, d.err = d.iDnSize()
	if d.err != nil {
		d.err = fmt.Errorf("error reading chunk header - %v", d.err)
		return nil, d.err
	}

	c := &Chunk{
		ID:   id,
		Size: int(size),
		R:    io.LimitReader(d.r, int64(size)),
	}
	return c, d.err
}

// Frames returns the audio frames contained in reader.
// Notes that this method allocates a lot of memory (depending on the duration of the underlying file).
// Consider using the decoder clip and reading/decoding using a buffer.
func (d *Decoder) Frames() (frames misc.AudioFrames, err error) {
	clip := d.Clip()
	totalFrames := int(clip.Size())
	readFrames := 0

	bufSize := 4096
	buf := make([]byte, bufSize)
	var tFrames misc.AudioFrames
	var n int
	for readFrames < totalFrames {
		n, err = clip.Read(buf)
		if err != nil || n == 0 {
			break
		}
		readFrames += n
		tFrames, err = d.DecodeFrames(buf)
		if err != nil {
			break
		}
		frames = append(frames, tFrames[:n]...)
	}
	return frames, err
}

// DecodeFrames decodes PCM bytes into audio frames based on the decoder context
func (d *Decoder) DecodeFrames(data []byte) (frames misc.AudioFrames, err error) {
	numChannels := int(d.NumChans)
	r := bytes.NewBuffer(data)

	bytesPerSample := int((d.BitDepth-1)/8 + 1)
	sampleBufData := make([]byte, bytesPerSample)

	frames = make(misc.AudioFrames, len(data)/bytesPerSample)
	for j := 0; j < int(numChannels); j++ {
		frames[j] = make([]int, numChannels)
	}
	n := 0

outter:
	for i := 0; (i + (bytesPerSample * numChannels)) <= len(data); {
		frame := make([]int, numChannels)
		for j := 0; j < numChannels; j++ {
			switch d.BitDepth {
			case 8:
				var v uint8
				err = binary.Read(r, binary.BigEndian, &v)
				if err != nil {
					if err == io.EOF {
						err = nil
					}
					break outter
				}
				frame[j] = int(v)
			case 16:
				var v int16
				binary.Read(r, binary.BigEndian, &v)
				frame[j] = int(v)
			case 24:
				_, err = r.Read(sampleBufData)
				if err != nil {
					if err == io.EOF {
						err = nil
					}
					break outter
				}
				// TODO: check if the conversion might not be inversed depending on
				// the encoding (BE vs LE)
				var output int32
				output |= int32(sampleBufData[2]) << 0
				output |= int32(sampleBufData[1]) << 8
				output |= int32(sampleBufData[0]) << 16
				frame[j] = int(output)
			case 32:
				var v int32
				binary.Read(r, binary.BigEndian, &v)
				frame[j] = int(v)
			default:
				err = fmt.Errorf("%v bit depth not supported", d.BitDepth)
				break outter
			}
			i += bytesPerSample
		}
		frames[n] = frame
		n++
	}

	return frames, err
}

// Duration returns the time duration for the current AIFF container
func (d *Decoder) Duration() (time.Duration, error) {
	if d == nil {
		return 0, errors.New("can't calculate the duration of a nil pointer")
	}
	d.readInfo()
	if err := d.Err(); err != nil {
		return 0, err
	}
	duration := time.Duration(float64(d.numSampleFrames) / float64(d.SampleRate) * float64(time.Second))
	return duration, nil
}

// String implements the Stringer interface.
func (d *Decoder) String() string {
	out := fmt.Sprintf("Format: %s - ", d.Format)
	if d.Format == aifcID {
		out += fmt.Sprintf("%s - ", d.EncodingName)
	}
	if d.SampleRate != 0 {
		out += fmt.Sprintf("%d channels @ %d / %d bits - ", d.NumChans, d.SampleRate, d.BitDepth)
		dur, _ := d.Duration()
		out += fmt.Sprintf("Duration: %f seconds\n", dur.Seconds())
	}
	return out
}

// iDnSize returns the next ID + block size
func (d *Decoder) iDnSize() ([4]byte, uint32, error) {
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

// readHeaders is safe to call multiple times
// byte size of the header: 12
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

// readInfo reads the underlying reader until the comm header is parsed.
// This method is safe to call multiple times.
func (d *Decoder) readInfo() {
	if d == nil || d.SampleRate > 0 {
		return
	}
	if d.err = d.readHeaders(); d.err != nil {
		d.err = fmt.Errorf("failed to read header - %v", d.err)
		return
	}

	var (
		id          [4]byte
		size        uint32
		rewindBytes int64
	)
	for d.err != io.EOF {
		id, size, d.err = d.iDnSize()
		if d.err != nil {
			d.err = fmt.Errorf("error reading chunk header - %v", d.err)
			break
		}
		switch id {
		case COMMID:
			d.parseCommChunk(size)
			// if we found other chunks before the COMM,
			// we need to rewind the reader so we can properly
			// read the rest later.
			if rewindBytes > 0 {
				d.r.Seek(-(rewindBytes + int64(size)), 1)
				break
			}
			return
		default:
			// we haven't read the COMM chunk yet, we need to track location to rewind
			if d.SampleRate == 0 {
				rewindBytes += int64(size)
			}
			if d.err = d.jumpTo(int(size)); d.err != nil {
				return
			}
		}
	}
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
	if d.err = binary.Read(d.r, binary.BigEndian, &d.numSampleFrames); d.err != nil {
		d.err = fmt.Errorf("num of sample frames failed to parse - %s", d.err)
		return d.err
	}
	if d.err = binary.Read(d.r, binary.BigEndian, &d.BitDepth); d.err != nil {
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

// jumpTo advances the reader to the amount of bytes provided
func (d *Decoder) jumpTo(bytesAhead int) error {
	var err error
	if bytesAhead > 0 {
		_, err = io.CopyN(ioutil.Discard, d.r, int64(bytesAhead))
	}
	return err
}
