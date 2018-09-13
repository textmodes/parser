package vga

import (
	"image"
	"image/color"

	"golang.org/x/text/encoding/charmap"
)

// Text implements a virtual VGA text mode mode buffer. The buffer operations
// are not concurrency safe.
type Text struct {
	// AutoExpand will let the height grow, in stead of scrolling the
	// buffer up.
	AutoExpand bool

	// Palette, if omitted, it will default to the VGA palette.
	Palette color.Palette

	// Buffer is the underlying raw buffer.
	Buffer TextBuffer

	// Padding is the number of pixels added in between each glyph.
	Padding int

	// DisableBlink disables blinking and enabled high intensity background colors.
	DisableBlink bool

	width, height       uint
	scrollRegion        [2]uint
	scrollRegionActive  bool
	cursor, savedCursor *textCursor
	progressFunc        func(float64)
}

// NewText allocates a new (width * height) buffer. You may also call
// new() on the struct and then call Resize.
func NewText(width, height uint) *Text {
	return &Text{
		Buffer:      newCharacters(width * height),
		width:       width,
		height:      height,
		cursor:      newTextCursor(),
		savedCursor: newTextCursor(),
	}
}

// Width of the buffer.
func (text Text) Width() int { return int(text.width) }

// Height of the buffer.
func (text Text) Height() int { return int(text.height) }

// Crop the buffer, returning a new buffer.
func (text *Text) Crop(r image.Rectangle) *Text {
	visible := image.Rect(0, 0, int(text.width), int(text.height)).Intersect(r)
	if visible.Empty() {
		return &Text{
			cursor:      newTextCursor(),
			savedCursor: newTextCursor(),
		}
	}

	area := visible.Size()
	tracef("vga: crop %s -> %s area of %s", r, visible, area)
	cropped := &Text{
		Buffer:      newCharacters(uint(area.X * area.Y)),
		width:       uint(area.X),
		height:      uint(area.Y),
		cursor:      newTextCursor(),
		savedCursor: newTextCursor(),
	}

	/*
		Say, for a 10x10 text buffer:

		If the requested crop is (0, 0, 2, 10):
			* Allocate 2 * 10 = 20 characters
			* Copy src[20:40] to dst
			* Done

		If the requested crop is (2, 2, 8, 8):
			* Allocate 6 * 36 = 36 characters
			* Copy line-by-line
			* Done
	*/
	if uint(area.X) == text.width {
		// Fast path: single copy
		var (
			so = visible.Min.Y * area.X // src offset
		)
		tracef("copy %d -> %d", so, 0)
		copy(cropped.Buffer, text.Buffer[so:so+area.X*area.Y])
	} else {
		for y := visible.Min.Y; y < visible.Max.Y; y++ {
			var (
				do = (y - visible.Min.Y) * area.X
				so = y*int(text.width) + visible.Min.X
			)
			tracef("copy %d -> %d", so, do)
			copy(cropped.Buffer[do:], text.Buffer[so:so+area.X])
		}
	}
	return cropped
}

// Resize the buffer, if the buffer is growing, the canvas will be expanded to
// the right and bottom. If the buffer is shrinking, the canvas will be cropped
// on the right and bottom. This implies, that if AutoExpand is enabled, the
// expanded area may disappear.
//
// If the cursor is outside of the new buffer, it will be moved up and left
// until it is within the buffer's bounding box.
func (text *Text) Resize(width, height uint) {
	if width < 1 || height < 1 || (width == text.width && height == text.height) {
		// fast path
		return
	}

	tracef("resize to (%d, %d)", width, height)

	// resize buffer
	buffer := make(TextBuffer, width*height)

	// copy tiles to align with new dimensions
	for y := int(height - 1); y >= 0; y-- {
		for x := int(width - 1); x >= 0; x-- {
			var (
				oldOffset = uint(y)*text.width + uint(x)
				newOffset = uint(y)*width + uint(x)
			)
			if uint(y) < text.height && uint(x) < text.width {
				buffer[newOffset] = text.Buffer[oldOffset]
			} else {
				buffer[newOffset] = BlankCharacter
			}
		}
	}

	// zero old buffer
	text.Buffer = text.Buffer[:0]

	// swap buffer
	text.Buffer = buffer
	text.width, text.height = width, height

	// move the cursor (if required)
	if text.cursor.Y > height {
		text.cursor.Y = height - 1
	}
	if text.cursor.X > width {
		text.cursor.X = width - 1
	}
}

// ScrollUp removes the top line.
func (text *Text) ScrollUp() {
	tracef("scroll up")
	for i := 0; i < int(text.width); i++ {
		text.Buffer[i] = BlankCharacter
	}
	text.Buffer = append(
		text.Buffer[text.width:],    /* one times width removed, so first line */
		text.Buffer[:text.width]..., /* total area minus one line, so first line */
	)
	if text.cursor.Y > 0 {
		text.cursor.Y--
	}
}

// ScrollDown removes the bottom line.
func (text *Text) ScrollDown() {
	var (
		area = uint(len(text.Buffer))
		size = area - text.width
	)
	text.Buffer = append(
		text.Buffer[size:],
		text.Buffer[:size]...,
	)
	for i := 0; i < int(text.width); i++ {
		text.Buffer[i] = BlankCharacter
	}
	if text.cursor.Y+2 < text.width {
		text.cursor.Y++
	} else {
		text.cursor.Y = text.width - 1
	}
}

// SetScrollRegion sets the scrolling region of the buffer.
func (text *Text) SetScrollRegion(top, bot uint) {
	if top == 1 && bot >= text.height {
		// ignore & disable
		text.scrollRegionActive = false
		return
	}

	if top == 0 && bot == 0 {
		// disable
		text.scrollRegionActive = false
		return
	}

	if bot > text.height {
		bot = text.height
	}

	text.scrollRegion[0] = top
	text.scrollRegion[1] = bot
	text.scrollRegionActive = true
}

// ClearAttribute clears the cursor attribute a.
func (text *Text) ClearAttribute(a Attribute) {
	text.cursor.ClearAttribute(a)
	tracef("attr to %s", text.cursor.Attributes())
}

// ClearAttributes clears all cursor attributes.
func (text *Text) ClearAttributes() {
	text.cursor.ClearAttributes()
	tracef("attr to %s", text.cursor.Attributes())
}

// ResetAttributes resets all cursor attributes and colors.
func (text *Text) ResetAttributes() {
	text.cursor.Character = BlankCharacter
	tracef("attr to %s", text.cursor.Attributes())
}

// SetAttribute sets the cursor attribute a.
func (text *Text) SetAttribute(a Attribute) {
	text.cursor.SetAttribute(a)
	tracef("attr to %s", text.cursor.Attributes())
}

// SetForegroundColor sets the cursor foreground color.
func (text *Text) SetForegroundColor(c color.Color) {
	tracef("fg to %#+v", c)
	text.cursor.SetForegroundColor(c)
}

// SetBackgroundColor sets the cursor background color.
func (text *Text) SetBackgroundColor(c color.Color) {
	tracef("bg to %#+v", c)
	text.cursor.SetBackgroundColor(c)
}

// Goto moves the cursor to (x, y).
func (text *Text) Goto(x, y uint) {
	var ox, oy = text.cursor.X, text.cursor.Y
	text.cursor.X = umax(0, umin(x, text.width-1))
	if y >= text.height-1 && text.AutoExpand {
		// WriteCharacter will take care of expanding once a char is written at the
		// new location, it may be that only the cursor moved here.
		text.cursor.Y = y
	} else {
		text.cursor.Y = umax(0, umin(y, text.height-1))
	}
	tracef("goto(%d, %d): (%d, %d) -> (%d, %d)", x, y, ox, oy, text.cursor.X, text.cursor.Y)
}

// Move moves the cursor relative to (x, y).
func (text *Text) Move(x, y int) {
	var ox, oy = text.cursor.X, text.cursor.Y
	if x != 0 {
		text.cursor.X = uint(max(0, min(int(text.cursor.X)+x, int(text.width)-1)))
	}
	if y != 0 {
		text.cursor.Y = uint(max(0, min(int(text.cursor.Y)+y, int(text.height)-1)))
	}
	tracef("move(%d, %d): (%d, %d) -> (%d, %d)", x, y, ox, oy, text.cursor.X, text.cursor.Y)
}

// Position of the cursor.
func (text *Text) Position() (x, y uint) {
	return text.cursor.X, text.cursor.Y
}

// SaveCursor saves the cursor position.
func (text *Text) SaveCursor() {
	text.savedCursor.X, text.savedCursor.Y = text.cursor.X, text.cursor.Y
}

// LoadCursor restores a previously saved cursor position.
func (text *Text) LoadCursor() {
	text.cursor.X, text.cursor.Y = text.savedCursor.X, text.savedCursor.Y
}

// WriteCodePoint writes a code point byte to the screen and advance the
// cursor. If the cursor would move beyond the screen buffer and AutoExpand is
// enabled, a new row is added to the buffer; if not enabled, the buffer will
// scroll up a line before adding the code point.
func (text *Text) WriteCodePoint(cp uint16) {
	offset := text.cursor.Offset(text.width)
	tracef("write %d/%d", offset, len(text.Buffer))
	if offset >= uint(len(text.Buffer)) {
		if text.AutoExpand {
			tracef("auto-expanding with %d more tiles to %d",
				text.width, uint(len(text.Buffer))+text.width)
			text.Buffer = append(text.Buffer, newCharacters(text.width)...)
			text.height++
			text.WriteCodePoint(cp)
			return
			//} else if text.DisableScrolling {
			//	return
		}
		text.ScrollUp()
		text.WriteCodePoint(cp)
		return
	}

	text.Buffer[offset] = text.cursor.Character // copy attributes
	text.Buffer[offset] &= ^Character(charMask) // clear char
	text.Buffer[offset] |= Character(cp)        // set char
	tracef("text at (%d, %d) [%d]: %q fg=%s bg=%s attr=%s",
		text.cursor.X, text.cursor.Y, offset, cp,
		text.Buffer[offset].ForegroundColor(),
		text.Buffer[offset].BackgroundColor(),
		text.Buffer[offset].Attributes())

	text.cursor.X++
	if text.cursor.X == text.width {
		text.cursor.X = 0
		text.cursor.Y++
		tracef("advanced to line %d", text.cursor.Y)
	}
}

// WriteCharacter is a helper for writing 8-bit code points.
func (text *Text) WriteCharacter(char uint8) {
	text.WriteCodePoint(uint16(char))
}

// WriteString writes a string to the screen.
func (text *Text) WriteString(s string) {
	for i, l := 0, len(s); i < l; i++ {
		text.WriteCharacter(s[i])
	}
}

// Bytes is the buffer text.
func (text *Text) Bytes() []byte {
	var (
		s = make([]byte, (text.width+1)*text.height)
		o int
	)
	for y := uint(0); y < text.height; y++ {
		for x := uint(0); x < text.width; x++ {
			s[o] = text.Buffer[y*text.width+x].CodePoint()
			o++
		}
		s[o] = '\n'
		o++
	}
	return s
}

// String is like Bytes but translated Code Page 437 to UTF-8.
func (text *Text) String() string {
	s := text.Bytes()
	d := charmap.CodePage437.NewDecoder()
	//s, _ = d.String(s)
	b, _ := d.Bytes(s)
	return string(b)
}

type textCursor struct {
	Character
	X, Y uint
}

func newTextCursor() *textCursor {
	return &textCursor{
		Character: BlankCharacter,
	}
}

func (cursor *textCursor) Offset(width uint) uint {
	return cursor.Y*width + cursor.X
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

func umin(a, b uint) uint {
	if a < b {
		return a
	}
	return b
}

func umax(a, b uint) uint {
	if a > b {
		return a
	}
	return b
}
