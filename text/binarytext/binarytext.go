package binarytext

import (
	"errors"
	"image"
	"io"
	"io/ioutil"
	"strings"

	"github.com/textmodes/parser"
	"github.com/textmodes/parser/chargen"
	"github.com/textmodes/parser/format/sauce"
	"github.com/textmodes/parser/format/vga"
)

// Errors.
var (
	ErrSAUCEDataType = errors.New("binarytext: file has SAUCE record, but it's not BinaryText")
	ErrSAUCEFileType = errors.New("binarytext: width in SAUCE record is 0")
)

// BinaryText is a raw VGA text mode buffer.
type BinaryText struct {
	// Text buffer.
	*vga.Text

	// Record is the contained SAUCE record (may be nil).
	Record *sauce.Record

	// Font used to render the BinaryText (may be nil, defaults to IBM VGA).
	Font *chargen.Font
}

// Decode a BinaryText image.
func Decode(r io.Reader) (*BinaryText, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	record, err := sauce.ParseBytes(b)
	if err != nil {
		return nil, err
	}

	if record.DataType != sauce.BinaryText {
		return nil, ErrSAUCEDataType
	}

	font, err := sauce.Font(record.Info)
	if err != nil {
		return nil, err
	}

	if i := strings.Index(string(b), "\x1aSAUCE"); i > -1 {
		b = b[:i]
	}

	if record.FileType == 0 {
		return nil, ErrSAUCEFileType
	}

	var (
		w = uint(record.FileType) << 1
		h = uint(len(b)>>1) / w
	)

	bin := &BinaryText{
		Text:   vga.NewText(uint(record.FileType)<<1, h),
		Record: record,
		Font:   font,
	}
	bin.AutoExpand = true
	bin.DisableBlink = record.Flags.NonBlink
	if err = bin.decode(b); err != nil {
		return nil, err
	}
	return bin, nil
}

func (bin *BinaryText) decode(b []byte) (err error) {
	for i, l := 0, len(b); i < l; i += 2 {
		bin.SetBackgroundColor(vga.Palette[(b[i+1]&0xf0)>>4])
		bin.SetForegroundColor(vga.Palette[(b[i+1]&0x0f)>>0])
		bin.WriteCharacter(b[i])
	}
	return
}

// Image renders the BinaryText to an image.
func (bin *BinaryText) Image() (image.Image, error) {
	return bin.Text.Image(bin.Font, true)
}

// ImageBlink renders the BinaryText to an image; blink indicates if we're in
// blink state.
func (bin *BinaryText) ImageBlink(blink bool) (image.Image, error) {
	return bin.Text.Image(bin.Font, blink)
}

// Interface checks
var (
	_ parser.Parser = (*BinaryText)(nil)
	_ parser.Image  = (*BinaryText)(nil)
)
