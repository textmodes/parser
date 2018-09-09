package bgi

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"os"

	"github.com/textmodes/parser/image/drawing"
)

const (
	SUB = 0x1a
)

var (
	scale = struct {
		up, down []int
	}{
		[]int{1, 6, 2, 3, 1, 4, 5, 2, 5, 3, 4},
		[]int{1, 10, 3, 4, 1, 3, 3, 1, 2, 1, 1},
	}
)

const (
	/*
		strokeOpEnd = iota
		strokeOpScan
		strokeOpMove
		strokeOpDraw
	*/
	strokeOpMask = 0x8080
	strokeOpEnd  = 0x0000
	strokeOpMove = 0x0080
	strokeOpScan = 0x8000
	strokeOpDraw = 0x8080
)

type stroke uint16

func (s stroke) Y() int {
	/*
		if s&0x4000 != 0 {
			return -int((s & 0x3f00) >> 8)
		}
		return int(s&0x3f00) >> 8
	*/
	return int(int8(s>>7) >> 1)
}

func (s stroke) X() int {
	/*
		if s&0x40 != 0 {
			return -int(s & 0x3f)
		}
		return int(s & 0x3f)
	*/
	return int(int8(s<<1) >> 1)
}

func (s stroke) ImagePoint() image.Point {
	return image.Pt(s.X(), s.Y())
}

func (s stroke) Point() drawing.Point {
	return drawing.Point{
		X: float64(s.X()),
		Y: float64(s.Y()),
	}
}

func (s stroke) String() string {
	switch s & strokeOpMask {
	case strokeOpEnd:
		return "end"
	case strokeOpScan:
		return "scan"
	case strokeOpMove:
		return fmt.Sprintf("move %s (%#04x)", s.Point(), uint16(s)&^strokeOpMask)
	case strokeOpDraw:
		return fmt.Sprintf("draw %s (%#04x)", s.Point(), uint16(s)&^strokeOpMask)
	default:
		return "invalid"
	}
}

type fontFileHeader struct {
	HeaderSize uint16
	Name       [4]byte
	FileSize   uint16 // File size in bytes
	FontMajor  uint8
	FontMinor  uint8
	BGIMajor   uint8
	BGIMinor   uint8
}

type fontHeader struct {
	Signature uint8
	Chars     uint16
	Unused1   uint8
	FirstChar uint8
	CharDefs  uint16
	ScanFlag  uint8
	OrgToCap  int8
	OrgToBase int8
	OrgToDec  int8
	Unused2   [5]byte
}

type Font struct {
	fileHeader fontFileHeader
	header     fontHeader
	name       string
	start      uint16
	offsets    []uint16
	wtable     []uint8
	xoffset    []int
	widths     []int
	yoffset    int
	height     int
	vectors    []stroke
}

func NewFont(r io.ReadSeeker) (*Font, error) {
	var (
		f       Font
		ident   [8]byte
		err     error
		oneByte [1]byte
	)
	if _, err = io.ReadFull(r, ident[:]); err != nil {
		return nil, err
	}

	// read ident
	if !bytes.Equal(ident[:], []byte("PK\b\bBGI ")) {
		return nil, errors.New("bgi: not a BGI font")
	}

	// read description
	for {
		if _, err = r.Read(oneByte[:]); err != nil {
			return nil, err
		}
		if oneByte[0] == SUB {
			break
		}
		f.name += string(oneByte[0])
	}

	// read file header
	if err = binary.Read(r, binary.LittleEndian, &f.fileHeader); err != nil {
		return nil, err
	}

	// read font header
	// skip header size - len(ident) + len(name) + 1 + sizeof(fontFileHeader)
	skip := int64(f.fileHeader.HeaderSize) - 20 - int64(len(f.name)) - 1
	if skip > 0 {
		if _, err = r.Seek(skip, os.SEEK_CUR); err != nil {
			return nil, err
		}
	}
	if err = binary.Read(r, binary.LittleEndian, &f.header); err != nil {
		return nil, err
	}
	if f.header.Signature != '+' {
		return nil, errors.New("bgi: not a stroked font")
	}

	// read offsets
	f.offsets = make([]uint16, f.header.Chars)
	if err = binary.Read(r, binary.LittleEndian, &f.offsets); err != nil {
		return nil, err
	}

	// read widths
	f.wtable = make([]uint8, f.header.Chars)
	if err = binary.Read(r, binary.LittleEndian, &f.wtable); err != nil {
		return nil, err
	}

	// read glyphs
	var b []byte
	if b, err = ioutil.ReadAll(r); err != nil {
		return nil, err
	}
	if len(b)%2 != 0 {
		return nil, errors.New("bgi: uneven number of bytes left")
	}
	f.start = uint16(skip) + 3*uint16(f.header.Chars)
	f.vectors = make([]stroke, len(b)>>1)
	if err = binary.Read(bytes.NewBuffer(b), binary.LittleEndian, &f.vectors); err != nil {
		return nil, err
	}

	f.xoffset = make([]int, f.header.Chars)
	f.widths = make([]int, f.header.Chars)
	f.parseWidths()

	return &f, nil
}

func (f *Font) String() string {
	return fmt.Sprintf("font name=%q, description=%q", f.Name(), f.name)
}

func (f *Font) Name() string {
	return string(f.fileHeader.Name[:])
}

func (f *Font) parseWidths() {
	var (
		i          uint16
		ymin       = 32000
		ymax       = -32000
		origHeight = f.header.OrgToCap - f.header.OrgToDec
	)
	for i = 0; i < f.header.Chars; i++ {
		var (
			xpos, ypos, xend, yend int
			xmin                   = 32000
			xmax                   = -32000
			offset                 = f.offsets[i] >> 1
			strokes                = f.vectors[offset:]
		)
		//log.Printf("char %d at %d of %d (offset: %d-%d); first stroke: %s",
		//	i, offset, len(f.vectors), f.offsets[i], f.start, strokes[0])
	reading:
		for _, stroke := range strokes {
			//log.Println(stroke)
			switch stroke & strokeOpMask {
			case strokeOpMove:
				xpos = stroke.X()
				ypos = int(f.header.OrgToCap) - stroke.Y()
			case strokeOpDraw:
				xend = stroke.X()
				yend = int(f.header.OrgToCap) - stroke.Y()
				xmin = min(xmin, min(xpos, xend))
				ymin = min(ymin, min(ypos, yend))
				xmax = max(xmax, max(xpos, xend))
				ymax = max(ymax, max(ypos, yend))
				xpos = xend
				ypos = yend
			case strokeOpScan:
			default:
				break reading
			}
		}
		f.xoffset[i] = max(0, -xmin)
		f.widths[i] = max(1, max(xmax+1, int(f.wtable[i]))+f.xoffset[i])
	}

	f.yoffset = max(0, -ymin)
	f.height = max(ymax+1, int(origHeight)) + f.yoffset
}

func (f *Font) charWidth(char byte) int {
	char -= f.header.FirstChar
	if uint16(char) >= f.header.Chars {
		return -1
	}
	return f.widths[char]
}

func (f *Font) charDraw(c *drawing.Context, x, y, size int, char byte) int {
	char -= f.header.FirstChar
	if uint16(char) >= f.header.Chars || size < 0 {
		return -1
	}
	var (
		offset           = f.offsets[char] >> 1
		strokes          = f.vectors[offset:]
		ratio            = float64(scale.up[size]) / float64(scale.down[size])
		ox               = float64(x)
		oy               = float64(y)
		xpos, ypos, xmax int
	)
reading:
	for _, stroke := range strokes {
		switch stroke & strokeOpMask {
		case strokeOpEnd:
			break reading
		case strokeOpMove:
			xpos = stroke.X() + f.xoffset[char]
			ypos = int(f.header.OrgToCap) - stroke.Y() + f.yoffset
			xmax = max(xmax, xpos)
			c.NewSubPath()
			c.LineTo(
				ox+float64(xpos)*ratio,
				oy+float64(ypos)*ratio,
			)
		case strokeOpDraw:
			xpos = stroke.X() + f.xoffset[char]
			ypos = int(f.header.OrgToCap) - stroke.Y() + f.yoffset
			xmax = max(xmax, xpos)
			c.LineTo(
				ox+float64(xpos)*ratio,
				oy+float64(ypos)*ratio,
			)
		}
	}
	//c.Stroke()
	c.Fill()
	return xmax
}

// Draw text onto an image
func (f *Font) Draw(dst image.Image, x, y, size int, text string, textColor color.Color) {
	c := drawing.NewContext(dst)
	c.Color(color.White)
	c.FillStyle(drawing.NewSolidPattern(color.White))
	o := x
	for i, l := 0, len(text); i < l; i++ {
		char := text[i]
		move := f.charDraw(c, o, y, size, char)
		o += move
	}
}

func abs(i int) int {
	if i < 0 {
		return i * -1
	}
	return i
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
