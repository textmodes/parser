package iff

import (
	"io"
	"log"
)

// ReadAtSeeker encapsulates the same functionality as io.SectionReader.
// Conveniently, it is *also* implemented by os.File!
type ReadAtSeeker interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}

type Decoder struct {
	builtin map[string]ChunkDecoder
	custom  map[string]ChunkDecoder
}

func NewDecoder(custom map[string]ChunkDecoder) *Decoder {
	if custom == nil {
		custom = make(map[string]ChunkDecoder)
	}
	return &Decoder{
		builtin: map[string]ChunkDecoder{
			formType: formDecoder{},
		},
		custom: custom,
	}
}

func (decoder *Decoder) Decode(r ReadAtSeeker) (Chunk, error) {
	kind, err := readString(r, 4)
	if err != nil {
		return nil, err
	}

	size, err := readUint32(r)
	if err != nil {
		return nil, err
	}

	off, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	dec, ok := decoder.custom[kind]
	if !ok {
		dec, ok = decoder.builtin[kind]
	}
	if !ok {
		log.Printf("ilbm: unknown type %q", kind)
		dec = unknownChunkDecoder{}
	}

	return dec.Decode(decoder, io.NewSectionReader(r, off, int64(size)), kind)

}
