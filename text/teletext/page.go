package teletext

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"time"

	"github.com/textmodes/parser"
	"github.com/textmodes/parser/chargen"
	"github.com/textmodes/parser/data"
)

const DefaultHeader = "XXXXXXXXGOLANG mpp DAY dd MTH \x03 hh:nn.ss"

type Page struct {
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
	coding      Coding
	function    Function
	lastPacket  uint8
	data        [25][40]byte
	attr        [25][40]attr
}

const defaultPage = 0x1ff00

func NewPage() *Page {
	page := &Page{Number: defaultPage}
	page.Clear()
	return page
}

// Clear page.
func (page *Page) Clear() {
	for y := 0; y < len(page.data); y++ {
		for x := 0; x < len(page.data[y]); x++ {
			page.data[y][x] = 0x20
			page.attr[y][x].fg = 7
			page.attr[y][x].bg = 0
			page.attr[y][x].doubleWidth = false
			page.attr[y][x].doubleHeight = false
		}
	}
	copy(page.data[0][:], []byte(DefaultHeader))
}

// Line returns the row bytes.
func (page Page) Line(row int) [40]byte {
	if row == 0 {
		return page.Header()
	}
	return page.data[row]
}

// LineAt returns the row bytes.
func (page Page) LineAt(row int, now time.Time) [40]byte {
	if row == 0 {
		return page.HeaderAt(now)
	}
	return page.data[row]
}

// SetLine sets the row bytes.
func (page *Page) SetLine(row int, line [40]byte) {
	if row < 0 || row >= len(page.data) {
		return
	}
	copy(page.data[row][:], line[:])
}

// SetLineBytes sets the row from escaped bytes.
func (page *Page) SetLineBytes(row int, line []byte) {
	if row < 0 || row >= len(page.data) {
		return
	}
	var i, col int
	for i = 0; i < len(line) && col < 40; i++ {
		switch c := line[i] & 0x7f; c {
		case 0x1b: // Escaped
			i++
			if i >= len(line) {
				break
			}
			page.data[row][col] = line[i] & 0x3f
			col++
		case 0x00: // NULL
			page.data[row][col] = 0x80 // Black text
			col++
		default:
			page.data[row][col] = c
			col++
		}
	}
	for ; col < 40; col++ {
		page.data[row][col] = 0x20
	}
	return
}

// Header bytes.
func (page Page) Header() [40]byte {
	return page.HeaderAt(time.Now())
}

// HeaderAt are the header bytes at time now.
func (page Page) HeaderAt(now time.Time) (line [40]byte) {
	raw := make([]byte, 40)

	// Page number
	var k = page.Number / 0x100
	if k < 0x100 || k > 0x8ff {
		k = 0x100
	}

	copy(raw[:], page.data[0][:])

	replace := func(b, sub []byte, formats ...string) []byte {
		for _, format := range formats {
			if i := bytes.Index(b, []byte(format)); i != -1 {
				return append(append(b[:i], sub...), b[i+len(sub):]...)
			}
		}
		return b
	}

	// Magazine and page number
	raw = replace(raw, []byte(fmt.Sprintf("P%03x    ", k)), "XXXXXXXX")
	raw = replace(raw, []byte(fmt.Sprintf("%03x", k)), "mpp", "%%#", "%%£")

	// Day number and name
	raw = replace(raw, []byte(now.Format("02")), "dd", "%d")
	raw = replace(raw, []byte(now.Format("Mon")), "DAY", "%%a")

	// Month name and seconds
	raw = replace(raw, []byte(now.Format("Jan")), "MTH", "%%b")
	raw = replace(raw, []byte(now.Format("01")), "uu", "%m")

	// Year, hours, minutes and seconds
	raw = replace(raw, []byte(now.Format("06")), "yy", "%Y")
	raw = replace(raw, []byte(now.Format("15")), "hh", "%H")
	raw = replace(raw, []byte(now.Format("04")), "nn", "%M")
	raw = replace(raw, []byte(now.Format("05")), "ss", "%S")

	copy(line[:], raw)
	return
}

var palette = color.Palette{
	color.RGBA{0x00, 0x00, 0x00, 0xff}, // 0b000 Black
	color.RGBA{0xff, 0x00, 0x00, 0xff}, // 0b100 Red
	color.RGBA{0x00, 0xff, 0x00, 0xff}, // 0b010 Green
	color.RGBA{0xff, 0xff, 0x00, 0xff}, // 0b110 Yellow
	color.RGBA{0x00, 0x00, 0xff, 0xff}, // 0b001 Blue
	color.RGBA{0xff, 0x00, 0xff, 0xff}, // 0b101 Magenta
	color.RGBA{0x00, 0xff, 0xff, 0xff}, // 0b011 Cyan
	color.RGBA{0xff, 0xff, 0xff, 0xff}, // 0b111 White
}

func (page Page) Image() (image.Image, error) {
	return page.ImageAt(time.Now())
}

func (page Page) ImageAt(now time.Time) (image.Image, error) {
	return page.render(true, now)
}

type charType uint8

const (
	alphanumeric charType = iota
	contiguousGraphics
	separatedGraphics
)

func (page Page) render(flashing bool, now time.Time) (*image.Paletted, error) {
	var (
		im                 = image.NewPaletted(image.Rect(0, 0, 40*12, 25*20), palette)
		rom, err           = data.Bytes(fmt.Sprintf("font/chargen/saa505%d.bin", page.Language))
		doubleHeightBottom bool
		nextCharType       charType
	)
	if err != nil {
		return nil, err
	}
	draw.Draw(im, im.Bounds(), image.NewUniform(palette[0]), image.ZP, draw.Src)
	var (
		opts = chargen.MaskOptions{Size: image.Pt(8, 10)}
		mask = chargen.RoundCharacters(chargen.NewBytesMask(rom, opts))
		//mask = chargen.NewBytesMask(rom, opts)
		font = chargen.New(mask)
	)
	for row := 0; row < 25; row++ {
		var (
			separated        bool
			graphics         bool
			graphicsHold     bool
			doubleHeight     bool
			doubleHeightNext bool
			doubleWidth      bool
			doubleWidthNext  bool
			doubleWidthRight bool
			heldCharType     charType
			flash            bool
			fg               = palette[7]
			bg               = palette[0]
			graphicsLast     uint8
		)
		nextCharType = alphanumeric
		heldCharType = alphanumeric
		debugf("line %d: %q", row, page.LineAt(row, now))
		for col, code := range page.LineAt(row, now) {
			var (
				ctrl              = code & 0x7f // 7-bit
				graphicsHoldOff   bool
				graphicsHoldClear bool
				currCharType      = nextCharType
			)
			tracef("(%d, %d) code %#02x (%s, %#02x)", col, row, code, codeName(code), ctrl)
			if ctrl < 0x20 {
				switch ctrl {
				case
					controlAlphaBlack,
					controlAlphaRed,
					controlAlphaGreen,
					controlAlphaYellow,
					controlAlphaBlue,
					controlAlphaMagenta,
					controlAlphaCyan,
					controlAlphaWhite:
					fg = &color.RGBA{
						0xff * ((ctrl & 1) >> 0),
						0xff * ((ctrl & 2) >> 1),
						0xff * ((ctrl & 4) >> 2),
						0xff,
					}
					graphics = false
					graphicsHoldClear = true
					tracef("(%d, %d) alpha, fg %+v", col, row, fg)
					nextCharType = alphanumeric
				case controlFlash:
					flash = true
					tracef("(%d, %d) flash on", col, row)
				case controlSteady:
					flash = false
					tracef("(%d, %d) flash off", col, row)
				case controlEndBox, controlStartBox:
				case controlNormalHeight:
					doubleHeight = false
				case controlDoubleHeight:
					doubleHeight = true
					if !doubleHeightBottom {
						doubleHeightNext = true
					}
				case controlDoubleWidth:
					doubleWidth = true
					if !doubleWidthRight {
						doubleWidthNext = true
					}
				case controlDoubleSize:
					if col < 39 {
						doubleHeight = true
						if !doubleHeightBottom {
							doubleHeightNext = true
						}
					}
					if row < 23 {
						doubleWidth = true
						if !doubleWidthRight {
							doubleWidthNext = true
						}
					}
				case
					controlMosaicBlack,
					controlMosaicRed,
					controlMosaicGreen,
					controlMosaicYellow,
					controlMosaicBlue,
					controlMosaicMagenta,
					controlMosaicCyan,
					controlMosaicWhite:
					fg = &color.RGBA{
						0xff * ((ctrl & 1) >> 0),
						0xff * ((ctrl & 2) >> 1),
						0xff * ((ctrl & 4) >> 2),
						0xff,
					}
					graphics = true
					if separated {
						nextCharType = separatedGraphics
					} else {
						nextCharType = contiguousGraphics
					}
					tracef("(%d, %d) graphics, fg %+v", col, row, fg)
				case controlConcealDisplay:
					fg = bg
					tracef("(%d, %d) conceal, fg %+v", col, row, fg)
				case controlContiguousMosaic:
					separated = false
					nextCharType = contiguousGraphics
				case controlSeparatedMosaic:
					separated = true
					nextCharType = separatedGraphics
				case controlESC:
				case controlBlackBackground:
					bg = palette[0]
					tracef("(%d, %d) black background, bg %+v", col, row, bg)
				case controlNewBackground:
					bg = fg
					tracef("(%d, %d) new background, bg %+v", col, row, bg)
				case controlHoldMosaic:
					graphicsHold = true
				case controlReleaseMosaic:
					graphicsHoldOff = true
				}

				if graphics && graphicsHold {
					tracef("graphics held: %q -> %q", code, graphicsLast)
					code = graphicsLast
					if code >= 0x40 && code < 0x60 {
						code = 0x20
					}
					currCharType = heldCharType
				} else {
					code = 0x20
				}
			} else /* code < ' ' */ if graphics {
				graphicsLast = code
				heldCharType = currCharType
			} else if code == 0x20 {
				graphics = false
			}

			var skip bool

			if graphicsHoldOff {
				graphicsHold = false
				graphicsLast = ' '
			}
			if graphicsHoldClear {
				graphicsLast = ' '
			}

			// Only display char if we're *not* flashing
			if flash && !flashing {
				tracef("flash but not flashing: %q -> %q", code, graphicsLast)
				code = ' '
			}

			if doubleHeight {

			} else if doubleHeightBottom {
				skip = true
			}
			if doubleWidth {

			} else if doubleWidthRight {
				skip = true
			}

			if !skip {
				if graphics {
					tracef("(%d, %d) graphics %#02x", col, row, code)
					page.renderMosaic(im, font, col, row, fg, bg, code, separated, doubleHeight, doubleWidth)
				} else {
					tracef("(%d, %d) alpha %q (%d) color %v on %v", col, row, code&0x7f, code, fg, bg)
					page.renderChar(im, font, col, row, fg, bg, code, doubleHeight, doubleWidth)
				}
			} else {
				tracef("(%d, %d) skip", col, row)
			}
			doubleWidthRight = doubleWidthNext
		}
		doubleHeightBottom = doubleHeightNext
	}

	return im, nil
}

func (page Page) renderChar(im *image.Paletted, font *chargen.Font, col, row int, fg, bg color.Color, code byte, doubleHeight, doubleWidth bool) {
	var (
		char = code & 0x7f // 7-bit
		w    = 12
		h    = 20
	)
	if doubleHeight {
		h <<= 1
	}
	if doubleWidth {
		w <<= 1
	}
	var (
		x = col * w
		y = row * h
	)
	draw.Draw(im, image.Rect(x, y, x+w, y+h), image.NewUniform(bg), image.ZP, draw.Src)
	if char >= 0x20 {
		mask, sp := font.CharMask(uint16(char - 0x20))
		sp = sp.Add(image.Pt(4, 0)) // First two columns are empty in the font ROM
		draw.DrawMask(im, im.Bounds().Add(image.Pt(x, y)), image.NewUniform(fg), image.ZP, mask, sp, draw.Over)
	} else {
		debugf("can't render char %q (%#02x)", char, char)
	}
}

func (page Page) renderMosaic(im *image.Paletted, font *chargen.Font, col, row int, fg, bg color.Color, code byte, separated, doubleHeight, doubleWidth bool) {
	var (
		char = code - 0x20
		w    = 12
		h    = 20
	)
	if doubleHeight {
		h <<= 1
	}
	if doubleWidth {
		w <<= 1
	}
	draw.Draw(im, image.Rect(w*col, h*row, w*(col+1), h*(row+1)), image.NewUniform(bg), image.ZP, draw.Src)
	if code < 0x20 {
		return
	}
	var (
		b1 = (char & 0x01) == 0x01
		b2 = (char & 0x02) == 0x02
		b3 = (char & 0x04) == 0x04
		b4 = (char & 0x08) == 0x08
		b5 = (char & 0x10) == 0x10
		b6 = (char & 0x40) == 0x40
	)
	for y := 0; y < h; y++ {
		var (
			r  = y
			oy = y + h*row
		)
		if doubleHeight {
			r /= 2
		}
		for x := 0; x < w; x++ {
			var (
				c  = x
				ox = x + w*col
			)
			if doubleWidth {
				c /= 2
			}
			if (c < 6 && r < 6 && b1) ||
				(c > 5 && r < 6 && b2) ||
				(c < 6 && r > 5 && r < 14 && b3) ||
				(c > 5 && r > 5 && r < 14 && b4) ||
				(c < 6 && r > 13 && b5) ||
				(c > 5 && r > 13 && b6) {
				im.Set(ox, oy, fg)
			} else {
				im.Set(ox, oy, bg)
			}

			/*
				if (x+y)%4 == 3 {
					im.Set(ox, oy, &color.RGBA{0xff, 0x80, 0x00, 0xff})
				}
			*/
		}
	}
}

func (page Page) renderMosaicChar(im *image.RGBA, font *chargen.Font, col, row int, fg, bg color.Color, code byte, doubleHeight, doubleWidth, separated bool) {
	var (
		char = code // 5-bit
		w    = 12
		h    = 20
	)
	if doubleHeight {
		h <<= 1
	}
	if doubleWidth {
		w <<= 1
	}
	var (
		x = col * w
		y = row * h
	)
	draw.Draw(im, image.Rect(x, y, x+w, y+h), image.NewUniform(bg), image.ZP, draw.Src)
	if char >= 0x20 {
		tracef("mosaic (%d, %d) %d: %08b -> %08b", col, row, code, code, code-0x20)
		mask, sp := font.CharMask(uint16(char-0x20) & 0x3f)
		sp = sp.Add(image.Pt(4, 0)) // First two columns are empty in the font ROM
		draw.DrawMask(im, im.Bounds().Add(image.Pt(x, y)), image.NewUniform(fg), image.ZP, mask, sp, draw.Over)
	}
}

func (page Page) generateMosaic(data uint8, row int, separated, doubleHeight, doubleWidth bool) (code uint16) {
	mask := uint(1)
	if doubleHeight {
		mask++
	}
	switch row >> mask {
	case 0, 1:
		if data&0x01 == 0x01 {
			code += 0xfc0
		}
		if data&0x02 == 0x02 {
			code += 0x03f
		}
		if separated {
			code &= 0x3cf
		}
	case 2:
		if separated {
			break
		}
		if data&0x01 == 0x01 {
			code += 0xfc0
		}
		if data&0x02 == 0x02 {
			code += 0x03f
		}
	case 3, 4, 5:
		if data&0x04 == 0x04 {
			code += 0xfc0
		}
		if data&0x08 == 0x08 {
			code += 0x03f
		}
		if separated {
			code &= 0x3cf
		}
	case 6:
		if separated {
			break
		}
		if data&0x04 == 0x04 {
			code += 0xfc0
		}
		if data&0x08 == 0x08 {
			code += 0x03f
		}
	case 7, 8:
		if data&0x10 == 0x10 {
			code += 0xfc0
		}
		if data&0x40 == 0x40 {
			code += 0x03f
		}
		if separated {
			code &= 0x3cf
		}
	case 9:
		if separated {
			break
		}
		if data&0x10 == 0x10 {
			code += 0xfc0
		}
		if data&0x40 == 0x40 {
			code += 0x03f
		}
	}
	return
}

// Pages are multiple pages.
type Pages []*Page

func (pages Pages) Image() (image.Image, error) {
	if len(pages) == 0 {
		return nil, errors.New("teletext: no pages")
	}
	return pages[0].Image()
}

func (pages Pages) AnimateDelay(delay time.Duration) (*gif.GIF, error) {
	var (
		g = new(gif.GIF)
		d = int(delay / (time.Millisecond * 100))
	)
	for _, page := range pages {
		i, err := page.Image()
		if err != nil {
			return nil, err
		}
		g.Image = append(g.Image, i.(*image.Paletted))
		g.Delay = append(g.Delay, d)
	}
	return g, nil
}

// Language code.
type Language int

func (lang Language) String() string {
	switch lang {
	case English:
		return "en"
	case French:
		return "fr"
	case Swedish:
		return "se"
	case Czech:
		return "cz/si"
	case German:
		return "de"
	case Spanish:
		return "es/pt"
	case Italian:
		return "it"
	default:
		return "unknown"
	}
}

// Languages.
const (
	English Language = iota
	French
	Swedish
	Czech
	German
	Spanish
	Italian

	// Aliases
	Slovac     = Czech
	Portuguese = Spanish
)

// Coding is the page coding.
type Coding uint8

// Codings
const (
	Coding7BitText Coding = iota
	Coding8BitData
	Coding13Triplets
	CodingHamming8_4
)

// Function is the page function.
type Function uint8

// Functions
const (
	LOP Function = iota
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

// Interface checks
var (
	_ parser.Image = (*Page)(nil)
	_ parser.Image = (*Pages)(nil)
)
