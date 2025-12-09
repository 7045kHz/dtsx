// templates.go - DTSX Package Templates
//
// This file implements a template system for reusable DTSX package patterns.
// Templates provide common ETL workflows that can be customized with parameters.

package dtsx

import (
	"fmt"
	"regexp"
	"strings"

	schema "github.com/7045kHz/dtsx/schemas"
)

// PackageTemplate represents a reusable DTSX package template
type PackageTemplate struct {
	Name        string
	Description string
	BasePackage *Package
	Parameters  map[string]string // parameter name -> description
}

// Instantiate creates a new package from the template with parameter substitution
func (pt *PackageTemplate) Instantiate(params map[string]interface{}) (*Package, error) {
	if pt.BasePackage == nil {
		return nil, fmt.Errorf("template has no base package")
	}

	// Deep copy the base package (simplified - in production would need proper deep copy)
	pkg := &Package{
		ExecutableTypePackage: &schema.ExecutableTypePackage{},
	}

	// Copy basic properties
	if pt.BasePackage.Property != nil {
		pkg.Property = make([]*schema.Property, len(pt.BasePackage.Property))
		copy(pkg.Property, pt.BasePackage.Property)
	}

	// Substitute parameters in properties
	err := pt.substituteParameters(pkg, params)
	if err != nil {
		return nil, err
	}

	return pkg, nil
}

// substituteParameters replaces {{ParamName}} placeholders with actual values
func (pt *PackageTemplate) substituteParameters(pkg *Package, params map[string]interface{}) error {
	// Substitute in package properties
	if pkg.Property != nil {
		for _, prop := range pkg.Property {
			if prop.Value != "" {
				newValue, err := pt.substituteString(prop.Value, params)
				if err != nil {
					return err
				}
				prop.Value = newValue
			}
		}
	}

	// Substitute in variables
	if pkg.Variables != nil && pkg.Variables.Variable != nil {
		for _, v := range pkg.Variables.Variable {
			if v.VariableValue != nil && v.VariableValue.Value != "" {
				newValue, err := pt.substituteString(v.VariableValue.Value, params)
				if err != nil {
					return err
				}
				v.VariableValue.Value = newValue
			}
		}
	}

	// Substitute in connections
	if pkg.ConnectionManagers != nil && pkg.ConnectionManagers.ConnectionManager != nil {
		for _, cm := range pkg.ConnectionManagers.ConnectionManager {
			if cm.Property != nil {
				for _, prop := range cm.Property {
					if prop.Value != "" {
						newValue, err := pt.substituteString(prop.Value, params)
						if err != nil {
							return err
						}
						prop.Value = newValue
					}
				}
			}
		}
	}

	return nil
}

// substituteString replaces {{Param}} with values from params map
func (pt *PackageTemplate) substituteString(s string, params map[string]interface{}) (string, error) {
	re := regexp.MustCompile(`\{\{(\w+)\}\}`)
	result := re.ReplaceAllStringFunc(s, func(match string) string {
		paramName := strings.Trim(match, "{}")
		if value, ok := params[paramName]; ok {
			return fmt.Sprintf("%v", value)
		}
		// Leave unsubstituted if parameter not provided
		return match
	})
	return result, nil
}

// TemplateRegistry manages a collection of package templates
type TemplateRegistry struct {
	templates map[string]*PackageTemplate
}

// NewTemplateRegistry creates a new template registry
func NewTemplateRegistry() *TemplateRegistry {
	return &TemplateRegistry{
		templates: make(map[string]*PackageTemplate),
	}
}

// Register adds a template to the registry
func (tr *TemplateRegistry) Register(template *PackageTemplate) {
	if template != nil && template.Name != "" {
		tr.templates[template.Name] = template
	}
}

// Get retrieves a template by name
func (tr *TemplateRegistry) Get(name string) *PackageTemplate {
	return tr.templates[name]
}

// List returns all available template names
func (tr *TemplateRegistry) List() []string {
	var names []string
	for name := range tr.templates {
		names = append(names, name)
	}
	return names
}

// CreateBasicETLTemplate creates a basic ETL template
func CreateBasicETLTemplate() *PackageTemplate {
	template := &PackageTemplate{
		Name:        "Basic ETL",
		Description: "Basic Extract-Transform-Load workflow template",
		Parameters: map[string]string{
			"SourceConnection":      "Source database connection string",
			"DestinationConnection": "Destination database connection string",
			"SourceQuery":           "SQL query for data extraction",
			"DestinationTable":      "Destination table name",
			"BatchSize":             "Batch size for data loading",
		},
		BasePackage: &Package{
			ExecutableTypePackage: &schema.ExecutableTypePackage{
				Property: []*schema.Property{
					{
						NameAttr: stringPtr("Name"),
						PropertyElementBaseType: &schema.PropertyElementBaseType{
							AnySimpleType: &schema.AnySimpleType{
								Value: "{{PackageName}}",
							},
						},
					},
				},
				Variables: &schema.VariablesType{
					Variable: []*schema.VariableType{
						{
							NamespaceAttr:  stringPtr("User"),
							ObjectNameAttr: stringPtr("BatchSize"),
							VariableValue: &schema.VariableValue{
								Value: "{{BatchSize}}",
							},
						},
					},
				},
				ConnectionManagers: &schema.ConnectionManagersType{
					ConnectionManager: []*schema.ConnectionManagerType{
						{
							ObjectNameAttr: stringPtr("SourceDB"),
							Property: []*schema.Property{
								{
									NameAttr: stringPtr("ConnectionString"),
									PropertyElementBaseType: &schema.PropertyElementBaseType{
										AnySimpleType: &schema.AnySimpleType{
											Value: "{{SourceConnection}}",
										},
									},
								},
							},
						},
						{
							ObjectNameAttr: stringPtr("DestinationDB"),
							Property: []*schema.Property{
								{
									NameAttr: stringPtr("ConnectionString"),
									PropertyElementBaseType: &schema.PropertyElementBaseType{
										AnySimpleType: &schema.AnySimpleType{
											Value: "{{DestinationConnection}}",
										},
									},
								},
							},
						},
					},
				},
				Executable: []*schema.AnyNonPackageExecutableType{
					{
						ExecutableTypeAttr: "ExecuteSQLTask",
						Property: []*schema.Property{
							{
								NameAttr: stringPtr("Connection"),
								PropertyElementBaseType: &schema.PropertyElementBaseType{
									AnySimpleType: &schema.AnySimpleType{
										Value: "SourceDB",
									},
								},
							},
							{
								NameAttr: stringPtr("SqlStatementSource"),
								PropertyElementBaseType: &schema.PropertyElementBaseType{
									AnySimpleType: &schema.AnySimpleType{
										Value: "{{SourceQuery}}",
									},
								},
							},
						},
					},
					{
						ExecutableTypeAttr: "DataFlowTask",
						Property: []*schema.Property{
							{
								NameAttr: stringPtr("Connection"),
								PropertyElementBaseType: &schema.PropertyElementBaseType{
									AnySimpleType: &schema.AnySimpleType{
										Value: "DestinationDB",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return template
}

// CreateFileProcessingTemplate creates a file processing template
func CreateFileProcessingTemplate() *PackageTemplate {
	template := &PackageTemplate{
		Name:        "File Processing",
		Description: "Template for processing files with logging and error handling",
		Parameters: map[string]string{
			"InputPath":      "Path to input files",
			"OutputPath":     "Path for processed output",
			"ArchivePath":    "Path for archiving processed files",
			"ErrorThreshold": "Maximum number of errors before failing",
		},
		BasePackage: &Package{
			ExecutableTypePackage: &schema.ExecutableTypePackage{
				Variables: &schema.VariablesType{
					Variable: []*schema.VariableType{
						{
							NamespaceAttr:  stringPtr("User"),
							ObjectNameAttr: stringPtr("InputPath"),
							VariableValue: &schema.VariableValue{
								Value: "{{InputPath}}",
							},
						},
						{
							NamespaceAttr:  stringPtr("User"),
							ObjectNameAttr: stringPtr("OutputPath"),
							VariableValue: &schema.VariableValue{
								Value: "{{OutputPath}}",
							},
						},
						{
							NamespaceAttr:  stringPtr("User"),
							ObjectNameAttr: stringPtr("ErrorCount"),
							VariableValue: &schema.VariableValue{
								Value: "0",
							},
						},
					},
				},
				Executable: []*schema.AnyNonPackageExecutableType{
					{
						ExecutableTypeAttr: "ForEachLoop",
						Property: []*schema.Property{
							{
								NameAttr: stringPtr("Directory"),
								PropertyElementBaseType: &schema.PropertyElementBaseType{
									AnySimpleType: &schema.AnySimpleType{
										Value: "{{InputPath}}",
									},
								},
							},
						},
					},
					{
						ExecutableTypeAttr: "ExecuteProcessTask",
						Property: []*schema.Property{
							{
								NameAttr: stringPtr("Executable"),
								PropertyElementBaseType: &schema.PropertyElementBaseType{
									AnySimpleType: &schema.AnySimpleType{
										Value: "cmd.exe",
									},
								},
							},
							{
								NameAttr: stringPtr("Arguments"),
								PropertyElementBaseType: &schema.PropertyElementBaseType{
									AnySimpleType: &schema.AnySimpleType{
										Value: "/c move \"{{InputPath}}\\*\" \"{{OutputPath}}\"",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return template
}

// GetDefaultTemplateRegistry returns a registry with built-in templates
func GetDefaultTemplateRegistry() *TemplateRegistry {
	registry := NewTemplateRegistry()
	registry.Register(CreateBasicETLTemplate())
	registry.Register(CreateFileProcessingTemplate())
	return registry
}
