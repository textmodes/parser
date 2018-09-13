package vga

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/textmodes/parser/chargen"
)

// Progress callback.
func (text *Text) Progress(fn func(float64)) {
	text.progressFunc = fn
}

// Image renders an RGBA image of the buffer with the specified chargen font,
// if blink is true, blinking character will be rendered and otherwise omitted.
func (text *Text) Image(regular *chargen.Font, blink bool) (*image.Paletted, error) {
	if regular == nil {
		return nil, fmt.Errorf("vga: font can't be nil")
	}

	var (
		italics = chargen.New(chargen.Italics(regular.Mask))
		size    = regular.Size
		stridex = (size.X + text.Padding)
		stridey = size.Y
		bounds  = image.Rect(0, 0, stridex*text.Width(), stridey*text.Height())
		palette color.Palette
		colors  = map[RGB]*image.Uniform{}
	)

	// Phase 1 is to scan the palette.
	if text.Palette == nil {
		palette = make(color.Palette, len(Palette))
		copy(palette, Palette)
	} else {
		palette = make(color.Palette, len(text.Palette))
		copy(palette, text.Palette)
	}
	for _, char := range text.Buffer {
		fg := char.BackgroundColor()
		if i := ColorIndex(fg, palette); i == -1 {
			palette = append(palette, fg)
		}
		bg := char.BackgroundColor()
		if i := ColorIndex(bg, palette); i == -1 {
			palette = append(palette, bg)
		}
	}

	// Create paletted image.
	im := image.NewPaletted(bounds, palette)

	// Black canvas
	colors[palette[0].(RGB)] = image.NewUniform(palette[0])
	draw.Draw(im, bounds, colors[palette[0].(RGB)], image.ZP, draw.Over)

	// log.Printf("buffer: %#+v", text.Buffer)

	if text.progressFunc != nil {
		text.progressFunc(0)
	}
	for y := 0; y < text.Height(); y++ {
		for x := 0; x < text.Width(); x++ {
			var (
				offset = y*text.Width() + x
				char   = text.Buffer[offset]
				attr   = char.Attributes()
				font   = regular
				bg     = ToRGB(char.BackgroundColor())
				fg     = ToRGB(char.ForegroundColor())
				r      = image.Rect(x*stridex, y*stridey, (x+1)*stridex, (y+1)*stridey)
				i      *image.Uniform
				ok     bool
			)

			// Parse attributes
			if attr&Conceal == Conceal {
				fg = bg

			} else { /* not Conceal */
				if attr&Reverse == Reverse {
					fg, bg = bg, fg
				}
				if attr&Bold == Bold {
					if j := ColorIndex(fg, palette); j > -1 && j < 8 {
						fg = palette[j+8].(RGB)
					}
				}
				if attr&Blink == Blink && text.DisableBlink {
					if j := ColorIndex(bg, palette); j > -1 && j < 8 {
						bg = palette[j+8].(RGB)
					}
				}
				if attr&Standout == Standout {
					font = italics
				}
			} /* Conceal */

			//fmt.Printf("vga: image (%d, %d) %q fg=%s bg=%s attr=%s\n",x, y, char.CodePoint(), fg, bg, attr)

			// Draw background rectangle
			if i, ok = colors[bg]; !ok {
				i = image.NewUniform(bg)
				colors[bg] = i
			}
			draw.Draw(im, r, i, image.ZP, draw.Src)

			// Draw character
			if i, ok = colors[fg]; !ok {
				i = image.NewUniform(fg)
				colors[fg] = i
			}
			if text.DisableBlink || attr&Blink != Blink || blink {
				font.Draw(im, r.Min, i, uint16(char&charMask))
			}

			// CrossedOut
			if attr&CrossedOut == CrossedOut {
				line := image.Rect(r.Min.X, r.Min.Y+size.Y/2, r.Max.X, r.Min.Y+size.Y/2+1)
				draw.Draw(im, line, image.NewUniform(fg), image.ZP, draw.Src)
			}
			// Underline
			if attr&Underline == Underline {
				line := image.Rect(r.Min.X, r.Max.Y-2, r.Max.X, r.Max.Y-1)
				draw.Draw(im, line, image.NewUniform(fg), image.ZP, draw.Src)
			}
		}
		if text.progressFunc != nil {
			text.progressFunc(float64(y) / float64(text.height))
		}
	}

	if text.progressFunc != nil {
		text.progressFunc(1)
	}

	return im, nil
}
