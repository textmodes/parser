package ilbm

import (
	"encoding/binary"
	"fmt"
	"image/color"
	"io"

	"github.com/textmodes/parser/format/iff"
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

// ColorMap chunk.
type ColorMap []uint8

// Type of chunk.
func (chunk ColorMap) Type() string { return colorMapType }

// Len is the length of the chunk in bytes.
func (chunk ColorMap) Len() int { return len(chunk) }

// Palette of the color map, it is assumed the color map consists of 3-byte
// RGB triplets.
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

// ColorRange chunk.
type ColorRange struct {
	Padding   int16
	Rate      int16
	Flags     int16
	Low, High uint8
}

// Type of chunk.
func (chunk ColorRange) Type() string { return colorRangeType }

// Len is the length of the chunk in bytes.
func (chunk ColorRange) Len() int { return colorRangeSize }
