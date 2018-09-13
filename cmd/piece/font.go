package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/textmodes/parser/format/sauce"
)

var fontAlias = map[string]string{
	"amiga":        "Amiga Topaz 1",
	"topaz":        "Amiga Topaz 1",
	"topaz1":       "Amiga Topaz 1",
	"topaz1+":      "Amiga Topaz 1+",
	"topaz2":       "Amiga Topaz 2",
	"topaz2+":      "Amiga Topaz 2+",
	"mosoul":       "Amiga mOsOul",
	"microknight":  "Amiga MicroKnight",
	"microknight+": "Amiga MicroKnight+",
	"potnoodle":    "Amiga P0T-NOoDLE",

	"atari":   "Atari ATASCII",
	"atascii": "Atari ATASCII",

	"8x8":   "IBM VGA50",
	"8x16":  "IBM VGA",
	"ega":   "IBM EGA",
	"ega43": "IBM EGA43",
	"vga":   "IBM VGA",
	"vga50": "IBM VGA50",
	"dos":   "IBM VGA 437",
	"msdos": "IBM VGA 437",
	"cp437": "IBM VGA 437",
	"cp737": "IBM VGA 737",
	"cp775": "IBM VGA 775",
	"cp850": "IBM VGA 850",
	"cp852": "IBM VGA 852",
	"cp855": "IBM VGA 855",
	"cp857": "IBM VGA 857",
	"cp860": "IBM VGA 860",
	"cp861": "IBM VGA 861",
	"cp862": "IBM VGA 862",
	"cp863": "IBM VGA 863",
	"cp865": "IBM VGA 865",
	"cp866": "IBM VGA 866",
	"cp869": "IBM VGA 869",
}

func listfonts() {
	names := make([]string, 0, len(sauce.Fonts))
	for name := range sauce.Fonts {
		names = append(names, name)
	}
	sort.Strings(names)

	alias := make([]string, 0, len(fontAlias))
	for name := range fontAlias {
		alias = append(alias, name)
	}
	sort.Strings(alias)

	fmt.Println("Supported fonts:")
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 20, 8, 0, '\t', 0)
	for _, name := range names {
		fmt.Fprint(w, name)
		var aliases []string
		for _, aka := range alias {
			if fontAlias[aka] == name {
				aliases = append(aliases, aka)
			}
		}
		if len(aliases) > 0 {
			fmt.Fprintf(w, "\t(aka %s)", strings.Join(aliases, ", "))
		}
		fmt.Fprintln(w, "")
	}
	w.Flush()
}
