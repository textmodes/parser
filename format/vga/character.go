package vga

import (
	"image/color"
	"strings"
)

// BlankCharacter is a space with black background and white foreground and no
// attributes set.
var BlankCharacter = MakeCharacter(' ', RGB(0xaaaaaa), RGB(0x000000))

const (
	bgMask   = 0xffffff0000000000
	fgMask   = 0x000000ffffff0000
	attrMask = 0x000000000000ff00
	charMask = 0x00000000000000ff
)

// Attribute for Character memory.
type Attribute uint8

func (attr Attribute) String() string {
	var s []string
	for _, a := range []Attribute{
		Bold,
		Faint,
		Standout,
		Underline,
		Blink,
		Reverse,
		Conceal,
	} {
		if attr&a == a {
			switch a {
			case Bold:
				s = append(s, "bold")
			case Faint:
				s = append(s, "faint")
			case Standout:
				s = append(s, "standout")
			case Underline:
				s = append(s, "underline")
			case Blink:
				s = append(s, "blink")
			case CrossedOut:
				s = append(s, "crossed out")
			case Reverse:
				s = append(s, "reverse")
			case Conceal:
				s = append(s, "conceal")
			}
		}
	}
	if len(s) == 0 {
		return "<none>"
	}
	return strings.Join(s, ",")
}

// Attributes
const (
	Bold Attribute = 1 << iota
	Faint
	Standout
	Underline
	Blink
	CrossedOut
	Reverse
	Conceal
	Invisible = Conceal
)

/*
Character is screen character; unlike 16-bit VGA screen characters we also
support 24-bit colors and extended attributes. The layout is as follows:

   76543210 76543210 76543210 76543210 76543210 76543210 76543210 76543210
	+--------+--------+--------+--------+--------+--------+--------+--------+
	|  back  |  back  |  back  |  fore  |  fore  |  fore  |  attr  |  char  |
	|  red   |  green |  blue  |  red   |  green |  blue  |CR_XUSFB|  code  |
	+--------+--------+--------+--------+--------+--------+--------+--------+

  char code: code point
	attr:      attributes
						 B = bold
						 F = faint
						 S = standout
						 U = underline
						 X = blink
						 _ = crossed out
						 R = reverse
						 C = conceal
	fore:      foreground color (24-bit RGB)
	back:      background color (24-bit RGB)
*/
type Character uint64

// MakeCharacter returns a Character with colors.
func MakeCharacter(cp uint8, fg, bg color.Color) Character {
	char := Character(cp)
	char |= color24Bit(fg) << 16
	char |= color24Bit(bg) << 40
	return char
}

// Reset the character to space with selected colors.
func (char *Character) Reset(fg, bg color.Color) {
	*char = Character(' ') | color24Bit(fg)<<16 | color24Bit(bg)<<40
}

// BackgroundColor returns the current background color.
func (char Character) BackgroundColor() color.Color {
	return RGB((char & bgMask) >> 40)
}

// SetBackgroundColor updates the background color.
func (char *Character) SetBackgroundColor(c color.Color) {
	*char &= ^Character(bgMask) // clear
	*char |= color24Bit(c) << 40
}

// ForegroundColor returns the current foreground color.
func (char Character) ForegroundColor() color.Color {
	return RGB((char & fgMask) >> 16)
}

// SetForegroundColor updates the foreground color.
func (char *Character) SetForegroundColor(c color.Color) {
	*char &= ^Character(fgMask) // clear
	*char |= color24Bit(c) << 16
}

// Attributes returns all attributes.
func (char Character) Attributes() Attribute {
	return Attribute((char & attrMask) >> 8)
}

// ClearAttributes clears all attributes.
func (char *Character) ClearAttributes() {
	*char &= ^Character(attrMask)
}

// ClearAttribute clears attribute a.
func (char *Character) ClearAttribute(a Attribute) {
	*char &= ^Character(a) << 8
}

// SetAttribute sets attribute a.
func (char *Character) SetAttribute(a Attribute) {
	*char |= Character(a) << 8
}

// SetAttributes updates all attributes with value a.
func (char *Character) SetAttributes(a Attribute) {
	*char &= ^Character(attrMask)
	*char |= Character(a) << 8
}

// CodePoint of the character.
func (char Character) CodePoint() uint8 {
	return uint8(char)
}

// SetCodePoint updates the code point of the character.
func (char *Character) SetCodePoint(v uint8) {
	*char = (*char & 0xff00) | Character(v)
}

// TextBuffer is a slice of Character.
type TextBuffer []Character

func newCharacters(size uint) TextBuffer {
	buffer := make(TextBuffer, size)
	for i := range buffer {
		buffer[i] = BlankCharacter
	}
	return buffer
}

func color24Bit(c color.Color) (o Character) {
	r, g, b, _ := c.RGBA()
	o |= (Character(r>>8) & 0xff) << 16
	o |= (Character(g>>8) & 0xff) << 8
	o |= (Character(b>>8) & 0xff)
	return
}
