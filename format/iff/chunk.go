package iff

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// Chunk is an IFF chunk.
type Chunk interface {
	Type() string
	Len() int
}

// ChunkDecoder can decode an IFF chunk.
type ChunkDecoder interface {
	// Decode an IFF chunk.
	Decode(*Decoder, *io.SectionReader, string) (Chunk, error)
}

func readBytes(r io.Reader, size int) ([]byte, error) {
	b := make([]byte, size)
	if _, err := io.ReadFull(r, b); err != nil {
		return nil, err
	}
	return b, nil
}

func readString(r io.Reader, size int) (string, error) {
	b, err := readBytes(r, size)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func readUint8(r io.Reader) (uint8, error) {
	var v uint8
	if err := binary.Read(r, binary.BigEndian, &v); err != nil {
		return 0, err
	}
	return v, nil
}

func readUint16(r io.Reader) (uint16, error) {
	var v uint16
	if err := binary.Read(r, binary.BigEndian, &v); err != nil {
		return 0, err
	}
	return v, nil
}

func readUint32(r io.Reader) (uint32, error) {
	var v uint32
	if err := binary.Read(r, binary.BigEndian, &v); err != nil {
		return 0, err
	}
	return v, nil
}

type unknownChunkDecoder struct{}

func (decoder unknownChunkDecoder) Decode(dec *Decoder, r *io.SectionReader, kind string) (Chunk, error) {
	size := r.Size()
	if size > math.MaxUint32 {
		return nil, fmt.Errorf("ilbm: chunk of size %d exceeds 32-bit limit", size)
	}

	var chunk = &Unknown{
		kind: kind,
		size: size,
	}
	if _, err := io.ReadFull(r, chunk.data); err != nil {
		return nil, err
	}
	return chunk, nil
}

// Unknown chunk holds the raw bytes.
type Unknown struct {
	kind string
	size int64
	data []byte
}

// Type of chunk.
func (chunk Unknown) Type() string { return chunk.kind }

// Len is the length of the chunk in bytes.
func (chunk Unknown) Len() int { return int(chunk.size) }

// Bytes are the raw chunk bytes.
func (chunk Unknown) Bytes() []byte { return chunk.data }
