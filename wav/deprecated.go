package wav

import (
	"bytes"
	"fmt"

	"github.com/mattetti/audio"
	"github.com/mattetti/audio/riff"
)

// PCM returns an audio.PCM compatible value to consume the PCM data
// contained in the underlying wav data.
// DEPRECATED
func (d *Decoder) PCM() *PCM {
	if d.pcmClip != nil {
		return d.pcmClip
	}
	d.err = d.readHeaders()
	if d.err != nil {
		return nil
	}

	var chunk *riff.Chunk
	for d.err == nil {
		chunk, d.err = d.parser.NextChunk()
		if d.err != nil {
			break
		}
		if chunk.ID == riff.DataFormatID {
			break
		}
		chunk.Drain()
	}
	if chunk == nil {
		return nil
	}

	d.pcmClip = &PCM{
		r:          d.r,
		byteSize:   chunk.Size,
		channels:   int(d.NumChans),
		bitDepth:   int(d.BitDepth),
		sampleRate: int64(d.SampleRate),
		blockSize:  chunk.Size,
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

// DecodeFrames decodes PCM bytes into audio frames based on the decoder context.
// This function is usually used in conjunction with Clip.Read which returns the amount
// of frames read into the buffer. It's highly recommended to slice the returned frames
// of this function by the amount of total frames reads into the buffer.
// The reason being that if the buffer didn't match the exact size of the frames,
// some of the data might be garbage but will still be converted into frames.
// DEPRECATED
func (d *Decoder) DecodeFrames(data []byte) (frames audio.Frames, err error) {
	numChannels := int(d.NumChans)
	r := bytes.NewBuffer(data)

	bytesPerSample := int((d.BitDepth-1)/8 + 1)
	sampleBufData := make([]byte, bytesPerSample)
	decodeF, err := sampleDecodeFunc(int(d.BitDepth))
	if err != nil {
		return nil, fmt.Errorf("could not get sample decode func %v", err)
	}

	frames = make(audio.Frames, len(data)/bytesPerSample)
	for j := 0; j < int(numChannels); j++ {
		frames[j] = make([]int, numChannels)
	}
	n := 0

outter:
	for i := 0; (i + (bytesPerSample * numChannels)) <= len(data); {
		frame := make([]int, numChannels)
		for j := 0; j < numChannels; j++ {
			_, err = r.Read(sampleBufData)
			if err != nil {
				break outter
			}
			frame[j] = decodeF(sampleBufData)
			i += bytesPerSample
		}
		frames[n] = frame
		n++
	}

	return frames, err
}
