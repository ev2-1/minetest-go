package main

import (
	"bytes"
	"fmt"
	_ "github.com/anon55555/mt"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

const pkg = "github.com/anon55555/mt"

func init() {
	gopath := filepath.Join(os.Getenv("HOME"), "go")

	if os.Getenv("GOPATH") != "" {
		gopath = os.Getenv("GOPATH")
	}

	path := filepath.Join(gopath, "src", pkg)

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	p := pkgs[filepath.Base(pkg)]

	dp := doc.New(p, pkg, doc.AllDecls)

	w := bytes.NewBufferString("")

	writeDoc(w, dp)

	fmt.Println(w.String())
}

func writeDoc(w *bytes.Buffer, mt *doc.Package) {
	for _, tp := range mt.Types {
		w.WriteString(tp.Name)
	}
}
