package caf

import (
	"encoding/binary"
	"fmt"
	"io"
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
	r io.Reader

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
}

func (d *Decoder) Parse() error {
	var err error

	if err = d.Read(&d.Format); err != nil {
		return err
	}
	if d.Format != fileHeaderID {
		return fmt.Errorf("%s %s", string(d.Format[:]), ErrFmtNotSupported)
	}

	if err = d.Read(&d.Version); err != nil {
		return err
	}

	if d.Version > 1 {
		return fmt.Errorf("CAF v%s - %s", d.Version, ErrFmtNotSupported)
	}

	// ignore the flags value
	if err = d.Read(&d.Flags); err != nil {
		return err
	}

	// The Audio Description chunk is required and must appear in a CAF file immediately following the file header. It describes the format of the audio data in the Audio Data chunk.
	cType, _, err := d.chunkHeader()
	if err != nil {
		return err
	}
	if string(cType) != StreamDescriptionChunkID {
		return fmt.Errorf("%s - Expected description chunk", ErrUnexpectedData)
	}
	if err := d.parseDescChunk(); err != nil {
		return err
	}

	// parse the actual content
	for err == nil {
		err = d.parseChunk()
	}

	if err != io.EOF {
		return err
	}

	return nil
}

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

func (d *Decoder) chunkHeader() ([]byte, int64, error) {
	var err error
	var cSize int64
	cType := make([]byte, 4)

	if err = d.Read(&cType); err != nil {
		return nil, 0, err
	}
	if err = d.Read(&cSize); err != nil {
		return nil, 0, err
	}

	return cType, cSize, err
}

func (d *Decoder) parseChunk() error {

	cType, cSize, err := d.chunkHeader()
	if err != nil {
		return err
	}

	t := string(cType)
	switch t {
	default:
		fmt.Println(t)
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
