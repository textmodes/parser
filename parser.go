/*
Package parser is the foundation for all content parsers on textmod.es.


Defaults

Although many parsers have different options and variants, the parser expects
the parsers to output an Image, Animation or Scroller with sensible defaults.

Some parser variants will provide additional parameters such as delay; it is
up to the implementer to check for most-specific to least specific rendition
interface where applicable.
*/
package parser

import (
	"image"
	"image/gif"
	"time"
)

// Parser base interface.
type Parser interface{}

// Image can generate images.
type Image interface {
	Image() (image.Image, error)
}

// Animation can generate animations.
type Animation interface {
	Animate() (*gif.GIF, error)
}

// AnimationDelay is like Animation with custom frame delay.
type AnimationDelay interface {
	AnimateDelay(time.Duration) (*gif.GIF, error)
}

// Scroller can generate scrolling animations. In the context of a text file,
// a scroller is different in that it will scroll the text within the
// limitations of the emulated display buffer.
type Scroller interface {
	Scroller() (*gif.GIF, error)
}

// ScrollerDelay is like Scroller with custom frame delay.
type ScrollerDelay interface {
	ScrollerDelay(time.Duration) (*gif.GIF, error)
}
