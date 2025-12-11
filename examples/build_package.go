//go:build ignore

// build_package.go - DTSX Package Builder API Demo
//
// This example demonstrates how to programmatically create DTSX packages
// using the fluent builder pattern.
//
// Features:
// - Create packages from scratch
// - Add variables and connection managers
// - Configure package properties
// - Validate and save packages

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/7045kHz/dtsx"
)

func main() {
	fmt.Println("DTSX Package Builder API Demo")
	fmt.Println("=============================")

	// Example 1: Basic ETL Package
	fmt.Println("1. Creating Basic ETL Package...")

	etlPackage := dtsx.NewPackageBuilder().
		// Variables of different types
		AddVariableWithType("User", "SourcePath", "C:\\data\\input\\customers.csv", "String").
		AddVariableWithType("User", "TargetPath", "C:\\data\\output\\customers_processed.csv", "String").
		AddVariableWithType("User", "BatchSize", "1000", "Int32").
		AddVariableWithType("User", "ProcessDate", "2025-01-01", "DateTime").
		AddVariableWithType("User", "EnableLogging", "true", "Boolean").

		// Connection managers
		AddConnection("SourceDB", "OLEDB", "Server=myserver;Database=SourceDB;Trusted_Connection=True;Provider=SQLNCLI11.1;").
		AddConnection("TargetDB", "OLEDB", "Server=myserver;Database=TargetDB;Trusted_Connection=True;Provider=SQLNCLI11.1;").
		AddConnection("LogFile", "FLATFILE", "C:\\logs\\etl_log.txt").
		Build()

	fmt.Printf("✓ ETL Package created with %d variables and %d connections\n", len(etlPackage.Variables.Variable), len(etlPackage.ConnectionManagers.ConnectionManager))

	// Example 2: Data Validation Package
	fmt.Println("\n2. Creating Data Validation Package...")

	validationPackage := dtsx.NewPackageBuilder().
		// Validation variables of different types
		AddVariableWithType("User", "InputFile", "C:\\data\\validation\\input.csv", "String").
		AddVariableWithType("User", "ErrorThreshold", "0.05", "Double"). // 5% error tolerance
		AddVariableWithType("User", "MaxErrors", "100", "Int32").
		AddVariableWithType("User", "ValidationRules", "NOT_NULL,UNIQUE,RANGE_CHECK", "String").
		AddVariableWithType("User", "ReportPath", "C:\\reports\\validation_report.txt", "String").
		AddVariableWithType("User", "SendEmail", "true", "Boolean").

		// Connections
		AddConnection("ValidationDB", "OLEDB", "Server=validator;Database=Validation;Trusted_Connection=True;").
		AddConnection("ReportSMTP", "SMTP", "SmtpServer=mail.company.com;UseWindowsAuthentication=True;").
		Build()

	fmt.Printf("✓ Validation Package created with %d variables\n", len(validationPackage.Variables.Variable))

	// Example 3: File Processing Package
	fmt.Println("\n3. Creating File Processing Package...")

	filePackage := dtsx.NewPackageBuilder().
		// File processing variables of different types
		AddVariableWithType("User", "InputDirectory", "C:\\incoming\\files\\", "String").
		AddVariableWithType("User", "ProcessedDirectory", "C:\\processed\\files\\", "String").
		AddVariableWithType("User", "ErrorDirectory", "C:\\error\\files\\", "String").
		AddVariableWithType("User", "FilePattern", "*.csv", "String").
		AddVariableWithType("User", "MaxFileSize", "104857600", "Int64").  // 100MB
		AddVariableWithType("User", "ProcessingTimeout", "3600", "Int32"). // 1 hour
		AddVariableWithType("User", "EnableArchive", "true", "Boolean").

		// Connections
		AddConnection("FileSystem", "FILE", "C:\\file_operations\\").
		AddConnection("ArchiveDB", "OLEDB", "Server=archive;Database=FileArchive;Trusted_Connection=True;").
		Build()

	fmt.Printf("✓ File Processing Package created\n")

	// Example 4: Dynamic Configuration Package
	fmt.Println("\n4. Creating Dynamic Configuration Package...")

	dynamicPackage := dtsx.NewPackageBuilder().
		// Configuration variables of different types
		AddVariableWithType("User", "ServerName", "default-server", "String").
		AddVariableWithType("User", "DatabaseName", "default-db", "String").
		AddVariableWithType("User", "ConnectionStringVar", "", "String").
		AddVariableWithType("User", "PortNumber", "1433", "Int32").
		AddVariableWithType("User", "UseIntegratedSecurity", "true", "Boolean").

		// Connection with dynamic properties
		AddConnection("DynamicDB", "OLEDB", "Server=placeholder;Database=placeholder;Trusted_Connection=True;").
		AddConnectionExpression("DynamicDB", "ConnectionString",
			`"Server=" + @[User::ServerName] + ";Database=" + @[User::DatabaseName] + ";Trusted_Connection=True;"`).
		Build()

	fmt.Printf("✓ Dynamic Configuration Package created\n")
	fmt.Printf("  Connection 'DynamicDB' has dynamic connection string expression\n")

	// Validate all packages
	fmt.Println("\n5. Validating Created Packages...")

	packages := []*dtsx.Package{etlPackage, validationPackage, filePackage, dynamicPackage}
	packageNames := []string{"ETL Package", "Validation Package", "File Processing Package", "Dynamic Configuration Package"}

	for i, pkg := range packages {
		fmt.Printf("\nValidating %s:\n", packageNames[i])

		errors := pkg.Validate()
		if len(errors) == 0 {
			fmt.Println("  ✓ No validation errors")
		} else {
			fmt.Printf("  ⚠️  Found %d validation issues:\n", len(errors))
			for _, err := range errors {
				fmt.Printf("    - [%s] %s: %s\n", err.Severity, err.Path, err.Message)
			}
		}

		// Check for unused variables
		unused := pkg.GetUnusedVariables()
		if len(unused) > 0 {
			fmt.Printf("  ℹ️  Unused variables: %v\n", unused)
		} else {
			fmt.Println("  ✓ No unused variables")
		}
	}

	// Save one of the packages as an example
	fmt.Println("\n6. Saving ETL Package to File...")

	outputFile := "built_etl_package.dtsx"
	data, err := dtsx.Marshal(etlPackage)
	if err != nil {
		log.Fatalf("Failed to serialize package: %v", err)
	}
	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		log.Fatalf("Failed to save package: %v", err)
	}
	if err != nil {
		log.Fatalf("Failed to save package: %v", err)
	}

	fmt.Printf("✓ Package saved to: %s\n", outputFile)

	// Verify the saved package can be loaded
	fmt.Println("\n7. Verifying Saved Package...")

	loadedPkg, err := dtsx.UnmarshalFromFile(outputFile)
	if err != nil {
		log.Fatalf("Failed to load saved package: %v", err)
	}

	fmt.Printf("✓ Package loaded successfully\n")
	if len(loadedPkg.Property) > 0 {
		fmt.Printf("  Name: %s\n", loadedPkg.Property[0].Value)
	} else {
		fmt.Printf("  Name: (no properties set)\n")
	}
	fmt.Printf("  Variables: %d\n", len(loadedPkg.Variables.Variable))
	fmt.Printf("  Connections: %d\n", len(loadedPkg.ConnectionManagers.ConnectionManager))

	// Demonstrate dependency analysis on the created package
	fmt.Println("\n8. Analyzing Package Dependencies...")

	graph := etlPackage.BuildDependencyGraph()
	fmt.Printf("Dependency Analysis Results:\n")
	fmt.Printf("  Variables defined: %d\n", len(etlPackage.Variables.Variable))
	fmt.Printf("  Connections defined: %d\n", len(etlPackage.ConnectionManagers.ConnectionManager))
	fmt.Printf("  Variable relationships: %d\n", len(graph.VariableDependencies))

	// Get optimization suggestions
	suggestions := etlPackage.GetOptimizationSuggestions()
	if len(suggestions) > 0 {
		fmt.Printf("  Optimization suggestions: %d\n", len(suggestions))
		for _, s := range suggestions {
			fmt.Printf("    - %s\n", s.Message)
		}
	} else {
		fmt.Println("  ✓ No optimization suggestions")
	}

	fmt.Println("\n=== Package Builder Demo Complete ===")
	fmt.Println("\nThe Package Builder API allows you to:")
	fmt.Println("• Create DTSX packages programmatically")
	fmt.Println("• Add variables and connection managers fluently")
	fmt.Println("• Set package properties and configuration")
	fmt.Println("• Validate packages for common issues")
	fmt.Println("• Save packages to disk for deployment")
}
