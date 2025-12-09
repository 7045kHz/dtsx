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
		fmt.Println("Usage: go run examples/read_dtsx.go <path-to-dtsx-file>")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Validate and read the DTSX file using the new IsDTSXPackage function
	pkg, ok := dtsx.IsDTSXPackage(filename)
	if !ok {
		log.Fatalf("File is not a valid DTSX package: %s", filename)
	}

	// Print basic package information
	fmt.Printf("Successfully loaded DTSX package from: %s\n", filename)

	if pkg.ExecutableTypePackage != nil {
		if pkg.Property != nil && len(pkg.Property) > 0 {
			fmt.Printf("\nPackage Properties (%d):\n", len(pkg.Property))
			for i, prop := range pkg.Property {
				if i < 5 { // Show first 5 properties
					name := "unknown"
					if prop.NameAttr != nil {
						name = *prop.NameAttr
					}
					fmt.Printf("  - %s\n", name)
				}
			}
			if len(pkg.Property) > 5 {
				fmt.Printf("  ... and %d more\n", len(pkg.Property)-5)
			}
		}

		if pkg.ConnectionManagers != nil && pkg.ConnectionManagers.ConnectionManager != nil && len(pkg.ConnectionManagers.ConnectionManager) > 0 {
			fmt.Printf("\nConnection Managers: %d\n", len(pkg.ConnectionManagers.ConnectionManager))
		}

		if pkg.Variables != nil && pkg.Variables.Variable != nil && len(pkg.Variables.Variable) > 0 {
			fmt.Printf("Variables: %d\n", len(pkg.Variables.Variable))
		}

		if pkg.Executable != nil && len(pkg.Executable) > 0 {
			fmt.Printf("Executables: %d\n", len(pkg.Executable))
		}
	}

	// Optionally, marshal back to XML and save to a new file
	// outputFile := filename + ".output.dtsx"
	// err = dtsx.MarshalToFile(outputFile, pkg)
	// if err != nil {
	// 	log.Fatalf("Error writing DTSX file: %v", err)
	// }
	// fmt.Printf("\nSaved to: %s\n", outputFile)
}
