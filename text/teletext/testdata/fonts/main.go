package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"os"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

var (
	backgroundWidth  = 650
	backgroundHeight = 150
	utf8FontFile     = "./ttf/MODE7GX3.TTF"
	utf8FontSize     = float64(15.0)
	spacing          = float64(1.5)
	dpi              = float64(72)
	ctx              = new(freetype.Context)
	utf8Font         = new(truetype.Font)
	red              = color.RGBA{255, 0, 0, 255}
	blue             = color.RGBA{0, 0, 255, 255}
	white            = color.RGBA{255, 255, 255, 255}
	black            = color.RGBA{0, 0, 0, 255}
	background       *image.RGBA
	// more color at https://github.com/golang/image/blob/master/colornames/table.go
)

func main() {

	// download font from http://www.slackware.com/~alien/slackbuilds/wqy-zenhei-font-ttf/build/wqy-zenhei-0.4.23-1.tar.gz
	// extract wqy-zenhei.ttf to the same folder as this program

	// Read the font data - for this example, we load the Chinese fontfile wqy-zenhei.ttf,
	// but it will display any utf8 fonts such as Russian, Japanese, Korean, etc as well.
	// some utf8 fonts cannot be displayed. You need to use your own language .ttf file
	fontBytes, err := ioutil.ReadFile(utf8FontFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	utf8Font, err = freetype.ParseFont(fontBytes)
	if err != nil {
		fmt.Println(err)
		return
	}

	fontForeGroundColor, fontBackGroundColor := image.NewUniform(black), image.NewUniform(white)

	background = image.NewRGBA(image.Rect(0, 0, backgroundWidth, backgroundHeight))

	draw.Draw(background, background.Bounds(), fontBackGroundColor, image.ZP, draw.Src)

	ctx = freetype.NewContext()
	ctx.SetDPI(dpi) //screen resolution in Dots Per Inch
	ctx.SetFont(utf8Font)
	ctx.SetFontSize(utf8FontSize) //font size in points
	ctx.SetClip(background.Bounds())
	ctx.SetDst(background)
	ctx.SetSrc(fontForeGroundColor)

	var UTF8text = []string{
		`English - Hello, Chinese - 你好, Russian - Здравствуйте, Korean - 여보세요, Greek - Χαίρετε`,
		`Tajik - Салом, Japanese - こんにちは, Icelandic - Halló, Belarusian - добры дзень`,
		`symbols - © Ø ® ß ◊ ¥ Ô º ™ € ¢ ∞ § Ω`,
	}

	// Draw the text to the background
	pt := freetype.Pt(10, 10+int(ctx.PointToFixed(utf8FontSize)>>6))

	// not all utf8 fonts are supported by wqy-zenhei.ttf
	// use your own language true type font file if your language cannot be printed

	for _, str := range UTF8text {
		_, err := ctx.DrawString(str, pt)
		if err != nil {
			fmt.Println(err)
			return
		}
		pt.Y += ctx.PointToFixed(utf8FontSize * spacing)
	}

	// Save
	outFile, err := os.Create("utf8text.png")
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	defer outFile.Close()
	buff := bufio.NewWriter(outFile)

	err = png.Encode(buff, background)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	// flush everything out to file
	err = buff.Flush()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	fmt.Println("Save to utf8text.png")
}
