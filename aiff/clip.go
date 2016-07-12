package aiff

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/mattetti/audio"
)

// Clip represents the PCM data contained in the aiff stream.
type Clip struct {
	r            io.ReadSeeker
	channels     int
	bitDepth     int
	sampleRate   int64
	sampleFrames int
	readFrames   int

	// decoder info
	offset     uint32
	blockSize  uint32
	offsetRead bool
}

// Next reads up to n frames from the clip.
// The frames as well as the number of full frames read are returned.
// This API is somewhat similar to https://golang.org/pkg/bytes/#Buffer.Next
func (c *Clip) Next(nFrames int) (frames audio.Frames, n int, err error) {
	if c == nil || c.sampleFrames == 0 {
		return nil, 0, nil
	}
	if err := c.readOffsetBlockSize(); err != nil {
		return nil, 0, err
	}
	// TODO(mattetti): respect offset and block size

	bytesPerSample := (c.bitDepth-1)/8 + 1
	sampleBufData := make([]byte, bytesPerSample)
	frames = make(audio.Frames, nFrames)
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

// Read reads frames into the passed buffer and returns the number of full frames
// read.
func (c *Clip) Read(buf []byte) (n int, err error) {
	if c == nil || c.sampleFrames == 0 {
		return n, nil
	}
	if err := c.readOffsetBlockSize(); err != nil {
		return n, err
	}
	// TODO(mattetti): respect offset and block size

	bytesPerSample := (c.bitDepth-1)/8 + 1
	sampleBufData := make([]byte, bytesPerSample)

	frameSize := (bytesPerSample * c.channels)
	// track how many frames we previously read so we don't
	// read past the chunk
	startingAtFrame := c.readFrames
	if startingAtFrame >= c.sampleFrames {
		return 0, nil
	}
outter:
	for i := 0; i+frameSize < len(buf); {
		for j := 0; j < c.channels; j++ {
			_, err := c.r.Read(sampleBufData)
			if err != nil {
				if err == io.EOF {
					err = nil
				}
				break outter
			}
			for _, b := range sampleBufData {
				buf[i] = b
				i++
			}
		}
		c.readFrames++
		if c.readFrames >= c.sampleFrames {
			break
		}
	}

	n = c.readFrames - startingAtFrame
	return n, err
}

// Size returns the total number of frames available in this clip.
func (c *Clip) Size() int64 {
	if c == nil {
		return 0
	}
	return int64(c.sampleFrames)
}

// Seek seeks into the clip
// TODO(mattetti): Seek offset should be in frames, not bytes
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

func (c *Clip) readOffsetBlockSize() error {
	// reading the offset and blocksize should only happen once per chunk
	if c == nil || c.offsetRead == true {
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

	c.offsetRead = true
	return nil
}
