//go:build ignore

// custom_templates.go - Custom Template Directory Example
//
// This example demonstrates how to load templates from a custom directory
// and work with the file-based template system.

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/7045kHz/dtsx"
	schema "github.com/7045kHz/dtsx/schemas"
)

func main() {
	fmt.Println("Custom Template Directory Example")
	fmt.Println("=================================")

	// Example 1: Load templates from the default directory
	fmt.Println("1. Loading templates from default directory:")
	registry, err := dtsx.LoadTemplatesFromDirectory(dtsx.GetTemplatesDirectory())
	if err != nil {
		log.Printf("Failed to load templates from default directory: %v", err)
		fmt.Println("   (This is expected if no templates exist yet)")
	} else {
		templates := registry.List()
		fmt.Printf("   Loaded %d templates: %v\n", len(templates), templates)
	}

	// Example 2: Create a custom templates directory
	customDir := "custom_templates"
	fmt.Printf("\n2. Creating custom templates directory: %s\n", customDir)

	err = os.MkdirAll(customDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create custom directory: %v", err)
	}

	// Initialize it with built-in templates
	err = dtsx.InitializeTemplatesDirectory(customDir)
	if err != nil {
		log.Fatalf("Failed to initialize templates directory: %v", err)
	}
	fmt.Printf("   ✓ Initialized with built-in templates\n")

	// Example 3: Load from custom directory
	fmt.Printf("\n3. Loading templates from custom directory: %s\n", customDir)
	customRegistry, err := dtsx.LoadTemplatesFromDirectory(customDir)
	if err != nil {
		log.Fatalf("Failed to load templates from custom directory: %v", err)
	}

	templates := customRegistry.List()
	fmt.Printf("   Loaded %d templates: %v\n", len(templates), templates)

	// Example 4: Demonstrate template usage from custom directory
	fmt.Println("\n4. Using templates from custom directory:")
	for _, name := range templates {
		template := customRegistry.Get(name)
		fmt.Printf("   - %s: %s\n", template.Name, template.Description)
	}

	// Example 5: Add a custom template
	fmt.Println("\n5. Adding a custom template:")
	name := "Name"
	customTemplate := &dtsx.PackageTemplate{
		Name:        "Custom Example",
		Description: "A custom template example",
		Parameters: map[string]string{
			"CustomParam": "A custom parameter",
		},
		BasePackage: &dtsx.Package{
			ExecutableTypePackage: &schema.ExecutableTypePackage{
				Property: []*schema.Property{
					{
						NameAttr: &name,
						PropertyElementBaseType: &schema.PropertyElementBaseType{
							AnySimpleType: &schema.AnySimpleType{
								Value: "CustomPackage",
							},
						},
					},
				},
			},
		},
	}

	// Save the custom template
	customPath := customDir + "/custom_example.json"
	err = customTemplate.SaveToFile(customPath)
	if err != nil {
		log.Fatalf("Failed to save custom template: %v", err)
	}
	fmt.Printf("   ✓ Saved custom template to: %s\n", customPath)

	// Reload to include the custom template
	fmt.Println("\n6. Reloading directory with custom template:")
	updatedRegistry, err := dtsx.LoadTemplatesFromDirectory(customDir)
	if err != nil {
		log.Fatalf("Failed to reload templates: %v", err)
	}

	updatedTemplates := updatedRegistry.List()
	fmt.Printf("   Now loaded %d templates: %v\n", len(updatedTemplates), updatedTemplates)

	fmt.Println("\nCustom template directory example complete!")
}
