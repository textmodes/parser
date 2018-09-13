package xbin

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"strings"

	"github.com/textmodes/parser"
	"github.com/textmodes/parser/chargen"
	"github.com/textmodes/parser/format/sauce"
	"github.com/textmodes/parser/format/vga"
)

// Flag accumulator.
type Flag uint8

func (flag Flag) String() string {
	var s []string
	if flag&FlagPalette == FlagPalette {
		s = append(s, "palette")
	}
	if flag&FlagFont == FlagFont {
		s = append(s, "font")
	}
	if flag&FlagCompression == FlagCompression {
		s = append(s, "compressed")
	}
	if flag&FlagNonBlink == FlagNonBlink {
		s = append(s, "no blink")
	}
	if flag&Flag512Chars == Flag512Chars {
		s = append(s, "512 chars")
	}
	if len(s) == 0 {
		return "none"
	}
	return strings.Join(s, ",")
}

// Flags.
const (
	FlagPalette = 1 << iota
	FlagFont
	FlagCompression
	FlagNonBlink
	Flag512Chars
)

// Compression.
const (
	CompressNone = iota
	CompressChar
	CompressAttr
	CompressBoth
)

// Header for an XBin file.
type Header struct {
	ID       [4]byte
	EOFChar  byte
	Width    uint16
	Height   uint16
	FontSize uint8
	Flags    Flag
}

// XBin is an eXtended BinaryText file.
type XBin struct {
	Header
	*vga.Text

	// Font for this XBin.
	Font *chargen.Font

	// Record for this XBin.
	Record *sauce.Record
}

// Decode an XBin from reader r.
func Decode(r io.Reader) (*XBin, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	record, err := sauce.ParseBytes(b)
	if err != nil && err != sauce.ErrNoRecord {
		return nil, err
	}

	// Parse header
	xbin := &XBin{
		Record: record,
	}
	if err = binary.Read(bytes.NewBuffer(b), binary.LittleEndian, &xbin.Header); err != nil {
		return nil, err
	}
	if !bytes.Equal(xbin.Header.ID[:], []byte("XBIN")) {
		return nil, fmt.Errorf("xbin: header ID %q is invalid", xbin.Header.ID[:])
	}

	// Remove header and everything from EOF char onwards
	b = b[11:]
	if i := bytes.Index(b, []byte{xbin.Header.EOFChar, 'S', 'A', 'U', 'C', 'E'}); i > -1 {
		b = b[:i]
	}

	xbin.Text = vga.NewText(uint(xbin.Header.Width), uint(xbin.Header.Height))

	if b, err = xbin.decodePalette(b); err != nil {
		return nil, err
	}
	if b, err = xbin.decodeFont(b); err != nil {
		return nil, err
	}
	if err = xbin.decode(b); err != nil {
		return nil, err
	}

	return xbin, nil
}

func (xbin *XBin) decodePalette(b []byte) (remain []byte, err error) {
	if xbin.Header.Flags&FlagPalette == 0 {
		xbin.Text.Palette = vga.Palette
		return b, nil
	}

	if len(b) < 16*3 {
		return nil, fmt.Errorf("xbin: expected %d palette bytes, got %d", 16*3, len(b))
	}
	for i := 0; i < 16*3; i += 3 {
		xbin.Text.Palette = append(xbin.Text.Palette, vga.NewRGB(
			(b[0]&0x3f)<<2|(b[0]&0x3f)>>4,
			(b[1]&0x3f)<<2|(b[1]&0x3f)>>4,
			(b[2]&0x3f)<<2|(b[2]&0x3f)>>4,
		))
		b = b[3:]
	}

	return b, nil
}

func (xbin *XBin) decodeFont(b []byte) (remain []byte, err error) {
	if xbin.Header.Flags&FlagFont == 0 {
		xbin.Font, err = sauce.Font("IBM VGA")
		return
	}

	if xbin.Header.FontSize == 0 {
		xbin.Header.FontSize = 16
	}

	chars := 256
	if xbin.Header.Flags&Flag512Chars == Flag512Chars {
		chars <<= 1
	}

	size := int(xbin.Header.FontSize) * chars
	data := make([]byte, size)
	if len(b) < size {
		return nil, fmt.Errorf("xbin: expected %d font bytes, got %d", size, len(b))
	}
	b = b[copy(data, b):]

	xbin.Font = chargen.New(chargen.NewBytesMask(data, chargen.MaskOptions{
		Size: image.Pt(8, int(xbin.Header.FontSize)),
	}))

	return b, nil
}

func (xbin *XBin) decode(b []byte) (err error) {
	if xbin.Header.Flags&FlagCompression == FlagCompression {
		return xbin.decodeCompressed(b)
	}
	return xbin.decodeUncompressed(b)
}

func (xbin *XBin) decodeCompressed(b []byte) (err error) {
	var code, attr uint8
	for o, l := 0, len(b); o < l; {
		var (
			repeat = (b[o] & 0xc0) >> 6
			counts = int(b[o]&0x3f) + 1
		)

		o++
		switch repeat {
		case CompressNone:
			if o+(counts*2) > l {
				return io.ErrUnexpectedEOF
			}
			for ; counts > 0; counts, o = counts-1, o+2 {
				xbin.write(b[o], b[o+1])
			}

		case CompressChar:
			if o+(counts+1) > l {
				return io.ErrUnexpectedEOF
			}
			code, o = b[o], o+1
			for ; counts > 0; counts, o = counts-1, o+1 {
				xbin.write(code, b[o])
			}

		case CompressAttr:
			if o+(counts+1) > l {
				return io.ErrUnexpectedEOF
			}
			attr, o = b[o], o+1
			for ; counts > 0; counts, o = counts-1, o+1 {
				xbin.write(b[o], attr)
			}

		case CompressBoth:
			if o+2 > l {
				return io.ErrUnexpectedEOF
			}
			code, o = b[o], o+1
			attr, o = b[o], o+1
			for ; counts > 0; counts-- {
				//xbin.Text.Buffer = append(xbin.Text.Buffer, char)
				xbin.write(code, attr)
			}
		}
	}
	return
}

func (xbin *XBin) decodeUncompressed(b []byte) (err error) {
	l := xbin.Text.Height() * xbin.Text.Width() * 2
	if l > len(b) {
		return fmt.Errorf("xbin: need %d bytes, only %d bytes remaining", l, len(b))
	}

	for i := 0; i < l; i += 2 {
		xbin.write(b[i], b[i+1])
	}

	return
}

func (xbin *XBin) write(code, attr uint8) {
	xbin.Text.SetForegroundColor(xbin.Text.Palette[(attr&0x0f)>>0])
	xbin.Text.SetBackgroundColor(xbin.Text.Palette[(attr&0xf0)>>4])
	xbin.Text.WriteCharacter(code)
}

// Image renders the XBin to an image.
func (xbin *XBin) Image() (image.Image, error) {
	return xbin.Text.Image(xbin.Font, true)
}

// ImageBlink renders the XBin to an image; blink indicates if we're in blink state.
func (xbin *XBin) ImageBlink(blink bool) (image.Image, error) {
	return xbin.Text.Image(xbin.Font, blink)
}

// Interface checks.
var (
	_ parser.Parser = (*XBin)(nil)
	_ parser.Image  = (*XBin)(nil)
)
