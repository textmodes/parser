package chargen

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"os"
	"testing"
)

func TestFont(t *testing.T) {
	testMaskFont(t, testFontImage5x7, image.Pt(5, 7))
	testMaskFont(t, testFontImage8x8, image.Pt(8, 8))
	testBytesMaskFont(t, testFontBytes8x8, image.Pt(8, 8))
}

func testMaskFont(t *testing.T, data []byte, size image.Point) {
	t.Helper()
	t.Run(fmt.Sprintf("%dx%d", size.X, size.Y), func(t *testing.T) {
		i, err := png.Decode(bytes.NewBuffer(data))
		if err != nil {
			t.Fatal(err)
		}
		mask := NewMask(i, MaskOptions{Size: size})
		testFont(t, mask, size, "mask")
	})
}

func testBytesMaskFont(t *testing.T, data []byte, size image.Point) {
	t.Helper()
	t.Run(fmt.Sprintf("%dx%d", size.X, size.Y), func(t *testing.T) {
		testFont(t, NewBytesMask(data, MaskOptions{Size: size}), size, "bytes-mask")
	})
}

func testFont(t *testing.T, mask Mask, size image.Point, name string) {
	t.Helper()

	if model := mask.ColorModel(); model != color.AlphaModel {
		t.Fatal("expected color.AlphaModel ColorModel")
	}

	var (
		font = New(mask)
		im   = image.NewRGBA(image.Rect(0, 0, size.X*6+4, size.Y+4))
		bg   = image.NewUniform(color.Black)
		fg   = image.NewUniform(color.RGBA{0xff, 0, 0, 0xff} /* color.White */)
	)

	draw.Draw(im, im.Bounds(), bg, image.ZP, draw.Src)
	font.Draw(im, image.Pt(2, 2), fg, 'G')
	font.DrawString(im, image.Pt(2+size.X, 2), fg, "opher")
	font.Draw(im, image.Pt(2, 2), fg, 0x100) // Not in bitmap, should be ignored

	if os.Getenv("TEST_WRITE_PNG") != "" {
		f, err := os.Create(fmt.Sprintf("testdata/test-font-%s-%dx%d.png", name, size.X, size.Y))
		if err != nil {
			t.Skip(err)
		}
		defer f.Close()
		if err = png.Encode(f, im); err != nil {
			t.Fatal(err)
		}
		if err = ioutil.WriteFile(fmt.Sprintf("testdata/test-font-%s-%dx%d.bin", name, size.X, size.Y), mask.(*bitmap).data, 0644); err != nil {
			t.Fatal(err)
		}
	}
}
