package ilbm

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/textmodes/parser/format/iff"
)

const (
	bitmapHeaderType = "BMHD"
	bitmapHeaderSize = 20
)

const (
	maskNone = iota
	maskMask
	maskTransparentColor
	maskLasso
)

type bitmapHeaderDecoder struct {
}

func (decoder bitmapHeaderDecoder) Decode(context *iff.Decoder, r *io.SectionReader, kind string) (iff.Chunk, error) {
	if kind != bitmapHeaderType {
		return nil, fmt.Errorf("ilbm: expected tag %q, got %q", bitmapHeaderType, kind)
	}

	var chunk = new(BitmapHeader)
	if err := binary.Read(r, binary.BigEndian, chunk); err != nil {
		return nil, err
	}

	//log.Printf("ilbm: %#+v", chunk)
	return chunk, nil
}

/*
UWORD w, h;             /* raster width & height in pixels
WORD  x, y;             /* pixel position for this image
UBYTE nPlanes;          /* # source bitplanes
Masking masking;
Compression compression;
UBYTE pad1;             /* unused; ignore on read, write as 0
UWORD transparentColor; /* transparent "color number" (sort of)
UBYTE xAspect, yAspect; /* pixel aspect, a ratio width : height
WORD  pageWidth, pageHeight;  /* source "page" size in pixels
*/

// BitmapHeader chunk.
type BitmapHeader struct {
	Width, Height         uint16
	X, Y                  int16
	Planes                uint8
	Masking               uint8
	Compression           uint8
	Flags                 uint8
	Transparent           uint16
	XAspect, YAspect      uint8
	PageWidth, PageHeight int16
}

// Type of chunk.
func (chunk BitmapHeader) Type() string { return bitmapHeaderType }

// Len is the length of the chunk in bytes.
func (chunk BitmapHeader) Len() int { return bitmapHeaderSize }

func (chunk BitmapHeader) rowSize() int {
	var words = uint(chunk.Width) / 16
	if chunk.Width%16 != 0 {
		words++
	}
	return int(words << 1)
}
