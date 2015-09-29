package aiff

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"log"
)

var (
	formID = [4]byte{'F', 'O', 'R', 'M'}
	aiffID = [4]byte{'A', 'I', 'F', 'F'}
	aifcID = [4]byte{'A', 'I', 'F', 'C'}
	commID = [4]byte{'C', 'O', 'M', 'M'}

	// AIFC encodings
	encNone = [4]byte{'N', 'O', 'N', 'E'}
	// inverted byte order LE instead of BE (not really compression)
	encSowt = [4]byte{'s', 'o', 'w', 't'}
	// inverted byte order LE instead of BE (not really compression)
	encTwos = [4]byte{'t', 'w', 'o', 's'}
	encRaw  = [4]byte{'r', 'a', 'w', ' '}
	encIn24 = [4]byte{'i', 'n', '2', '4'}
	enc42n1 = [4]byte{'4', '2', 'n', '1'}
	encIn32 = [4]byte{'i', 'n', '3', '2'}
	enc23ni = [4]byte{'2', '3', 'n', 'i'}

	encFl32 = [4]byte{'f', 'l', '3', '2'}
	encFL32 = [4]byte{'F', 'L', '3', '2'}
	encFl64 = [4]byte{'f', 'l', '6', '4'}
	encFL64 = [4]byte{'F', 'L', '6', '4'}

	envUlaw = [4]byte{'u', 'l', 'a', 'w'}
	encULAW = [4]byte{'U', 'L', 'A', 'W'}
	encAlaw = [4]byte{'a', 'l', 'a', 'w'}
	encALAW = [4]byte{'A', 'L', 'A', 'W'}

	encDwvw = [4]byte{'D', 'W', 'V', 'W'}
	encGsm  = [4]byte{'G', 'S', 'M', ' '}
	encIma4 = [4]byte{'i', 'm', 'a', '4'}

	// ErrFmtNotSupported is a generic error reporting an unknown format.
	ErrFmtNotSupported = errors.New("format not supported")
	// ErrUnexpectedData is a generic error reporting that the parser encountered unexpected data.
	ErrUnexpectedData = errors.New("unexpected data content")
)

// New is the entry point to this package.
func New(r io.Reader) *Parser {
	return &Parser{r: r}
}

// NewParser lets a dev pass a channel to receive audio data and raw chunks.
func NewParser(r io.Reader, c chan *Chunk) *Parser {
	return &Parser{r: r, Chan: c}
}

// Read processes the reader and returns the basic data and LPCM audio frames.
// TODO: change the API to take a channel and write the frames to the channel.
func ReadFrames(r io.Reader) (sampleRate, sampleSize, numChans int, frames [][]int) {
	ch := make(chan *Chunk)
	c := NewParser(r, ch)
	var sndDataFrames [][]int
	go func() {
		if err := c.Parse(); err != nil {
			panic(err)
		}
	}()

	for chunk := range ch {
		if sndDataFrames == nil {
			sndDataFrames = make([][]int, c.NumSampleFrames, c.NumSampleFrames)
		}
		id := string(chunk.ID[:])
		if id == "SSND" {
			var offset uint32
			var blockSize uint32
			// TODO: BE might depend on the encoding used to generate the aiff data.
			// check encSowt or encTwos
			chunk.ReadBE(&offset)
			chunk.ReadBE(&blockSize)

			// TODO: might want to use io.NewSectionReader
			bufData := make([]byte, chunk.Size-8)
			chunk.ReadBE(bufData)
			buf := bytes.NewReader(bufData)

			bytesPerSample := (c.SampleSize-1)/8 + 1
			frameCount := int(c.NumSampleFrames)

			if c.NumSampleFrames == 0 {
				chunk.Done()
				continue
			}

			for i := 0; i < frameCount; i++ {
				sampleBufData := make([]byte, bytesPerSample)
				frame := make([]int, c.NumChans)

				for j := uint16(0); j < c.NumChans; j++ {
					_, err := buf.Read(sampleBufData)
					if err != nil {
						if err == io.EOF {
							break
						}
						log.Println("error reading the buffer")
						log.Fatal(err)
					}

					sampleBuf := bytes.NewBuffer(sampleBufData)
					switch c.SampleSize {
					case 8:
						var v uint8
						binary.Read(sampleBuf, binary.BigEndian, &v)
						frame[j] = int(v)
					case 16:
						var v uint16
						binary.Read(sampleBuf, binary.BigEndian, &v)
						frame[j] = int(v)
					case 24:
						// TODO: check if the conversion might not be inversed depending on
						// the encoding (BE vs LE)
						var output uint32
						output |= uint32(sampleBufData[2]) << 0
						output |= uint32(sampleBufData[1]) << 8
						output |= uint32(sampleBufData[0]) << 16
						frame[j] = int(output)
					case 32:
						var v uint32
						binary.Read(sampleBuf, binary.BigEndian, &v)
						frame[j] = int(v)
					default:
						log.Fatalf("%v bitrate not supported", c.SampleSize)
					}
				}
				sndDataFrames[i] = frame

			}
		}

		chunk.Done()
	}

	return int(c.SampleRate), int(c.SampleSize), int(c.NumChans), sndDataFrames
}
