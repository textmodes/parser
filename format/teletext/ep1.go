package teletext

import (
	"bytes"
	"errors"
	"io"
)

const (
	ep1Prefix    = "\xfe\x01\x09\x00\x00\x00"
	ep1PrefixLen = 6
	ep1Suffix    = "\x00\x00"
	ep1SuffixLen = 2
	ep1Size      = 25*40 + ep1PrefixLen + ep1SuffixLen
)

// "basically "JWC", (number of frames in file), 00 00, then that many frames
// (25x40, each prefixed by FE 01 09 00 00 00 and terminated with 00 00).
// Individual frames in that format - i.e. without the six-byte 'JWC' header -
// are 'EP1' files, as used by various editors"

// DecodeEP1 decodes an EP1 encoded teletext page.
func DecodeEP1(r io.Reader) (Pages, error) {
	var (
		pages      Pages
		prev, page *Page
		buf        [40]byte
		row        uint8
		err        error
	)

	for {
		// Read prefix
		if _, err = io.ReadFull(r, buf[:ep1PrefixLen]); err != nil {
			if err != io.EOF || len(pages) == 0 {
				return nil, err
			}
			break
		}
		if bytes.Compare(buf[:ep1PrefixLen], []byte(ep1Prefix)) != 0 {
			if len(pages) == 0 {
				return nil, errors.New("teletext: EP1 header mark not found")
			}
			break
		}

		// Read 24 lines of 40 bytes
		if prev, page = page, NewPage(); prev != nil {
			page.Number = prev.Number + 1
		} else {
			page.Number = 0x10001
		}
		for row = 0; row < 24; row++ {
			if _, err = io.ReadFull(r, buf[:]); err != nil {
				return nil, err
			}
			page.SetRow(row, buf[:])
		}

		// Read suffix
		if _, err = io.ReadFull(r, buf[:ep1SuffixLen]); err != nil {
			return nil, err
		}

		pages = append(pages, page)
	}

	return pages, nil
}
