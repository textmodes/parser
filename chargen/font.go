/*
Package chargen implements bitmapped monospaced "character generator" fonts.
*/
package chargen

import (
	"image"
	"image/draw"
)

// Font can draw characters from a Mask.
type Font struct {
	mask *Mask
}

// New Font from an image, where the width of the font must be specified and
// the height of the font is the height of the image.
func New(mask *Mask) *Font {
	return &Font{
		mask: mask,
	}
}

// Draw a character from the font onto dst with mask applied to src.
func (font Font) Draw(dst draw.Image, p image.Point, src image.Image, char uint16) {
	// fast path
	if char >= font.mask.characters {
		return
	}

	offs := font.mask.size.X * int(char)
	rect := image.Rect(offs, 0, offs+font.mask.size.X, font.mask.size.Y)
	mask := font.mask.SubMask(rect)
	draw.DrawMask(dst, dst.Bounds().Add(p), src, image.ZP, mask, rect.Min, draw.Over)
}

// DrawString draws a string using characters from the font onto dst with mask
// applied to src. The string is interpreted as 8-bit values (bytes).
func (font Font) DrawString(dst draw.Image, p image.Point, src image.Image, s string) {
	for _, char := range []byte(s) {
		// fast path
		if uint16(char) >= font.mask.characters {
			return
		}

		offs := font.mask.size.X * int(char)
		rect := image.Rect(offs, 0, offs+font.mask.size.X, font.mask.size.Y)
		mask := font.mask.SubMask(rect)
		draw.DrawMask(dst, dst.Bounds().Add(p), src, image.ZP, mask, rect.Min, draw.Over)
		p.X += int(font.mask.size.X)
	}
}
