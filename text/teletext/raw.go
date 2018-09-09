package teletext

import (
	"io"
)

// DecodeRaw decodes a raw TeleText page (of 25 * 40 = 1000 bytes)
func DecodeRaw(r io.Reader) (Pages, error) {
	var (
		pages      Pages
		prev, page *Page
		buf        [1000]byte
		row        uint8
		err        error
	)

	for {
		// Read 25 lines of 40 bytes
		if prev, page = page, NewPage(); prev != nil {
			page.Number = prev.Number + 1
		} else {
			page.Number = 0x10001
		}
		if _, err = io.ReadFull(r, buf[:]); err != nil {
			if err != io.EOF {
				return nil, err
			} else if len(pages) == 0 {
				return nil, err
			}
			break
		}
		for row = 0; row < 25; row++ {
			page.Lines[row] = NewLine(buf[row*40:])
		}
		pages = append(pages, page)
	}

	return pages, nil
}
