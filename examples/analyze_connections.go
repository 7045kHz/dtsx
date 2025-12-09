//go:build ignore

// analyze_connections.go - Comprehensive DTSX connection analysis
//
// This example demonstrates how to analyze all connection managers in a DTSX package,
// including their drivers, associated variables, expressions, and evaluated values.
//
// Usage: go run examples/analyze_connections.go <dtsx_file>
//
// Features:
// - Shows connection details (name, type, driver)
// - Lists property expressions and referenced variables
// - Displays variable default values
// - Attempts to evaluate simple expressions by substituting variables
package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/7045kHz/dtsx"
	schema "github.com/7045kHz/dtsx/schemas"
)

type ConnectionAnalysis struct {
	Name        string
	Type        string
	Driver      string
	Properties  map[string]string
	Expressions map[string]string
	Variables   []string
	Evaluated   map[string]string
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run analyze_connections.go <dtsx_file>")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Load the DTSX package
	pkg, err := dtsx.UnmarshalFromFile(filename)
	if err != nil {
		log.Fatalf("Failed to load DTSX file: %v", err)
	}

	fmt.Printf("=== DTSX Connection Analysis ===\n")
	fmt.Printf("File: %s\n\n", filename)

	// Get all variables for reference
	variables := make(map[string]string)
	if pkg.Variables != nil && pkg.Variables.Variable != nil {
		for _, v := range pkg.Variables.Variable {
			varName := "unnamed"
			if v.ObjectNameAttr != nil {
				varName = *v.ObjectNameAttr
			}
			var namespace string
			if v.NamespaceAttr != nil {
				namespace = *v.NamespaceAttr
			}
			fullName := fmt.Sprintf("%s::%s", namespace, varName)
			if v.VariableValue != nil {
				variables[fullName] = v.VariableValue.Value
			}
		}
	}

	// Analyze connections
	connections := pkg.GetConnections()
	if connections.Count == 0 {
		fmt.Println("No connection managers found in package.")
		return
	}

	fmt.Printf("Found %d connection manager(s):\n\n", connections.Count)

	connMgrs := connections.Results.([]*schema.ConnectionManagerType)
	for i, cm := range connMgrs {
		analysis := analyzeConnection(cm, variables, pkg)
		fmt.Printf("%d. %s (%s)\n", i+1, analysis.Name, analysis.Type)
		fmt.Printf("   Driver: %s\n", analysis.Driver)

		if len(analysis.Properties) > 0 {
			fmt.Println("   Properties:")
			for k, v := range analysis.Properties {
				fmt.Printf("     %s: %s\n", k, v)
			}
		}

		if len(analysis.Expressions) > 0 {
			fmt.Println("   Expressions:")
			for prop, expr := range analysis.Expressions {
				fmt.Printf("     %s: %s\n", prop, expr)
			}
		}

		if len(analysis.Variables) > 0 {
			fmt.Println("   Referenced Variables:")
			for _, v := range analysis.Variables {
				fmt.Printf("     %s", v)
				if value, exists := variables[v]; exists {
					if len(value) > 50 {
						value = value[:47] + "..."
					}
					fmt.Printf(" = \"%s\"", value)
				}
				fmt.Println()
			}
		}

		if len(analysis.Evaluated) > 0 {
			fmt.Println("   Evaluated Values:")
			for prop, value := range analysis.Evaluated {
				if len(value) > 80 {
					value = value[:77] + "..."
				}
				fmt.Printf("     %s: %s\n", prop, value)
			}
		}

		fmt.Println()
	}
}

func analyzeConnection(cm *schema.ConnectionManagerType, variables map[string]string, pkg *dtsx.Package) *ConnectionAnalysis {
	analysis := &ConnectionAnalysis{
		Name:        "Unknown",
		Type:        "Unknown",
		Driver:      "Unknown",
		Properties:  make(map[string]string),
		Expressions: make(map[string]string),
		Evaluated:   make(map[string]string),
	}

	// Extract attributes
	if cm.ObjectNameAttr != nil {
		analysis.Name = *cm.ObjectNameAttr
	}
	if cm.CreationNameAttr != nil {
		analysis.Driver = *cm.CreationNameAttr
		analysis.Type = getConnectionType(*cm.CreationNameAttr)
	}

	// Extract properties
	if cm.Property != nil {
		for _, prop := range cm.Property {
			if prop.NameAttr != nil && prop.PropertyElementBaseType != nil &&
				prop.PropertyElementBaseType.AnySimpleType != nil {
				name := *prop.NameAttr
				value := prop.PropertyElementBaseType.AnySimpleType.Value
				analysis.Properties[name] = value
			}
		}
	}

	// Extract expressions
	if cm.PropertyExpression != nil {
		for _, expr := range cm.PropertyExpression {
			if expr.NameAttr != "" && expr.AnySimpleType != nil {
				propName := expr.NameAttr
				expression := expr.AnySimpleType.Value
				analysis.Expressions[propName] = expression

				// Extract variable references
				vars := extractVariables(expression)
				analysis.Variables = append(analysis.Variables, vars...)

				// Try to evaluate expressions using the advanced evaluator
				if evaluated := evaluateExpressionAdvanced(expression, pkg); evaluated != "" {
					analysis.Evaluated[propName] = evaluated
				}
			}
		}
	}

	return analysis
}

func getConnectionType(creationName string) string {
	switch strings.ToUpper(creationName) {
	case "OLEDB":
		return "OLE DB Database"
	case "FLATFILE":
		return "Flat File"
	case "ADO.NET":
		return "ADO.NET Database"
	case "EXCEL":
		return "Excel File"
	case "HTTP":
		return "HTTP Connection"
	case "FTP":
		return "FTP Connection"
	case "SMTP":
		return "SMTP Connection"
	case "MSMQ":
		return "MSMQ Connection"
	case "WMI":
		return "WMI Connection"
	default:
		return creationName
	}
}

func extractVariables(expression string) []string {
	// Simple regex to find @[Namespace::Variable] patterns
	re := regexp.MustCompile(`@\[([^\]]+)\]`)
	matches := re.FindAllStringSubmatch(expression, -1)

	var vars []string
	for _, match := range matches {
		if len(match) > 1 {
			vars = append(vars, match[1])
		}
	}
	return vars
}

func evaluateExpressionAdvanced(expression string, pkg *dtsx.Package) string {
	// Use the advanced SSIS expression evaluator
	result, err := dtsx.EvaluateExpression(expression, pkg)
	if err != nil {
		// Return empty string if evaluation fails
		return ""
	}

	// Convert result to string
	switch v := result.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v) // Remove decimal for integers
	default:
		return fmt.Sprintf("%v", v)
	}
}
