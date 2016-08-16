package aiff

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/mattetti/audio"
)

// PCM returns an audio.PCM compatible value to consume the PCM data
// contained in the underlying aiff data.
// DEPRECATED
func (d *Decoder) PCM() *PCM {
	if d.pcmClip != nil {
		return d.pcmClip
	}
	if d.err = d.readHeaders(); d.err != nil {
		d.err = fmt.Errorf("failed to read header - %v", d.err)
		return nil
	}

	d.pcmClip = &PCM{}

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
			d.pcmClip.channels = int(d.NumChans)
			d.pcmClip.bitDepth = int(d.BitDepth)
			d.pcmClip.sampleRate = int64(d.SampleRate)
			d.pcmClip.sampleFrames = int64(d.numSampleFrames)
			d.pcmClip.blockSize = size
			// if we found the sound data before the COMM,
			// we need to rewind the reader so we can properly
			// set the clip reader.
			if rewindBytes > 0 {
				d.r.Seek(-rewindBytes, 1)
				break
			}
		case SSNDID:
			d.pcmClip.blockSize = size
			// if we didn't read the COMM, we are going to need to come back
			if d.pcmClip.sampleRate == 0 {
				rewindBytes += int64(size)
				if d.err = d.jumpTo(int(size)); d.err != nil {
					return nil
				}
			}
			d.pcmClip.r = d.r
			return d.pcmClip

		default:
			// if we read SSN but didn't read the COMM, we need to track location
			if d.pcmClip.sampleRate == 0 {
				rewindBytes += int64(size)
			}
			if d.err = d.jumpTo(int(size)); d.err != nil {
				return nil
			}
		}
	}

	return d.pcmClip
}

// FramesInt returns the audio frames contained in reader.
// Notes that this method allocates a lot of memory (depending on the duration of the underlying file).
// Consider using the decoder clip and reading/decoding using a buffer.
// DEPRECATED
func (d *Decoder) FramesInt() (frames audio.FramesInt, err error) {
	pcm := d.PCM()
	if pcm == nil {
		return nil, fmt.Errorf("no PCM data available")
	}
	totalFrames := int(pcm.Size()) * int(d.NumChans)
	frames = make(audio.FramesInt, totalFrames)
	n, err := pcm.Ints(frames)
	return frames[:n*int(d.NumChans)], err
}

// Frames returns the audio frames contained in reader.
// Notes that this method allocates a lot of memory (depending on the duration of the underlying file).
// Consider using the decoder clip and reading/decoding using a buffer.
// DEPRECATED
func (d *Decoder) Frames() (frames audio.Frames, err error) {
	clip := d.PCM()
	totalFrames := int(clip.Size())
	readFrames := 0

	bufSize := 4096
	buf := make([]byte, bufSize)
	var tFrames audio.Frames
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
// DEPRECATED
func (d *Decoder) DecodeFrames(data []byte) (frames audio.Frames, err error) {
	numChannels := int(d.NumChans)
	r := bytes.NewBuffer(data)

	bytesPerSample := int((d.BitDepth-1)/8 + 1)
	sampleBufData := make([]byte, bytesPerSample)

	frames = make(audio.Frames, len(data)/bytesPerSample)
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
