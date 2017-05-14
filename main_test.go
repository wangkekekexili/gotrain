package main

import (
	"testing"
)

func callCallerFunctionNameInside() string {
	return callerFunctionName(1)
}

func TestCallerFunctionName(t *testing.T) {
	// The result should be deterministic.
	if callerFunctionName(0) != callerFunctionName(0) {
		t.Fatal("callerFunctionName should be deterministic")
	}

	if callerFunctionName(0) != callCallerFunctionNameInside() {
		t.Fatalf("expected to get %v; got %v", callerFunctionName(0), callCallerFunctionNameInside())
	}

}
