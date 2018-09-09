package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type dir struct {
	name string
}

type file struct {
	name    string
	size    int64
	modTime int64
	data    []byte
}

func include(b []byte, dirs ...string) (out []byte, err error) {
	out = b

	var (
		outDirs  = make(map[string]*dir)
		outFiles = make(map[string]*file)
		outLinks = make(map[string]string)
	)

	for _, root := range dirs {
		if err = filepath.Walk(root, func(name string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			/*
				var rel string
				if rel, err = filepath.Rel(root, name); err != nil {
					return err
				}
			*/
			if info.IsDir() {
				outDirs[name] = &dir{name}
			} else if info.Mode()&os.ModeSymlink != 0 {
				var link string
				if link, err = os.Readlink(name); err != nil {
					return err
				}
				outLinks[name] = filepath.Join(filepath.Dir(name), link)
				fmt.Fprintf(os.Stderr, "%s: link to %s\n", name, outLinks[name])
			} else {
				var f = &file{
					name:    name,
					size:    info.Size(),
					modTime: info.ModTime().Unix(),
				}
				if f.data, err = ioutil.ReadFile(name); err != nil {
					return err
				}
				outFiles[f.name] = f
				fmt.Fprintf(os.Stderr, "%s: file\n", f.name)
			}
			return nil
		}); err != nil {
			return
		}
	}

	var (
		bufDirs  = new(bytes.Buffer)
		bufFiles = new(bytes.Buffer)
		bufLinks = new(bytes.Buffer)
	)
	for name, d := range outDirs {
		fmt.Fprintf(bufDirs, "%q: %#+v,\n", name, d)
	}
	for name, f := range outFiles {
		fmt.Fprintf(bufFiles, "%q: %#+v,\n", name, f)
	}
	for name, t := range outLinks {
		fmt.Fprintf(bufLinks, "files[%q] = files[%q]\n", name, t)
	}

	out = bytes.Replace(out, []byte("\t// INJECT DIRS HERE\n"), bufDirs.Bytes(), 1)
	out = bytes.Replace(out, []byte("\t// INJECT FILES HERE\n"), bufFiles.Bytes(), 1)
	out = bytes.Replace(out, []byte("\t// INJECT LINKS HERE\n"), bufLinks.Bytes(), 1)
	out = bytes.Replace(out, []byte("main.dir"), []byte("dir"), -1)
	out = bytes.Replace(out, []byte("main.file"), []byte("file"), -1)

	return
}

func main() {
	outputFile := flag.String("o", "", "output file")
	flag.Parse()

	var w io.Writer = os.Stdout
	if *outputFile != "" {
		f, err := os.Create(*outputFile)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		w = f
	}

	b, err := ioutil.ReadFile("data_template.go")
	if err != nil {
		panic(err)
	}

	// Replace build tag with generate tag
	b = bytes.Replace(b,
		[]byte("// +build ignore"),
		[]byte("//go:generate go run testdata/data_pack.go -o data.go"),
		1,
	)

	if b, err = include(b, "font"); err != nil {
		panic(err)
	}

	formatted, err := format.Source(b)
	if err != nil {
		fmt.Fprintln(os.Stderr, string(b))
		panic(err)
	}

	if _, err = w.Write(formatted); err != nil {
		panic(err)
	}
}
