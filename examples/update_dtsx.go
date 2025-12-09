//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/7045kHz/dtsx"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run examples/update_dtsx.go <path-to-dtsx-file>")
		fmt.Println("Example: go run examples/update_dtsx.go SSIS_EXAMPLES/Expressions.dtsx")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Load the DTSX package
	pkg, ok := dtsx.IsDTSXPackage(filename)
	if !ok {
		log.Fatalf("File is not a valid DTSX package: %s", filename)
	}

	fmt.Printf("Loaded DTSX package: %s\n", filename)

	// Display current variables
	fmt.Printf("\n=== Current Variables ===\n")
	if pkg.Variables != nil && pkg.Variables.Variable != nil {
		for _, v := range pkg.Variables.Variable {
			if v.NamespaceAttr != nil && v.ObjectNameAttr != nil && v.VariableValue != nil {
				fmt.Printf("  %s::%s = %s\n", *v.NamespaceAttr, *v.ObjectNameAttr, v.VariableValue.Value)
			}
		}
	}

	// Display current connections
	fmt.Printf("\n=== Current Connection Managers ===\n")
	if pkg.ConnectionManagers != nil && pkg.ConnectionManagers.ConnectionManager != nil {
		fmt.Printf("Found %d connection managers\n", len(pkg.ConnectionManagers.ConnectionManager))
		for _, cm := range pkg.ConnectionManagers.ConnectionManager {
			// Find the connection name from properties
			var connName string
			var connString string
			for _, prop := range cm.Property {
				if prop.NameAttr != nil {
					if *prop.NameAttr == "ObjectName" && prop.PropertyElementBaseType != nil &&
						prop.PropertyElementBaseType.AnySimpleType != nil {
						connName = prop.PropertyElementBaseType.AnySimpleType.Value
					}
					if *prop.NameAttr == "ConnectionString" && prop.PropertyElementBaseType != nil &&
						prop.PropertyElementBaseType.AnySimpleType != nil {
						connString = prop.PropertyElementBaseType.AnySimpleType.Value
					}
				}
			}
			if connName != "" {
				fmt.Printf("  %s", connName)
				if connString != "" {
					fmt.Printf(" - %s", connString)
				}
				fmt.Printf("\n")
			}
		}
	}

	// Demonstrate updating a variable (if it exists)
	fmt.Printf("\n=== Demonstrating Variable Update ===\n")
	if pkg.Variables != nil && pkg.Variables.Variable != nil && len(pkg.Variables.Variable) > 0 {
		v := pkg.Variables.Variable[0]
		if v.NamespaceAttr != nil && v.ObjectNameAttr != nil {
			oldValue := ""
			if v.VariableValue != nil {
				oldValue = v.VariableValue.Value
			}
			fmt.Printf("Updating variable %s::%s from '%s' to 'UpdatedValue'\n",
				*v.NamespaceAttr, *v.ObjectNameAttr, oldValue)

			err := pkg.UpdateVariable(*v.NamespaceAttr, *v.ObjectNameAttr, "UpdatedValue")
			if err != nil {
				fmt.Printf("Error updating variable: %v\n", err)
			} else {
				fmt.Printf("Variable updated successfully!\n")
			}
		}
	}

	// Demonstrate updating a connection string (if it exists)
	fmt.Printf("\n=== Demonstrating Connection Update ===\n")
	if pkg.ConnectionManagers != nil && pkg.ConnectionManagers.ConnectionManager != nil && len(pkg.ConnectionManagers.ConnectionManager) > 0 {
		// Find the first connection with a name
		var connName string
		for _, cm := range pkg.ConnectionManagers.ConnectionManager {
			for _, prop := range cm.Property {
				if prop.NameAttr != nil && *prop.NameAttr == "ObjectName" &&
					prop.PropertyElementBaseType != nil && prop.PropertyElementBaseType.AnySimpleType != nil {
					connName = prop.PropertyElementBaseType.AnySimpleType.Value
					break
				}
			}
			if connName != "" {
				break
			}
		}

		if connName != "" {
			fmt.Printf("Updating connection '%s' to 'UpdatedConnectionString'\n", connName)

			err := pkg.UpdateConnectionString(connName, "UpdatedConnectionString")
			if err != nil {
				fmt.Printf("Error updating connection: %v\n", err)
			} else {
				fmt.Printf("Connection updated successfully!\n")
			}
		} else {
			fmt.Printf("No named connections found\n")
		}
	}

	// Demonstrate updating expressions (if any exist)
	fmt.Printf("\n=== Demonstrating Expression Updates ===\n")
	expressions := pkg.GetExpressions()
	if expressions.Count > 0 {
		fmt.Printf("Found %d expressions in the package\n", expressions.Count)
		exprList := expressions.Results.([]*dtsx.ExpressionInfo)

		// Try to update the first variable expression found
		for _, expr := range exprList {
			if expr.Location == "Variable" {
				// For variables, we need to extract the variable name from context
				// Context format: "Variable[0] (Namespace::Name)"
				context := expr.Context
				if strings.Contains(context, "(") && strings.Contains(context, ")") {
					start := strings.Index(context, "(") + 1
					end := strings.Index(context, ")")
					if start < end {
						varName := context[start:end]
						fmt.Printf("Updating expression on variable %s:\n", varName)
						fmt.Printf("  Property: %s\n", expr.Name)
						fmt.Printf("  Old: %s\n", expr.Expression)
						fmt.Printf("  New: @[System::StartTime]\n")

						err := pkg.UpdateExpression("Variable", varName, expr.Name, "@[System::StartTime]")
						if err != nil {
							fmt.Printf("Error updating expression: %v\n", err)
						} else {
							fmt.Printf("Expression updated successfully!\n")
						}
						break
					}
				}
			}
		}
	} else {
		fmt.Printf("No expressions found in this package\n")
	}

	// Demonstrate UpdateProperty method
	fmt.Printf("\n=== Demonstrating UpdateProperty ===\n")

	// Update a package property
	fmt.Printf("Updating package property 'CreatorName' to 'UpdatedByDTSXLibrary'\n")
	err := pkg.UpdateProperty("package", "", "CreatorName", "UpdatedByDTSXLibrary")
	if err != nil {
		fmt.Printf("Error updating package property: %v\n", err)
	} else {
		fmt.Printf("Package property updated successfully!\n")
	}

	// Update an executable property (if executables exist)
	if len(pkg.Executable) > 0 {
		// Find the first executable with a name
		var execName string
		for _, exec := range pkg.Executable {
			for _, prop := range exec.Property {
				if prop.NameAttr != nil && *prop.NameAttr == "ObjectName" &&
					prop.PropertyElementBaseType != nil && prop.PropertyElementBaseType.AnySimpleType != nil {
					execName = prop.PropertyElementBaseType.AnySimpleType.Value
					break
				}
			}
			if execName != "" {
				break
			}
		}

		if execName != "" {
			fmt.Printf("Updating executable '%s' property 'Description' to 'Updated via UpdateProperty'\n", execName)
			err := pkg.UpdateProperty("executable", execName, "Description", "Updated via UpdateProperty")
			if err != nil {
				fmt.Printf("Error updating executable property: %v\n", err)
			} else {
				fmt.Printf("Executable property updated successfully!\n")
			}
		} else {
			fmt.Printf("No named executables found\n")
		}
	}

	// Save the updated package
	outputFile := filename + ".updated.dtsx"
	err = dtsx.MarshalToFile(outputFile, pkg)
	if err != nil {
		log.Fatalf("Error saving updated package: %v", err)
	}

	fmt.Printf("\n=== Package Updated and Saved ===\n")
	fmt.Printf("Updated package saved to: %s\n", outputFile)
	fmt.Printf("You can compare the original and updated files to see the changes.\n")
}
