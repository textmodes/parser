package vga

import "testing"

func TestText(t *testing.T) {
	text := new(Text)
	text.Resize(4, 4)
	if w, h := text.Width(), text.Height(); w != 4 || h != 4 {
		t.Fatalf("expected 4x4, got %dx%d", w, h)
	}
}

func TestTextResize(t *testing.T) {
	text := new(Text)
	text.Resize(4, 4)
	if w, h := text.Width(), text.Height(); w != 4 || h != 4 {
		t.Fatalf("expected 4x4, got %dx%d", w, h)
	}
	text.WriteString("abcdefghijklmnop")
	t.Run("Resize(3,3)", func(t *testing.T) {
		text.Resize(3, 3)
		want := "abc\nefg\nijk\n"
		if got := text.String(); got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}

		t.Run("Resize(3,5)", func(t *testing.T) {
			text.Resize(3, 5)
			want := "abc\nefg\nijk\n   \n   \n"
			if got := text.String(); got != want {
				t.Fatalf("expected %q, got %q", want, got)
			}

			t.Run("Resize(5,5)", func(t *testing.T) {
				text.Resize(5, 5)
				want := "abc  \nefg  \nijk  \n     \n     \n"
				if got := text.String(); got != want {
					t.Fatalf("expected %q, got %q", want, got)
				}

				t.Run("Resize(3,3)", func(t *testing.T) {
					text.Resize(3, 3)
					want := "abc\nefg\nijk\n"
					if got := text.String(); got != want {
						t.Fatalf("expected %q, got %q", want, got)
					}
				})
			})
		})
	})
}

func TestTextScroll(t *testing.T) {
	text := new(Text)
	text.Resize(4, 4)
	if w, h := text.Width(), text.Height(); w != 4 || h != 4 {
		t.Fatalf("expected 4x4, got %dx%d", w, h)
	}
	text.WriteString("abcdefghijklmnop")
	t.Run("ScrollUp", func(t *testing.T) {
		text.ScrollUp()
		want := "efgh\nijkl\nmnop\n    \n"
		if got := text.String(); got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
		if x, y := text.Position(); x != 0 || y != 3 {
			t.Fatalf("expected cursor at (0, 3), got (%d, %d)", x, y)
		}

		t.Run("Goto", func(t *testing.T) {
			text.Goto(0, 0)
			text.WriteString("qrst")
			want := "qrst\nijkl\nmnop\n    \n"
			if got := text.String(); got != want {
				t.Fatalf("expected %q, got %q", want, got)
			}
			if x, y := text.Position(); x != 0 || y != 1 {
				t.Fatalf("expected cursor at (0, 1), got (%d, %d)", x, y)
			}

			t.Run("Goto", func(t *testing.T) {
				text.Goto(0, 2)
				text.WriteString("uvwxyz01")
				want := "qrst\nijkl\nuvwx\nyz01\n"
				if got := text.String(); got != want {
					t.Fatalf("expected %q, got %q", want, got)
				}
				if x, y := text.Position(); x != 0 || y != 4 {
					t.Fatalf("expected cursor at (0, 4), got (%d, %d)", x, y)
				}

				t.Run("ScrollDown", func(t *testing.T) {
					text.ScrollDown()
					want := "    \nqrst\nijkl\nuvwx\n"
					if got := text.String(); got != want {
						t.Fatalf("expected %q, got %q", want, got)
					}
					if x, y := text.Position(); x != 0 || y != 3 {
						t.Fatalf("expected cursor at (0, 3), got (%d, %d)", x, y)
					}
				})
			})
		})
	})
}

func TestTextWrites(t *testing.T) {
	t.Run("WriteRune", func(t *testing.T) {
		text := NewText(4, 4)
		for i := 0; i < 4*4; i++ {
			text.WriteCharacter('a' + byte(i))
		}
		want := "abcd\nefgh\nijkl\nmnop\n"
		if got := text.String(); got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("WriteRune, scroll, no auto expand", func(t *testing.T) {
		text := NewText(4, 4)
		for i := 0; i < 4*6; i++ {
			text.WriteCharacter('a' + byte(i))
		}
		want := "ijkl\nmnop\nqrst\nuvwx\n"
		if got := text.String(); got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("WriteRune, no scroll, no auto expand", func(t *testing.T) {
		text := NewText(4, 4)
		//text.DisableScrolling = true
		for i := 0; i < 4*5; i++ {
			text.WriteCharacter('a' + byte(i))
		}
		want := "abcd\nefgh\nijkl\nmnop\n"
		if got := text.String(); got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("WriteRune, scroll, auto expand", func(t *testing.T) {
		text := NewText(4, 4)
		text.AutoExpand = true
		for i := 0; i < 4*5; i++ {
			text.WriteCharacter('a' + byte(i))
		}
		want := "abcd\nefgh\nijkl\nmnop\nqrst\n"
		if got := text.String(); got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("WriteString", func(t *testing.T) {
		text := NewText(4, 4)
		text.WriteString("abcdefghijklmnop")
		want := "abcd\nefgh\nijkl\nmnop\n"
		if got := text.String(); got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})
}
