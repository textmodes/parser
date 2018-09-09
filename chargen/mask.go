package chargen

import (
	"image"
	"image/color"
)

// Mask alpha colors.
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

// Mask has a bitmap pixel mask.
type Mask struct {
	data       []byte
	size       image.Point     // size of each character
	bounds     image.Rectangle // size of image
	characters uint16
	stride     int
}

// NewMask returns a mask from image data with character dimensions as
// specified with size. The image data is converted to gray scale values; all
// values brigher than 50% will be opaque, others will be transparent.
func NewMask(im image.Image, size image.Point) *Mask {
	if size.X < 1 || size.Y < 1 {
		return nil
	}
	var (
		r          = im.Bounds()
		stride     = size.X * size.Y
		characters = r.Max.X / size.X
		mask       = &Mask{
			// im:         image.NewAlpha(r),
			data:       make([]byte, (characters*stride+7)>>3),
			size:       size,
			bounds:     image.Rect(0, 0, size.X*characters, size.Y),
			characters: uint16(characters),
			stride:     stride,
		}
		bits uint
	)
	for c := 0; c < characters; c++ {
		offset := c * size.X
		for y := 0; y < size.Y; y++ {
			for x := 0; x < size.X; x++ {
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
func NewBytesMask(data []byte, size image.Point) *Mask {
	if size.X < 1 || size.Y < 1 {
		return nil
	}
	var (
		stride     = size.X * size.Y
		characters = (len(data) * 8) / stride
		mask       = &Mask{
			data:       data,
			size:       size,
			bounds:     image.Rect(0, 0, characters*size.X, size.Y),
			characters: uint16(characters),
			stride:     stride,
		}
	)
	return mask
}

// At returns the alpha mask of the pixel at (x, y).
func (mask *Mask) At(x, y int) color.Color {
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
	var (
		char   = x / mask.size.X
		stride = mask.size.X * mask.size.Y
		start  = char * stride
		bits   = start + y*mask.size.X + (x % mask.size.X)
		buf    = bits / 8
		bit    = uint(7 - (bits % 8))
	)
	if mask.data[buf]&(1<<bit) == 0 {
		return Transparent
	}
	return Opaque
}

// Bounds returns the domain for which At can return non-zero color.
func (mask *Mask) Bounds() image.Rectangle {
	return mask.bounds
}

// ColorModel returns color.AlphaModel (8-bit alpha values).
func (mask *Mask) ColorModel() color.Model {
	return color.AlphaModel
}

// SubMask returns a mask representing the portion of the mask visible
// through r. The returned value shares pixels with the original image.
func (mask *Mask) SubMask(r image.Rectangle) *Mask {
	return &Mask{
		data:   mask.data,
		size:   mask.size,
		bounds: r,
	}
}

var _ image.Image = (*Mask)(nil)
