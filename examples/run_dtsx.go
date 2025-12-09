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
		fmt.Println("Usage: go run examples/run_dtsx.go <path-to-dtsx-file> [dtexec-path]")
		fmt.Println("\nExample:")
		fmt.Println(`  go run examples/run_dtsx.go MyPackage.dtsx`)
		fmt.Println(`  go run examples/run_dtsx.go MyPackage.dtsx "C:\Program Files\Microsoft SQL Server\160\DTS\Binn\DTExec.exe"`)
		os.Exit(1)
	}

	dtsxFile := os.Args[1]

	// Default DTExec path (SQL Server 2022)
	dtexecPath := `C:\Program Files\Microsoft SQL Server\160\DTS\Binn\DTExec.exe`

	// Allow custom DTExec path as second argument
	if len(os.Args) >= 3 {
		dtexecPath = os.Args[2]
	}

	// Verify DTExec exists
	if _, err := os.Stat(dtexecPath); os.IsNotExist(err) {
		log.Fatalf("DTExec.exe not found at: %s\nPlease specify the correct path as the second argument", dtexecPath)
	}

	// Validate the DTSX package first
	fmt.Printf("Validating package: %s\n", dtsxFile)
	_, ok := dtsx.IsDTSXPackage(dtsxFile)
	if !ok {
		log.Fatalf("File is not a valid DTSX package: %s", dtsxFile)
	}
	fmt.Println("✓ Package is valid\n")

	// Configure execution options
	opts := &dtsx.RunOptions{
		ReportingLevel: "V", // Verbose reporting
		WarnAsError:    false,
	}

	// Execute the package
	fmt.Printf("Executing package with DTExec: %s\n\n", dtexecPath)
	fmt.Println("--- Package Execution Output ---")

	output, err := dtsx.RunPackage(dtexecPath, dtsxFile, opts)

	fmt.Println(output)
	fmt.Println("--- End of Output ---\n")

	if err != nil {
		log.Fatalf("Package execution failed: %v", err)
	}

	fmt.Println("✓ Package executed successfully!")
}
