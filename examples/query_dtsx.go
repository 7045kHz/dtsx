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
		fmt.Println("Usage: go run examples/query_dtsx.go <path-to-dtsx-file>")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Validate and read the DTSX file
	pkg, ok := dtsx.IsDTSXPackage(filename)
	if !ok {
		log.Fatalf("File is not a valid DTSX package: %s", filename)
	}

	fmt.Printf("=== DTSX Package Query Examples ===\n")
	fmt.Printf("File: %s\n\n", filename)

	// Example 1: Get all connection managers
	fmt.Println("1. GetConnections() - All connection managers:")
	connections := pkg.GetConnections()
	fmt.Printf("   Found %d connection managers\n", connections.Count)
	if connections.Count > 0 {
		connMgrs := connections.Results.([]*schema.ConnectionManagerType)
		for i, cm := range connMgrs {
			fmt.Printf("   %d. Connection Manager with %d properties\n", i+1, len(cm.Property))
		}
	}
	fmt.Println()

	// Example 2: Get all variables
	fmt.Println("2. GetVariables() - All variables:")
	variables := pkg.GetVariables()
	fmt.Printf("   Found %d variables\n", variables.Count)
	if variables.Count > 0 {
		vars := variables.Results.([]*schema.VariableType)
		for i, v := range vars {
			// Try to get variable name from attributes
			fullName := dtsx.GetVariableName(v)

			fmt.Printf("   %d. %s", i+1, fullName)

			// Show variable value if available
			value := dtsx.GetVariableValue(v)
			if value != "" {
				// Truncate long values for display
				if len(value) > 50 {
					value = value[:47] + "..."
				}
				fmt.Printf(" = \"%s\"", value)
			}

			// Show if it has an expression
			if len(v.PropertyExpression) > 0 {
				fmt.Printf(" (has expression)")
				// Show the first expression
				if len(v.PropertyExpression) > 0 && v.PropertyExpression[0].AnySimpleType != nil {
					expr := v.PropertyExpression[0].AnySimpleType.Value
					if len(expr) > 30 {
						expr = expr[:27] + "..."
					}
					fmt.Printf(": %s", expr)
				}
			}

			fmt.Println()
		}
	}
	fmt.Println()

	// Example 3: Get a specific variable by name
	fmt.Println("3. GetVariableByName() - Find specific variable:")
	varNames := []string{"User::CSV_LOCATION", "User::DB_CS", "User::NonExistentVar"}
	for _, varName := range varNames {
		variable, err := pkg.GetVariableByName(varName)
		if err != nil {
			fmt.Printf("   Variable '%s': Not found\n", varName)
		} else {
			fmt.Printf("   Variable '%s': Found", varName)

			// Show variable value if available
			value := dtsx.GetVariableValue(variable)
			if value != "" {
				if len(value) > 30 {
					value = value[:27] + "..."
				}
				fmt.Printf(" = \"%s\"", value)
			}

			// Show if it has an expression
			if len(variable.PropertyExpression) > 0 {
				fmt.Printf(" (has expression)")
			}

			fmt.Println()
		}
	}
	fmt.Println()

	// Example 4: Query executables with filters
	fmt.Println("4. QueryExecutables() - Filter executables:")

	// Get all executables
	allExecutables := pkg.QueryExecutables(func(*schema.AnyNonPackageExecutableType) bool { return true })
	fmt.Printf("   Total executables: %d\n", len(allExecutables))

	// Filter for SQL tasks
	sqlTasks := pkg.QueryExecutables(func(exec *schema.AnyNonPackageExecutableType) bool {
		return exec.ExecutableTypeAttr == "STOCK:SQLTask"
	})
	fmt.Printf("   SQL Tasks: %d\n", len(sqlTasks))

	// Filter for data flow tasks
	dataFlowTasks := pkg.QueryExecutables(func(exec *schema.AnyNonPackageExecutableType) bool {
		return exec.ExecutableTypeAttr == "STOCK:Pipeline"
	})
	fmt.Printf("   Data Flow Tasks: %d\n", len(dataFlowTasks))

	// Filter for tasks with specific properties
	tasksWithExpressions := pkg.QueryExecutables(func(exec *schema.AnyNonPackageExecutableType) bool {
		return len(exec.PropertyExpression) > 0
	})
	fmt.Printf("   Tasks with expressions: %d\n", len(tasksWithExpressions))
	fmt.Println()

	// Example 5: Get all expressions
	fmt.Println("5. GetExpressions() - All expressions in package:")
	expressions := pkg.GetExpressions()
	fmt.Printf("   Found %d expressions\n", expressions.Count)
	if expressions.Count > 0 {
		exprs := expressions.Results.([]*dtsx.ExpressionInfo)
		for i, expr := range exprs {
			details := dtsx.GetExpressionDetails(expr, pkg)
			fmt.Printf("   %d. [%s] %s\n", i+1, details.Location, details.Expression)
			if details.Context != "" {
				fmt.Printf("      Context: %s\n", details.Context)
			}
			if details.EvaluatedValue != "" {
				fmt.Printf("      Evaluated: %s\n", details.EvaluatedValue)
			}
			if len(details.Dependencies) > 0 {
				fmt.Printf("      Dependencies: %v\n", details.Dependencies)
			}
		}
	}
	fmt.Println()

	// Example 6: Advanced queries - combine multiple query methods
	fmt.Println("6. Advanced Query Examples:")

	// Find variables that have expressions
	varsWithExpressions := 0
	if variables.Count > 0 {
		vars := variables.Results.([]*schema.VariableType)
		for _, v := range vars {
			if len(v.PropertyExpression) > 0 {
				varsWithExpressions++
			}
		}
	}
	fmt.Printf("   Variables with expressions: %d\n", varsWithExpressions)

	// Find executables that have both properties and expressions
	execsWithPropsAndExprs := pkg.QueryExecutables(func(exec *schema.AnyNonPackageExecutableType) bool {
		return len(exec.Property) > 0 && len(exec.PropertyExpression) > 0
	})
	fmt.Printf("   Executables with both properties and expressions: %d\n", len(execsWithPropsAndExprs))

	fmt.Println("\n=== Query Examples Complete ===")
	fmt.Println("\nThese query methods provide a clean API for analyzing DTSX packages:")
	fmt.Println("- GetConnections(): Get all connection managers")
	fmt.Println("- GetVariables(): Get all variables")
	fmt.Println("- GetVariableByName(name): Find a specific variable")
	fmt.Println("- QueryExecutables(filter): Find executables matching criteria")
	fmt.Println("- GetExpressions(): Get all expressions with context")
}
