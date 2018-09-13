package tundradraw

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image/color"
	"io"
	"io/ioutil"

	"github.com/textmodes/parser/chargen"
	"github.com/textmodes/parser/format/sauce"
	"github.com/textmodes/parser/format/vga"
	"github.com/textmodes/parser/text/ansi"
)

const (
	tundraDrawID         = "\x18TUNDRA24"
	tundraDrawPos        = 0x01
	tundraDrawForeground = 0x02
	tundraDrawBackground = 0x04
	tundraDrawColors     = tundraDrawBackground | tundraDrawForeground
)

// TundraDraw parser
type TundraDraw struct {
	// Text buffer.
	*vga.Text

	// Record is the contained SAUCE record (may be nil).
	Record *sauce.Record

	// Font used to render the BinaryText (may be nil, defaults to IBM VGA).
	Font *chargen.Font

	is24bit bool
}

// Decode a TundraDraw file.
func Decode(r io.Reader) (*TundraDraw, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if string(b[:len(tundraDrawID)]) != tundraDrawID {
		return nil, errors.New("text: not a 24-bit TundraDraw file")
	}

	var (
		record *sauce.Record
		bits   = 8
	)
	if record, err = sauce.ParseBytes(b); err != nil {
		if err != sauce.ErrNoRecord {
			return nil, err
		}
	}
	if record == nil {
		record = new(sauce.Record)
	}
	if record.Flags == nil {
		record.Flags = &sauce.ANSiFlags{}
	} else if record.Flags.LetterSpacing == sauce.LetterSpacing9Pixel {
		bits = 9
	}
	// It appears PabloDraw happily ignores all flags :-|
	record.Flags.NonBlink = true
	//record.Flags.AspectRatio = sauce.AspectRatioLegacy
	//record.Flags.LetterSpacing = sauce.LetterSpacing9Pixel

	// Remove header and SAUCE header (if present).
	b = b[len(tundraDrawID):]
	if i := bytes.Index(b, []byte{0x1a, 'S', 'A', 'U', 'C', 'E'}); i > -1 {
		b = b[:i]
	}

	tnd := new(TundraDraw)
	tnd.Record = record
	if tnd.Font, err = sauce.Font("IBM VGA"); err != nil {
		return nil, err
	}

	width := uint(80)
	if record.DataType == sauce.Character && record.FileType == sauce.TundraDraw && record.TypeInfo[0] > 0 {
		width = uint(record.TypeInfo[0])
	}
	tnd.Text = vga.NewText(width, 25)
	tnd.AutoExpand = true
	tnd.Palette = make(color.Palette, len(palette))
	copy(tnd.Palette, palette)

	if bits > 8 {
		tnd.Font = chargen.New(chargen.AddColumn(tnd.Font.Mask))
	}

	if err = tnd.decode(b); err != nil {
		return nil, err
	}

	return tnd, nil
}

func (tnd *TundraDraw) decode(b []byte) (err error) {
	var (
		op, ch uint8
		c      color.Color
	)
	for len(b) > 0 {
		op, b = b[0], b[1:]
		switch op {
		case ansi.SUB:
			return

		case tundraDrawPos:
			// Read the position
			var x, y uint
			if x, y, b, err = tnd.decodePosition(b); err != nil {
				return
			}
			tnd.Goto(x, y)

		case tundraDrawForeground, tundraDrawBackground, tundraDrawColors:
			if len(b) == 0 {
				return io.EOF
			}
			ch, b = b[0], b[1:]
			if op&tundraDrawForeground != 0 {
				if c, b, err = tnd.decodeColor(b); err != nil {
					return err
				}
				tnd.SetForegroundColor(c)
			}
			if op&tundraDrawBackground != 0 {
				if c, b, err = tnd.decodeColor(b); err != nil {
					return err
				}
				tnd.SetBackgroundColor(c)
			}
			tnd.WriteCodePoint(uint16(ch))

		default:
			tnd.WriteCodePoint(uint16(op))
		}
	}

	return
}

func (tnd *TundraDraw) decodeColor(b []byte) (c color.Color, remain []byte, err error) {
	if len(b) < 4 {
		return nil, nil, io.EOF
	}
	c = &color.RGBA{
		A: b[0],
		R: b[1],
		G: b[2],
		B: b[3],
	}
	remain = b[4:]
	return
}

func (tnd *TundraDraw) decodePosition(b []byte) (x, y uint, remain []byte, err error) {
	if len(b) < 8 {
		return 0, 0, nil, io.EOF
	}
	y = uint(binary.BigEndian.Uint32(b[:4]))
	x = uint(binary.BigEndian.Uint32(b[4:]))
	remain = b[8:]
	return
}

var palette = color.Palette{
	color.RGBA{0, 0, 0, 255},
	color.RGBA{173, 0, 0, 255},
	color.RGBA{0, 170, 0, 255},
	color.RGBA{173, 85, 0, 255},
	color.RGBA{0, 0, 173, 255},
	color.RGBA{173, 0, 173, 255},
	color.RGBA{0, 170, 173, 255},
	color.RGBA{173, 170, 173, 255},
	color.RGBA{82, 85, 82, 255},
	color.RGBA{255, 82, 85, 255},
	color.RGBA{82, 255, 82, 255},
	color.RGBA{255, 255, 82, 255},
	color.RGBA{82, 85, 255, 255},
	color.RGBA{255, 85, 255, 255},
	color.RGBA{82, 255, 255, 255},
	color.RGBA{255, 255, 255, 255},
}
