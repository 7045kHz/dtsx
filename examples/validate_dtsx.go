//go:build ignore

// validate_dtsx.go - Comprehensive DTSX Package Validation
//
// This example demonstrates how to validate DTSX packages for common issues,
// orphaned variables, structural problems, and best practices.
//
// Usage: go run examples/validate_dtsx.go <dtsx_file>
//
// Features:
// - Comprehensive validation with error/warning/info severities
// - Orphaned variable detection
// - Dependency analysis
// - Optimization suggestions

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/7045kHz/dtsx"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run examples/validate_dtsx.go <path-to-dtsx-file>")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Validate if the file is a DTSX package
	pkg, ok := dtsx.IsDTSXPackage(filename)
	if !ok {
		log.Fatalf("File is not a valid DTSX package: %s", filename)
	}

	fmt.Printf("✓ Valid DTSX package loaded: %s\n\n", filename)

	// Display basic package information
	if pkg.ExecutableTypePackage != nil {
		fmt.Printf("Package Details:\n")
		fmt.Printf("  Properties: %d\n", len(pkg.Property))
		if pkg.ConnectionManagers != nil {
			fmt.Printf("  Connection Managers: %d\n", len(pkg.ConnectionManagers.ConnectionManager))
		}
		if pkg.Variables != nil {
			fmt.Printf("  Variables: %d\n", len(pkg.Variables.Variable))
		}
		fmt.Printf("  Executables/Tasks: %d\n", len(pkg.Executable))
		fmt.Printf("  Precedence Constraints: %d\n", len(pkg.PrecedenceConstraint))

		// Show executable type if available
		if pkg.ExecutableTypeAttr != nil {
			fmt.Printf("  Executable Type: %s\n", *pkg.ExecutableTypeAttr)
		}
	}

	fmt.Println("\n=== Comprehensive Package Validation ===")

	// Run comprehensive validation using PackageValidator
	validator := dtsx.NewPackageValidator(pkg)
	errors := validator.Validate()

	if len(errors) == 0 {
		fmt.Println("✓ No validation issues found!")
		fmt.Println("  Package structure is valid and follows best practices.")
	} else {
		// Group errors by severity
		errorCount := 0
		warningCount := 0
		infoCount := 0

		for _, err := range errors {
			switch err.Severity {
			case "error":
				errorCount++
			case "warning":
				warningCount++
			case "info":
				infoCount++
			}
		}

		fmt.Printf("Validation Results: %d errors, %d warnings, %d info\n\n", errorCount, warningCount, infoCount)

		// Display errors
		if errorCount > 0 {
			fmt.Println("❌ ERRORS (must be fixed):")
			for _, err := range errors {
				if err.Severity == "error" {
					fmt.Printf("  • %s: %s\n", err.Path, err.Message)
				}
			}
			fmt.Println()
		}

		// Display warnings
		if warningCount > 0 {
			fmt.Println("⚠️  WARNINGS (should be reviewed):")
			for _, err := range errors {
				if err.Severity == "warning" {
					fmt.Printf("  • %s: %s\n", err.Path, err.Message)
				}
			}
			fmt.Println()
		}

		// Display info
		if infoCount > 0 {
			fmt.Println("ℹ️  INFO (optimization suggestions):")
			for _, err := range errors {
				if err.Severity == "info" {
					fmt.Printf("  • %s: %s\n", err.Path, err.Message)
				}
			}
			fmt.Println()
		}
	}

	// Additional analysis
	fmt.Println("=== Additional Analysis ===")

	// Check for unused variables
	unused := pkg.GetUnusedVariables()
	if len(unused) > 0 {
		fmt.Printf("Unused Variables (%d):\n", len(unused))
		for _, varName := range unused {
			fmt.Printf("  • %s\n", varName)
		}
	} else {
		fmt.Println("✓ No unused variables found")
	}

	// Build dependency graph
	graph := pkg.BuildDependencyGraph()
	fmt.Printf("\nDependency Analysis:\n")
	fmt.Printf("  Variables defined: %d\n", len(pkg.Variables.Variable))
	fmt.Printf("  Connections used: %d\n", len(graph.ConnectionDependencies))

	// Show some dependency relationships
	if len(graph.VariableDependencies) > 0 {
		fmt.Printf("  Variable relationships: %d\n", len(graph.VariableDependencies))
	}

	// Get optimization suggestions
	suggestions := pkg.GetOptimizationSuggestions()
	if len(suggestions) > 0 {
		fmt.Printf("\nOptimization Suggestions (%d):\n", len(suggestions))
		for _, s := range suggestions {
			fmt.Printf("  • [%s] %s\n", s.Severity, s.Message)
		}
	} else {
		fmt.Println("\n✓ No optimization suggestions")
	}

	fmt.Println("\n=== Validation Complete ===")
	if len(errors) == 0 {
		fmt.Println("🎉 Package is fully validated and ready for production!")
	} else {
		fmt.Printf("📋 Package has %d issues that should be addressed.\n", len(errors))
	}
}
