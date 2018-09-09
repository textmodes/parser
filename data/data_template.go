// +build ignore

package data

import (
	"bytes"
	"os"
	"path"
	"time"
)

// Bytes of a named file.
func Bytes(name string) ([]byte, error) {
	f, ok := files[path.Clean(name)]
	if !ok {
		return nil, os.ErrNotExist
	}
	return f.data, nil
}

// Open a named file.
func Open(name string) (*bytes.Reader, error) {
	f, ok := files[path.Clean(name)]
	if !ok {
		return nil, os.ErrNotExist
	}
	return bytes.NewReader(f.data), nil
}

// Stat a directory or file with name.
func Stat(name string) (os.FileInfo, error) {
	name = path.Clean(name)
	if d, ok := dirs[name]; ok {
		return d, nil
	}
	if f, ok := files[name]; ok {
		return f, nil
	}
	return nil, os.ErrNotExist
}

type dir struct {
	name string
}

func (dir) IsDir() bool          { return true }
func (d dir) Name() string       { return d.name }
func (d dir) Size() int64        { return 0 }
func (d dir) ModTime() time.Time { return time.Unix(0, 0) }
func (d dir) Mode() os.FileMode  { return 0755 }
func (d dir) Sys() interface{}   { return d }

type file struct {
	name    string
	size    int64
	modTime int64
	data    []byte
}

func (file) IsDir() bool          { return false }
func (f file) Name() string       { return f.name }
func (f file) Size() int64        { return f.size }
func (f file) ModTime() time.Time { return time.Unix(f.modTime, 0) }
func (f file) Mode() os.FileMode  { return 0644 }
func (f file) Sys() interface{}   { return f }

var (
	// Compile time interface checks
	_ os.FileInfo = (*dir)(nil)
	_ os.FileInfo = (*file)(nil)

	dirs = map[string]*dir{
		// INJECT DIRS HERE
	}
	files = map[string]*file{
		// INJECT FILES HERE
	}
)

func init() {
	// INJECT LINKS HERE
}
