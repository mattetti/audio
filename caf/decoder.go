package caf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/go-audio/chunk"
)

type AudioDescChunk struct {
}

/*
Decoder
CAF files begin with a file header, which identifies the file type and the CAF version,
followed by a series of chunks. A chunk consists of a header, which defines the type of the chunk and
indicates the size of its data section, followed by the chunk data.
The nature and format of the data is specific to each type of chunk.

The only two chunk types required for every CAF file are the Audio Data chunk and the Audio Description chunk,
which specifies the audio data format.

The Audio Description chunk must be the first chunk following the file header.
The Audio Data chunk can appear anywhere else in the file, unless the size of its data section has not been determined.
In that case, the size field in the Audio Data chunk header is set to -1 and the Audio Data chunk must come last in the file
so that the end of the audio data chunk is the same as the end of the file.
This placement allows you to determine the data section size when that information is not available in the size field.

Audio is stored in the Audio Data chunk as a sequential series of packets. An audio packet in a CAF file contains one or more frames of audio data.

Every chunk consists of a chunk header followed by a data section. Chunk headers contain two fields:
* A four-character code indicating the chunk’s type
* A number indicating the chunk size in bytes

The format of the data in a chunk depends on the chunk type.
It consists of a series of sections, typically called fields.
The format of the audio data depends on the data type. All of the other fields in a CAF file are in big-endian (network) byte order.


*/
type Decoder struct {
	r io.ReadSeeker

	// Ch chan *TBD

	// Format: the file type. This value must be set to 'caff'.
	// You should consider only files with the Type field set to 'caff' to be valid CAF files.
	Format [4]byte
	// Version: The file version. For CAF files conforming to this specification, the version must be set to 1.
	// If Apple releases a substantial revision of this specification, files compliant with that revision will have their Version
	// field set to a number greater than 1.
	Version uint16
	// Flags reserved by Apple for future use. For CAF v1 files, must be set to 0. You should ignore any value of this field you don’t understand,
	// and you should accept the file as a valid CAF file as long as the version and file type fields are valid.
	Flags uint16

	// The number of sample frames per second of the data. You can combine this value with the frames per packet to determine the amount of time represented by a packet. This value must be nonzero.
	SampleRate float64

	// A four-character code indicating the general kind of data in the stream.
	FormatID [4]byte

	// Flags specific to each format. May be set to 0 to indicate no format flags.
	// Detailed specification linear PCM, MPEG-4 AAC, and AC-3
	FormatFlags uint32

	// The number of bytes in a packet of data. For formats with a variable packet size,
	// this field is set to 0. In that case, the file must include a Packet Table chunk Packet Table Chunk.
	// Packets are always aligned to a byte boundary. For an example of an Audio Description chunk for a format with a variable packet size
	BytesPerPacket uint32

	// The number of sample frames in each packet of data. For compressed formats,
	// this field indicates the number of frames encoded in each packet. For formats with a variable number of frames per packet,
	// this field is set to 0 and the file must include a Packet Table chunk Packet Table Chunk.
	FramesPerPacket uint32

	// The number of channels in each frame of data. This value must be nonzero.
	ChannelsPerFrame uint32

	// The number of bits of sample data for each channel in a frame of data.
	// This field must be set to 0 if the data format (for instance any compressed format) does not contain separate samples for each channel
	BitsPerChannel uint32

	// Size of the audio data
	//A size value of -1 indicates that the size of the data section for this chunk is unknown. In this case, the Audio Data chunk must appear last in the file
	// so that the end of the Audio Data chunk is the same as the end of the file.
	// This placement allows you to determine the data section size.
	AudioDataSize int64

	err error
}

// ReadInfo reads the underlying reader finds the data it needs.
// This method is safe to call multiple times.
func (d *Decoder) ReadInfo() error {
	if d == nil || d.SampleRate > 0 {
		return nil
	}
	if d.err = d.readHeaders(); d.err != nil {
		d.err = fmt.Errorf("failed to read header - %v", d.err)
		return d.err
	}

	var (
		id          [4]byte
		size        int64
		rewindBytes int64
	)
	for d.err != io.EOF {
		id, size, d.err = d.iDnSize()
		if d.err != nil {
			if d.err != io.EOF {
				d.err = fmt.Errorf("error reading chunk header - %v", d.err)

			}
			break
		}
		switch id {
		default:
			if d.SampleRate == 0 {
				rewindBytes += int64(size) + 8 // we add 8 for the ID and size of this chunk
			}
			if d.err = d.jumpTo(int(size)); d.err != nil {
				break
			}
		}
	}
	return d.Err()
}

// Err returns the last non-EOF error that was encountered by the Decoder.
func (d *Decoder) Err() error {
	if d.err == io.EOF {
		return nil
	}
	return d.err
}

// String implements the stringer interface
func (d *Decoder) String() string {
	out := fmt.Sprintf("Format: %s - %s", string(d.Format[:]), string(d.FormatID[:]))
	out += fmt.Sprintf("%d channels @ %d - ", d.ChannelsPerFrame, int(d.SampleRate))
	out += fmt.Sprintf("data size: %d", d.AudioDataSize)

	return out
}

// readHeaders is safe to call multiple times
// byte size of the header: 12
func (d *Decoder) readHeaders() error {
	// prevent the headers to be re-read
	if d.Version > 0 {
		return nil
	}
	var n int64
	size := 8 // 4 + 2 + 2
	src := bytes.NewBuffer(make([]byte, 0, size))
	n, d.err = io.CopyN(src, d.r, int64(size))
	if n < int64(size) {
		src.Truncate(int(n))
	}

	// format
	if _, d.err = src.Read(d.Format[:]); d.err != nil {
		return d.err
	}
	if d.Format != fileHeaderID {
		return fmt.Errorf("%s %s", string(d.Format[:]), ErrFmtNotSupported)
	}

	// version
	if d.err = binary.Read(src, binary.BigEndian, &d.Version); d.err != nil {
		return d.err
	}
	if d.Version > 1 {
		return fmt.Errorf("CAF v%d - %v", d.Version, ErrFmtNotSupported)
	}

	// The Audio Description chunk is required and must appear in a CAF file immediately following the file header. It describes the format of the audio data in the Audio Data chunk.
	cType, _, err := d.iDnSize()
	if err != nil {
		return err
	}
	if cType != StreamDescriptionChunkID {
		return fmt.Errorf("%s - Expected description chunk", ErrUnexpectedData)
	}
	if err := d.parseDescChunk(); err != nil {
		return err
	}

	return d.err
}

// NextChunk returns the next available chunk
func (d *Decoder) NextChunk() (*chunk.Reader, error) {
	var err error

	if err = d.readHeaders(); err != nil {
		return nil, err
	}

	var (
		id   [4]byte
		size int64
	)

	id, size, d.err = d.iDnSize()
	if d.err != nil {
		if d.err == io.EOF || d.err == io.ErrUnexpectedEOF {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("error reading chunk header - %v", d.err)
	}

	c := &chunk.Reader{
		ID:   id,
		Size: int(size),
		R:    io.LimitReader(d.r, int64(size)),
	}

	return c, d.err
}

// parseDescChunk parses the first chunk called description chunk.
func (d *Decoder) parseDescChunk() error {
	if err := d.Read(&d.SampleRate); err != nil {
		return err
	}
	if err := d.Read(&d.FormatID); err != nil {
		return err
	}
	if err := d.Read(&d.FormatFlags); err != nil {
		return err
	}
	if err := d.Read(&d.BytesPerPacket); err != nil {
		return err
	}
	if err := d.Read(&d.FramesPerPacket); err != nil {
		return err
	}
	if err := d.Read(&d.ChannelsPerFrame); err != nil {
		return err
	}
	if err := d.Read(&d.BitsPerChannel); err != nil {
		return err
	}

	return nil
}

func (d *Decoder) Duration() time.Duration {
	//duration := time.Duration((float64(p.Size) / float64(p.AvgBytesPerSec)) * float64(time.Second))
	//duration := time.Duration(float64(p.NumSampleFrames) / float64(p.SampleRate) * float64(time.Second))

	return 0
}

func (d *Decoder) iDnSize() ([4]byte, int64, error) {
	var err error
	var cSize int64
	var cType [4]byte

	if err = d.Read(&cType); err != nil {
		return cType, 0, err
	}
	if err = d.Read(&cSize); err != nil {
		return cType, 0, err
	}

	return cType, cSize, err
}

func (d *Decoder) parseChunk() error {

	cType, cSize, err := d.iDnSize()
	if err != nil {
		return err
	}

	t := cType
	switch t {
	case AudioDataChunkID:
		d.AudioDataSize = cSize
		// TODO:
		// editCount uint32
		// The modification status of the data section. You should initially set this field to 0, and should increment it each time the audio data in the file is modified.
		// the rest of the data is the actual audio data.
		var err error
		bytesToSkip := cSize
		for bytesToSkip > 0 {
			readSize := bytesToSkip
			if readSize > 4000 {
				readSize = 4000
			}
			buf := make([]byte, readSize)
			if err = binary.Read(d.r, binary.LittleEndian, &buf); err != nil {
				return nil
			}
			bytesToSkip -= readSize
		}
	case InfoStringsChunkID:
		chunks := &stringsChunk{}
		if err = binary.Read(d.r, binary.BigEndian, &chunks.numEntries); err != nil {
			return nil
		}

	default:
		// kuki
		// The Magic Cookie chunk contains supplementary (“magic cookie”) data required by certain audio data formats, such as MPEG-4 AAC, for decoding of the audio data. If the audio data format contained in a CAF file requires magic cookie data, the file must have this chunk.
		// https://developer.apple.com/library/content/documentation/MusicAudio/Reference/CAFSpec/CAF_spec/CAF_spec.html#//apple_ref/doc/uid/TP40001862-CH210-BCGFCCFA

		// strg
		// The optional Strings chunk contains any number of textual
		// strings, along with an index for accessing them. These strings serve
		// as labels for other chunks, such as Marker or Region chunks.

		// free
		// The optional Free chunk is for reserving space, or providing
		// padding, in a CAF file. The contents of the Free chunk data section
		// have no significance and should be ignored.

		// info
		// You can use the optional Information chunk to contain any number
		// of human-readable text strings. Each string is accessed through a
		// standard or application-defined key. You should consider information
		// in this chunk to be secondary when the same information appears in
		// other chunks. For example, both the Information chunk and the MIDI
		// chunk (MIDI Chunk) may specify key signature and tempo. In that case,
		// the MIDI chunk values overrides the values in the Information chunk.

		// uuid
		//
		// You can define your own chunk type to extend the CAF file
		// specification. For this purpose, this specification includes the
		// User-Defined chunk type, which you can use to provide a unique
		// universal identifier for your custom chunk. When parsing a CAF file,
		// you should ignore any chunk with a UUID that you do not recognize.

		// ovvw
		//
		// You can use the optional Overview chunk to hold sample descriptions
		// that you can use to draw a graphical view of the audio data in a CAF
		// file. A CAF file can include multiple Overview chunks to represent
		// the audio at multiple graphical resolutions.

		// peak
		//
		// You can use the optional Peak chunk to describe the peak amplitude
		// present in each channel of a CAF file and to indicate in which frame
		// the peak occurs for each channel.

		fmt.Println(string(t[:]))
		buf := make([]byte, cSize)
		return d.Read(buf)
	}

	return nil
}

func (d *Decoder) ReadByte() (byte, error) {
	var b byte
	err := binary.Read(d.r, binary.BigEndian, &b)
	return b, err
}

// read reads n bytes from the parser's reader and stores them into the provided dst,
// which must be a pointer to a fixed-size value.
func (d *Decoder) Read(dst interface{}) error {
	return binary.Read(d.r, binary.BigEndian, dst)
}

// jumpTo advances the reader to the amount of bytes provided
func (d *Decoder) jumpTo(bytesAhead int) error {
	var err error
	if bytesAhead > 0 {
		_, err = d.r.Seek(int64(bytesAhead), io.SeekCurrent)
		// TODO: benchmark against
		// _, err = io.CopyN(ioutil.Discard, d.r, int64(bytesAhead))
	}
	return err
}
