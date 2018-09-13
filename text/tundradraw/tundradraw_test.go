package tundradraw

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/textmodes/parser/format/sauce"
	"github.com/textmodes/parser/format/vga"
)

func TestDecode(t *testing.T) {
	names, err := filepath.Glob("testdata/*.tnd")
	if err != nil {
		if os.IsNotExist(err) {
			t.Skip(err)
		}
		t.Fatal(err)
	}

	for _, name := range names {
		t.Run(filepath.Base(name), func(t *testing.T) {
			testDecode(t, name)
		})
	}
}

func testDecode(t *testing.T, name string) {
	t.Helper()

	f, err := os.Open(name)
	if err != nil {
		t.Fatalf("%s open error: %v", name, err)
	}
	defer f.Close()

	if _, err = Decode(f); err != nil {
		t.Fatal(err)
	}
}

func TestDecodeError(t *testing.T) {
	t.Run("ID", func(t *testing.T) {
		b := new(bytes.Buffer)
		r := &sauce.Record{
			DataType: sauce.Character,
			FileType: sauce.TundraDraw,
		}
		b.WriteByte(0x1a)
		r.WriteTo(b)

		f := bytes.NewReader(b.Bytes())
		if _, err := Decode(f); err != nil {
			t.Logf("expected error: %v", err)
			return
		}
		t.Fatal("expected error")
	})

	t.Run("Position", func(t *testing.T) {
		b := new(bytes.Buffer)
		r := &sauce.Record{
			DataType: sauce.Character,
			FileType: sauce.TundraDraw,
		}
		b.WriteString(tundraDrawID)
		b.WriteByte(tundraDrawPos)
		b.WriteByte(0x1a)
		r.WriteTo(b)

		vga.Trace = true

		f := bytes.NewReader(b.Bytes())
		if _, err := Decode(f); err != io.EOF {
			t.Fatalf("expected error %q, got %q", io.EOF, err)
		} else {
			t.Logf("expected error: %v", err)
		}
	})

	t.Run("Foreground", func(t *testing.T) {
		b := new(bytes.Buffer)
		r := &sauce.Record{
			DataType: sauce.Character,
			FileType: sauce.TundraDraw,
		}
		b.WriteString(tundraDrawID)
		b.WriteByte(tundraDrawForeground)
		b.WriteByte(0x1a)
		r.WriteTo(b)

		vga.Trace = true

		f := bytes.NewReader(b.Bytes())
		if _, err := Decode(f); err != io.EOF {
			t.Fatalf("expected error %q, got %q", io.EOF, err)
		} else {
			t.Logf("expected error: %v", err)
		}
	})

	t.Run("DecodeForeground", func(t *testing.T) {
		b := new(bytes.Buffer)
		r := &sauce.Record{
			DataType: sauce.Character,
			FileType: sauce.TundraDraw,
		}
		b.WriteString(tundraDrawID)
		b.WriteByte(tundraDrawForeground)
		b.WriteByte(0x00)
		b.WriteByte(0x1a)
		r.WriteTo(b)

		vga.Trace = true

		f := bytes.NewReader(b.Bytes())
		if _, err := Decode(f); err != io.EOF {
			t.Fatalf("expected error %q, got %q", io.EOF, err)
		} else {
			t.Logf("expected error: %v", err)
		}
	})

	t.Run("DecodeBackground", func(t *testing.T) {
		b := new(bytes.Buffer)
		r := &sauce.Record{
			DataType: sauce.Character,
			FileType: sauce.TundraDraw,
		}
		b.WriteString(tundraDrawID)
		b.WriteByte(tundraDrawBackground)
		b.WriteByte(0x00)
		b.WriteByte(0x1a)
		r.WriteTo(b)

		vga.Trace = true

		f := bytes.NewReader(b.Bytes())
		if _, err := Decode(f); err != io.EOF {
			t.Fatalf("expected error %q, got %q", io.EOF, err)
		} else {
			t.Logf("expected error: %v", err)
		}
	})
}
