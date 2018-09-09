package ilbm

import (
	"fmt"
	"io"

	"github.com/textmodes/parser/image/iff"
)

const (
	bodyType = "BODY"
)

type bodyDecoder struct{}

func (decoder bodyDecoder) Decode(context *iff.Decoder, r *io.SectionReader, kind string) (iff.Chunk, error) {
	if kind != bodyType {
		return nil, fmt.Errorf("ilbm: expected type %q, got %q", bodyType, kind)
	}

	body := &Body{size: r.Size(), Data: make([]byte, r.Size())}
	if _, err := io.ReadFull(r, body.Data); err != nil {
		return nil, err
	}

	return body, nil
}

type Body struct {
	Data []byte
	size int64
}

func (body Body) Type() string { return bodyType }
func (body Body) Len() int     { return int(body.size) }
