package vga

import (
	"fmt"
	"image/color"
)

// RGB is a 24-bit RGB triplet.
type RGB uint32

// NewRGB returns a 24-bit RGB triplet.
func NewRGB(r, g, b uint8) RGB {
	return RGB(uint32(r)<<16 | uint32(g)<<8 | uint32(b))
}

// ToRGB converts a color to RGB.
func ToRGB(c color.Color) RGB {
	if c, ok := c.(RGB); ok {
		return c
	}

	r, g, b, _ := c.RGBA()
	return NewRGB(uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

// RGBA returns the 16-bit RGBA values for the color.
func (rgb RGB) RGBA() (r, g, b, a uint32) {
	r = uint32(rgb&0xff0000) >> 16
	r |= r << 8
	g = uint32(rgb&0x00ff00) >> 8
	g |= g << 8
	b = uint32(rgb & 0x0000ff)
	b |= b << 8
	a = 0xffff
	return
}

func (rgb RGB) String() string {
	return fmt.Sprintf("#%02x%02x%02x",
		uint8((rgb&0xff0000)>>16),
		uint8((rgb&0x00ff00)>>8),
		uint8((rgb&0x0000ff)>>0))
}

// Color names for the first 16 colors.
var (
	Black         = RGB(0x000000)
	Red           = RGB(0xaa0000)
	Green         = RGB(0x00aa00)
	Yellow        = RGB(0xaa5500)
	Blue          = RGB(0x0000aa)
	Magenta       = RGB(0xaa00aa)
	Cyan          = RGB(0x00aaaa)
	White         = RGB(0xaaaaaa)
	BrightBlack   = RGB(0x555555)
	BrightRed     = RGB(0xff5555)
	BrightGreen   = RGB(0x55ff55)
	BrightYellow  = RGB(0xffff55)
	BrightBlue    = RGB(0x5555ff)
	BrightMagenta = RGB(0xff55ff)
	BrightCyan    = RGB(0x55ffff)
	BrightWhite   = RGB(0xffffff)

	// Convenience aliases
	Brown       = Yellow
	BrightBrown = BrightYellow
	Gray        = White
	Grey        = White
)

// Palette is the standard 256-color VGA palette.
var Palette = color.Palette{
	// CGA or Color Graphics Adapter palette
	Black,
	Red,
	Green,
	Yellow,
	Blue,
	Magenta,
	Cyan,
	White,
	BrightBlack,
	BrightRed,
	BrightGreen,
	BrightYellow,
	BrightBlue,
	BrightMagenta,
	BrightCyan,
	BrightWhite,
}

// ColorIndex returns the index in the palette.
func ColorIndex(c color.Color, p color.Palette) int {
	for i, o := range p {
		if colorEqual(c, o) {
			return i
		}
	}
	return -1
}

func colorEqual(a, b color.Color) bool {
	var (
		ar, ag, ab, aa = a.RGBA()
		br, bg, bb, ba = b.RGBA()
	)
	return ar == br && ag == bg && ab == bb && aa == ba
}

func init() {
	// Generate VGA palette
	for r := uint8(0); r < 6; r++ {
		for g := uint8(0); g < 6; g++ {
			for b := uint8(0); b < 6; b++ {
				Palette = append(Palette, color.RGBA{
					0x37 + r*0x28,
					0x37 + g*0x28,
					0x37 + b*0x28,
					0xff,
				})
			}
		}
	}
	for i := uint8(0); i < 24; i++ {
		v := 10*i + 8
		Palette = append(Palette, color.RGBA{v, v, v, 0xff})
	}
}
