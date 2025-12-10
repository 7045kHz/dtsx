//go:build ignore

// evaluate_expressions.go - Advanced SSIS Expression Evaluation
//
// This example demonstrates the comprehensive SSIS expression evaluation
// capabilities including variables, arithmetic, functions, conditionals,
// and type casting.
//
// Usage: go run examples/evaluate_expressions.go <dtsx_file>
//
// Features:
// - Variable substitution
// - Arithmetic and string operations
// - Built-in SSIS functions
// - Conditional expressions
// - Type casting
// - Comparison and logical operators

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/7045kHz/dtsx"
	schema "github.com/7045kHz/dtsx/schemas"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run examples/evaluate_expressions.go <dtsx_file>")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Load the DTSX package
	pkg, err := dtsx.UnmarshalFromFile(filename)
	if err != nil {
		log.Fatalf("Failed to load DTSX file: %v", err)
	}

	// Create parser for expression evaluation with caching
	parser := dtsx.NewPackageParser(pkg)

	// Show available variables first
	variables := pkg.GetVariables()
	if variables.Count > 0 {
		fmt.Println("Available Variables:")
		vars := variables.Results.([]*schema.VariableType)
		for _, v := range vars {
			varName := "unnamed"
			if v.ObjectNameAttr != nil {
				varName = *v.ObjectNameAttr
			}
			namespace := "User"
			if v.NamespaceAttr != nil {
				namespace = *v.NamespaceAttr
			}
			fmt.Printf("  %s::%s", namespace, varName)

			if v.VariableValue != nil && v.VariableValue.Value != "" {
				value := v.VariableValue.Value
				if len(value) > 30 {
					value = value[:27] + "..."
				}
				fmt.Printf(" = \"%s\"", value)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	// Test cases for expression evaluation
	testCases := []struct {
		expression  string
		description string
	}{
		// Basic variable access
		{"@[User::MyVar]", "Basic variable access"},
		{"123", "Literal number"},
		{"\"Hello World\"", "Literal string"},

		// Arithmetic operations
		{"@[User::MyVar] + 1", "Arithmetic addition"},
		{"@[User::MyVar] * 2", "Arithmetic multiplication"},
		{"@[User::MyVar] / 2.0", "Arithmetic division"},

		// String operations
		{"@[User::StrVar] + \"_suffix\"", "String concatenation"},
		{"LEN(@[User::StrVar])", "String length function"},
		{"UPPER(@[User::StrVar])", "Uppercase function"},
		{"LOWER(@[User::StrVar])", "Lowercase function"},
		{"SUBSTRING(@[User::StrVar], 1, 3)", "Substring function"},

		// Math functions
		{"ABS(-5)", "Absolute value"},
		{"CEILING(3.2)", "Ceiling function"},
		{"FLOOR(3.8)", "Floor function"},

		// Date functions
		{"GETDATE()", "Current date/time"},
		{"YEAR(GETDATE())", "Year extraction"},
		{"MONTH(GETDATE())", "Month extraction"},
		{"DATEADD(\"DAY\", 7, GETDATE())", "Date addition"},
		{"DATEDIFF(\"DAY\", GETDATE(), DATEADD(\"DAY\", 7, GETDATE()))", "Date difference"},

		// Conditional expressions
		{"@[User::MyVar] > 10 ? \"High\" : \"Low\"", "Conditional expression"},
		{"@[User::MyVar] == 42 ? \"Answer\" : \"Not Answer\"", "Equality conditional"},

		// Type casting
		{"(DT_STR) @[User::MyVar]", "Cast to string"},
		{"(DT_INT) \"123\"", "Cast string to int"},

		// Comparison and logical operators
		{"@[User::MyVar] > 40 && @[User::MyVar] < 50", "Logical AND"},
		{"@[User::MyVar] < 30 || @[User::MyVar] > 50", "Logical OR"},
		{"!(@[User::MyVar] == 42)", "Logical NOT"},
		{"@[User::MyVar] != 0", "Inequality comparison"},
		{"@[User::MyVar] >= 40", "Greater than or equal"},
		{"@[User::MyVar] <= 50", "Less than or equal"},
	}

	fmt.Println("Expression Evaluation Results:")
	fmt.Println("==============================")

	for i, test := range testCases {
		fmt.Printf("%d. %s\n", i+1, test.description)
		fmt.Printf("   Expression: %s\n", test.expression)

		result, err := parser.EvaluateExpression(test.expression)
		if err != nil {
			fmt.Printf("   ❌ Error: %v\n", err)
		} else {
			fmt.Printf("   ✅ Result: %v (%T)\n", result, result)
		}
		fmt.Println()
	}

	// Demonstrate expression extraction from package
	expressions := pkg.GetExpressions()
	if expressions.Count > 0 {
		fmt.Println("Expressions Found in Package:")
		fmt.Println("=============================")

		exprs := expressions.Results.([]*dtsx.ExpressionInfo)
		for i, expr := range exprs {
			fmt.Printf("%d. Location: %s\n", i+1, expr.Location)
			fmt.Printf("   Expression: %s\n", expr.Expression)
			if expr.Name != "" {
				fmt.Printf("   Property: %s\n", expr.Name)
			}
			if expr.Context != "" {
				fmt.Printf("   Context: %s\n", expr.Context)
			}

			// Try to evaluate the expression
			result, err := parser.EvaluateExpression(expr.Expression)
			if err != nil {
				fmt.Printf("   ❌ Evaluation failed: %v\n", err)
			} else {
				fmt.Printf("   ✅ Evaluates to: %v\n", result)
			}
			fmt.Println()
		}
	}

	fmt.Println("=== Expression Evaluation Demo Complete ===")
}
