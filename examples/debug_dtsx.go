//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/7045kHz/dtsx"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run examples/debug_dtsx.go <path-to-dtsx-file>")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Validate and read the DTSX file using the new IsDTSXPackage function
	pkg, ok := dtsx.IsDTSXPackage(filename)
	if !ok {
		log.Fatalf("File is not a valid DTSX package: %s", filename)
	}

	// Convert to JSON for inspection
	jsonData, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		log.Fatalf("Error converting to JSON: %v", err)
	}

	fmt.Println(string(jsonData))
}
