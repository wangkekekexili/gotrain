package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)

func main() {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		logger.Fatal("GOPATH must be set")
	}
	if len(os.Args) != 2 {
		logger.Fatal("usage: gotrain [package]")
	}
	importPath := os.Args[1]

	dependencies := make(map[string][]string)
	if err := getDependencies(filepath.Join(goPath, "src"), importPath, dependencies); err != nil {
		logger.Fatal(err)
	}

	printGraph(os.Stdout, dependencies)
}

// getDependencies populates dependencies recursively.
//
// srcDir is the root directory for all golang source code.
// importPath is like github.com/google/btree. The getDependencies call will populate dependencies which btree package depends on.
func getDependencies(srcDir, importPath string, dependencies map[string][]string) error {
	if dependencies[importPath] != nil {
		return nil
	}
	dependencies[importPath] = []string{}

	// Stop if the directory doesn't exist.
	// It could be because it's an built-in package or the package hasn't been downloaded.
	directory := filepath.Join(srcDir, importPath)
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return err
	}

	next := make(map[string]bool)
	for _, file := range files {
		if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
			continue
		}
		filename := file.Name()
		if filepath.Ext(filename) == ".go" {
			ast, err := parser.ParseFile(token.NewFileSet(), filepath.Join(directory, filename), nil, parser.ImportsOnly)
			if err != nil {
				return err
			}
			for _, im := range ast.Imports {
				nextImportPath := im.Path.Value
				dependencies[strconv.Quote(importPath)] = append(dependencies[strconv.Quote(importPath)], nextImportPath)

				nextImportPathUnquoted, err := strconv.Unquote(nextImportPath)
				if err != nil {
					logger.Println(err)
					continue
				}
				next[nextImportPathUnquoted] = true
			}
		}
	}

	for n := range next {
		if err := getDependencies(srcDir, n, dependencies); err != nil {
			return err
		}
	}

	return nil
}

func printGraph(w io.Writer, dependencies map[string][]string) {
	for from, tos := range dependencies {
		if len(tos) == 0 {
			continue
		}
		fmt.Fprint(w, from, " ")
		for _, to := range tos {
			fmt.Fprint(w, to, " ")
		}
		fmt.Fprintln(w)
	}
}
