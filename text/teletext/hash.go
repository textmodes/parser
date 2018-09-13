package teletext

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

func DecodeHash(hash string) (*Page, error) {
	// #d:A0GLVkgQB4cWLGg2EDJAxaskE3fuQMmSCfj6A2DR00ar2LAvqRs9epGz16kbPWUOk0CBAg0NNDT4s-aHjz5oeJECBAgQIC-pAk16kCTXqQJ9ZQ6TQIECDU31tf_D_qatf-ruxQIECBAgLokaNGiRo0aJGjRlF69evXr1q5auWLFi1cuWLVi5evXr168HQ37NOPKgy7MuPpp37uaDZlw5tmXpzQcN_PplyIECBAgQIEEzLhzbMvTmgw4t_Xog3ZfHRBt37umhPzQZdmXH00793NAgQZt_JBw37NOPKgw7siDHy07cqDHv27dPPnp37svLmgw8sqBBiy6d2dBzy7uiDpvQMmKDbp2bNO_cg0b-vPLo37MnNcgQIECBAgQIECBAgQIECBAgQIECBAgQIECBAgQIECBAgQIECBAgDVt_TLy5oNO5Bh2bEHDDy6c0G_Mgi7s-zDuyIMO7IgQIECANXw7MvNBv69OenJlQTN-7Jv3IO-nZsQZdmXH0QYUCBAgQIA3DLy579yDpvQb-2Xlzy5UGXDj0IOG_Zpx5UGbfyx5UG_cgDMWqCdv7ZduLLyXIECBAgQIECBAgQIECBAgQIECBAgQIECAMgQIECBAgQIECBAgQIECBAgQIECBAgQIECBAgQIECBAgQIA1TRlQbsvdBj37dunnz0792XlzQd9OzYg5ZeGzDjyoECBAgDcN-zTjyoMPXpo38tPTTl5oNO5A0YoMPLLh5rkCBAgQIECAMgQIECBAgQIECBAgQIECBAgQIECBAgQIECBAgQIECBAgQIA1TRlQRdmXH038sOxBSy5t_Lagp78enL08oNmXDm2ZenNAgDZfHDZh07kG3fyyoMOLf16IOmjKg5b9mVBh3ZEHTRlQIECAN239NO7Og5-efTLtQYsundnQdeeXIsQYdunIgx792PLy3cw2_MgwoNm_ug6deW7f16IM3LftQdt_TLy5rkCBAgQIECBAgQIECBAgQIECBAgQIECBAgQIECBAgQIECBAgQIECBAgQIECAodByN-3Kg3Ze_NBk058vPoHYtGIOvv5bMiDJpz5efQOxaMiZ0JI390EPLlzYfCDn309MejLzQdN6CJpz6emHYgqVkDZy3QIECBAgQIECBAgQIECBAgQIECBAgQIECBAgQIECBAgQIECA
	if len(hash) == 0 || hash[0] != '#' {
		return nil, errors.New("teletext: not a hash")
	}
	if i := strings.IndexByte(hash, ':'); i != 2 {
		return nil, errors.New("teletext: expedted \"#x:...\"")
	}

	// Extract base-64 encoded data
	data := []byte(hash[3:])
	if i := bytes.IndexByte(data, ':'); i > -1 {
		// Get rid of metadata
		data = data[:i]
	}

	switch len(data) {
	case 1167, 1172:
	default:
		return nil, fmt.Errorf("teletext: unexpected hash length %d", len(data))
	}

	// Decode base64 to 7-bit data
	dst7 := make([]byte, base64.RawURLEncoding.DecodedLen(len(data)))
	if _, err := base64.RawURLEncoding.Decode(dst7, data); err != nil {
		return nil, err
	}

	// Decode 7-bit data to 8-bit data
	dst8 := make([]byte, (len(dst7)/7)*8)
	decode7bits(dst8, dst7)
	// log.Printf("hash decode:\n%s", hex.Dump(dst8))
	// log.Printf("hash decoded to %d (%d lines)", len(dst8), len(dst8)/40)

	page := NewPage()
	for row, rows := 0, len(dst8)/40; row < rows; row++ {
		page.SetLineBytes(row, dst8[row*40:])
	}

	return page, nil
}

func decode7bits(dst, src []byte) {
	for d, s := 0, 0; s < len(src); d, s = d+8, s+7 {
		dst[d+0] |= (src[s+0]>>1)&0x7f | 0
		dst[d+1] |= (src[s+0]<<6)&0x40 | (src[s+1]>>2)&0x3f
		dst[d+2] |= (src[s+1]<<5)&0x60 | (src[s+2]>>3)&0x1f
		dst[d+3] |= (src[s+2]<<4)&0x70 | (src[s+3]>>4)&0x0f
		dst[d+4] |= (src[s+3]<<3)&0x78 | (src[s+4]>>5)&0x07
		dst[d+5] |= (src[s+4]<<2)&0x7c | (src[s+5]>>6)&0x03
		dst[d+6] |= (src[s+5]<<1)&0x7e | (src[s+6]>>7)&0x01
		dst[d+7] |= (src[s+6]<<0)&0x7f | 0
	}
}
