package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGetDependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		depth           int
		expDependencies map[string]map[string]bool
	}{
		{0, make(map[string]map[string]bool)},
		{
			depth: 1,
			expDependencies: map[string]map[string]bool{
				"github.com/wangkekekexili/gotrain/test_apple": {`"github.com/wangkekekexili/gotrain/test_apple/test_banana"`: true},
			},
		},
		{
			depth: 2,
			expDependencies: map[string]map[string]bool{
				"github.com/wangkekekexili/gotrain/test_apple":             {`"github.com/wangkekekexili/gotrain/test_apple/test_banana"`: true},
				"github.com/wangkekekexili/gotrain/test_apple/test_banana": {`"fmt"`: true},
			},
		},
		{
			depth: 3,
			expDependencies: map[string]map[string]bool{
				"github.com/wangkekekexili/gotrain/test_apple":             {`"github.com/wangkekekexili/gotrain/test_apple/test_banana"`: true},
				"github.com/wangkekekexili/gotrain/test_apple/test_banana": {`"fmt"`: true},
				"fmt": {},
			},
		},
		{
			depth: 4,
			expDependencies: map[string]map[string]bool{
				"github.com/wangkekekexili/gotrain/test_apple":             {`"github.com/wangkekekexili/gotrain/test_apple/test_banana"`: true},
				"github.com/wangkekekexili/gotrain/test_apple/test_banana": {`"fmt"`: true},
				"fmt": {},
			},
		},
	}

	for _, test := range tests {
		gotDependencies := make(map[string]map[string]bool)
		srcDir := filepath.Join(os.Getenv("GOPATH"), "src")
		if err := getDependencies(srcDir, "github.com/wangkekekexili/gotrain/test_apple", gotDependencies, test.depth); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(test.expDependencies, gotDependencies) {
			t.Fatalf("expected to get %v; got %v", test.expDependencies, gotDependencies)
		}
	}
}

func callCallerFunctionNameInside() string {
	return callerFunctionName(1)
}

func TestCallerFunctionName(t *testing.T) {
	t.Parallel()

	// The result should be deterministic.
	if callerFunctionName(0) != callerFunctionName(0) {
		t.Fatal("callerFunctionName should be deterministic")
	}

	if callerFunctionName(0) != callCallerFunctionNameInside() {
		t.Fatalf("expected to get %v; got %v", callerFunctionName(0), callCallerFunctionNameInside())
	}
}

func TestPrintDigraph(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	dep := map[string]map[string]bool{
		"apple": {`"banana"`: true, `"peach"`: true},
	}
	expStr := "\"apple\" \"banana\" \"peach\" \n"

	printDigraph(&buf, dep)

	gotStr := buf.String()
	if gotStr != expStr {
		t.Fatalf("expected %v; got %v", expStr, gotStr)
	}
}

func TestPrintGraphviz(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	dep := map[string]map[string]bool{
		"apple": {`"banana"`: true, `"peach"`: true},
	}
	expStr := `digraph G {
"apple"->"banana";
"apple"->"peach";
}
`

	printGraphviz(&buf, dep)

	gotStr := buf.String()
	if gotStr != expStr {
		t.Fatalf("expected %v; got %v", expStr, gotStr)
	}
}
