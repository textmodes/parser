package teletext

import (
	"bufio"
	"io"
)

func DecodeM7(r io.Reader) (*Page, error) {
	var (
		page = NewPage()
		br   = bufio.NewReader(r)
		line []byte
		row  int
		err  error
	)
	for row < 25 {
		if line, err = br.ReadBytes(0x0a); err != nil {
			if err != io.EOF {
				return nil, err
			}
		}
		page.SetLineBytes(row, line)
		row++
	}
	return page, nil
}
