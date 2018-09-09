package iff

import (
	"fmt"
	"io"
)

const formType = "FORM"

type formDecoder struct{}

func (decoder formDecoder) Decode(context *Decoder, r *io.SectionReader, kind string) (Chunk, error) {
	if kind != formType {
		return nil, fmt.Errorf("ilbm: expected tag %q, got %q", formType, kind)
	}

	var (
		chunk = new(Form)
		err   error
	)
	if chunk.kind, err = readString(r, 4); err != nil {
		return nil, err
	}

	var (
		off   int64 = 4
		limit       = r.Size()
	)
	for off < limit {
		child, err := context.Decode(r)
		if err != nil {
			return nil, err
		}
		chunk.Chunks = append(chunk.Chunks, child)

		off += 8
		off += int64(child.Len())
		if off%2 != 0 {
			off++ // Align even
		}
		if off, err = r.Seek(off, io.SeekStart); err != nil {
			return nil, err
		}
	}

	return chunk, nil
}

type Form struct {
	kind   string
	size   int64
	Chunks []Chunk
}

func (form Form) Type() string { return form.kind }
func (form Form) Len() int     { return int(form.size) }

func (form Form) Chunk(kind string) Chunk {
	for _, chunk := range form.Chunks {
		if chunk.Type() == kind {
			return chunk
		}
	}
	return nil
}
