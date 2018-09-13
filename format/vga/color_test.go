package vga

import (
	"image/color"
	"testing"
)

func TestColorIndex(t *testing.T) {
	if i := ColorIndex(color.Black, Palette); i != 0 {
		t.Fatalf("expected black to be at 0, got %d", i)
	}
	if i := ColorIndex(color.White, Palette); i != 15 {
		t.Fatalf("expected black to be at 15, got %d", i)
	}
}

func TestToRGB(t *testing.T) {
	if c := ToRGB(color.Black); c != 0x000000 {
		t.Fatalf("expected white to be 0x000000, got %#06x", c)
	}
	if c := ToRGB(color.White); c != 0xffffff {
		t.Fatalf("expected white to be 0xffffff, got %#06x", c)
	}
}

func TestColorEqual(t *testing.T) {
	testColorEqual(t, RGB(0x000000), color.Black)
	testColorEqual(t, RGB(0xffffff), color.White)
}

func testColorEqual(t *testing.T, a, b color.Color) {
	t.Helper()
	if !colorEqual(a, b) {
		t.Fatalf("expected %v to equal %v", a, b)
	}
}
