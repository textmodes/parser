package ilbm

import (
	"encoding/binary"
	"fmt"
	"image/color"
	"io"

	"github.com/textmodes/parser/image/iff"
)

const (
	colorMapType   = "CMAP"
	colorRangeType = "CRNG"
	colorRangeSize = 8
)

type colorMapDecoder struct{}

func (decoder colorMapDecoder) Decode(context *iff.Decoder, r *io.SectionReader, kind string) (iff.Chunk, error) {
	if kind != colorMapType {
		return nil, fmt.Errorf("ilbm: expected type %q, got %q", colorMapType, kind)
	}

	var (
		size  = r.Size()
		chunk = make(ColorMap, size)
	)
	if _, err := io.ReadFull(r, chunk); err != nil {
		return nil, err
	}
	return chunk, nil
}

type ColorMap []uint8

func (chunk ColorMap) Type() string { return colorMapType }
func (chunk ColorMap) Len() int     { return len(chunk) }

func (chunk ColorMap) Palette() color.Palette {
	var palette color.Palette
	for i, l := 0, len(chunk); i < l; i += 3 {
		palette = append(palette, color.RGBA{
			R: chunk[i+0],
			G: chunk[i+1],
			B: chunk[i+2],
			A: 0xff,
		})

	}
	return palette
}

type colorRangeDecoder struct{}

func (decoder colorRangeDecoder) Decode(context *iff.Decoder, r *io.SectionReader, kind string) (iff.Chunk, error) {
	if kind != colorRangeType {
		return nil, fmt.Errorf("ilbm: expected type %q, got %q", colorRangeType, kind)
	}

	chunk := new(ColorRange)
	if err := binary.Read(r, binary.BigEndian, chunk); err != nil {
		return nil, err
	}
	return chunk, nil
}

type ColorRange struct {
	Padding   int16
	Rate      int16
	Flags     int16
	Low, High uint8
}

func (chunk ColorRange) Type() string { return colorRangeType }
func (chunk ColorRange) Len() int     { return colorRangeSize }
