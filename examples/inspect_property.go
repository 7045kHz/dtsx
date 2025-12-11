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
		fmt.Println("Usage: go run examples/inspect_property.go <path-to-dtsx-file>")
		os.Exit(1)
	}
	filename := os.Args[1]
	pkg, err := dtsx.UnmarshalFromFile(filename)
	if err != nil {
		log.Fatalf("Error parsing DTSX: %v", err)
	}
	var dataFlowExec *schema.AnyNonPackageExecutableType
	for _, exec := range pkg.Executable {
		if exec.ObjectNameAttr != nil && *exec.ObjectNameAttr == "production_product_color_to_csv" {
			dataFlowExec = exec
			break
		}
	}
	if dataFlowExec == nil {
		log.Fatalf("could not find data flow")
	}
	var target *schema.PipelineComponentType
	for _, comp := range dataFlowExec.ObjectData.Pipeline.Components.Component {
		if comp.NameAttr != nil && *comp.NameAttr == "get_colors_all" {
			target = comp
			break
		}
	}
	if target == nil || target.Properties == nil {
		log.Fatalf("could not find component or properties")
	}
	for _, p := range target.Properties.Property {
		if p.NameAttr != nil && *p.NameAttr == "SqlCommand" {
			fmt.Printf("Found SqlCommand property. name=%v\n", *p.NameAttr)
			fmt.Printf("  DataType: %v\n", p.DataTypeAttr)
			fmt.Printf("  Description: %v\n", p.DescriptionAttr)
			fmt.Printf("  UITypeEditor: %v\n", p.UITypeEditorAttr)
			fmt.Printf("  Value: %q\n", p.Value)
			return
		}
	}
	fmt.Println("SqlCommand property not found")
}
