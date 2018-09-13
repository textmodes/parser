package xbin

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/textmodes/parser/format/sauce"
)

func TestDecode(t *testing.T) {
	names, err := filepath.Glob("testdata/*.xb")
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
			DataType: sauce.XBIN,
			FileType: 80 >> 1,
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

	t.Run("Palette", func(t *testing.T) {
		b := new(bytes.Buffer)
		h := Header{
			ID:      [4]byte{'X', 'B', 'I', 'N'},
			Flags:   FlagPalette,
			EOFChar: 0x1a,
		}
		binary.Write(b, binary.LittleEndian, h)
		r := &sauce.Record{
			DataType: sauce.XBIN,
			FileType: 80 >> 1,
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

	t.Run("Font", func(t *testing.T) {
		b := new(bytes.Buffer)
		h := Header{
			ID:      [4]byte{'X', 'B', 'I', 'N'},
			Flags:   FlagFont,
			EOFChar: 0x1a,
		}
		binary.Write(b, binary.LittleEndian, h)
		r := &sauce.Record{
			DataType: sauce.XBIN,
			FileType: 80 >> 1,
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

	t.Run("Uncompressed", func(t *testing.T) {
		b := new(bytes.Buffer)
		h := Header{
			ID:      [4]byte{'X', 'B', 'I', 'N'},
			EOFChar: 0x1a,
			Width:   80,
			Height:  25,
		}
		binary.Write(b, binary.LittleEndian, h)
		r := &sauce.Record{
			DataType: sauce.XBIN,
			FileType: 80 >> 1,
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
}
