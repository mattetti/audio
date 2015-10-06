package aiff

import (
	"encoding/binary"
	"errors"
	"io"
	"sync"
)

// Chunk is a struct representing a data chunk
// the reader is shared with the container but convenience methods
// are provided.
// The reader always starts at the beggining of the data.
// SSND chunk is the sound chunk
// Chunk specs:
// http://www.onicos.com/staff/iz/formats/aiff.html
// AFAn seems to be an OS X specific chunk, meaning & format TBD
type Chunk struct {
	ID     [4]byte
	Size   int
	R      io.Reader
	okChan chan bool
	Pos    int
	Wg     *sync.WaitGroup
}

// Done signals the parent parser that we are done reading the chunk
// if the chunk isn't fully read, this code will do so before signaling.
func (ch *Chunk) Done() {
	if !ch.IsFullyRead() {
		ch.drain()
	}
	ch.Wg.Done()
}

func (ch *Chunk) drain() {
	var err error
	bytesAhead := ch.Size - ch.Pos
	for bytesAhead > 0 {
		readSize := bytesAhead
		if readSize > 4000 {
			readSize = 4000
		}

		buf := make([]byte, readSize)
		err = binary.Read(ch.R, binary.LittleEndian, &buf)
		if err != nil {
			return
		}
		bytesAhead -= readSize
	}
}

// ReadLE reads the Little Endian chunk data into the passed struct
func (ch *Chunk) ReadLE(dst interface{}) error {
	if ch == nil || ch.R == nil {
		return errors.New("nil chunk/reader pointer")
	}
	if ch.IsFullyRead() {
		return io.EOF
	}
	ch.Pos += binary.Size(dst)
	return binary.Read(ch.R, binary.LittleEndian, dst)
}

// ReadBE reads the Big Endian chunk data into the passed struct
func (ch *Chunk) ReadBE(dst interface{}) error {
	if ch.IsFullyRead() {
		return io.EOF
	}
	ch.Pos += binary.Size(dst)
	return binary.Read(ch.R, binary.BigEndian, dst)
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

// IsFullyRead checks if we're finished reading the chunk
func (ch *Chunk) IsFullyRead() bool {
	if ch == nil || ch.R == nil {
		return true
	}
	return ch.Size <= ch.Pos
}
