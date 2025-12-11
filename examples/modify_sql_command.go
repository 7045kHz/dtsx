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
		fmt.Println("Usage: go run examples/modify_sql_command.go <path-to-dtsx-file>")
		fmt.Println("Example: go run examples/modify_sql_command.go SSIS_EXAMPLES/ConfigFile.dtsx")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Load the DTSX package
	pkg, ok := dtsx.IsDTSXPackage(filename)
	if !ok {
		log.Fatalf("File is not a valid DTSX package: %s", filename)
	}

	fmt.Printf("Loaded DTSX package: %s\n", filename)

	// Find the data flow executable "production_product_color_to_csv"
	var dataFlowExec *schema.AnyNonPackageExecutableType
	for _, exec := range pkg.Executable {
		if exec.ObjectNameAttr != nil && *exec.ObjectNameAttr == "production_product_color_to_csv" {
			dataFlowExec = exec
			break
		}
	}

	if dataFlowExec == nil {
		log.Fatalf("Could not find data flow executable 'production_product_color_to_csv'")
	}

	fmt.Printf("Found data flow executable: %s\n", *dataFlowExec.ObjectNameAttr)

	// Check if it has pipeline data
	if dataFlowExec.ObjectData == nil || dataFlowExec.ObjectData.Pipeline == nil ||
		dataFlowExec.ObjectData.Pipeline.Components == nil {
		log.Fatalf("Data flow does not contain pipeline components")
	}

	// Find the component named "get_colors_all"
	var targetComponent *schema.PipelineComponentType
	for _, comp := range dataFlowExec.ObjectData.Pipeline.Components.Component {
		if comp.NameAttr != nil && *comp.NameAttr == "get_colors_all" {
			targetComponent = comp
			break
		}
	}

	if targetComponent == nil {
		log.Fatalf("Could not find component 'get_colors_all' in the data flow")
	}

	fmt.Printf("Found component: %s\n", *targetComponent.NameAttr)

	// Find and update the SqlCommand property
	found := false
	oldSQL := ""
	newSQL := "select * from production.product where color='Blue'"

	if targetComponent.Properties != nil {
		for _, prop := range targetComponent.Properties.Property {
			if prop.NameAttr != nil && *prop.NameAttr == "SqlCommand" {
				oldSQL = prop.Value
				prop.Value = newSQL
				found = true
				break
			}
		}
	}

	if !found {
		log.Fatalf("Could not find SqlCommand property in component 'get_colors_all'")
	}

	fmt.Printf("Updated SQL command:\n")
	fmt.Printf("  Old: %s\n", oldSQL)
	fmt.Printf("  New: %s\n", newSQL)

	// Save the modified package
	outputFile := "ConfigFile2.dtsx"
	data, err := dtsx.Marshal(pkg)
	if err != nil {
		log.Fatalf("Error serializing modified package: %v", err)
	}
	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		log.Fatalf("Error saving modified package: %v", err)
	}

	fmt.Printf("\nModified package saved to: %s\n", outputFile)
	fmt.Printf("You can now run the modified package with DTExec or compare it with the original.\n")
}
