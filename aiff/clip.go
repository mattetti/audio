package aiff

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/mattetti/audio"
	"github.com/mattetti/audio/misc"
)

type Clip struct {
	r            io.ReadSeeker
	size         int64
	channels     int
	bitDepth     int
	sampleRate   int64
	sampleFrames int

	// decoder info
	offset    uint32
	blockSize uint32
}

// ReadPCM reads up to n frames from the clip.
// The frwames as well as the number of frames/items read are returned.
// TODO(mattetti): misc.AudioFrames is a temporary solution that needs to be improved.
// TODO(mattetti): we might want to keep track of the postion in the reader so we can easily check if
// the reader has been reset.
func (c *Clip) ReadPCM(nFrames int) (frames misc.AudioFrames, n int, err error) {
	if c == nil || c.sampleFrames == 0 {
		return nil, 0, nil
	}
	if err := c.readOffsetBlockSise(); err != nil {
		return nil, 0, err
	}

	bytesPerSample := (c.bitDepth-1)/8 + 1
	sampleBufData := make([]byte, bytesPerSample)
	frames = make(misc.AudioFrames, nFrames)
	for i := 0; i < c.channels; i++ {
		frames[i] = make([]int, c.channels)
	}

outter:
	for frameIDX := 0; frameIDX < nFrames; frameIDX++ {
		if frameIDX > len(frames) {
			break
		}

		frame := make([]int, c.channels)
		for j := 0; j < c.channels; j++ {
			_, err := c.r.Read(sampleBufData)
			if err != nil {
				if err == io.EOF {
					err = nil
				}
				break outter
			}

			sampleBuf := bytes.NewBuffer(sampleBufData)
			switch c.bitDepth {
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
				err = fmt.Errorf("%v bit depth not supported", c.bitDepth)
				break outter
			}
		}
		frames[frameIDX] = frame
		n++
	}

	return frames, n, err
}

func (c *Clip) Read(p []byte) (n int, err error) {
	// TODO(mattetti): should this return raw bytes or PCM data
	// TODO(mattetti): the underlying reader might pass the size limit, we probably
	// need to use some sort of limitreader.
	return
}

// Seek seeks into the clip
func (c *Clip) Seek(offset int64, whence int) (int64, error) {
	if c == nil {
		return 0, nil
	}

	return c.r.Seek(offset, whence)
}

func (c *Clip) FrameInfo() audio.FrameInfo {
	return audio.FrameInfo{
		Channels:   c.channels,
		BitDepth:   c.bitDepth,
		SampleRate: c.sampleRate,
	}
}

func (c *Clip) Size() int64 {
	return c.size
}

func (c *Clip) readOffsetBlockSise() error {
	if c == nil || c.blockSize > 0 {
		return nil
	}

	// TODO: endianness might depend on the encoding used to generate the aiff data.
	// check encSowt or encTwos

	if err := binary.Read(c.r, binary.BigEndian, &c.offset); err != nil {
		return err
	}
	if err := binary.Read(c.r, binary.BigEndian, &c.blockSize); err != nil {
		return err
	}

	return nil
}
