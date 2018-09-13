package main

import (
	"fmt"
	"image"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/textmodes/parser"
	"github.com/textmodes/parser/format/sauce"
	"github.com/textmodes/parser/text/ansi"
	"github.com/textmodes/parser/text/binarytext"
	"github.com/textmodes/parser/text/teletext"
	"github.com/textmodes/parser/text/xbin"
)

type imagerWithBlink interface {
	Image(blink bool) (image.Image, error)
}

func ansiParser(font string) func(io.Reader) (parser.Parser, error) {
	return func(r io.Reader) (parser.Parser, error) {
		var record *sauce.Record
		if s, ok := r.(io.ReadSeeker); ok {
			var err error
			if record, err = sauce.Parse(s); err != nil && err != sauce.ErrNoRecord {
				fatalf("error scanning for SAUCE record: %v", err)
			}
			if _, err = s.Seek(0, io.SeekStart); err != nil {
				fatalf("error rewinding: %v", err)
			}
		}

		d := ansi.NewDecoder()
		d.AutoExpand = true
		d.Progress(progress)

		if record != nil {
			if record.DataType == sauce.Character && record.FileType <= sauce.ANSiMation {
				if font == "" {
					font = record.Info
				}
				if record.TypeInfo[0] > 0 {
					d.Resize(uint(record.TypeInfo[0]), uint(record.TypeInfo[1]))
				}
			}
		}

		if font == "" {
			fmt.Fprintf(os.Stderr, "%s: using default font\n", program)
		} else if record != nil && font == record.Info {
			fmt.Fprintf(os.Stderr, "%s: using font %q (from SAUCE)\n", program, font)
		} else {
			fmt.Fprintf(os.Stderr, "%s: using font %q\n", program, font)
		}

		f, err := sauce.Font(font)
		if err != nil {
			return nil, err
		}
		d.Font = f

		if err = d.Decode(r); err != nil {
			return nil, err
		}

		return d, nil
	}
}

var types = map[string]string{
	"ansi":       ".ans",
	"ascii":      ".asc",
	"xbin":       ".xb",
	"bin":        ".bin",
	"binarytext": ".bin",
	"ep1":        ".ep1",
	"tti":        ".tti",
}

func listtypes() {
	names := make([]string, 0, len(types))
	for name := range types {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Println("Supported types:")
	for _, name := range names {
		fmt.Println("  ", name)
	}
}

func parseFor(kind, name, font string) (func(io.Reader) (parser.Parser, error), error) {
	if kind == "auto" {
		switch org := strings.ToLower(filepath.Ext(name)); org {
		case ".ans", ".asc", ".diz":
			infof("parsing as ANSi")
			return ansiParser(font), nil

		case ".ep1":
			infof("parsing as TeleText (EP1)")
			return func(r io.Reader) (parser.Parser, error) { return teletext.DecodeEP1(r) }, nil

		case ".tti":
			infof("parsing as TeleText (TTI)")
			return func(r io.Reader) (parser.Parser, error) { return teletext.DecodeTTI(r) }, nil

		case ".bin":
			infof("parsing as BinaryText")
			return func(r io.Reader) (parser.Parser, error) { return binarytext.Decode(r) }, nil

		case ".xb":
			infof("parsing as XBin")
			return func(r io.Reader) (parser.Parser, error) { return xbin.Decode(r) }, nil
		}

		fmt.Fprintf(os.Stderr, "%s: no parser detected for %s; assuming it's ANSi\n",
			program, name)

		return ansiParser(font), nil
	}

	if ext, ok := types[kind]; ok {
		return parseFor("auto", ext, font)
	}

	return nil, fmt.Errorf(`unknown type %q (try "list" ?)`, kind)
}
