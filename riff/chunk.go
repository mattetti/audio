package riff

import (
	"encoding/binary"
	"io"
)

// Chunk represents the header and containt of a sub block
type Chunk struct {
	ID   [4]byte
	Size uint32
	R    *ChunkReader
}

// ChunkReader is a convenient type around chunk data
type ChunkReader struct {
	r    io.Reader
	size int
	pos  int
}

// ReadLE reads the Little Endian chunk data into the passed struct
func (ch *ChunkReader) ReadLE(dst interface{}) error {
	if ch.size <= ch.pos {
		return io.EOF
	}
	ch.pos += intDataSize(dst)
	return binary.Read(ch.r, binary.LittleEndian, dst)
}

// ReadBE reads the Big Endian chunk data into the passed struct
func (ch *ChunkReader) ReadBE(dst interface{}) error {
	if ch.IsFullyRead() {
		return io.EOF
	}
	ch.pos += intDataSize(dst)
	return binary.Read(ch.r, binary.LittleEndian, dst)
}

// ReadByte reads and returns a single byte
func (ch *ChunkReader) ReadByte() (byte, error) {
	if ch.IsFullyRead() {
		return 0, io.EOF
	}
	var r byte
	err := ch.ReadLE(&r)
	return r, err
}

// IsFullyRead checks if we finished reading the chunk
func (ch *ChunkReader) IsFullyRead() bool {
	return ch.size <= ch.pos
}
