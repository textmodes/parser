package teletext

import (
	"bytes"
	"fmt"
	"time"
)

var (
	// Blank line
	Blank = make(Line, 40)

	// BlankHeader is a blank header line
	BlankHeader = Line("XXXXXXXXGOLANG mpp DAY dd MTH \x03 hh:nn.ss")
)

// Line of Page data.
type Line []byte

// NewLine returns a new line filled with data, with at least 40 characters
// right padded with spaces.
func NewLine(data []byte) *Line {
	var line = make(Line, 40)
	for i := range line {
		line[i] = 0x20
	}
	copy(line, data)
	return &line
}

// Header parses the line as the page header.
func (line Line) Header(page *Page) string {
	return line.HeaderTime(page, time.Now())
}

// HeaderTime parses the line as the page header at time now.
func (line Line) HeaderTime(page *Page, now time.Time) string {
	if page == nil {
		page = NewPage()
		page.Number = 0x10001
	}

	var (
		size = len(line)
		raw  []byte
	)
	if size < 40 {
		size = 40
	}
	raw = make([]byte, size)
	for i := range raw {
		raw[i] = ' '
	}
	copy(raw, line)

	// First 8 characters are discarded
	for i := 0; i < 8; i++ {
		raw[i] = ' '
	}

	// Page number
	var k = page.Number / 0x100
	if k < 0x100 || k > 0x8ff {
		k = 0x100
	}
	raw[0] = 'P'
	for i, c := range fmt.Sprintf("%03x", k) {
		raw[1+i] = byte(c)
	}

	replace := func(b, sub []byte, formats ...string) []byte {
		for _, format := range formats {
			if i := bytes.Index(b, []byte(format)); i != -1 {
				return append(append(b[:i], sub...), b[i+len(sub):]...)
			}
		}
		return b
	}

	// Magazine and page number
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

	return string(raw)
}

// Set the line data, if escaped then the line will first be unescaped.
func (line *Line) Set(data []byte, escaped bool) {
	if escaped {
		line.Set(line.unescape(data), false)
		return
	}

	if l := len(data); len(*line) < l {
		*line = make(Line, l)
		for i := range *line {
			(*line)[i] = 0x20
		}
	}
	copy(*line, data)
}

func (line Line) unescape(src []byte) (dst []byte) {
	for i := 0; i < len(src); i++ {
		switch c := src[i] & 0x7f; c {
		case 0x1b: // Escaped
			i++
			if i >= len(src) {
				break
			}
			dst = append(dst, src[i]&0x3f)
			continue
		case 0x00: // NULL
			dst = append(dst, 0x80) // Black text
		default:
			dst = append(dst, c)
		}
	}
	if j := len(dst) - 1; j > 0 && dst[j] == '\n' {
		dst = dst[:j-1]
	}
	if j := len(dst) - 1; j > 0 && dst[j] == '\r' {
		dst = dst[:j-1]
	}
	return
}

// IsAlpha checks if this line is Alpha mode.
func (line Line) IsAlpha(col int) bool {
	if col > 39 {
		return true
	}
	for i := 0; i < col; i++ {
		if line[i] >= CodeAlphaBlack && line[i] <= CodeAlphaWhite {
			return true
		}
		if line[i] >= CodeGraphicsBlack && line[i] <= CodeGraphicsWhite {
			return false
		}
	}
	return true
}

// IsBlank checks if this is a blank line.
func (line Line) IsBlank() bool {
	if len(line) == 0 {
		// Fast path
		return true
	}
	for _, c := range line {
		if c != ' ' {
			return false
		}
	}
	return true
}

// IsDoubleHeight returns if this line has double height alpha.
func (line Line) IsDoubleHeight() bool {
	for _, c := range line {
		if c == '\r' || c == 0x10 {
			return true
		}
	}
	return false
}

func (line Line) String() string {
	return string(line)
}

// Lines are a full frame of (max 29) lines.
type Lines [29]*Line

// Append a line to the next empty slot, line will be discarded if all slots
// are occupied.
func (lines Lines) Append(line *Line) {
	for i, l := 0, len(lines); i < l; i++ {
		if lines[i] == nil {
			lines[i] = line
			return
		}
	}
}
