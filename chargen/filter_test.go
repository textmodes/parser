package chargen

import (
	"image"
	"image/color"
	"testing"
)

func TestFilterAddColumn(t *testing.T) {
	mask := NewBytesMask([]byte{0xff, 0xff, 0xff, 0xff}, MaskOptions{
		Size: image.Pt(2, 4), /* 2x4 font */
	})

	test := AddColumn(mask)

	if size := test.CharacterSize(); size.X != 3 || size.Y != 4 {
		t.Fatalf("expected size of (3, 4) got %s", size)
	}

	if area := test.Bounds().Size(); area.X != 12 || area.Y != 4 {
		t.Fatalf("expected bounds size of (12, 4) got %s", area)
	}

	for x := 0; x < 6; x++ {
		c := test.At(x, 2)
		_, _, _, a := c.RGBA()
		switch x {
		case
			0, 1, // character 3
			3, 4: // character 4
			if a == 0x00 {
				t.Fatalf("expected pixel at (%d, 2) to be opaque", x)
			}
		default:
			if a != 0x00 {
				t.Fatalf("expected pixel at (%d, 2) to be transparent", x)
			}
		}
	}
}

func testFilterItalicsMask() Mask {
	return NewBytesMask([]byte{
		// First
		0x00, /* ________ 0 */
		0x00, /* ________ 1 */
		0x00, /* ________ 2 */
		0x18, /* ___##___ 3 */
		0x00, /* ________ 4 */
		0x00, /* ________ 5 */
		0x18, /* ___##___ 6 */
		0x18, /* ___##___ 7 */
		0x18, /* ___##___ 8 */
		0x18, /* ___##___ 9 */
		0x18, /* ___##___ a */
		0x18, /* ___##___ b */
		0x00, /* ________ c */
		0x00, /* ________ d */
		0x00, /* ________ e */
		0x00, /* ________ f */
		// Second
		0x00, /* ________ 0 */
		0x00, /* ________ 1 */
		0x00, /* ________ 2 */
		0x00, /* ________ 3 */
		0x00, /* ________ 4 */
		0x00, /* ________ 5 */
		0x00, /* ________ 6 */
		0x00, /* ________ 7 */
		0x00, /* ________ 8 */
		0x00, /* ________ 9 */
		0xff, /* ######## a */
		0xff, /* ######## b */
		0x00, /* ________ c */
		0x00, /* ________ d */
		0x00, /* ________ e */
		0x00, /* ________ f */
	}, MaskOptions{
		Size: image.Pt(8, 16), /* 8x16 font */
	})
}

func TestFilterItalics(t *testing.T) {
	mask := testFilterItalicsMask()
	test := Italics(mask)

	t.Logf("mask character size: %s", mask.CharacterSize())
	t.Logf("test character size: %s", test.CharacterSize())

	if size := test.CharacterSize(); size.X != 8 || size.Y != 16 {
		t.Fatalf("expected size of (8, 16) got %s", size)
	}

	if area := test.Bounds().Size(); area.X != 16 || area.Y != 16 {
		t.Fatalf("expected bounds size of (16, 16) got %s", area)
	}

	if model := test.ColorModel(); model != color.AlphaModel {
		t.Fatalf("expected %T, got %T", color.AlphaModel, model)
	}

	// Test pixels
	want := NewBytesMask([]byte{
		// First
		0x00, /* ________ 0 */
		0x00, /* ________ 1 */
		0x00, /* ________ 2 */
		0x06, /* _____##_ 3 */
		0x00, /* ________ 4 */
		0x00, /* ________ 5 */
		0x0c, /* ____##__ 6 */
		0x0c, /* ____##__ 7 */
		0x18, /* ___##___ 8 */
		0x18, /* ___##___ 9 */
		0x18, /* ___##___ a */
		0x18, /* ___##___ b */
		0x00, /* ________ c */
		0x00, /* ________ d */
		0x00, /* ________ e */
		0x00, /* ________ f */
		// Second
		0x00, /* ________ 0 */
		0x00, /* ________ 1 */
		0x00, /* ________ 2 */
		0x00, /* ________ 3 */
		0x00, /* ________ 4 */
		0x00, /* ________ 5 */
		0x00, /* ________ 6 */
		0x00, /* ________ 7 */
		0x00, /* ________ 8 */
		0x00, /* ________ 9 */
		0xff, /* ######## a */
		0xff, /* ######## b */
		0x00, /* ________ c */
		0x00, /* ________ d */
		0x00, /* ________ e */
		0x00, /* ________ f */
	}, MaskOptions{
		Size: image.Pt(8, 16), /* 8x16 font */
	})

	t.Logf("mask:")
	testMaskDump(t, mask, mask.Bounds())
	testMaskEqual(t, test, want)
}

func testFilterRoundCharactersMask() Mask {
	return NewBytesMask([]byte{
		// First
		0x00, /* ________ */
		0x01, /* _______# */
		0x02, /* ______#_ */
		0x00, /* ________ */
		// Second
		0x00, /* ________ */
		0x80, /* #_______ */
		0x40, /* _#______ */
		0x00, /* ________ */
	}, MaskOptions{
		Size: image.Pt(8, 4), /* 8x4 font */
	})
}

func TestFilterRoundCharacters(t *testing.T) {
	mask := testFilterRoundCharactersMask()
	test := RoundCharacters(mask)

	t.Logf("mask character size: %s", mask.CharacterSize())
	t.Logf("test character size: %s", test.CharacterSize())

	if size := test.CharacterSize(); size.X != 16 || size.Y != 8 {
		t.Fatalf("expected size of (16, 8) got %s", size)
	}

	if area := test.Bounds().Size(); area.X != 32 || area.Y != 8 {
		t.Fatalf("expected bounds size of (32, 8) got %s", area)
	}

	if model := test.ColorModel(); model != color.AlphaModel {
		t.Fatalf("expected %T, got %T", color.AlphaModel, model)
	}

	// Test pixels
	want := NewBytesMask([]byte{
		// First
		0x00, 0x00, /* ________________ */
		0x00, 0x00, /* ________________ */
		0x00, 0x03, /* ______________## */
		0x00, 0x07, /* _____________### */
		0x00, 0x0e, /* ____________###_ */
		0x00, 0x0c, /* ____________##__ */
		0x00, 0x00, /* ________________ */
		0x00, 0x00, /* ________________ */
		// Second
		0x00, 0x00, /* ________________ */
		0x00, 0x00, /* ________________ */
		0xc0, 0x00, /* ##______________ */
		0xe0, 0x00, /* ###_____________ */
		0x70, 0x00, /* _###____________ */
		0x30, 0x00, /* __##____________ */
		0x00, 0x00, /* ________________ */
		0x00, 0x00, /* ________________ */
	}, MaskOptions{
		Size: image.Pt(16, 8), /* 16x8 font */
	})

	t.Logf("mask:")
	testMaskDump(t, mask, mask.Bounds())
	testMaskEqual(t, test, want)
}

// Test pixels of submask
func TestFilterRoundCharactersSubMask(t *testing.T) {
	mask := testFilterRoundCharactersMask()
	size := mask.CharacterSize()
	offs := size.X * 2
	rect := image.Rect(offs, 0, offs+size.X*2, size.Y*2) // Second character only
	temp := RoundCharacters(mask)
	test := temp.SubMask(rect)
	t.Logf("%s ∩ %s = %s", temp.Bounds(), rect, test.Bounds())
	want := NewBytesMask([]byte{
		// First
		0x00, 0x00, /* ________________ */
		0x00, 0x00, /* ________________ */
		0x00, 0x00, /* ________________ */
		0x00, 0x00, /* ________________ */
		0x00, 0x00, /* ________________ */
		0x00, 0x00, /* ________________ */
		0x00, 0x00, /* ________________ */
		0x00, 0x00, /* ________________ */
		// Second
		0x00, 0x00, /* ________________ */
		0x00, 0x00, /* ________________ */
		0xc0, 0x00, /* ##______________ */
		0xe0, 0x00, /* ###_____________ */
		0x70, 0x00, /* _###____________ */
		0x30, 0x00, /* __##____________ */
		0x00, 0x00, /* ________________ */
		0x00, 0x00, /* ________________ */
	}, MaskOptions{
		Size: image.Pt(16, 8), /* 16x8 font */
	})

	t.Logf("mask character size: %s", mask.CharacterSize())
	t.Logf("test character size: %s", test.CharacterSize())
	t.Logf("mask:")
	testMaskDump(t, mask, mask.Bounds())
	// FIXME
	//t.Logf("want:")
	//testMaskDump(t, want, want.Bounds())
	testMaskEqual(t, test, want)
}

func testMaskEqual(t *testing.T, test, want Mask) {
	t.Helper()

	if test == nil {
		if want == nil {
			return
		}
		t.Fatal("test Mask is nil")
	} else if want == nil {
		t.Fatal("want Mask is nil, test mask is not")
	}

	bounds := test.Bounds()
	for y := bounds.Min.X; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			a, b := test.At(x, y), want.At(x, y)
			if isOpaque(a) != isOpaque(b) {
				t.Logf("want: %s", want.Bounds())
				testMaskDump(t, want, want.Bounds())
				t.Logf("test: %s", test.Bounds())
				testMaskDump(t, test, test.Bounds())
				t.Fatalf("pixel at (%d, %d) differs", x, y)
			}
		}
	}

	t.Logf("pass: %s", test.Bounds())
	testMaskDump(t, test, test.Bounds())
}

func testMaskDump(t *testing.T, mask Mask, bounds image.Rectangle) {
	if bounds.Min.X > 0 {
		for y := 0; y < bounds.Min.Y; y++ {
			row := make([]byte, bounds.Max.X)
			for i := range row {
				row[i] = '_'
			}
			t.Log(string(row))
		}
	}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		row := make([]byte, 0, bounds.Size().X)
		for x := 0; x < bounds.Min.X; x++ {
			row = append(row, '_')
		}
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if isOpaque(mask.At(x, y)) {
				row = append(row, '#')
			} else {
				row = append(row, '_')
			}
		}
		t.Log(string(row))
	}
}
