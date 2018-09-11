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
	mask       Mask
	characters uint16
	size       image.Point
}

// New Font from an image, where the width of the font must be specified and
// the height of the font is the height of the image.
func New(mask Mask) *Font {
	//log.Printf("chargen.New(%#v)", mask)
	return &Font{
		mask:       mask,
		characters: mask.Characters(),
		size:       mask.CharacterSize(),
	}
}

// CharMask returns a mask and a source point for the requested char.
func (font Font) CharMask(char uint16) (Mask, image.Point) {
	if char >= font.characters {
		return nil, image.Point{}
	}
	var (
		offs = font.size.X * int(char)
		rect = image.Rect(offs, 0, offs+font.size.X, font.size.Y)
		mask = font.mask.SubMask(rect)
	)
	return mask, rect.Min
}

// Draw a character from the font onto dst with mask applied to src.
func (font Font) Draw(dst draw.Image, p image.Point, src image.Image, char uint16) {
	// fast path
	if char >= font.characters {
		return
	}
	/*
		var (
			offs = font.size.X * int(char)
			rect = image.Rect(offs, 0, offs+font.size.X, font.size.Y)
			mask = font.mask.SubMask(rect)
		)
	*/
	mask, sp := font.CharMask(char)
	draw.DrawMask(dst, dst.Bounds().Add(p), src, image.ZP, mask, sp, draw.Over)
}

// DrawString draws a string using characters from the font onto dst with mask
// applied to src. The string is interpreted as 8-bit values (bytes).
func (font Font) DrawString(dst draw.Image, p image.Point, src image.Image, s string) {
	for _, char := range []byte(s) {
		// fast path
		if uint16(char) >= font.characters {
			return
		}

		/*
			var (
				offs = font.size.X * int(char)
				rect = image.Rect(offs, 0, offs+font.size.X, font.size.Y)
				mask = font.mask.SubMask(rect)
			)
		*/
		mask, sp := font.CharMask(uint16(char))
		draw.DrawMask(dst, dst.Bounds().Add(p), src, image.ZP, mask, sp, draw.Over)
		p.X += int(font.size.X)
	}
}
