//go:build ignore

// list_dataflows.go - List all data flows in a DTSX file
//
// This example demonstrates how to list all data flow tasks in a DTSX package.
//
// Usage: go run examples/list_dataflows.go <dtsx_file>
//
// Data flows are identified by ExecutableTypeAttr == "STOCK:Pipeline"
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/7045kHz/dtsx"
	schema "github.com/7045kHz/dtsx/schemas"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run list_dataflows.go <dtsx_file>")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Load the DTSX package
	pkg, err := dtsx.UnmarshalFromFile(filename)
	if err != nil {
		log.Fatalf("Failed to load DTSX file: %v", err)
	}

	if len(pkg.Executable) == 0 {
		fmt.Printf("No executables found in %s\n", filename)
		return
	}

	// Query for data flow tasks (try different possible types)
	dataFlows := pkg.QueryExecutables(func(exec *schema.AnyNonPackageExecutableType) bool {
		return exec.ExecutableTypeAttr == "STOCK:Pipeline" || exec.ExecutableTypeAttr == "Microsoft.Pipeline"
	})

	fmt.Printf("Found %d data flow(s) in %s:\n", len(dataFlows), filename)
	for i, df := range dataFlows {
		// Extract the data flow name from attributes
		name := "unnamed"
		if df.ObjectNameAttr != nil {
			name = *df.ObjectNameAttr
		}
		fmt.Printf("%d. %s\n", i+1, name)
	}
}
