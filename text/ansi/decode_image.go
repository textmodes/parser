package ansi

import (
	"image"
	"image/gif"
	"time"
)

// Progress callback.
func (decoder *Decoder) Progress(fn func(float64)) {
	decoder.progressFunc = fn
	decoder.Text.Progress(fn)
}

// Image renders the BinaryText to an image.
func (decoder *Decoder) Image() (image.Image, error) {
	return decoder.Text.Image(decoder.Font, true)
}

// ImageBlink renders the ANSi to an image; blink indicates if we're in blink state.
func (decoder *Decoder) ImageBlink(blink bool) (image.Image, error) {
	return decoder.Text.Image(decoder.Font, blink)
}

// Animate returns a rendered buffer.
func (decoder *Decoder) Animate() (*gif.GIF, error) {
	return decoder.AnimateDelay(time.Millisecond * 400)
}

// AnimateDelay returns a rendered buffer with the selected font as animated GIF.
func (decoder *Decoder) AnimateDelay(delay time.Duration) (*gif.GIF, error) {
	var (
		src [2]*image.Paletted
		err error
		d   = int(delay / (time.Second / 100))
	)
	if src[0], err = decoder.Text.Image(decoder.Font, false); err != nil {
		return nil, err
	}
	if decoder.Text.DisableBlink {
		// Oh guess what, we're not even blinking...
		return &gif.GIF{
			Image: src[:1],
		}, nil
	}
	// TODO(maze): we can optimize a lot here, by only drawing the glyphs that
	//             didn't draw (because they're blinking) in the first pass over
	//             the second image
	if src[1], err = decoder.Text.Image(decoder.Font, true); err != nil {
		return nil, err
	}
	return &gif.GIF{
		Image: src[:],
		Delay: []int{d, d},
	}, nil
}

// Scroller returns an animated GIF of the piece scrolling.
func (decoder *Decoder) Scroller() (*gif.GIF, error) {
	return decoder.ScrollerDelay(time.Millisecond * 400)
}

// ScrollerDelay returns an animated GIF of the piece scrolling.
func (decoder *Decoder) ScrollerDelay(delay time.Duration) (*gif.GIF, error) {
	// Pass 1: encode entire file to image, so we have the palette.

	var (
		src [2]*image.Paletted
		err error
	)
	if src[0], err = decoder.Text.Image(decoder.Font, false); err != nil {
		return nil, err
	}
	if src[1], err = decoder.Text.Image(decoder.Font, true); err != nil {
		return nil, err
	}

	// Pass 2: cut the frames
	out := new(gif.GIF)
	if decoder.Text.Height() < 25 {
		// We only have one frame.
		out.Image = append(out.Image, src[0])
		out.Image = append(out.Image, src[1])
		out.Delay = []int{5, 5}
	} else {
		// Crop each frame, render it and append them to the gif.
		var (
			w      = int(decoder.Text.Width())
			h      = int(decoder.Text.Height())
			s      = src[0].Bounds().Size()
			cx     = s.X / w // char width in pixels
			cy     = s.Y / h // char height in pixels
			ox     = cx * w  // output image width
			oy     = cy * 25 // output image height
			pixels = ox * cy // pixels stride
			bounds = image.Rect(0, 0, ox, oy)
			d      = int(delay / (time.Second / 100))
		)
		if d < 5 {
			d = 5
		}
		if decoder.progressFunc != nil {
			decoder.progressFunc(0)
		}
		for y := 0; y < h-25; y++ {
			//for i := 0; i < 2; i++ {
			out.Image = append(out.Image, &image.Paletted{
				Pix:     src[(y/2)%2].Pix[y*pixels:],
				Stride:  src[(y/2)%2].Stride,
				Rect:    bounds,
				Palette: src[(y/2)%2].Palette,
			})
			out.Delay = append(out.Delay, d)
			if decoder.progressFunc != nil {
				decoder.progressFunc(float64(y) / float64(h-25))
			}
			//}
		}
		// Last frame is repeated, so the final image stays on longer
		for r := 0; r < 10; r++ {
			l := len(out.Image)
			out.Image = append(out.Image, out.Image[l-1])
			out.Delay = append(out.Delay, d)
		}
	}

	if decoder.progressFunc != nil {
		decoder.progressFunc(1)
	}

	return out, nil
}
