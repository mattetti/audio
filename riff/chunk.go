package riff

import (
	"encoding/binary"
	"errors"
	"io"
)

// Chunk represents the header and containt of a sub block
type Chunk struct {
	ID   [4]byte
	Size int
	Pos  int
	R    io.Reader
}

// ReadLE reads the Little Endian chunk data into the passed struct
func (ch *Chunk) ReadLE(dst interface{}) error {
	if ch == nil || ch.R == nil {
		return errors.New("nil chunk/reader pointer")
	}
	if ch.IsFullyRead() {
		return io.EOF
	}
	ch.Pos += intDataSize(dst)
	return binary.Read(ch.R, binary.LittleEndian, dst)
}

// ReadBE reads the Big Endian chunk data into the passed struct
func (ch *Chunk) ReadBE(dst interface{}) error {
	if ch.IsFullyRead() {
		return io.EOF
	}
	ch.Pos += intDataSize(dst)
	return binary.Read(ch.R, binary.LittleEndian, dst)
}

// ReadByte reads and returns a single byte
func (ch *Chunk) ReadByte() (byte, error) {
	if ch.IsFullyRead() {
		return 0, io.EOF
	}
	var r byte
	err := ch.ReadLE(&r)
	return r, err
}

// IsFullyRead checks if we finished reading the chunk
func (ch *Chunk) IsFullyRead() bool {
	if ch == nil || ch.R == nil {
		return true
	}
	return ch.Size <= ch.Pos
}
