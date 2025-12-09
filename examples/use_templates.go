//go:build ignore

// use_templates.go - DTSX Template Usage Example
//
// This example demonstrates how to use the DTSX template system
// to create packages from reusable templates.

package main

import (
	"fmt"
	"log"

	"github.com/7045kHz/dtsx"
)

func main() {
	fmt.Println("DTSX Template System Example")
	fmt.Println("============================")

	// Get the default template registry
	registry := dtsx.GetDefaultTemplateRegistry()

	// List available templates
	templates := registry.List()
	fmt.Printf("Available templates (%d):\n", len(templates))
	for i, name := range templates {
		template := registry.Get(name)
		fmt.Printf("  %d. %s - %s\n", i+1, template.Name, template.Description)
	}
	fmt.Println()

	// Demonstrate Basic ETL template
	etlTemplate := registry.Get("Basic ETL")
	if etlTemplate != nil {
		fmt.Printf("Using template: %s\n", etlTemplate.Name)
		fmt.Printf("Description: %s\n", etlTemplate.Description)
		fmt.Println("Parameters:")
		for param, desc := range etlTemplate.Parameters {
			fmt.Printf("  - %s: %s\n", param, desc)
		}
		fmt.Println()

		// Instantiate with sample parameters
		params := map[string]interface{}{
			"SourceConnection":      "Server=localhost;Database=SourceDB;Trusted_Connection=True",
			"DestinationConnection": "Server=localhost;Database=DestDB;Trusted_Connection=True",
			"SourceQuery":           "SELECT CustomerID, CustomerName, Email FROM Customers WHERE Status = 'Active'",
			"DestinationTable":      "Customers_Backup",
			"BatchSize":             "10000",
			"PackageName":           "CustomerDataMigration",
		}

		fmt.Println("Instantiating template with parameters...")
		pkg, err := etlTemplate.Instantiate(params)
		if err != nil {
			log.Fatalf("Failed to instantiate template: %v", err)
		}

		fmt.Printf("✓ Template instantiated successfully\n")
		fmt.Printf("✓ Package name: %s\n", pkg.Property[0].Value)

		// Validate the generated package
		fmt.Println("\nValidating generated package...")
		errors := pkg.Validate()
		if len(errors) > 0 {
			fmt.Printf("Validation issues found (%d):\n", len(errors))
			for _, err := range errors {
				fmt.Printf("  [%s] %s: %s\n", err.Severity, err.Path, err.Message)
			}
		} else {
			fmt.Println("✓ Package validation passed")
		}

		// Save the package
		outputFile := "generated_etl_package.dtsx"
		err = dtsx.MarshalToFile(outputFile, pkg)
		if err != nil {
			log.Fatalf("Failed to save package: %v", err)
		}
		fmt.Printf("✓ Package saved to: %s\n", outputFile)

		// Analyze dependencies
		fmt.Println("\nAnalyzing package dependencies...")
		graph := pkg.BuildDependencyGraph()
		fmt.Printf("Connections used: %d\n", len(graph.ConnectionDependencies))
		fmt.Printf("Variables defined: %d\n", len(pkg.Variables.Variable))

		unused := pkg.GetUnusedVariables()
		if len(unused) > 0 {
			fmt.Printf("Unused variables: %v\n", unused)
		} else {
			fmt.Println("✓ No unused variables")
		}
	}

	fmt.Println("\nTemplate system demonstration complete!")
}
