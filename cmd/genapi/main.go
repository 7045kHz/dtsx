package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// This tool scans the dtsx package files (excluding schemas/ and examples/) and generates
// docs/API_symbols.md listing exported types/functions/methods with signatures and comments.

func main() {
	out := flag.String("out", "docs/API_symbols.md", "output file")
	dir := flag.String("dir", ".", "directory to scan (repo root)")
	flag.Parse()

	fset := token.NewFileSet()
	files := map[string]*ast.File{}
	var filenames []string

	// Walk files in dir
	err := filepath.WalkDir(*dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip directories we don't want to parse
		if d.IsDir() {
			name := d.Name()
			if name == "examples" || name == "schemas" || name == "docs" || name == "cmd" || name == "vendor" || name == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}
		// Read/parsethe file
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			// Skip files that do not parse
			return nil
		}
		files[path] = file
		filenames = append(filenames, path)
		return nil
	})
	if err != nil {
		log.Fatalf("walk error: %v", err)
	}

	// Collect exported symbols
	types := map[string]string{}
	funcs := map[string]string{}
	methods := map[string]map[string]string{}
	docs := map[string]string{}

	for _, file := range files {
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				if d.Tok == token.TYPE {
					for _, spec := range d.Specs {
						spec := spec.(*ast.TypeSpec)
						name := spec.Name.Name
						if ast.IsExported(name) {
							var b bytes.Buffer
							_ = printer.Fprint(&b, fset, &ast.GenDecl{Tok: token.TYPE, Specs: []ast.Spec{spec}})
							types[name] = strings.TrimSpace(b.String())
							if d.Doc != nil {
								docs[name] = strings.TrimSpace(d.Doc.Text())
							}
						}
					}
				}
			case *ast.FuncDecl:
				name := d.Name.Name
				if d.Recv == nil {
					// standalone function
					if ast.IsExported(name) {
						var b bytes.Buffer
						_ = printer.Fprint(&b, fset, &ast.FuncDecl{Type: d.Type, Name: d.Name})
						funcs[name] = strings.TrimSpace(b.String())
						if d.Doc != nil {
							docs[name] = strings.TrimSpace(d.Doc.Text())
						}
					}
				} else {
					// method
					recv := d.Recv.List[0].Type
					var recvType string
					switch r := recv.(type) {
					case *ast.StarExpr:
						if id, ok := r.X.(*ast.Ident); ok {
							recvType = id.Name
						}
					case *ast.Ident:
						recvType = r.Name
					}
					if recvType != "" && ast.IsExported(d.Name.Name) {
						if _, ok := methods[recvType]; !ok {
							methods[recvType] = map[string]string{}
						}
						var b bytes.Buffer
						_ = printer.Fprint(&b, fset, d)
						methods[recvType][d.Name.Name] = strings.TrimSpace(b.String())
						if d.Doc != nil {
							docs[recvType+"."+d.Name.Name] = strings.TrimSpace(d.Doc.Text())
						}
					}
				}
			}
		}
	}

	// Build output
	var outBuf bytes.Buffer
	outBuf.WriteString("# Generated API symbols (dtsx package)\n\n")
	outBuf.WriteString("This file is auto-generated. To regenerate, run: `go generate` in the repo root.\n\n")

	// Types
	if len(types) > 0 {
		outBuf.WriteString("## Exported types\n\n")
		// sort keys
		keys := make([]string, 0, len(types))
		for k := range types {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			outBuf.WriteString(fmt.Sprintf("### %s\n\n", k))
			if docs[k] != "" {
				outBuf.WriteString(docs[k] + "\n\n")
			}
			outBuf.WriteString("```go\n")
			outBuf.WriteString(types[k] + "\n")
			outBuf.WriteString("```\n\n")
		}
	}

	// Functions
	if len(funcs) > 0 {
		outBuf.WriteString("## Exported functions\n\n")
		keys := make([]string, 0, len(funcs))
		for k := range funcs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			outBuf.WriteString(fmt.Sprintf("### %s\n\n", k))
			if docs[k] != "" {
				outBuf.WriteString(docs[k] + "\n\n")
			}
			outBuf.WriteString("```go\n")
			outBuf.WriteString(funcs[k] + "\n")
			outBuf.WriteString("```\n\n")
		}
	}

	// Methods
	if len(methods) > 0 {
		outBuf.WriteString("## Methods on exported types\n\n")
		recvKeys := make([]string, 0, len(methods))
		for k := range methods {
			recvKeys = append(recvKeys, k)
		}
		sort.Strings(recvKeys)
		for _, recv := range recvKeys {
			outBuf.WriteString(fmt.Sprintf("### %s\n\n", recv))
			mkeys := make([]string, 0, len(methods[recv]))
			for mk := range methods[recv] {
				mkeys = append(mkeys, mk)
			}
			sort.Strings(mkeys)
			for _, mk := range mkeys {
				outBuf.WriteString(fmt.Sprintf("#### %s\n\n", mk))
				if docs[recv+"."+mk] != "" {
					outBuf.WriteString(docs[recv+"."+mk] + "\n\n")
				}
				outBuf.WriteString("```go\n")
				outBuf.WriteString(methods[recv][mk] + "\n")
				outBuf.WriteString("```\n\n")
			}
		}
	}

	// Write file
	if err := os.MkdirAll(filepath.Dir(*out), 0755); err != nil {
		log.Fatalf("mkdir error: %v", err)
	}
	if err := os.WriteFile(*out, outBuf.Bytes(), 0644); err != nil {
		log.Fatalf("write error: %v", err)
	}
	fmt.Printf("generated: %s\n", *out)
}
