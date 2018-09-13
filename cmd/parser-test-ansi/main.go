package main

import (
	"flag"
	"fmt"
	"image/color"
	"os"
	"strconv"
	"strings"

	"github.com/textmodes/parser/format/sauce"
	"github.com/textmodes/parser/format/vga"
	"github.com/textmodes/parser/text/ansi"
)

func genHead(name string) {
	pad := make([]rune, 40-len(name)/2-3)
	for i := range pad {
		pad[i] = ' '
	}
	box := make([]rune, len(name)+4)
	for i := range box {
		box[i] = ' '
	}
	fmt.Printf("\n\r%s\x1b[0;44m%s\x1b[0;1;30m\xdc\x1b[0m\n\r", string(pad), string(box))
	fmt.Printf("%s\x1b[0;44m  \x1b[1m%s\x1b[0;44m  \x1b[0;1;30m\xdb\x1b[0m\n\r", string(pad), name)
	fmt.Printf("%s\x1b[0;44m%s\x1b[0;1;30m\xdb\x1b[0m\n\r", string(pad), string(box))
	fmt.Printf("%s \x1b[1;30m", string(pad))
	for _ = range box {
		fmt.Print("\xdf")
	}
	fmt.Print("\x1b[0m\n\r\n\r")
}

func genAttr(name string, sgr ...int) {
	seq := make([]string, len(sgr))
	for i, c := range sgr {
		seq[i] = strconv.Itoa(c)
	}
	fmt.Printf("\t%30s \xb3 \x1b[%sm%s\x1b[0m\n\r", name, strings.Join(seq, ";"), name)
}

func genColor(fg, bg int, sgr ...int) {
	seq := make([]string, 2+len(sgr))
	seq[0] = strconv.Itoa(fg + 30)
	seq[1] = strconv.Itoa(bg + 40)
	for i, c := range sgr {
		seq[i+2] = strconv.Itoa(c)
	}
	fmt.Printf("\x1b[%sm\xb0\xb1\xb2\xdb\x1b[0m", strings.Join(seq, ";"))
}

func genVGAColor(fg, bg int, sgr ...int) {
	seq := make([]string, 6+len(sgr))
	seq[0] = "38"
	seq[1] = "5"
	seq[2] = strconv.Itoa(fg)
	seq[3] = "48"
	seq[4] = "5"
	seq[5] = strconv.Itoa(bg)
	for i, c := range sgr {
		seq[i+6] = strconv.Itoa(c)
	}
	fmt.Printf("\x1b[%sm\xb0\xb1\xb2\xdb\x1b[0m", strings.Join(seq, ";"))
}

func gen24BitColor(fg, bg color.Color) {
	out := func(t int, c color.Color) {
		r, g, b, _ := c.RGBA()
		seq := []string{
			strconv.Itoa(t),
			strconv.Itoa(int(r >> 8)),
			strconv.Itoa(int(g >> 8)),
			strconv.Itoa(int(b >> 8)),
		}
		fmt.Printf("\x1b[%st", strings.Join(seq, ";"))
	}
	out(0, bg)
	out(1, fg)
	fmt.Print("\xb0\xb1\xb2\xdb\x1b[0m")
}

func genSAUCE(font string) {
	record := &sauce.Record{
		DataType: sauce.Character,
		FileType: sauce.ANSi,
		Info:     font,
	}
	record.Flags = &sauce.ANSiFlags{
		NonBlink: true,
	}
	os.Stdout.Write([]byte{0x1a})
	os.Stdout.Write(record.Bytes())
}

func main() {
	font := flag.String("font", "IBM VGA", "font name")
	flag.Parse()

	genHead("Attributes")
	genAttr("normal", 0)
	genAttr("bold", 1)
	genAttr("faint", 2)
	genAttr("italic", 3)
	genAttr("underline", 4)
	genAttr("slow blink", 5)
	genAttr("fast blink", 6)
	genAttr("reverse video", 7)
	genAttr("conceal", 8)
	genAttr("crossed out", 9)
	genAttr("primary font", 10)
	genAttr("bold italic", 1, 3)
	genAttr("bold underline", 1, 4)
	genAttr("bold slow blink", 1, 5)
	genAttr("bold fast blink", 1, 6)
	genAttr("bold reverse video", 1, 7)
	genAttr("bold conceal", 1, 8)
	genAttr("bold crossed out", 1, 9)
	genAttr("yellow", 0, 33)
	genAttr("bold yellow", 1, 33)
	genAttr("faint yellow", 2, 33)
	genAttr("italic yellow", 3, 33)
	genAttr("underline yellow", 4, 33)
	genAttr("slow blink yellow", 5, 33)
	genAttr("fast blink yellow", 6, 33)
	genAttr("reverse video yellow", 7, 33)
	genAttr("conceal yellow", 8, 33)
	genAttr("crossed out yellow", 9, 33)
	genAttr("primary font yellow", 10, 33)
	genAttr("bold italic yellow", 1, 3, 33)
	genAttr("bold underline yellow", 1, 4, 33)
	genAttr("bold slow blink yellow", 1, 5, 33)
	genAttr("bold fast blink yellow", 1, 6, 33)
	genAttr("bold reverse video yellow", 1, 7, 33)
	genAttr("bold conceal yellow", 1, 8, 33)
	genAttr("bold crossed out yellow", 1, 9, 33)

	fmt.Print("\r\n")
	genHead("Default colors")

	fmt.Print("\t  ")
	for xx := 0; xx < 8; xx++ {
		fmt.Printf(" %02x ", xx)
	}
	for xx := 0; xx < 8; xx++ {
		fmt.Printf(" %02x ", xx)
	}
	fmt.Print("\r\n")
	for bg := 0; bg < 8; bg++ {
		fmt.Printf("\t%02x ", bg)
		for fg := 0; fg < 8; fg++ {
			genColor(fg, bg)
		}
		for fg := 0; fg < 8; fg++ {
			genColor(fg, bg, 1)
		}
		fmt.Print("\r\n")
	}
	for bg := 0; bg < 8; bg++ {
		fmt.Printf("\t%02x ", bg)
		for fg := 0; fg < 8; fg++ {
			genColor(fg, bg, 5)
		}
		for fg := 0; fg < 8; fg++ {
			genColor(fg, bg, 5, 1)
		}
		fmt.Print("\r\n")
	}

	fmt.Print("\r\n")
	genHead("VGA colors")

	fmt.Print("\t   ")
	for xx := 0; xx < 16; xx++ {
		fmt.Printf("%02x  ", xx)
	}
	fmt.Print("\r\n")
	for b := 0; b < 16; b++ {
		fmt.Printf("\t%02x ", b*16)
		for a := 0; a < 16; a++ {
			genVGAColor(b*16+a, 0)
		}
		fmt.Print("\r\n")
	}
	fmt.Print("\r\n")
	fmt.Print("\t   ")
	for xx := 0; xx < 16; xx++ {
		fmt.Printf("%02x  ", xx)
	}
	fmt.Print("\r\n")
	for b := 0; b < 16; b++ {
		fmt.Printf("\t%02x ", b*16)
		for a := 0; a < 16; a++ {
			genVGAColor(0, b*16+a)
		}
		fmt.Print("\r\n")
	}

	fmt.Print("\r\n")
	genHead("24-bit colors")

	fmt.Print("\t   ")
	for xx := 0; xx < 16; xx++ {
		fmt.Printf("%02x  ", xx)
	}
	fmt.Print("\r\n")
	for b := 0; b < 16; b++ {
		fmt.Printf("\t%02x ", b*16)
		for a := 0; a < 16; a++ {
			gen24BitColor(vga.Palette[b*16+a], color.Black)
		}
		fmt.Print("\r\n")
	}
	fmt.Print("\r\n")
	fmt.Print("\t   ")
	for xx := 0; xx < 16; xx++ {
		fmt.Printf("%02x  ", xx)
	}
	fmt.Print("\r\n")
	for b := 0; b < 16; b++ {
		fmt.Printf("\t%02x ", b*16)
		for a := 0; a < 16; a++ {
			gen24BitColor(color.Black, vga.Palette[b*16+a])
		}
		fmt.Print("\r\n")
	}
	fmt.Print("\r\n")

	fmt.Print("\r\n")
	genHead("Character set")

	fmt.Print("\t      ")
	for x := 0; x < 16; x++ {
		fmt.Printf("%02x ", x)
	}
	fmt.Print("\r\n")
	for y := 0; y < 16; y++ {
		fmt.Printf("\t   %02x \x1b[1m", y<<4)
		for x := 0; x < 16; x++ {
			switch b := byte(x + y*16); b {
			case ansi.BS:
				os.Stdout.Write([]byte("BS "))
			case ansi.TAB:
				os.Stdout.Write([]byte("   "))
			case ansi.LF:
				os.Stdout.Write([]byte("LF "))
			case ansi.VT:
				os.Stdout.Write([]byte("VT "))
			case ansi.FF:
				os.Stdout.Write([]byte("FF "))
			case ansi.CR:
				os.Stdout.Write([]byte("CR "))
			case ansi.SUB:
				os.Stdout.Write([]byte("   "))
			case ansi.ESC:
				os.Stdout.Write([]byte("   "))
			default:
				os.Stdout.Write([]byte{0x20, b, 0x20})
			}
		}
		fmt.Print("\x1b[0m\r\n")
	}
	fmt.Print("\r\n")

	fmt.Print("\r\n")
	genHead("Control sequences")

	// Test 1:
	fmt.Print("Come back here \x10 \x1b[1;31mTest 1 FAIL")
	// Move 23 up, 14 left, 23 down, 8 right, tab, 7 left
	fmt.Print("\x1b[23A\x1b[14D\x1b[23B\x1b[8C\t\x1b[7D\x1b[1;32mTest 1 PASS\x1b[0m\n\r")

	// Done later
	// Test 2:
	fmt.Print("Come back here \x10 \x1b[1;31mTest 2 FAIL\x1b[0m\r\n\r")

	// Test 3:
	fmt.Print("Come back here \x10 \x1b[1;31mTest 3 FAIL\x1b[0m")
	// Move 23 up, 30 left, 23 down, 17 right
	fmt.Print("\x1b[23A\x1b[30D\x1b[23B\x1b[17C\x1b[0m\x1b[1;32mTest 3 PASS\x1b[0m\n\r")

	// Test 4:
	fmt.Print("Backspace test \x10 \x1b[1;31mTest 4 FAIL \x1b[1;33m:-)\x1b[0m\x1b[4D\b\b\b\b\b\b\b\b\b\b\b\x1b[1;32mTest 4 PASS\x1b[0m\n\r")

	// Back to our second test
	//fmt.Print("\x1b[A\x1b[8A\x1b[17C\x0bTest 2 \x1b[1;32pass!\x1b[0m\r\n\r\r\n\r\r\n\r")

	fmt.Print("\r\n")
	fmt.Println("Fin.\r")

	genSAUCE(*font)
}
