//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/7045kHz/dtsx"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run examples/run_with_params.go <path-to-dtsx-file>")
		fmt.Println("\nThis example demonstrates running a package with parameters and options.")
		os.Exit(1)
	}

	dtsxFile := os.Args[1]
	dtexecPath := `C:\Program Files\Microsoft SQL Server\160\DTS\Binn\DTExec.exe`

	// Verify DTExec exists
	if _, err := os.Stat(dtexecPath); os.IsNotExist(err) {
		log.Fatalf("DTExec.exe not found at: %s", dtexecPath)
	}

	// Configure execution with various options
	opts := &dtsx.RunOptions{
		// Package parameters (customize based on your package)
		Parameters: []string{
			"$Package::MyParameter;MyValue",
			"$Project::ProjectParam(Int32);42",
		},

		// Environment variables
		EnvironmentVars: []string{
			"MyEnvVar=TestValue",
		},

		// Connection manager overrides (example)
		Connections: []string{
			// "ConnectionName;Data Source=localhost;Initial Catalog=MyDB;Integrated Security=SSPI;",
		},

		// Property overrides
		PropertySets: []string{
			// `\Package.Variables[User::MyVariable].Value;NewValue`,
		},

		// Execution options
		ReportingLevel: "I",   // Info level reporting
		WarnAsError:    false, // Don't treat warnings as errors
		MaxConcurrent:  -1,    // Auto-determine max concurrent executables
		Validate:       false, // Execute (not just validate)

		// Checkpoint options (if needed)
		// Checkpointing:  "on",
		// CheckpointFile: "checkpoint.xml",
		// Restart:        "ifPossible",
	}

	fmt.Printf("Executing package: %s\n", dtsxFile)
	fmt.Printf("DTExec path: %s\n\n", dtexecPath)

	fmt.Println("Configuration:")
	fmt.Printf("  Parameters: %d\n", len(opts.Parameters))
	fmt.Printf("  Environment Variables: %d\n", len(opts.EnvironmentVars))
	fmt.Printf("  Reporting Level: %s\n", opts.ReportingLevel)
	fmt.Printf("  Max Concurrent: %d\n\n", opts.MaxConcurrent)

	fmt.Println("--- Package Execution Output ---")
	output, err := dtsx.RunPackage(dtexecPath, dtsxFile, opts)

	fmt.Println(output)
	fmt.Println("--- End of Output ---\n")

	if err != nil {
		log.Fatalf("Package execution failed: %v", err)
	}

	fmt.Println("✓ Package executed successfully with parameters!")
}
