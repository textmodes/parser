package sauce

import (
	"fmt"
	"image"
	"strings"

	"github.com/textmodes/parser/chargen"
	"github.com/textmodes/parser/data"
)

// Fonts are the supported fonts (and their chargen file name).
var Fonts = map[string]FontInfo{
	// Amiga fonts copyright 2003-2009 dMG/t!s^dS!
	"Amiga MicroKnight": FontInfo{
		ROM:          "amiga_microknight.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   AmigaPixelRatio,
	},
	"Amiga MicroKnight+": FontInfo{
		ROM:          "amiga_microknight+.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   AmigaPixelRatio,
	},
	"Amiga mOsOul": FontInfo{
		ROM:          "amiga_mosoul.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   AmigaPixelRatio,
	},
	"Amiga P0T-NOoDLE": FontInfo{
		ROM:          "amiga_p0t-noodle.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   AmigaPixelRatio,
	},
	"Amiga Topaz 1": FontInfo{
		ROM:          "amiga_topaz_1.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   AmigaPixelRatio,
	},
	"Amiga Topaz 1+": FontInfo{
		ROM:          "amiga_topaz_1+.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   AmigaPixelRatio,
	},
	"Amiga Topaz 2": FontInfo{
		ROM:          "amiga_topaz_2.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   AmigaPixelRatio,
	},
	"Amiga Topaz 2+": FontInfo{
		ROM:          "amiga_topaz_2+.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   AmigaPixelRatio,
	},

	"Atari ATASCII": FontInfo{
		ROM:          "atari_atascii.bin",
		Size:         image.Pt(8, 8),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   Ratio{4, 5},
	},

	// IBM PC fonts copyright IBM corporation
	"IBM EGA": FontInfo{
		ROM:          "ibm_ega43.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   Ratio{35, 48},
	},
	"IBM EGA43": FontInfo{
		ROM:          "ibm_ega43.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   Ratio{35, 48},
	},
	"IBM VGA": FontInfo{
		ROM:          "ibm_vga.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   IBMPCPixelRatio,
	},
	"IBM VGA 437": FontInfo{
		ROM:          "ibm_vga_437.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   IBMPCPixelRatio,
	},
	"IBM VGA 737": FontInfo{
		ROM:          "ibm_vga_737.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   IBMPCPixelRatio,
	},
	"IBM VGA 775": FontInfo{
		ROM:          "ibm_vga_775.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   IBMPCPixelRatio,
	},
	"IBM VGA 850": FontInfo{
		ROM:          "ibm_vga_850.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   IBMPCPixelRatio,
	},
	"IBM VGA 852": FontInfo{
		ROM:          "ibm_vga_852.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   IBMPCPixelRatio,
	},
	"IBM VGA 855": FontInfo{
		ROM:          "ibm_vga_855.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   IBMPCPixelRatio,
	},
	"IBM VGA 857": FontInfo{
		ROM:          "ibm_vga_857.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   IBMPCPixelRatio,
	},
	"IBM VGA 860": FontInfo{
		ROM:          "ibm_vga_860.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   IBMPCPixelRatio,
	},
	"IBM VGA 861": FontInfo{
		ROM:          "ibm_vga_861.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   IBMPCPixelRatio,
	},
	"IBM VGA 862": FontInfo{
		ROM:          "ibm_vga_862.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   IBMPCPixelRatio,
	},
	"IBM VGA 863": FontInfo{
		ROM:          "ibm_vga_863.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   IBMPCPixelRatio,
	},
	"IBM VGA 865": FontInfo{
		ROM:          "ibm_vga_865.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   IBMPCPixelRatio,
	},
	"IBM VGA 866": FontInfo{
		ROM:          "ibm_vga_866.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   IBMPCPixelRatio,
	},
	"IBM VGA 866b": FontInfo{
		ROM:          "ibm_vga_866b.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   IBMPCPixelRatio,
	},
	"IBM VGA 866c": FontInfo{
		ROM:          "ibm_vga_866c.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   IBMPCPixelRatio,
	},
	"IBM VGA 866u": FontInfo{
		ROM:          "ibm_vga_866u.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   IBMPCPixelRatio,
	},
	"IBM VGA 869": FontInfo{
		ROM:          "ibm_vga_869.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   IBMPCPixelRatio,
	},
	"IBM VGA 1251": FontInfo{
		ROM:          "ibm_vga_1251.bin",
		Size:         image.Pt(8, 16),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   IBMPCPixelRatio,
	},
	"IBM VGA50": FontInfo{
		ROM:          "ibm_vga50.bin",
		Size:         image.Pt(8, 8),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   IBMPCPixelRatio,
	},
	"IBM VGA50 437": FontInfo{
		ROM:          "ibm_vga50_437.bin",
		Size:         image.Pt(8, 8),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   IBMPCPixelRatio,
	},
	"IBM VGA50 850": FontInfo{
		ROM:          "ibm_vga50_450.bin",
		Size:         image.Pt(8, 8),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   IBMPCPixelRatio,
	},
	"IBM VGA50 865": FontInfo{
		ROM:          "ibm_vga50_865.bin",
		Size:         image.Pt(8, 8),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   IBMPCPixelRatio,
	},
	"IBM VGA50 866": FontInfo{
		ROM:          "ibm_vga50_866.bin",
		Size:         image.Pt(8, 8),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   IBMPCPixelRatio,
	},
	"IBM VGA50 1251": FontInfo{
		ROM:          "ibm_vga50_1251.bin",
		Size:         image.Pt(8, 8),
		DisplayRatio: CRTDisplayRatio,
		PixelRatio:   IBMPCPixelRatio,
	},
}

// Ratio between width and height.
type Ratio struct {
	X, Y int
}

// Default ratios
var (
	// CRTDisplayRatio typically is 4:3.
	CRTDisplayRatio = Ratio{4, 3}

	// AmigaPixelRatio is the default pixel ratio on Commodore Amiga.
	AmigaPixelRatio = Ratio{5, 12}

	// IBMPCPixelRatio is the default pixel ratio on IBM PC.
	IBMPCPixelRatio = Ratio{20, 27}
)

// FontInfo has information about a font.
type FontInfo struct {
	// ROM is the name of the chargen binary.
	ROM string

	// Size of the characters.
	Size image.Point

	// DisplayRatio is the aspect ratio of the display device the font was
	// intended for. Up until around 2003 the common display device was either a
	// CRT computer monitor or a CRT TV which usually had a display aspect ratio
	// of 4:3, and occasionally 5:4. Those formats are being replaced by
	// “widescreen” displays commonly having a 16:10 or 16:9 aspect ratio.
	DisplayRatio Ratio

	// PixelRatio is the aspect ratio of the pixels. Modern display devices (LCD,
	// LED and plasma) tend to have square pixels or at least as near square as
	// is technically feasible, because square pixels make it easy to draw squares
	// and circles. For various technical reasons square pixels have not always
	// been the norm however.
	PixelRatio Ratio
}

// Font for a SAUCE font name.
func Font(name string) (*chargen.Font, error) {
	if strings.TrimSpace(name) == "" {
		name = "ibm_vga"
	}

	clean := cleanFontName(name)
	for other, info := range Fonts {
		if cleanFontName(other) == clean {
			data, err := data.Bytes(fmt.Sprintf("font/chargen/%s", info.ROM))
			if err != nil {
				return nil, err
			}
			opts := chargen.MaskOptions{Size: info.Size}
			return chargen.New(chargen.NewBytesMask(data, opts)), nil
		}
	}

	return nil, fmt.Errorf("sauce: font %q not supported", name)
}

func cleanFontName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.Replace(name, " ", "_", -1)
	return name
}
