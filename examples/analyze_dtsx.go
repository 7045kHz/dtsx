//go:build ignore

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
		fmt.Println("Usage: go run examples/analyze_dtsx.go <path-to-dtsx-file>")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Validate and read the DTSX file using the new IsDTSXPackage function
	pkg, ok := dtsx.IsDTSXPackage(filename)
	if !ok {
		log.Fatalf("File is not a valid DTSX package: %s", filename)
	}

	fmt.Printf("=== DTSX Package Analysis ===\n")
	fmt.Printf("File: %s\n\n", filename)

	if pkg.ExecutableTypePackage == nil {
		fmt.Println("No package data found")
		return
	}

	// Show executable type
	if pkg.ExecutableTypeAttr != nil {
		fmt.Printf("Executable Type: %s\n", *pkg.ExecutableTypeAttr)
	}

	// Use new query methods to analyze the package

	// Properties (still need to access directly since no query method for properties)
	fmt.Printf("\n--- Properties (%d) ---\n", len(pkg.Property))
	for i, prop := range pkg.Property {
		if prop.NameAttr != nil {
			value := "nil"
			if prop.PropertyElementBaseType != nil && prop.PropertyElementBaseType.AnySimpleType != nil {
				value = prop.PropertyElementBaseType.AnySimpleType.Value
				if len(value) > 50 {
					value = value[:50] + "..."
				}
			}
			fmt.Printf("%d. %s = %s\n", i+1, *prop.NameAttr, value)
		}
	}

	// Connection Managers - using new query method
	connections := pkg.GetConnections()
	fmt.Printf("\n--- Connection Managers (%d) ---\n", connections.Count)
	if connections.Count > 0 {
		connMgrs := connections.Results.([]*schema.ConnectionManagerType)
		for i, cm := range connMgrs {
			name := dtsx.GetConnectionName(cm)
			fmt.Printf("%d. %s (%d properties)\n", i+1, name, len(cm.Property))
		}
	}

	// Variables - using new query method
	variables := pkg.GetVariables()
	fmt.Printf("\n--- Variables (%d) ---\n", variables.Count)
	if variables.Count > 0 {
		vars := variables.Results.([]*schema.VariableType)
		for i, variable := range vars {
			name := dtsx.GetVariableName(variable)
			fmt.Printf("%d. %s\n", i+1, name)
		}
	}

	// Executables (Tasks/Containers) - using new query method
	fmt.Printf("\n--- Executables/Tasks ---\n")

	// Get all executables
	allExecutables := pkg.QueryExecutables(func(*schema.AnyNonPackageExecutableType) bool { return true })
	fmt.Printf("Total Executables: %d\n", len(allExecutables))

	// Filter for specific types
	sqlTasks := pkg.QueryExecutables(func(exec *schema.AnyNonPackageExecutableType) bool {
		return exec.ExecutableTypeAttr == "STOCK:SQLTask"
	})
	if len(sqlTasks) > 0 {
		fmt.Printf("SQL Tasks: %d\n", len(sqlTasks))
	}

	// Show details of all executables
	for i, exec := range allExecutables {
		name := dtsx.GetExecutableName(exec)
		execType := exec.ExecutableTypeAttr
		fmt.Printf("%d. %s [%s]\n", i+1, name, execType)
	}

	// Execution Flow Analysis
	fmt.Printf("\n--- Execution Flow Analysis ---\n")
	analyzer := dtsx.NewPrecedenceAnalyzer(pkg)
	flowDesc := analyzer.GetExecutionFlowDescription()
	fmt.Print(flowDesc)

	// Expressions - using new query method
	expressions := pkg.GetExpressions()
	fmt.Printf("\n--- Expressions (%d) ---\n", expressions.Count)
	if expressions.Count > 0 {
		exprs := expressions.Results.([]*dtsx.ExpressionInfo)
		for i, expr := range exprs {
			details := dtsx.GetExpressionDetails(expr, pkg)
			fmt.Printf("%d. [%s] %s\n", i+1, details.Location, details.Expression)
			if details.Name != "" {
				fmt.Printf("   Property: %s\n", details.Name)
			}
			if details.Context != "" {
				fmt.Printf("   Context: %s\n", details.Context)
			}
			if details.EvaluatedValue != "" {
				fmt.Printf("   Evaluated: %s\n", details.EvaluatedValue)
			}
			if details.EvaluationError != "" {
				fmt.Printf("   Error: %s\n", details.EvaluationError)
			}
			if len(details.Dependencies) > 0 {
				fmt.Printf("   Dependencies: %v\n", details.Dependencies)
			}
		}
	}

	// Precedence Constraints
	if len(pkg.PrecedenceConstraint) > 0 {
		fmt.Printf("\n--- Precedence Constraints (%d) ---\n", len(pkg.PrecedenceConstraint))
	}

	// Event Handlers
	if len(pkg.EventHandler) > 0 {
		fmt.Printf("\n--- Event Handlers (%d) ---\n", len(pkg.EventHandler))
	}

	// Configurations
	if len(pkg.Configuration) > 0 {
		fmt.Printf("\n--- Configurations (%d) ---\n", len(pkg.Configuration))
	}

	// Log Providers
	if len(pkg.LogProvider) > 0 {
		fmt.Printf("\n--- Log Providers (%d) ---\n", len(pkg.LogProvider))
	}

	fmt.Println("\n=== Analysis Complete ===")
}
