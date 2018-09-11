package teletext

import "fmt"

type saa505x struct {
	code              uint8
	heldChar          uint8
	nextCharType      uint8
	currCharType      uint8
	heldCharType      uint8
	charData          uint16
	bit               int
	crs               int
	ra                int
	bg, fg, pc, color uint8
	graphics          bool
	separated         bool
	flash             bool
	boxed             bool
	doubleHeight      bool
	doubleHeightOld   bool
	doubleHeightTop   bool
	doubleHeightBot   bool
	holdChar          bool
	holdClear         bool
	holdOff           bool
	frameCount        int
	cols, rows, size  int
	rom               []byte
}

func newSAA505x(rom []byte) *saa505x {
	dev := &saa505x{rom: rom}
	dev.reset()
	return dev
}

func (dev *saa505x) reset() {
	dev.ra = 0
	dev.doubleHeight = false
	dev.doubleHeightBot = false
}

func (dev *saa505x) process(b uint8) {
	dev.doubleHeightOld = dev.doubleHeight
	dev.pc = dev.fg
	dev.currCharType = dev.nextCharType

	if b < 0x20 {
		dev.processControl(b)
		if dev.holdChar && dev.doubleHeight == dev.doubleHeightOld {
			b = dev.heldChar
			if b >= 0x40 && b < 0x60 {
				b = 0x20
			}
			dev.currCharType = dev.heldCharType
		} else {
			b = 0x20
		}
	} else if dev.graphics {
		dev.heldChar = b
		dev.heldCharType = dev.currCharType
	}

	ra := dev.ra
	if dev.doubleHeightOld {
		ra >>= 1
		if dev.doubleHeightBot {
			ra += 10
		}
	}

	if dev.flash && dev.frameCount > 38 {
		b = 0x20
	}
	if dev.doubleHeightBot && !dev.doubleHeight {
		b = 0x20
	}

	if dev.holdOff {
		dev.holdChar = false
		dev.heldChar = 0x20
	}
	if dev.holdClear {
		dev.heldChar = 0x20
	}

	if dev.currCharType == typeAlphanumeric || !bitSet(b, 5) {
		rb := ra
		if ra&1 == 0 {
			rb--
		} else {
			rb++
		}
		dev.charData = roundCharacter(dev.romData(b, ra), dev.romData(b, rb))
	} else {
		dev.charData = dev.gfxData(b, ra, dev.currCharType == typeSeparated)
	}
}

func (dev *saa505x) processControl(b uint8) {
	dev.holdClear = false
	dev.holdOff = false

	switch b {
	case
		controlAlphaRed,
		controlAlphaGreen,
		controlAlphaYellow,
		controlAlphaBlue,
		controlAlphaMagenta,
		controlAlphaCyan,
		controlAlphaWhite:
		dev.graphics = false
		dev.holdClear = true
		dev.fg = b & 0x07
		dev.setNextCharType()

	case controlFlash:
		dev.flash = true

	case controlSteady:
		dev.flash = true

	case controlStartBox, controlEndBox:
		// TODO

	case controlNormalHeight, controlDoubleHeight:
		dev.doubleHeight = (b & 1) != 0
		if dev.doubleHeight {
			dev.doubleHeightOld = true
		}

	case
		controlMosaicRed,
		controlMosaicGreen,
		controlMosaicYellow,
		controlMosaicBlue,
		controlMosaicMagenta,
		controlMosaicCyan,
		controlMosaicWhite:
		dev.graphics = true
		dev.fg = b & 0x07
		dev.setNextCharType()

	case controlConcealDisplay:
		dev.fg = dev.bg
		dev.pc = dev.bg

	case controlContiguousMosaic:
		dev.separated = false
		dev.setNextCharType()

	case controlSeparatedMosaic:
		dev.separated = true
		dev.setNextCharType()

	case controlBlackBackground:
		dev.bg = 0

	case controlNewBackground:
		dev.bg = dev.fg

	case controlHoldMosaic:
		dev.holdChar = true

	case controlReleaseMosaic:
		dev.holdChar = false
	}
}

func (dev *saa505x) setNextCharType() {
	if dev.graphics {
		if dev.separated {
			dev.nextCharType = typeSeparated
		} else {
			dev.nextCharType = typeContiguous
		}
	} else {
		dev.nextCharType = typeAlphanumeric
	}
}

func (dev *saa505x) romData(b uint8, row int) (c uint16) {
	if row < 0 || row >= 20 {
	} else {
		c = uint16(dev.rom[int(b*10)+(row>>1)])
		c = ((c & 0x01) * 0x03) + ((c & 0x02) * 0x06) + ((c & 0x04) * 0x0c) + ((c & 0x08) * 0x18) + ((c & 0x10) * 0x30)
	}
	return
}

func (dev *saa505x) gfxData(b uint8, row int, separated bool) (c uint16) {
	switch row >> 1 {
	case 0, 1:
		if b&0x01 != 0 {
			c += 0xfc0 // bit 1 top left
		}
		if b&0x02 != 0 {
			c += 0x03f // bit 2 top right
		}
		if separated {
			c &= 0x3cf
		}
	case 2:
		if separated {
			break
		}
		if b&0x01 != 0 {
			c += 0xfc0 // bit 1 top left
		}
		if b&0x02 != 0 {
			c += 0x03f // bit 2 top right
		}
	case 3, 4, 5:
		if b&0x04 != 0 {
			c += 0xfc0 // bit 3 center left
		}
		if b&0x08 != 0 {
			c += 0x03f // bit 4 center right
		}
		if separated {
			c &= 0x3cf
		}
	case 6:
		if separated {
			break
		}
		if b&0x04 != 0 {
			c += 0xfc0 // bit 3 center left
		}
		if b&0x08 != 0 {
			c += 0x03f // bit 4 center right
		}
	case 7, 8:
		if b&0x10 != 0 {
			c += 0xfc0 // bit 3 bottom left
		}
		if b&0x40 != 0 {
			c += 0x03f // bit 4 bottom right
		}
		if separated {
			c &= 0x3cf
		}
	case 9:
		if separated {
			break
		}
		if b&0x10 != 0 {
			c += 0xfc0 // bit 3 bottom left
		}
		if b&0x40 != 0 {
			c += 0x03f // bit 4 bottom right
		}
	}
	return
}

func (dev *saa505x) dataEntryWindow() {
	dev.ra = 19

	dev.frameCount++
	if dev.frameCount > 50 {
		dev.frameCount = 0
	}
}

// load output shift register enable
func (dev *saa505x) lose() {
	dev.ra = (dev.ra + 1) % 20
	if dev.ra == 0 {
		if dev.doubleHeightBot {
			dev.doubleHeightBot = false
		} else {
			dev.doubleHeightBot = dev.doubleHeightOld
		}
	}

	dev.fg = 7
	dev.bg = 0
	dev.graphics = false
	dev.separated = false
	dev.flash = false
	dev.boxed = false
	dev.holdChar = false
	dev.heldChar = 0x20
	dev.nextCharType = typeAlphanumeric
	dev.heldCharType = typeAlphanumeric
	dev.doubleHeight = false
	dev.doubleHeightOld = false
	dev.bit = 11
}

func (dev *saa505x) pixelClock() {
	if bitSet16(dev.charData, uint8(dev.bit)) {
		dev.color = dev.pc
	} else {
		dev.color = dev.bg
	}
	dev.bit--

	if dev.bit < 0 {
		dev.bit = 11
	}
}

func bitSet(b, bit uint8) bool {
	return (b>>bit)&1 != 0
}

func bitSet16(b uint16, bit uint8) bool {
	return (b>>bit)&1 != 0
}

func roundCharacter(a, b uint16) uint16 {
	return a | ((a >> 1) & b & ^(b >> 1)) | ((a << 1) & b & ^(b << 1))
}

const (
	controlAlphaBlack = iota
	controlAlphaRed
	controlAlphaGreen
	controlAlphaYellow
	controlAlphaBlue
	controlAlphaMagenta
	controlAlphaCyan
	controlAlphaWhite
	controlFlash
	controlSteady
	controlEndBox
	controlStartBox
	controlNormalHeight
	controlDoubleHeight
	controlDoubleWidth
	controlDoubleSize
	controlMosaicBlack
	controlMosaicRed
	controlMosaicGreen
	controlMosaicYellow
	controlMosaicBlue
	controlMosaicMagenta
	controlMosaicCyan
	controlMosaicWhite
	controlConcealDisplay
	controlContiguousMosaic
	controlSeparatedMosaic
	controlESC
	controlBlackBackground
	controlNewBackground
	controlHoldMosaic
	controlReleaseMosaic
)

func codeName(code uint8) string {
	switch code {
	case controlAlphaBlack:
		return "AlphaBlack"
	case controlAlphaRed:
		return "AlphaRed"
	case controlAlphaGreen:
		return "AlphaGreen"
	case controlAlphaYellow:
		return "AlphaYellow"
	case controlAlphaBlue:
		return "AlphaBlue"
	case controlAlphaMagenta:
		return "AlphaMagenta"
	case controlAlphaCyan:
		return "AlphaCyan"
	case controlAlphaWhite:
		return "AlphaWhite"
	case controlFlash:
		return "Flash"
	case controlSteady:
		return "Steady"
	case controlEndBox:
		return "EndBox"
	case controlStartBox:
		return "StartBox"
	case controlNormalHeight:
		return "NormalHeight"
	case controlDoubleHeight:
		return "DoubleHeight"
	case controlDoubleWidth:
		return "DoubleWidth"
	case controlDoubleSize:
		return "DoubleSize"
	case controlMosaicBlack:
		return "MosaicBlack"
	case controlMosaicRed:
		return "MosaicRed"
	case controlMosaicGreen:
		return "MosaicGreen"
	case controlMosaicYellow:
		return "MosaicYellow"
	case controlMosaicBlue:
		return "MosaicBlue"
	case controlMosaicMagenta:
		return "MosaicMagenta"
	case controlMosaicCyan:
		return "MosaicCyan"
	case controlMosaicWhite:
		return "MosaicWhite"
	case controlConcealDisplay:
		return "ConcealDisplay"
	case controlContiguousMosaic:
		return "ContiguousMosaic"
	case controlSeparatedMosaic:
		return "SeparatedMosaic"
	case controlESC:
		return "ESC"
	case controlBlackBackground:
		return "BlackBackground"
	case controlNewBackground:
		return "NewBackground"
	case controlHoldMosaic:
		return "HoldMosaic"
	case controlReleaseMosaic:
		return "ReleaseMosaic"
	default:
		return fmt.Sprintf("%#02x", code)
	}
}

const (
	_ = iota
	typeAlphanumeric
	typeContiguous
	typeSeparated
)

const (
	charsetG0P = iota /* primary G0 */
	charsetG0S        /* secondary G0 */
	charsetG1C        /* G1 contiguous */
	charsetG1S        /* G1 separate */
	charsetG2
	charsetG3
	charsetOffsetDRCS = 32
)
