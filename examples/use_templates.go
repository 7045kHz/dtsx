//go:build ignore

// use_templates.go - DTSX Template Usage Example
//
// This example demonstrates how to use the DTSX template system
// to create packages from reusable templates.

package main

import (
	"fmt"
	"log"
	"strings"

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

		// Validate the generated package using PackageValidator
		fmt.Println("\nValidating generated package...")
		validator := dtsx.NewPackageValidator(pkg)
		errors := validator.Validate()
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

	// Demonstrate Table Copy template
	fmt.Println("\n" + strings.Repeat("=", 50))
	tableCopyTemplate := registry.Get("Table Copy")
	if tableCopyTemplate != nil {
		fmt.Printf("Using template: %s\n", tableCopyTemplate.Name)
		fmt.Printf("Description: %s\n", tableCopyTemplate.Description)
		fmt.Println("Parameters:")
		for param, desc := range tableCopyTemplate.Parameters {
			fmt.Printf("  - %s: %s\n", param, desc)
		}
		fmt.Println()

		// Instantiate with sample parameters
		tableCopyParams := map[string]interface{}{
			"SourceConnection":      "Server=prod-server;Database=ProductionDB;Trusted_Connection=True",
			"DestinationConnection": "Server=staging-server;Database=StagingDB;Trusted_Connection=True",
			"SourceTable":           "dbo.Customers",
			"DestinationTable":      "dbo.Customers_Backup",
			"PackageName":           "CustomerTableCopy",
		}

		fmt.Println("Instantiating Table Copy template with parameters...")
		tableCopyPkg, err := tableCopyTemplate.Instantiate(tableCopyParams)
		if err != nil {
			log.Fatalf("Failed to instantiate Table Copy template: %v", err)
		}

		fmt.Printf("✓ Table Copy template instantiated successfully\n")
		fmt.Printf("✓ Package name: %s\n", tableCopyPkg.Property[0].Value)

		// Show the generated SQL statements
		fmt.Println("\nGenerated SQL statements:")
		parser := dtsx.NewPackageParser(tableCopyPkg)
		sqlStatements := parser.GetSQLStatements()
		for _, stmt := range sqlStatements {
			fmt.Printf("  %s (%s): %s\n", stmt.TaskName, stmt.TaskType, stmt.SQL)
		}

		// Show execution order
		fmt.Println("\nExecution order:")
		analyzer := dtsx.NewPrecedenceAnalyzer(tableCopyPkg)
		orders, err := analyzer.GetAllExecutionOrders()
		if err != nil {
			fmt.Printf("  Error getting execution orders: %v\n", err)
		} else if len(orders) == 0 {
			fmt.Println("  No execution orders found")
		} else {
			for refId, order := range orders {
				fmt.Printf("  Order %d: %s\n", order, refId)
			}
		}

		// Save the package
		tableCopyOutputFile := "generated_table_copy_package.dtsx"
		err = dtsx.MarshalToFile(tableCopyOutputFile, tableCopyPkg)
		if err != nil {
			log.Fatalf("Failed to save Table Copy package: %v", err)
		}
		fmt.Printf("✓ Table Copy package saved to: %s\n", tableCopyOutputFile)
	}

	fmt.Println("\nTemplate system demonstration complete!")
}
