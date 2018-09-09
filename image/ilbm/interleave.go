package ilbm

func deinterleave(src []byte, header BitmapHeader) (dst []byte) {
	size := header.rowSize() * int(header.Height)
	dst = make([]byte, header.Planes)

	var (
		planes [][]byte
		offset int
	)
	for i := 0; i < int(header.Planes); i++ {
		planes = append(planes, src[offset:])
		offset += size
	}

	deinterleavePlanes(dst, planes, header)

	return
}

func deinterleavePlanes(dst []byte, planes [][]byte, header BitmapHeader) {
	var (
		count   int
		offset  int
		rowSize = header.rowSize()
	)

	for y := 0; y < int(header.Height); y++ {
		for p := 0; p < int(header.Planes); p++ {
			copy(dst[count:], planes[p][offset:offset+rowSize])
			count += rowSize
		}
		offset += rowSize
	}
}
