package ilbm

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"

	"github.com/textmodes/parser/image/iff"
)

var custom = map[string]iff.ChunkDecoder{
	bitmapHeaderType: bitmapHeaderDecoder{},
	colorMapType:     colorMapDecoder{},
	colorRangeType:   colorRangeDecoder{},
	bodyType:         bodyDecoder{},
}

const (
	idILBM = "ILBM"
	idPBM  = "PBM "
)

func Decode(r iff.ReadAtSeeker) (image.Image, error) {
	var (
		d      = iff.NewDecoder(custom)
		c, err = d.Decode(r)
	)
	if err != nil {
		return nil, err
	}
	if form, ok := c.(*iff.Form); ok {
		switch kind := form.Type(); kind {
		case "ACBM", "ILBM", "PBM ":
			return decodeImage(form)
		default:
			return nil, fmt.Errorf("ilbm: format %q not supported", kind)
		}
	}
	return nil, errors.New("ilbm: FORM tag missing")
}

func decodeImage(form *iff.Form) (image.Image, error) {
	var (
		header   *BitmapHeader
		colorMap ColorMap
		body     *Body
		ok       bool
	)
	if header, ok = form.Chunk(bitmapHeaderType).(*BitmapHeader); !ok {
		return nil, fmt.Errorf("ilbm: no %q chunk found", bitmapHeaderType)
	}
	if colorMap, ok = form.Chunk(colorMapType).(ColorMap); !ok {
		return nil, fmt.Errorf("ilbm: no %q chunk found", colorMapType)
	}
	if body, ok = form.Chunk(bodyType).(*Body); !ok {
		return nil, fmt.Errorf("ilbm: no %q chunk found", bodyType)
	}

	var (
		im      = image.NewRGBA(image.Rect(0, 0, int(header.Width), int(header.Height)))
		palette = colorMap.Palette()
		id      = form.Type()
		err     error
	)

	if id == idILBM {
		log.Printf("ilbm: dimensions %dx%d, %d planes", header.Width, header.Height, header.Planes)
	} else {
		log.Printf("ilbm: dimensions %dx%d", header.Width, header.Height)
	}

	if id == idILBM || id == idPBM {
		switch header.Compression {
		case 0:
			log.Println("ilbm: no compression")
		case 1:
			log.Println("ilbm: byterun1 compression")
		default:
			log.Println("ilbm: unknown compression")
		}
	}

	if err = decodeBody(im, id, header, body, palette); err != nil {
		return nil, err
	}

	return im, nil
}

func decodeBody(im *image.RGBA, id string, header *BitmapHeader, body *Body, palette color.Palette) error {
	switch id {
	case idILBM:
		if header.Planes > 8 {
			if header.Planes <= 16 {
				return decodeBodyStandard(im, header, body, palette)
			} else if header.Planes%3 == 0 {
				return decodeBodyDeep(im, header, body, palette)
			} else if header.Planes < 16 {
				// will be interpreted as grayscale
				return decodeBodyStandard(im, header, body, palette)
			}
			return fmt.Errorf("ilbm: don't know how to interpret %d-plane image", header.Planes)
		}
		return decodeBodyStandard(im, header, body, palette)

	case idPBM:
		return decodeBodyPBM(im, header, body, palette)

	default:
		return decodeBodyRGBN(im, header, body, palette)
	}
}

func decodeBodyStandard(im *image.RGBA, header *BitmapHeader, body *Body, palette color.Palette) error {
	return errors.New("not implemented")
}

func decodeBodyDeep(im *image.RGBA, header *BitmapHeader, body *Body, palette color.Palette) error {
	return errors.New("not implemented")
}

func decodeBodyPBM(im *image.RGBA, header *BitmapHeader, body *Body, palette color.Palette) (err error) {
	if header.Planes != 8 {
		return fmt.Errorf("ilbm: invalid number of planes for IFF-PBM: %d (must be 8)", header.Planes)
	}
	if header.Masking == maskMask {
		return errors.New("ilbm: invalid masking for IFF-PBM")
	}

	var (
		cols     = int(header.Width)
		rows     = int(header.Height)
		col, row int
		remain   = body.Len()
		r        = bytes.NewReader(body.Data)
		plane    []byte
	)
	for row = 0; row < rows; row++ {
		if plane, err = readILBMPlane(r, &remain, cols, header.Compression); err != nil {
			return
		}
		for col = 0; col < cols; col++ {
			im.Set(col, row, palette[plane[col]])
		}
	}

	return
}

func readILBMPlane(r *bytes.Reader, remain *int, size int, compression uint8) (plane []byte, err error) {
	switch compression {
	case 0:
		plane = make([]byte, size)
		_, err = io.ReadFull(r, plane)
		return

	case 1:
		var (
			v, b byte
			j    int
		)
		for remaining := size; remaining > 0; {
			if v, err = r.ReadByte(); err != nil {
				return
			}

			if int(v) <= 127 {
				j = int(v)
				remaining -= (j + 1)
				if remaining < 0 {
					return nil, fmt.Errorf("ilbm: error decompressing byterun1: need %d more bytes", -remaining)
				}
				for ; j >= 0; j-- {
					if b, err = r.ReadByte(); err != nil {
						return
					}
					plane = append(plane, b)
				}
			} else if int(v) != 128 {
				j = 256 - int(v)
				remaining -= (j + 1)
				if remaining < 0 {
					return nil, fmt.Errorf("ilbm: error decompressing byterun1: need %d more bytes", -remaining)
				}
				if v, err = r.ReadByte(); err != nil {
					return
				}
				for ; j >= 0; j-- {
					plane = append(plane, v)
				}
			}
		}
		return

	default:
		return nil, fmt.Errorf("ilbm: unknown compression type %d", compression)
	}
}

func decodeBodyRGBN(im *image.RGBA, header *BitmapHeader, body *Body, palette color.Palette) error {
	return errors.New("not implemented")
}
