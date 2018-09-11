package chargen

import (
	"image"
	"image/color"
)

// Bitmap mask alpha colors.
var (
	Opaque      = color.Alpha{A: 0xff}
	Transparent = color.Alpha{A: 0x00}
)

/*
Mask:

For a 4x6 font, each character takes up 24 bits: 4-bits per row, 6 rows, so
for the characters "Gopher", their visual representation may be as follows:

	_##_________#___________
	#____#__##__#____##_#_#_
	#_#_#_#_#_#_##__#_#_##__
	#_#_#_#_#_#_#_#_##__#___
	_##__#__##__#_#__##_#___
	________#_______________

Or in bits:

	011000000000100000000000 -> 0110 0000 0000 1000 0000 0000
	100001001100100001101010 -> 1000 0100 1100 1000 0110 1010
	101010101010110010101100 -> 1010 1010 1010 1100 1010 1100
	101010101010101011001000 -> 1010 1010 1010 1010 1100 1000
	011001001100101001101000 -> 0110 0100 1100 1010 0110 1000
	000000001000000000000000 -> 0000 0000 1000 0000 0000 0000

Now, for each character, the bits are aligned horizontally:

  bit 01234567 01234567 01234567
	G = 01101000 10101010 01100000 -> 0x68 0xaa 0x60
	o = 00000100 10101010 01000000 -> 0x04 0xaa 0x40
	p = 00001100 10101010 11001000 -> 0x0c 0xaa 0xc8
	h = 10001000 11001010 10100000 -> 0x88 0xca 0xa0
	e = 00000110 10101100 01100000 -> 0x06 0xac 0x60
	r = 00001010 11001000 10000000 -> 0x0a 0xc8 0x80

We can now use this bitmap as an image.Alpha image.
*/

// MaskOptions are the options for generating a Mask.
type MaskOptions struct {
	// Size of each character. If a coordinate is left empty the Mask function
	// returns nil.
	Size image.Point

	// StrideX is the amount of bits to advance for each character scan line. If
	// left empty, it defaults to the character width.
	StrideX int

	// Scale factor.
	Scale uint

	// Smoothing for scaling up.
	Smoothing bool
}

// Mask for a chargen font; each character is laid out horizontally in the
// mask image.
type Mask interface {
	image.Image

	// Characters is the number of characters in the mask.
	Characters() uint16

	// CharacterSize is the pixel size of each character.
	CharacterSize() image.Point

	// SubMask returns a mask representing the portion of the mask visible through the passed bounds.
	SubMask(image.Rectangle) Mask
}

type bitmap struct {
	opts       MaskOptions
	bounds     image.Rectangle // size of image
	characters uint16
	stride     int
	data       []byte
}

// NewMask returns a mask from image data with character dimensions as
// specified with size. The image data is converted to gray scale values; all
// values brigher than 50% will be opaque, others will be transparent.
func NewMask(im image.Image, opts MaskOptions) Mask {
	if opts.Size.X < 1 || opts.Size.Y < 1 {
		return nil
	}

	if opts.StrideX < 1 {
		opts.StrideX = opts.Size.X
	}
	var (
		r          = im.Bounds()
		characters = r.Max.X / opts.StrideX
		stride     = opts.StrideX * opts.Size.Y
		mask       = &bitmap{
			opts:       opts,
			data:       make([]byte, (characters*stride+7)>>3),
			bounds:     image.Rect(0, 0, opts.Size.X*characters, opts.Size.Y),
			characters: uint16(characters),
			stride:     stride,
		}
		bits uint
	)
	for c := 0; c < characters; c++ {
		offset := c * opts.StrideX
		for y := 0; y < opts.Size.Y; y++ {
			for x := 0; x < opts.Size.X; x++ {
				if gray := color.GrayModel.Convert(im.At(x+offset, y)).(color.Gray); gray.Y >= 0x80 {
					var (
						buf = bits / 8
						bit = bits % 8
					)
					mask.data[buf] |= 1 << uint(7-bit)
				}
				bits++
			}
		}
	}
	return mask
}

// NewBytesMask returns a mask from bitmap data with character dimensions as
// specified with size.
func NewBytesMask(data []byte, opts MaskOptions) Mask {
	if opts.Size.X < 1 || opts.Size.Y < 1 {
		return nil
	}

	if opts.StrideX < 1 {
		opts.StrideX = opts.Size.X
	}
	var (
		stride     = opts.StrideX * opts.Size.Y
		characters = (len(data) * 8) / stride
		mask       = &bitmap{
			opts:       opts,
			data:       data,
			bounds:     image.Rect(0, 0, characters*opts.Size.X, opts.Size.Y),
			characters: uint16(characters),
			stride:     stride,
		}
	)
	return mask
}

func (mask *bitmap) Characters() uint16 {
	if mask == nil {
		return 0
	}
	return mask.characters
}

func (mask *bitmap) CharacterSize() image.Point {
	if mask == nil {
		return image.Point{}
	}
	if mask.opts.Scale == 0 {
		return mask.opts.Size
	}
	return image.Pt(mask.opts.Size.X<<mask.opts.Scale, mask.opts.Size.Y<<mask.opts.Scale)
}

// At returns the alpha mask of the pixel at (x, y).
func (mask *bitmap) At(x, y int) color.Color {
	if mask == nil {
		return Transparent
	}
	/*
			Picture a 2x3 font (I know, TINY!), starting with char 'A', layed out
			in memory looks like:

				0         1         2         3         4         5         6         7
				012345678901234567890123456789012345678901234567890123456789012345678901
				AAAAAABBBBBBCCCCCCDDDDDDEEEEEEFFFFFFGGGGGGHHHHHHIIIIIIJJJJJJKKKKKKLLLLLL...
				0     1     2     3     4     5     6     7     8     9     10    11

			As image data, this font shall look like:

		     0         1         2
		     0123456789012345678901234
				0AABBCCDDEEFFGGHHIIJJKKLL...
				1AABBCCDDEEFFGGHHIIJJKKLL...
				2AABBCCDDEEFFGGHHIIJJKKLL...
				 0 1 2 3 4 5 6 7 8 9

			So, let's say we're rendering char 'B' on position (1, 1), the font mask
			drawing will be requesting pixel (3, 1), which resembles buffer offset:

				char width     = 2
				char height    = 3
				x              = 3
				x'             = x % char width = 1
				y              = 1
				bits per pixel = (char width * char height) = 6
				char           = (x / char width) = 1 -> 'B'
				start bit      = char * bits per pixel = 1 * 6 = 6
				bit position   = start bit + y * char width + x' = 6 + 1 * 2 + 1 = 9
				buffer offset  = start bit / 8 = 1
				bit            = 7 - bit position % 8 = 6

				buffer[offset] = BBBBCCCC
				                   ^ bit 6

			So, let's say we're rendering char 'H' on position (1, 1), the font mask
			drawing will be requesting pixel (15, 1), which resembles buffer offset:

		    char width     = 2
				char height    = 3
		    x              = 15
				x'             = x % char width = 1
				y              = 1
				bits per pixel = (char width * char height) = 6
				char           = (x / char width) = 7 -> 'H'
			  start bit      = char * bits per pixel = 7 * 6 = 42
				bit position   = start bit + y * char width + x' = 42 + 1 * 2 + 1 = 45
				buffer offset  = bit position / 8 = 5
				bit            = 7 - bit position % 8 = 2

				0         1         2         3         4         5         6         7
				AAAAAABBBBBBCCCCCCDDDDDDEEEEEEFFFFFFGGGGGGHHHHHHIIIIIIJJJJJJKKKKKKLLLLLL...

				buffer[offset] = GGHHHHHH
				                       ^ bit 2
	*/
	if mask.opts.Scale == 1 && mask.opts.Smoothing {
		/*

			Font smoothing goes like this, imagine a diagonal line in the original
			like such:

				+--+--+							+--+--+							+--+--+
				|  |##|             |##|  |             |##|  |
				+--+--+             +--+--+             +--+--+
			  |##|  |             |  |##|             |##|##|
				+--+--+             +--+--+             +--+--+

			Scaling this up without smoothing, would result in:

				+--+--+--+--+       +--+--+--+--+       +--+--+--+--+
				|  |  |##|##|       |##|##|  |  |       |##|##|  |  |
				+--+--+--+--+       +--+--+--+--+       +--+--+--+--+
				|  |  |##|##|       |##|##|  |  |       |##|##|  |  |
				+--+--+--+--+       +--+--+--+--+       +--+--+--+--+
				|##|##|  |  |       |  |  |##|##|       |##|##|##|##|
				+--+--+--+--+       +--+--+--+--+       +--+--+--+--+
				|##|##|  |  |       |  |  |##|##|       |##|##|##|##|
				+--+--+--+--+       +--+--+--+--+       +--+--+--+--+

			With smoothing enabled (notice the last block does *not* smooth):

				+--+--+--+--+       +--+--+--+--+       +--+--+--+--+
				|  |  |##|##|       |##|##|  |  |       |##|##|  |  |
				+--+--+--+--+       +--+--+--+--+       +--+--+--+--+
				|  |##|##|##|       |##|##|##|  |       |##|##|  |  |
				+--+--+--+--+       +--+--+--+--+       +--+--+--+--+
				|##|##|##|  |       |  |##|##|##|       |##|##|##|##|
				+--+--+--+--+       +--+--+--+--+       +--+--+--+--+
				|##|##|  |  |       |  |  |##|##|       |##|##|##|##|
				+--+--+--+--+       +--+--+--+--+       +--+--+--+--+

			Half dot is inserted before or after a whole dot in the presence of
			a diagonal in the matrix.

		*/
		var (
			/*
				+--+--+--+
				|nw|n |ne|
				+--+--+--+
				|w |xy|e |
				+--+--+--+
				|sw|s |se|
				+--+--+--+
			*/
			sx, sy = x >> 1, y >> 1
			n      = mask.bit(sx+0, sy-1)
			nw     = mask.bit(sx-1, sy-1)
			ne     = mask.bit(sx+1, sy-1)
			w      = mask.bit(sx-1, sy+0)
			xy     = mask.bit(sx+0, sy+0)
			e      = mask.bit(sx+1, sy+0)
			sw     = mask.bit(sx-1, sy+1)
			s      = mask.bit(sx+0, sy+1)
			se     = mask.bit(sx+1, sy+1)
		)
		if xy {
			return Opaque
		}
		switch x & 0x01 {
		case 0x00:
			switch y & 0x01 {
			case 0x00: // upper left
				if !nw && n && w {
					return Opaque
				}
			case 0x01: // lower left
				if !sw && s && w {
					return Opaque
				}
			}
		case 0x01:
			switch y & 0x01 {
			case 0x00: // upper right
				if !ne && n && e {
					return Opaque
				}
			case 0x01: // lower right
				if !se && s && e {
					return Opaque
				}
			}
		}
		/*
			var (
				w, e = mask.bit(sx-1, sy), mask.bit(sx+1, y)
				n, s = mask.bit(sx, sy-1), mask.bit(sx, sy+1)
			)
			if (w && e && n) || (e && n && s) || (n && s && w) || (s && w && e) {
				return Opaque
			}
		*/
		return Transparent
	} else if mask.opts.Scale > 0 {
		// Scale down coordinates according to scale rules.
		x >>= mask.opts.Scale
		y >>= mask.opts.Scale
	}
	var (
		bpp   = mask.opts.StrideX * mask.opts.Size.Y
		char  = x / mask.opts.StrideX
		start = char * bpp
		bits  = start + y*mask.opts.StrideX + (x % mask.opts.StrideX)
		buf   = bits >> 3
		bit   = uint(7 - (bits % 8))
	)
	if mask.data[buf]&(1<<bit) == 0 {
		return Transparent
	}
	return Opaque
}

func (mask *bitmap) bit(x, y int) bool {
	if x < 0 || y < 0 {
		return false
	}
	var (
		bpp   = mask.opts.StrideX * mask.opts.Size.Y
		char  = x / mask.opts.StrideX
		start = char * bpp
		bits  = start + y*mask.opts.StrideX + (x % mask.opts.StrideX)
		buf   = bits >> 3
		bit   = uint(7 - (bits % 8))
	)
	if buf >= len(mask.data) {
		return false
	}
	return mask.data[buf]&(1<<bit) != 0
}

// Bounds returns the domain for which At can return non-zero color.
func (mask *bitmap) Bounds() image.Rectangle {
	if mask == nil {
		return image.Rectangle{}
	}
	if mask.opts.Scale == 0 {
		return mask.bounds
	}
	return image.Rect(
		mask.bounds.Min.X<<mask.opts.Scale,
		mask.bounds.Min.Y<<mask.opts.Scale,
		mask.bounds.Max.X<<mask.opts.Scale,
		mask.bounds.Max.Y<<mask.opts.Scale,
	)
}

// ColorModel returns color.AlphaModel (8-bit alpha values).
func (mask bitmap) ColorModel() color.Model {
	return color.AlphaModel
}

// SubMask returns a mask representing the portion of the mask visible
// through r. The returned value shares pixels with the original image.
func (mask *bitmap) SubMask(r image.Rectangle) Mask {
	if mask == nil {
		return nil
	}
	if mask.opts.Scale > 0 {
		r.Min.X >>= mask.opts.Scale
		r.Min.Y >>= mask.opts.Scale
		r.Max.X >>= mask.opts.Scale
		r.Max.Y >>= mask.opts.Scale
	}
	r = r.Intersect(mask.bounds)
	if r.Empty() {
		return nil
	}
	return &bitmap{
		opts:       mask.opts,
		data:       mask.data,
		bounds:     r,
		characters: mask.characters,
		stride:     mask.stride,
	}
}

var _ image.Image = (*bitmap)(nil)
