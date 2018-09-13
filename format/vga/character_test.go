package vga

import (
	"image/color"
	"testing"
)

func TestAttribute(t *testing.T) {
	var a Attribute
	if s := a.String(); s != "<none>" {
		t.Fatalf("expected <none>, got %q", s)
	}

	a |= Bold
	if s := a.String(); s != "bold" {
		t.Fatalf(`expected "bold", got %q`, s)
	}

	a |= Faint
	if s := a.String(); s != "bold,faint" {
		t.Fatalf(`expected "bold,faint", got %q`, s)
	}

	a |= Standout
	if s := a.String(); s != "bold,faint,standout" {
		t.Fatalf(`expected "bold,faint,standout", got %q`, s)
	}

	a = Underline
	if s := a.String(); s != "underline" {
		t.Fatalf(`expected "underline", got %q`, s)
	}
	a |= Blink
	if s := a.String(); s != "underline,blink" {
		t.Fatalf(`expected "underline,blink", got %q`, s)
	}
	a ^= Reverse
	if s := a.String(); s != "underline,blink,reverse" {
		t.Fatalf(`expected "underline,blink,reverse", got %q`, s)
	}
	a |= Conceal
	if s := a.String(); s != "underline,blink,reverse,conceal" {
		t.Fatalf(`expected "underline,blink,reverse,conceal", got %q`, s)
	}
}

func TestCharacter(t *testing.T) {
	c := BlankCharacter
	if p := c.CodePoint(); p != 0x20 {
		t.Fatalf("expected code point %#02x, got %#02x", 0x20, p)
	}

	c.SetForegroundColor(color.Black)
	if v := c.ForegroundColor(); !colorEqual(v, color.Black) {
		t.Fatalf("expected black foreground, got %v", v)
	}

	c.SetBackgroundColor(color.White)
	if v := c.BackgroundColor(); !colorEqual(v, color.White) {
		t.Fatalf("expected white background, got %v", v)
	}

	c.Reset(color.White, color.Black)
	if p := c.CodePoint(); p != 0x20 {
		t.Fatalf("expected code point %#02x, got %#02x", 0x20, p)
	}
	if v := c.ForegroundColor(); !colorEqual(v, color.White) {
		t.Fatalf("expected white foreground, got %v", v)
	}
	if v := c.BackgroundColor(); !colorEqual(v, color.Black) {
		t.Fatalf("expected black background, got %v", v)
	}

	c.ClearAttributes()
	c.SetAttribute(Bold)
	if a := c.Attributes(); a != Bold {
		t.Fatalf("expected %s, got %s", Bold, a)
	}
	c.SetAttributes(Reverse)
	if a := c.Attributes(); a != Reverse {
		t.Fatalf("expected %s, got %s", Reverse, a)
	}
	c.SetAttribute(Bold)
	c.ClearAttribute(Reverse)
	if a := c.Attributes(); a != Bold {
		t.Fatalf("expected %s, got %s", Bold, a)
	}
	c.Reset(color.White, color.Black)
	if a := c.Attributes(); a != 0 {
		t.Fatalf("expected <none>, got %s", a)
	}

	c.SetCodePoint(0x2a)
	if p := c.CodePoint(); p != 0x2a {
		t.Fatalf("expected code point %#02x, got %#02x", 0x2a, p)
	}
}
