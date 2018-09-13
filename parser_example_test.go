package parser_test

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"os"

	"github.com/textmodes/parser"
)

type testImage struct {
	Bytes []byte
}

// Image decodes Bytes to Image.
func (i testImage) Image() (image.Image, error) {
	return png.Decode(bytes.NewBuffer(i.Bytes))
}

// RenderJPEG parses p as Image and returns an encoded JPEG to standard out.
func RenderJPEG(p parser.Parser) (err error) {
	if imager, ok := p.(parser.Image); ok {
		var im image.Image
		if im, err = imager.Image(); err != nil {
			return
		}
		return jpeg.Encode(os.Stdout, im, nil)
	}

	return fmt.Errorf("parser: %v does not implement parser.Image", p)
}

// ExampleImage renders a PNG to stdout as JPEG.
func ExampleImage() {
	b, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(fmt.Sprintf("error reading %s: %v", os.Args[1], err))
	}

	if err = RenderJPEG(testImage{Bytes: b}); err != nil {
		panic(err)
	}
}
