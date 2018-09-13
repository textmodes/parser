package ansi

import (
	"bufio"
	"bytes"
	"image/color"
	"io"

	"github.com/textmodes/parser"
	"github.com/textmodes/parser/chargen"
	"github.com/textmodes/parser/format/vga"
)

// Decoder for ANSI files.
type Decoder struct {
	*vga.Text

	// Font for Image() and Scroller()
	Font *chargen.Font

	// progressFunc will be called when generating a Scoller.
	progressFunc func(float64)
}

// NewDecoder returns a decoder with a 80x25 VGA text buffer.
func NewDecoder() *Decoder {
	return &Decoder{
		Text: vga.NewText(80, 25),
	}
}

// Decode an ANSi
func (decoder *Decoder) Decode(r io.Reader) error {
	var (
		br   = bufio.NewReader(r)
		b, n byte
		err  error
	)
	for {
		if b, err = br.ReadByte(); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		tracef("read: %q", b)
		switch b {
		case BS:
			decoder.backspace()
		case TAB:
			/*
				x, _ := decoder.Position()
				decoder.Move(int(8-(x%8)), 0)
			*/
			//decoder.Move(8, 0)
			decoder.WriteString("        ")
		case LF: // Line feed
			_, y := decoder.Position()
			decoder.Goto(0, y+1)
			//decoder.Move(0, 1)
		case VT: // Vertical tab
			debugf("not implemented: vertical tab")
		case FF: // Form feed
			debugf("not implemented: form feed")
		case CR: // Carriage return
			if n, err = br.ReadByte(); err != nil {
				if err != io.EOF {
					return err
				}
			} else if n == LF {
				_, y := decoder.Position()
				decoder.Goto(0, y+1)
			} else if err = br.UnreadByte(); err != nil {
				return err
			}
		case SUB: // Sub, end if next up is a SAUCE record
			var peek []byte
			if peek, err = br.Peek(7); err != nil {
				if err != io.EOF {
					return err
				}
			}
			if bytes.Equal(peek, []byte("SAUCE00")) {
				// SAUCE record next, done parsing
				return nil
			}
			tracef("SUB peek: %q (%d)", peek, len(peek))
			decoder.WriteCharacter(b)
		case ESC: // Escape
			if err = decoder.processEscape(br); err != nil {
				return err
			}
		default:
			tracef("char %q", b)
			decoder.WriteCharacter(b)
		}
	}
}

func (decoder *Decoder) backspace() {
	decoder.Move(-1, 0)
	decoder.WriteCharacter(' ')
	decoder.Move(-1, 0)
}

func (decoder *Decoder) tab(n int) {

}

func (decoder *Decoder) eraseLine(n int) {

}

func (decoder *Decoder) eraseScreen(n int) {

}

func (decoder *Decoder) processEscape(r *bufio.Reader) (err error) {
	var b byte
	if b, err = r.ReadByte(); err != nil {
		return
	}
	debugf("process <ESC>%c", b)
	switch b {
	case // Escaped control code
		BS,
		TAB,
		LF,
		VT,
		FF,
		CR,
		SO,
		SI,
		SUB,
		ESC:
		decoder.WriteCharacter(b)
		return
	case '7': // Save Cursor (VT100)
		decoder.SaveCursor()
	case '8': // Restore Cursor (VT100)
		decoder.LoadCursor()
	case '=', '>':
		return decoder.processPrivateMode(r)
	case '@':
		_, err = r.ReadByte()
		return
	case 'D': // Index
	case 'E': // Next Line
	case 'H': // Tab Set (HTS  is 0x88).
	case 'M': // Reverse Index (RI  is 0x8d).
	case 'N': // Single Shift Select of G2 Character Set (SS2  is 0x8e), VT220.
	case 'O': // Single Shift Select of G3 Character Set (SS3  is 0x8f), VT220.
	case 'P': // Device Control String (DCS  is 0x90).
	case 'V': // Start of Guarded Area (SPA  is 0x96).
	case 'W': // End of Guarded Area (EPA  is 0x97).
	case 'X': // Start of String (SOS  is 0x98).
	case 'Z': // Return Terminal ID (DECID is 0x9a).  Obsolete form of CSI c  (DA).
	case '#': // DEC
		return decoder.processDECSequence(r)
	case '[': // Control Sequence Introducer
		return decoder.processCSISequence(r)
	case ']': // Operating System Command
		return decoder.processOSCSequence(r)
	case '(': // Designate G0 Character Set (VT100, ISO 2022)
		if _, err = r.ReadByte(); err != nil {
			return
		}
		// decoder.SetCharacterSet(0, b)
	case
		')', // Designate G1 Character Set (ISO 2022, VT100)
		'-': // Designate G1 Character Set (VT300)
		if _, err = r.ReadByte(); err != nil {
			return
		}
		// decoder.SetCharacterSet(1, b)
	case
		'*', // Designate G2 Character Set (ISO 2022, VT220)
		'.': // Designate G2 Character Set (VT300)
		if _, err = r.ReadByte(); err != nil {
			return
		}
		// decoder.SetCharacterSet(2, b)
	case
		'+', // Designate G3 Character Set (ISO 2022, VT220)
		'/': // Designate G3 Character Set (VT300)
		if _, err = r.ReadByte(); err != nil {
			return
		}
	// decoder.SetCharacterSet(3, b)
	default:
		debugf("unknown escape: <ESC>%c", b)
	}
	return
}

func (decoder *Decoder) processPrivateMode(r *bufio.Reader) (err error) {
	return
}

func (decoder *Decoder) processDECSequence(r *bufio.Reader) (err error) {
	var b byte
	if b, err = r.ReadByte(); err != nil {
		return
	}
	_ = b
	return
}

// processCSISequence processes a Control Sequence Introducer (CSI) escape sequence.
func (decoder *Decoder) processCSISequence(r *bufio.Reader) (err error) {
	var b, p byte
	if b, err = r.ReadByte(); err != nil {
		return
	}
	tracef("read: %q", b)
	if b >= '<' && b <= '?' {
		p = b
		if b, err = r.ReadByte(); err != nil {
			return
		}
		tracef("read: %q", b)
	}

	// Read numeric sequence
	var (
		args = make([]int, 0, 32)
		n    int
	)
	for b >= ' ' && b < '@' {
		n = 0
		for isdigit(b) {
			n = n*10 + int(b-'0')
			if b, err = r.ReadByte(); err != nil {
				return
			}
			tracef("read: %q", b)
		}
		if len(args) < cap(args) {
			args = append(args, n)
		}

		debugf("process CSR %v %c", args, b)
		switch {
		case b == '\b':
			decoder.backspace()
		case b == ESC:
			return decoder.processEscape(r)
		case b < ' ':
			decoder.WriteCharacter(b)
			return
		case b < '@':
			if b, err = r.ReadByte(); err != nil {
				return
			}
			tracef("read: %q", b)
		}
	}

	if b == 0x1b {
		return decoder.processEscape(r)
	} else if b < ' ' {
		return
	}

	switch b {
	case 'A', 'e': // Cursor Up
		decoder.Move(0, -defaultInt(args, 1))
	case 'B': // Cursor Down
		decoder.Move(0, +defaultInt(args, 1))
	case 'C', 'a': // Cursor Right
		if len(args) < 1 || args[0] == 0 {
			decoder.Move(1, 0)
		} else {
			decoder.Move(+args[0], 0)
		}
	case 'D': // Cursor Left
		if len(args) < 1 || args[0] == 0 {
			decoder.Move(1, 0)
		} else {
			decoder.Move(-args[0], 0)
		}
	case 'E': // Cursor Next Line
		_, y := decoder.Position()
		decoder.Goto(0, y+uint((defaultInt(args, 1)-1)))
	case 'F': // Cursor Preceding Line
		_, y := decoder.Position()
		decoder.Goto(0, y-uint((defaultInt(args, 1)-1)))
	case 'G', '`': // Cursor Character Absolute  [column]
		_, y := decoder.Position()
		decoder.Goto(uint((defaultInt(args, 1) - 1)), y)
	case 'd': // Cursor Line Absolute  [row]
		x, _ := decoder.Position()
		decoder.Goto(x, uint((defaultInt(args, 1) - 1)))
	case 'H', 'f': // Cursor Position [row;column]
		switch len(args) {
		case 0:
			decoder.Goto(0, 0)
		case 1:
			decoder.Goto(uint(args[0]-1), 0)
		default:
			decoder.Goto(uint(args[0]-1), uint(args[1]-1))
		}
	case 'I': // Cursor Forward Tabulation
		decoder.tab(+defaultInt(args, 1))
	case 'Z': // Cursor Backward Tabulation
		decoder.tab(-defaultInt(args, 1))
	case 'J':
		decoder.eraseScreen(defaultInt(args, 0))
	case 'K':
		decoder.eraseLine(defaultInt(args, 0))
	case 'm':
		decoder.processSGRMode(args)
	case 't':
		decoder.processCustomMode(args)
	case 'r': // Set Scrolling Region [top;bottom] (default = full size of window)
		if p == '?' {
			if len(args) < 2 || args[0] >= args[1] {
				decoder.SetScrollRegion(0, 0)
			} else {
				decoder.SetScrollRegion(uint(args[0]), uint(args[1]))
			}
		}
	case 'g': // Tab Clear (TBC)
		switch ps := defaultInt(args, 0); ps {
		case 0: // Clear Current Column
		case 3: // Clear All
		}
	case 'W':
		switch ps := defaultInt(args, 0); ps {
		case 0: // <ESC>H
		case 2: // <ESC>[0g Clear Current Column Tabs
		case 3: // <ESC>[3g or Clear All Tabs
		}
	default:
		tracef("unknown CSI sequence ESC[...%c", b)
	}
	return
}

func (decoder *Decoder) processOSCSequence(r *bufio.Reader) (err error) {
	return
}

func (decoder *Decoder) processSGRMode(args []int) {
	for i, l := 0, len(args); i < l; i++ {
		switch args[i] {
		case 0: // reset
			decoder.ResetAttributes()
		case 1: // bold
			decoder.SetAttribute(vga.Bold)
		case 2: // faint
			decoder.SetAttribute(vga.Faint)
		case 3: // standout
			decoder.SetAttribute(vga.Standout)
		case 4: // underline
			decoder.SetAttribute(vga.Underline)
		case 5, 6: // blink
			decoder.SetAttribute(vga.Blink)
		case 7: // reverse
			decoder.SetAttribute(vga.Reverse)
		case 8: // invisible
			decoder.SetAttribute(vga.Conceal)
		case 9: // crossed out
			decoder.SetAttribute(vga.CrossedOut)
		case 22: // normal
			decoder.ClearAttributes()
		case 23: // not standout
			decoder.ClearAttribute(vga.Standout)
		case 24: // not underline
			decoder.ClearAttribute(vga.Underline)
		case 25: // not blink
			decoder.ClearAttribute(vga.Blink)
		case 27: // not reverse
			decoder.ClearAttribute(vga.Reverse)
		case 28: // reveal
			decoder.ClearAttribute(vga.Conceal)
		case 29: // not crossed out
			decoder.ClearAttribute(vga.CrossedOut)
		case 30, 31, 32, 33, 34, 35, 36, 37:
			decoder.SetForegroundColor(vga.Palette[args[i]-30])
		case 38: // color mode
			skip, c, ok := processSGRModeColor(args[i:])
			if ok {
				decoder.SetForegroundColor(c)
			}
			tracef("skip %d (%d -> %d)", skip, i, i+skip)
			i += skip
		case 39: // default foreground
			if i == 2 {
				// 256-color mode
				decoder.SetForegroundColor(vga.Palette[args[1]])
			} else {
				decoder.SetForegroundColor(vga.White)
			}
		case 40, 41, 42, 43, 44, 45, 46, 47:
			decoder.SetBackgroundColor(vga.Palette[args[i]-40])
		case 48: // color mode
			skip, c, ok := processSGRModeColor(args[i:])
			if ok {
				decoder.SetBackgroundColor(c)
			}
			i += skip
		case 49: // default background
			if i == 2 {
				// 256-color mode
				decoder.SetBackgroundColor(vga.Palette[args[1]])
			} else {
				decoder.SetBackgroundColor(vga.Black)
			}
		}
	}
}

func (decoder *Decoder) processCustomMode(args []int) {
	if len(args) == 4 {
		switch args[0] {
		case 0: // "24-bit ansi" by SyncTerm
			decoder.SetBackgroundColor(vga.NewRGB(uint8(args[1]), uint8(args[2]), uint8(args[3])))
		case 1:
			decoder.SetForegroundColor(vga.NewRGB(uint8(args[1]), uint8(args[2]), uint8(args[3])))
		}
	}
}

func processSGRModeColor(args []int) (skip int, c color.Color, ok bool) {
	tracef("process SGR mode color %v", args)
	if len(args) < 2 {
		return
	}
	switch args[0] {
	case 38, 48:
		args = args[1:]
		switch args[0] {
		case 2: // 24-bit color mode
			if len(args) < 4 {
				skip = 1 // dno
				return
			}
			args = args[1:]
			return 4, &color.RGBA{uint8(args[0]), uint8(args[1]), uint8(args[2]), 0xff}, true
		case 5: // 256 color mode
			if len(args) < 2 {
				skip = 2 // dno
				return
			}
			args = args[1:]
			if args[0] >= len(vga.Palette) {
				tracef("VGA color %d out of range", args[0])
				skip = 2
				return
			}
			tracef("VGA color %d", args[0])
			return 2, vga.Palette[args[0]], true
		}
	}
	// unknown mode
	return 1, nil, false
}

func defaultInt(v []int, d int) int {
	if len(v) > 0 && v[0] > 0 {
		return v[0]
	}
	return d
}

func defaultInts(v []int, d []int) []int {
	for i := 0; i < len(v); i++ {
		if v[i] == 0 {
			if i < len(d) {
				return append(v[:i], d[i:]...)
			}
			return v[:i]
		}
	}
	return v
}

// Interface checks
var (
	_ parser.Parser         = (*Decoder)(nil)
	_ parser.Image          = (*Decoder)(nil)
	_ parser.Animation      = (*Decoder)(nil)
	_ parser.AnimationDelay = (*Decoder)(nil)
)
