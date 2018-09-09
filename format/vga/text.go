package vga

import (
	"image/color"
	"log"
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

	width, height       uint
	scrollRegion        [2]uint
	scrollRegionActive  bool
	cursor, savedCursor textCursor
}

// NewText allocates a new (width * height) buffer. You may also call
// new() on the struct and then call Resize.
func NewText(width, height uint) *Text {
	return &Text{
		Buffer: newCharacters(width * height),
		width:  width,
		height: height,
	}
}

// Width of the buffer.
func (text Text) Width() uint { return text.width }

// Height of the buffer.
func (text Text) Height() uint { return text.height }

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
	var (
		area = uint(len(text.Buffer))
		size = area - text.width
	)
	text.Buffer = append(
		text.Buffer[text.width:],    /* one times width removed, so first line */
		text.Buffer[:text.width]..., /* total area minus one line, so first line */
	)
	for _, char := range text.Buffer[size:] {
		char.Reset(text.cursor.ForegroundColor(), text.cursor.BackgroundColor())
	}
	if text.cursor.Y > 0 {
		text.cursor.Y--
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
	for _, char := range text.Buffer[:text.width] {
		char.Reset(text.cursor.ForegroundColor(), text.cursor.BackgroundColor())
	}
	if text.cursor.Y+2 < text.width {
		text.cursor.Y++
	} else {
		text.cursor.Y = text.width - 1
	}
}

// ClearAttribute clears the cursor attribute a.
func (text *Text) ClearAttribute(a Attribute) {
	text.cursor.ClearAttribute(a)
}

// ClearAttributes clears all cursor attributes.
func (text *Text) ClearAttributes() {
	text.cursor.ClearAttributes()
}

// SetAttribute sets the cursor attribute a.
func (text *Text) SetAttribute(a Attribute) {
	text.cursor.SetAttribute(a)
}

// SetForegroundColor sets the cursor foreground color.
func (text *Text) SetForegroundColor(c color.Color) {
	text.cursor.SetForegroundColor(c)
}

// SetBackgroundColor sets the cursor background color.
func (text *Text) SetBackgroundColor(c color.Color) {
	text.cursor.SetBackgroundColor(c)
}

// Goto moves the cursor to (x, y).
func (text *Text) Goto(x, y uint) {
	log.Printf("vga: goto(%d,%d)", x, y)
	text.cursor.X = umax(0, umin(x, text.width-1))
	text.cursor.Y = umax(0, umin(y, text.height-1))
}

// Move moves the cursor relative to (x, y).
func (text *Text) Move(x, y int) {
	log.Printf("vga: move(%d,%d)", x, y)
	if x != 0 {
		text.cursor.X = uint(max(0, min(int(text.cursor.X)+x, int(text.width)-1)))
	}
	if y != 0 {
		text.cursor.Y = uint(max(0, min(int(text.cursor.Y)+y, int(text.height)-1)))
	}
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

// WriteCharacter writes a code point byte to the screen and advance the
// cursor. If the cursor would move beyond the screen buffer and AutoExpand is
// enabled, a new row is added to the buffer; if not enabled, the buffer will
// scroll up a line before adding the code point.
func (text *Text) WriteCharacter(b byte) {
	offset := text.cursor.Offset(text.width)
	if offset >= uint(len(text.Buffer)) {
		if text.AutoExpand {
			text.Buffer = append(text.Buffer, newCharacters(text.width)...)
			text.height++
			text.WriteCharacter(b)
			return
			//} else if text.DisableScrolling {
			//	return
		}
		text.ScrollUp()
		text.WriteCharacter(b)
		return
	}

	text.Buffer[offset].SetCodePoint(b)
	text.cursor.X++
	if text.cursor.X == text.width {
		text.cursor.X = 0
		text.cursor.Y++
	}
}

// WriteString writes a string to the screen.
func (text *Text) WriteString(s string) {
	for i, l := 0, len(s); i < l; i++ {
		text.WriteCharacter(s[i])
	}
}

// String is the buffer text.
func (text *Text) String() string {
	var s = make([]byte, (text.width+1)*text.height)
	for y := uint(0); y < text.height; y++ {
		for x := uint(0); x < text.width; x++ {
			s[y*(text.width+1)+x] = byte(text.Buffer[y*text.width+x])
		}
		s[(y+1)*(text.width+1)-1] = '\n'
	}
	return string(s)
}

type textCursor struct {
	Character
	X, Y uint
}

func (cursor textCursor) Offset(width uint) uint {
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
