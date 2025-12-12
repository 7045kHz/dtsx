//go:build ignore

// package_analysis.go - Comprehensive DTSX package analysis using new parser features
//
// This example demonstrates the new PackageParser, PrecedenceAnalyzer, and PackageValidator
// for advanced DTSX package analysis.
//
// Usage: go run examples/package_analysis.go <dtsx_file>
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/7045kHz/dtsx"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run package_analysis.go <dtsx_file>")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Load the DTSX package
	pkg, err := dtsx.UnmarshalFromFile(filename)
	if err != nil {
		log.Fatalf("Failed to load DTSX file: %v", err)
	}

	fmt.Printf("=== DTSX Package Analysis ===\n")
	fmt.Printf("File: %s\n\n", filename)

	// Create parser for centralized analysis
	parser := dtsx.NewPackageParser(pkg)

	// Analyze SQL statements
	fmt.Println("=== SQL Statements ===")
	sqlStatements := parser.GetSQLStatements()
	for _, stmt := range sqlStatements {
		fmt.Printf("Task: %s (%s)\n", stmt.TaskName, stmt.TaskType)
		fmt.Printf("  SQL: %s\n", stmt.SQL)
		if len(stmt.Connections) > 0 {
			fmt.Printf("  Connections: %v\n", stmt.Connections)
		}
		fmt.Println()
	}

	// Create precedence analyzer
	analyzer := dtsx.NewPrecedenceAnalyzer(pkg)

	// Get execution orders
	fmt.Println("=== Execution Order ===")
	orders, err := analyzer.GetAllExecutionOrders()
	if err != nil {
		fmt.Printf("Error getting execution orders: %v\n", err)
	} else {
		for refId, order := range orders {
			fmt.Printf("Order %d: %s\n", order, refId)
		}
	}
	fmt.Println()

	// Get execution flow description
	fmt.Println("=== Execution Flow Description ===")
	flowDesc := analyzer.GetExecutionFlowDescription()
	fmt.Print(flowDesc)
	fmt.Println()

	// Validate package
	fmt.Println("=== Package Validation ===")
	validator := dtsx.NewPackageValidator(pkg)
	validationErrors := validator.Validate()

	if len(validationErrors) == 0 {
		fmt.Println("✓ No validation errors found")
	} else {
		fmt.Printf("Found %d validation issues:\n", len(validationErrors))
		for _, err := range validationErrors {
			fmt.Printf("[%s] %s: %s\n", err.Severity, err.Path, err.Message)
		}
	}
	fmt.Println()

	// Demonstrate expression evaluation with caching
	fmt.Println("=== Expression Evaluation Demo ===")
	testExpr := "@[User::DB_CS]"
	if result, err := parser.EvaluateExpression(testExpr); err == nil {
		fmt.Printf("Expression '%s' evaluates to: %v\n", testExpr, result)
	} else {
		fmt.Printf("Expression evaluation failed: %v\n", err)
	}

	// Test caching by evaluating the same expression again
	if result, err := parser.EvaluateExpression(testExpr); err == nil {
		fmt.Printf("Cached evaluation result: %v\n", result)
	}
}
