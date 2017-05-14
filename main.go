package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)

const (
	formatDigraph  = "digraph"
	formatGraphviz = "graphviz"
)

func main() {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		logger.Fatal("GOPATH must be set")
	}

	depth := flag.Int("depth", 2, "Max depth of dependency tree.")
	format := flag.String("format", "digraph", "Output format for the dependency graph. Can be one of graphviz, digraph. Defaults to digraph.")
	flag.Parse()

	// Validate arguments.
	if len(flag.Args()) != 1 {
		logger.Fatal("package name must be specified")
	}
	importPath := flag.Arg(0)
	if *format != formatDigraph && *format != formatGraphviz {
		flag.Usage()
		return
	}

	dependencies := make(map[string][]string)
	if err := getDependencies(filepath.Join(goPath, "src"), importPath, dependencies, *depth); err != nil {
		logger.Fatal(err)
	}

	printGraph(*format, dependencies)
}

// getDependencies populates dependencies recursively.
//
// srcDir is the root directory for all golang source code.
// importPath is like github.com/google/btree. The getDependencies call will populate dependencies which btree package depends on.
func getDependencies(srcDir, importPath string, dependencies map[string][]string, maxDepth int) error {
	if callerFunctionName(maxDepth) == callerFunctionName(0) {
		return nil
	}
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
		if err := getDependencies(srcDir, n, dependencies, maxDepth); err != nil {
			return err
		}
	}

	return nil
}

func callerFunctionName(depth int) string {
	pc, _, _, ok := runtime.Caller(depth + 1)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		return details.Name()
	}
	return "<unknown>"
}

// printGraph outputs the dependency graph to standard output in the specified format.
func printGraph(format string, dependencies map[string][]string) {
	switch format {
	case formatDigraph:
		printDigraph(os.Stdout, dependencies)
	case formatGraphviz:
		printGraphviz(os.Stdout, dependencies)
	}
}

func printDigraph(w io.Writer, dependencies map[string][]string) {
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

func printGraphviz(w io.Writer, dependencies map[string][]string) {
	fmt.Fprintln(w, "digraph G {")
	for from, tos := range dependencies {
		if len(tos) == 0 {
			continue
		}
		for _, to := range tos {
			fmt.Fprintf(w, "%s->%s;\n", from, to)
		}
	}
	fmt.Fprintln(w, "}")
}
