package sauce

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Errors.
var (
	ErrShortRead = errors.New("sauce: short read")
	ErrNoRecord  = errors.New("sauce: no record")
)

// Record is a SAUCE record.
type Record struct {
	// Version is the SAUCE version. Only version 0 is known to be used.
	Version int

	// Title of the file.
	Title string

	// Author of the file.
	Author string

	// Group name,
	Group string

	// Date of publication.
	Date time.Time

	// FileSize in bytes.
	FileSize uint32

	// DataType is the type of data.
	DataType uint8

	// FileType is the type of file, based on DataType.
	FileType uint8

	// TypeInfo are file type specific numbers.
	TypeInfo [4]uint16

	// Comments.
	Comments []string

	// Flags.
	Flags *ANSiFlags

	// RawFlags is the unparsed Flags byte.
	RawFlags uint8

	// Info is file type specific.
	Info string
}

// ANSiFlags contains a parsed ANSiFlags structure.
type ANSiFlags struct {
	NonBlink      bool
	LetterSpacing LetterSpacing
	AspectRatio   AspectRatio
}

// LetterSpacing determines the spacing in between characters.
type LetterSpacing byte

// LetterSpacings.
const (
	// LetterSpacingLegacy enables legacy letter spacing
	LetterSpacingLegacy LetterSpacing = iota
	// LetterSpacing8Pixel enables 8 pixel letter spacing
	LetterSpacing8Pixel
	// LetterSpacing9Pixel enables 9 pixel letter spacing
	LetterSpacing9Pixel
	// LetterSpacingInvalid is unspecified
	LetterSpacingInvalid
)

func (ls LetterSpacing) String() string {
	switch ls {
	case LetterSpacingLegacy:
		return "legacy"
	case LetterSpacing8Pixel:
		return "8-pixel"
	case LetterSpacing9Pixel:
		return "9-pixel"
	default:
		return "invalid"
	}
}

// AspectRatio of the rendered piece.
type AspectRatio byte

// AspectRatios.
const (
	// AspectRatioLegacy enables legacy aspect ratio
	AspectRatioLegacy AspectRatio = iota
	// AspectRatioStretch enables stretching on displays with square pixels
	AspectRatioStretch
	// AspectRatioSquare enables optimization for non-square displays
	AspectRatioSquare
	// AspectRatioInvalid is unspecified
	AspectRatioInvalid
)

func (ar AspectRatio) String() string {
	switch ar {
	case AspectRatioLegacy:
		return "legacy"
	case AspectRatioStretch:
		return "stretch"
	case AspectRatioSquare:
		return "square"
	default:
		return "invalid"
	}
}

// Parse a SAUCE record from r. If r implements a io.ReadSeeker, the function
// will seek to EOF-128 and attempt to find a SAUCE record there. Otherwise,
// the function will read the file byte-by-byte until it encounters an ASCII
// SUB (0x1a) and expects to read a SAUCE record next.
func Parse(r io.Reader) (*Record, error) {
	var err error

	if s, ok := r.(io.ReadSeeker); ok {
		// ReadSeekers are a lot more easy to parse
		var size int64
		if size, err = s.Seek(0, io.SeekEnd); err != nil {
			return nil, fmt.Errorf("sauce: error while seeking to end: %v", err)
		} else if size < 128 {
			return nil, ErrShortRead
		}
		if _, err = s.Seek(size-128, io.SeekStart); err != nil {
			return nil, fmt.Errorf("sauce: error seeking to %d: %v", size-128, err)
		}

		b := make([]byte, 128)
		if _, err = s.Read(b); err != nil {
			return nil, fmt.Errorf("sauce: error while reading record: %v", err)
		}
		return ParseBytes(b)
	}

	b := bufio.NewReader(r)
	for {
		var c byte
		if c, err = b.ReadByte(); err != nil {
			return nil, fmt.Errorf("sauce: error while scanning for SUB: %v", err)
		} else if c == 0x1a {
			break
		}
	}

	var rest []byte
	if rest, err = ioutil.ReadAll(b); err != nil {
		return nil, fmt.Errorf("sauce: error while reading: %v", err)
	}
	return ParseBytes(rest)
}

// ParseBytes parses the ending (128) bytes as a SAUCE record.
func ParseBytes(b []byte) (*Record, error) {
	o := len(b)
	if o < 128 {
		return nil, ErrShortRead
	}
	o -= 128

	if !bytes.Equal(b[o:o+5], []byte("SAUCE")) {
		return nil, ErrNoRecord
	}

	record := new(Record)
	record.Version, _ = strconv.Atoi(string(b[o+5 : o+7]))
	record.Title = clean(string(b[o+7 : o+41]))
	record.Author = clean(string(b[o+41 : o+61]))
	record.Date = parseDate(string(b[o+82 : o+90]))
	record.FileSize = binary.LittleEndian.Uint32(b[o+90 : o+94])
	record.DataType = b[o+94]
	record.FileType = b[o+95]
	record.TypeInfo[0] = binary.LittleEndian.Uint16(b[o+96 : o+98])
	record.TypeInfo[1] = binary.LittleEndian.Uint16(b[o+98 : o+100])
	record.TypeInfo[2] = binary.LittleEndian.Uint16(b[o+100 : o+102])
	record.TypeInfo[3] = binary.LittleEndian.Uint16(b[o+102 : o+104])
	record.RawFlags = b[o+105]
	if hasANSiFlags(record) {
		record.Flags = &ANSiFlags{
			NonBlink:      (record.RawFlags & 0x01) == 0x01,
			LetterSpacing: LetterSpacing((record.RawFlags >> 1) & 3),
			AspectRatio:   AspectRatio((record.RawFlags >> 3) & 3),
		}
	}
	record.Info = clean(string(b[o+106:]))

	return record, nil
}

func clean(s string) string {
	return strings.TrimFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || r == 0x0000
	})
}

func parseDate(s string) time.Time {
	y, _ := strconv.Atoi(s[:4])
	m, _ := strconv.Atoi(s[4:6])
	d, _ := strconv.Atoi(s[6:8])
	if y < 100 {
		y += 1900
	}
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}

func hasANSiFlags(record *Record) bool {
	switch record.DataType {
	case Character:
		return record.FileType < 3
	case BinaryText:
		return true
	default:
		return false
	}
}
