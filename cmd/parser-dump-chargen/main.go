package main

import (
	"flag"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/textmodes/parser/chargen"
	"github.com/textmodes/parser/data"
)

func main() {
	x := flag.Int("x", 8, "x size")
	y := flag.Int("y", 8, "y size")
	stridex := flag.Int("stridex", 0, "x stride (default: auto detect)")
	showGrid := flag.Bool("grid", true, "show grid")
	flag.Parse()

	for _, arg := range flag.Args() {
		dump(*x, *y, *stridex, *showGrid, arg)
	}
}

var (
	red  = &color.RGBA{0xff, 0x00, 0x00, 0xff}
	none = new(color.RGBA)
)

type grid struct {
	x, y   int
	bounds image.Rectangle
}

func (g grid) At(x, y int) color.Color {
	if x == 0 || y == 0 {
		return red
	} else if (x+1)%g.x == 0 {
		return red
	} else if (y+1)%g.y == 0 {
		return red
	}
	return none
}

func (g grid) Bounds() image.Rectangle {
	return g.bounds
}

func (g grid) ColorModel() color.Model {
	return color.RGBAModel
}

func dump(x, y, strideX int, withGrid bool, name string) {
	b, err := data.Bytes(name)
	if err != nil {
		panic(err)
	}

	var (
		opts = chargen.MaskOptions{Size: image.Pt(x, y), StrideX: strideX}
		font = chargen.New(chargen.NewBytesMask(b, opts))
		size = image.Rect(0, 0, x*16, y*16)
	)

	if withGrid {
		size.Max.X = 1 + (x+1)*16
		size.Max.Y = 1 + (y+1)*16
	}

	var (
		im = image.NewRGBA(size)
		fg = image.NewUniform(color.White)
		bg = image.NewUniform(color.Black)
		gg = image.NewUniform(&color.RGBA{0xff, 0x00, 0x00, 0xff})
	)

	if withGrid {
		draw.Draw(im, im.Bounds(), gg, image.ZP, draw.Over)
	}

	for c := 0; c < 256; c++ {
		var (
			xs, ys = (c % 16), (c >> 4)
			xx, yy int
		)
		if withGrid {
			xx = 1 + (xs * (x + 1))
			yy = 1 + (ys * (y + 1))
			log.Printf("%d: (%d, %d) -> (%d, %d)", c, xs, ys, xx, yy)
		} else {
			xx = xs * x
			yy = ys * y
			log.Printf("%d: (%d, %d) -> (%d, %d)", c, xs, ys, xx, yy)
		}
		draw.Draw(im, image.Rect(xx, yy, xx+x, yy+y), bg, image.ZP, draw.Src)
		font.Draw(im, image.Pt(xx, yy), fg, uint16(c))
	}

	o, err := os.Create(filepath.Base(name) + ".png")
	if err != nil {
		panic(err)
	}
	defer o.Close()

	if err = png.Encode(o, im); err != nil {
		panic(err)
	}
}
