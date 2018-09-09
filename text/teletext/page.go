package teletext

import (
	"image"
	"image/color"
	"image/draw"
	"strings"

	"github.com/golang/freetype"
	"golang.org/x/image/math/fixed"
)

// Page constants
const (
	FirstPage = 0x1ff00
)

// PageCoding is the page coding.
type PageCoding uint8

// PageCodings
const (
	Coding7BitText PageCoding = iota
	Coding8BitData
	Coding13Triplets
	CodingHamming8_4
)

// PageFunction is the page function.
type PageFunction uint8

// PageFunctions
const (
	LOP PageFunction = iota
	DATABROADCAST
	GPOP
	POP
	GDRCS
	DRCS
	MOT
	MIP
	BTT
	AIT
	MPT
	MPTEX
)

// Page of TeleText data.
type Page struct {
	// Lines data.
	Lines Lines

	// CycleTime in seconds.
	CycleTime int // (CT) Seconds

	// CycleTimeType byte.
	CycleTimeType uint8 // (CT)

	// FastExtLinks to different pages.
	FastExtLinks [6]int // (FL)

	// Number of the page.
	Number int // (PN)

	// Language of the page, defaults to English.
	Language Language

	// Palette is used for the Image function and defaults to the TeleText Level 1
	// default palette.
	Palette color.Palette

	sub         *Page
	destination string // (DS)
	sourcePage  string // (SP)
	description string // (DE)
	subCode     uint   // (SC)
	status      int    // (PS)
	region      int    // (RE)
	coding      PageCoding
	function    PageFunction
	lastPacket  uint8
}

// NewPage returns a blank page with the first page number and the default
// palette assigned.
func NewPage() *Page {
	return &Page{
		Number:  FirstPage,
		Palette: Palette,
	}
}

const (
	imageStrideX = 6
	imageStrideY = 9
	imageScale   = 2
)

// Image of the page.
func (page *Page) Image() (image.Image, error) {
	// Load font
	fontBytes, err := FSByte(false, "/MODE7GX3.TTF")
	if err != nil {
		return nil, err
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, err
	}

	palette := make([]image.Image, len(page.Palette))
	for i, c := range page.Palette {
		palette[i] = image.NewUniform(c)
	}

	im := image.NewPaletted(image.Rect(0, 0, 40*imageStrideX*imageScale, 25*imageStrideY*imageScale), page.Palette)
	draw.Draw(im, im.Bounds(), palette[0], image.ZP, draw.Src)

	ctx := freetype.NewContext()
	ctx.SetDPI(72) //screen resolution in Dots Per Inch
	ctx.SetFont(font)
	ctx.SetFontSize(7.5 * imageScale) //font size in points
	ctx.SetClip(im.Bounds())
	ctx.SetDst(im)
	ctx.SetSrc(palette[7])

	var fg, bg int
	for row, rows := 0, len(page.Lines); row < rows; row++ {
		line := page.Lines[row]
		if line == nil {
			switch row {
			case 0:
				line = &BlankHeader
			default:
				line = &Blank
			}
		}

		if row == 0 {
			header := Line(line.Header(page))
			line = &header
		}

		var (
			graphicsMode bool
			doubleHeight bool
			skipNextRow  bool
			flashing     bool
			hold         bool
			holdChar     byte = ' '
			concealed    bool
			separated    bool
		)

		fg, bg = White, Black
		for col, c := range *line {
			var d = rune(' ')
			switch c {
			case CodeAlphaBlack,
				CodeAlphaRed,
				CodeAlphaGreen,
				CodeAlphaYellow,
				CodeAlphaBlue,
				CodeAlphaMagenta,
				CodeAlphaCyan,
				CodeAlphaWhite:
				hold = false
			case CodeFlash:
				flashing = true
			case CodeSteady:
				flashing = false
			case CodeEndBox, CodeStartBox:
			case CodeDoubleHeight, // Double height
				CodeGraphicsBlack,   // Graphics black (level 2.5+)
				CodeGraphicsRed,     // Graphics red
				CodeGraphicsGreen,   // Graphics green
				CodeGraphicsYellow,  // Graphics yellow
				CodeGraphicsBlue,    // Graphics blue
				CodeGraphicsMagenta, // Graphics magenta
				CodeGraphicsCyan,    // Graphics cyan
				CodeGraphicsWhite:   // Graphics white
			case CodeConcealDisplay: // Conceal display
				concealed = true
			case CodeContiguousGraphics: // Contiguous graphics
				separated = false
			case CodeSeparatedGraphics: // Separated gfx
				separated = true
			case CodeBlackBackground: // Background black
				bg = Black
			case CodeNewBackground: // New background
				bg = fg
			case CodeHoldGraphics: // Hold gfx (set at)
				hold = true
			case CodeReleaseGraphics: // Release gfx (set after)
			case 14, 15: // Ignore shift in/shift out and avoid them falling into default
			default:
				d = mapTextChar(page.region, page.Language, rune(c))
				c &= 0x7f

				/*
				   ch2=str[col];
				   ch2=mapTextChar(ch2);
				   // holdchar records the last mosaic character sent out
				*/
				if isMosaic(c) {
					holdChar = c // In case we encounter hold mosaics (Space doesn't count as a mosaic)
				}
			}

			if concealed {
				// Replace text with spaces
				c, d, holdChar = ' ', ' ', ' '
			}

			if graphicsMode && (hold || isMosaic(c)) { // Draw graphics. Either mosaic (but not capital A..Z) or in hold mode
				if hold {
					c = holdChar
				}
				/*
					d = 0xe200 | rune(c)
					if doubleHeight {
						if separated {
							d += 0x0100
						} else {
							d += 0x0040
						}
					} else {
						if separated {
							d += 0x00c0
						}
					}
					log.Printf("graphics mode char %#02x -> %#04x", c, d)
					ctx.DrawString(string(d), fixed.P(
						imageStrideX*imageScale*(col),
						imageStrideY*imageScale*(row+1),
					))
				*/
				if doubleHeight {
					page.imageGraphics(im, c, palette[fg], palette[bg], col, row, imageStrideX*imageScale, imageStrideY*imageScale*2, separated)
				} else {
					page.imageGraphics(im, c, palette[fg], palette[bg], col, row, imageStrideX*imageScale, imageStrideY*imageScale, separated)
				}
			} else { // Graphic block
				// Foreground color
				if !flashing {
					ctx.SetSrc(palette[fg])
				} else {
					ctx.SetSrc(palette[bg])
				}

				// Background color
				if doubleHeight {
					if d >= 0x0020 && d <= 0x00ff {
						ctx.DrawString(string(0xe000|d), fixed.P(
							imageStrideX*imageScale*(col),
							imageStrideY*imageScale*(row+1),
						))
					}
				} else {
					ctx.DrawString(string(d), fixed.P(
						imageStrideX*imageScale*(col),
						imageStrideY*imageScale*(row+1),
					))
				}
			}

			// Set-after codes implemented here
			switch (*line)[col] {
			case CodeAlphaBlack:
				fg = Black
				concealed = false // Side effect of colour. It cancels a conceal.
				graphicsMode = false
			case CodeAlphaRed:
				fg = Red
				concealed = false
				graphicsMode = false
			case CodeAlphaGreen:
				fg = Green
				concealed = false
				graphicsMode = false
			case CodeAlphaYellow:
				fg = Yellow
				concealed = false
				graphicsMode = false
			case CodeAlphaBlue:
				fg = Blue
				concealed = false
				graphicsMode = false
			case CodeAlphaMagenta:
				fg = Magenta
				concealed = false
				graphicsMode = false
			case CodeAlphaCyan:
				fg = Cyan
				concealed = false
				graphicsMode = false
			case CodeAlphaWhite:
				fg = White
				concealed = false
				graphicsMode = false
			case CodeFlash:
				flashing = true
			case CodeSteady:
			case CodeEndBox:
			case CodeStartBox:
			case CodeNormalHeight: // Normal height
				doubleHeight = false
			case CodeDoubleHeight: // Double height
				doubleHeight = true
				skipNextRow = true // ETSI: Don't use content from next row
			case CodeGraphicsBlack: // Graphics black
				fg = Black
				concealed = false
				graphicsMode = true
			case CodeGraphicsRed: // Graphics red
				fg = Red
				concealed = false
				graphicsMode = true
			case CodeGraphicsGreen: // Graphics green
				fg = Green
				concealed = false
				graphicsMode = true
			case CodeGraphicsYellow: // Graphics yellow
				fg = Yellow
				concealed = false
				graphicsMode = true
			case CodeGraphicsBlue: // Graphics blue
				fg = Blue
				concealed = false
				graphicsMode = true
			case CodeGraphicsMagenta: // Graphics magenta
				fg = Magenta
				concealed = false
				graphicsMode = true
			case CodeGraphicsCyan: // Graphics cyan
				fg = Cyan
				concealed = false
				graphicsMode = true
			case CodeGraphicsWhite: // Graphics white
				fg = White
				concealed = false
				graphicsMode = true
			case CodeConcealDisplay: // Conceal display
			case CodeContiguousGraphics: // Contiguous graphics
			case CodeSeparatedGraphics: // Separated gfx
			case CodeBlackBackground: // Background black
			case CodeNewBackground: // New background
			case CodeHoldGraphics: // Hold gfx
			case CodeReleaseGraphics: // Separated gfx
				hold = false
			default:
			}

			if skipNextRow {
				row++
			}
		}
	}

	return im, nil
}

/*

  Graphics are layed out in a 2x3 grid:

    +--+--+
    | 1| 2|
    +--+--+
    | 4| 8|
    +--+--+
    |16|64|
    +--+--+

  The base value for the codes is 160, so that they lie in the ranges 160 to
  191 and 224 to 255. For example:

    +--+--+
    |##|  |
    +--+--+
    |  |##|
    +--+--+
    |##|  |
    +--+--+

  has a code of 160 + 1 + 8 + 16 = 185

*/
func (page *Page) imageGraphics(im *image.Paletted, c byte, fg, bg image.Image, col, row, strideX, strideY int, separated bool) {
	var (
		ox = col * strideX
		oy = row * strideY
		dx = strideX / 2
		dy = strideY / 3
		sx int
		r  = image.Rect(0, 0, dx, dy)
		d  image.Image
	)
	if separated {
		if sx = dx / 3; sx == 0 {
			sx = 1
		}
		//r.Max.X -= sx
	}
	for i, j := byte(0), byte(1); i < 6; i++ {
		if c&j != 0 {
			d = fg
		} else {
			d = bg
		}

		j <<= 1
		if j == 0x20 {
			j <<= 1
		}

		draw.Draw(im, r.Add(image.Pt(ox+dx*int(i%2)-sx, oy+dy*int(i/2))), d, image.ZP, draw.Src)
	}
}

func (page Page) String() string {
	var part []string
	for _, line := range page.Lines {
		if line == nil {
			break
		}
		part = append(part, line.String())
	}
	return strings.Join(part, "\n")
}

// SetRow sets the line data for the selected row, rows above the line limit
// will be discarded.
func (page *Page) SetRow(row uint8, line []byte) {
	if row > uint8(len(page.Lines)) {
		return
	}

	if row == 28 && len(line) >= 40 {
		if dc := line[0] & 0x0f; dc == 0 || dc == 2 || dc == 3 || dc == 4 {
			// packet is X/28/0, X/28/2, X/28/3, or X/28/4
			triplet := uint32(line[1] & 0x3f)
			triplet |= uint32(line[2]&0x3f) << 6
			triplet |= uint32(line[3]&0x3f) << 12
			page.coding = PageCoding((triplet & 0x70) >> 4)
			page.function = PageFunction(uint8(triplet & 0x0f))
		}
	}

	if row == 26 && len(line) >= 40 {
		if dc := (line[0] & 0x0f) + 26; dc > page.lastPacket {
			page.lastPacket = dc
		}
	} else if row < 26 {
		if row > page.lastPacket {
			page.lastPacket = row
		}
	}

	if page.Lines[row] == nil {
		page.Lines[row] = NewLine(line)
	} else if row < 26 {
		page.Lines[row].Set(line, true)
	} else {
		// Enhanced packet
		page.Lines.Append(NewLine(line))
	}
}

// Pages slice of Page.
type Pages []*Page

// Color names
const (
	Black = iota
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

// Palette is the default TeleText Level 1 palette.
var Palette = color.Palette{
	color.RGBA{0x00, 0x00, 0x00, 0xff}, // 0b000 Black
	color.RGBA{0xff, 0x00, 0x00, 0xff}, // 0b100 Red
	color.RGBA{0x00, 0xff, 0x00, 0xff}, // 0b010 Green
	color.RGBA{0xff, 0xff, 0x00, 0xff}, // 0b110 Yellow
	color.RGBA{0x00, 0x00, 0xff, 0xff}, // 0b001 Blue
	color.RGBA{0xff, 0x00, 0xff, 0xff}, // 0b101 Magenta
	color.RGBA{0x00, 0xff, 0xff, 0xff}, // 0b011 Cyan
	color.RGBA{0xff, 0xff, 0xff, 0xff}, // 0b111 White
}

func isMosaic(c byte) bool {
	c &= 0x7f
	return (c >= 0x20 && c < 0x40) || c >= 0x60
}

func mapTextChar(region int, lang Language, r rune) rune {
	if reg, ok := regions[region]; ok {
		if c, ok := reg[lang]; ok {
			return c.Rune(r)
		}
	}
	return r
}
