package ilbm

import (
	"errors"
	"fmt"
	"log"
)

func decompress(src []byte, header *BitmapHeader) (dst []byte, err error) {
	switch header.Compression {
	case 0:
		// No compression
		log.Printf("ilbm: decompress none %d bytes", len(src))
		return src, nil

	case 1:
		return decompressByteRun(src, header)

	default:
		return nil, errors.New("ilbm: unsupported compression method")
	}
}

func decompressByteRun(src []byte, header *BitmapHeader) (dst []byte, err error) {
	size := header.rowSize() * int(header.Height) * int(header.Planes)
	dst = make([]byte, 0, size)

	var pos, j int
	for remaining := len(src); remaining > 0; {
		var v = int(src[pos])
		pos++

		if v <= 127 {
			j = v
			remaining -= (j + 1)
			if remaining < 0 {
				return nil, fmt.Errorf("ilbm: error during byte run decompression: need %d more bytes", -remaining)
			}
			for ; j >= 0; j-- {
				dst = append(dst, src[pos])
				pos++
			}
		} else if v != 128 {
			j = 256 - v
			remaining -= (j + 1)
			if remaining < 0 {
				return nil, fmt.Errorf("ilbm: error during byte run decompression: need %d more bytes", -remaining)
			}
			for ; j >= 0; j-- {
				dst = append(dst, uint8(v))
			}
		} /* 128 is a NOP */
	}

	log.Printf("ilbm: decompress RLE %d to %d bytes", len(src), len(dst))
	return
}
