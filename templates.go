// templates.go - DTSX Package Templates
//
// This file implements a template system for reusable DTSX package patterns.
// Templates provide common ETL workflows that can be customized with parameters.

package dtsx

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	schema "github.com/7045kHz/dtsx/schemas"
)

// PackageTemplate represents a reusable DTSX package template
type PackageTemplate struct {
	Name        string
	Description string
	BasePackage *Package
	Parameters  map[string]string // parameter name -> description
}

// SaveToFile saves the template to a JSON file
func (pt *PackageTemplate) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(pt, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	return nil
}

// LoadFromFile loads a template from a JSON file
func LoadTemplateFromFile(filename string) (*PackageTemplate, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	var template PackageTemplate
	err = json.Unmarshal(data, &template)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal template: %w", err)
	}

	return &template, nil
}

// Instantiate creates a new package from the template with parameter substitution
func (pt *PackageTemplate) Instantiate(params map[string]interface{}) (*Package, error) {
	if pt.BasePackage == nil {
		return nil, fmt.Errorf("template has no base package")
	}

	// Create package with required DTS attributes for SSIS execution
	pkg := &Package{
		RefIdAttr:                      stringPtr("Package"),
		CreationDateAttr:               stringPtr(time.Now().Format("1/2/2006 3:04:05 PM")),
		CreationNameAttr:               stringPtr("Microsoft.Package"),
		CreatorComputerNameAttr:        stringPtr("localhost"),
		CreatorNameAttr:                stringPtr("dtsx-generator"),
		DescriptionAttr:                stringPtr("Generated from DTSX template"),
		DTSIDAttr:                      stringPtr(generateGUID()),
		EnableConfigAttr:               stringPtr("False"),
		ExecutableTypeAttr:             stringPtr("Microsoft.Package"),
		LastModifiedProductVersionAttr: stringPtr("16.0.0.0"),
		LocaleIDAttr:                   stringPtr("1033"),
		ObjectNameAttr:                 stringPtr("{{PackageName}}"),
		PackageTypeAttr:                stringPtr("5"),
		VersionBuildAttr:               stringPtr("1"),
		VersionGUIDAttr:                stringPtr(generateGUID()),
		ExecutableTypePackage:          &schema.ExecutableTypePackage{},
	}

	// Copy variables
	if pt.BasePackage.Variables != nil {
		pkg.Variables = &schema.VariablesType{}
		if pt.BasePackage.Variables.Variable != nil {
			pkg.Variables.Variable = make([]*schema.VariableType, len(pt.BasePackage.Variables.Variable))
			copy(pkg.Variables.Variable, pt.BasePackage.Variables.Variable)
		}
	}

	// Copy connection managers and set required attributes
	if pt.BasePackage.ConnectionManagers != nil {
		pkg.ConnectionManagers = &schema.ConnectionManagersType{}
		if pt.BasePackage.ConnectionManagers.ConnectionManager != nil {
			pkg.ConnectionManagers.ConnectionManager = make([]*schema.ConnectionManagerType, len(pt.BasePackage.ConnectionManagers.ConnectionManager))
			for i, cm := range pt.BasePackage.ConnectionManagers.ConnectionManager {
				// Copy the connection manager
				copiedCM := &schema.ConnectionManagerType{
					RefIdAttr:          cm.RefIdAttr,
					CreationNameAttr:   cm.CreationNameAttr,
					DTSIDAttr:          cm.DTSIDAttr,
					ObjectNameAttr:     cm.ObjectNameAttr,
					Property:           cm.Property,
					PropertyExpression: cm.PropertyExpression,
					ObjectData:         cm.ObjectData,
				}
				// Set required attributes if not present
				if copiedCM.RefIdAttr == nil {
					refId := fmt.Sprintf("Package.ConnectionManagers[%s]", *cm.ObjectNameAttr)
					copiedCM.RefIdAttr = &refId
				}
				if copiedCM.CreationNameAttr == nil {
					copiedCM.CreationNameAttr = stringPtr("OLEDB")
				}
				if copiedCM.DTSIDAttr == nil {
					copiedCM.DTSIDAttr = stringPtr(generateGUID())
				}
				// Add basic ObjectData for OLEDB connection
				// if copiedCM.ObjectData == nil {
				// 	copiedCM.ObjectData = &schema.ConnectionManagerObjectDataType{
				// 		ConnectionManager: &schema.ConnectionManagerObjectDataConnectionManagerType{
				// 			Property: []*schema.Property{},
				// 		},
				// 	}
				// }
				pkg.ConnectionManagers.ConnectionManager[i] = copiedCM
			}
		}
	}

	// Copy basic properties
	if pt.BasePackage.Property != nil {
		pkg.Property = make([]*schema.Property, len(pt.BasePackage.Property))
		copy(pkg.Property, pt.BasePackage.Property)
	}

	// Copy executables and set required attributes
	if pt.BasePackage.Executable != nil {
		pkg.Executable = make([]*schema.AnyNonPackageExecutableType, len(pt.BasePackage.Executable))
		for i, exec := range pt.BasePackage.Executable {
			// Copy the executable
			copiedExec := &schema.AnyNonPackageExecutableType{
				RefIdAttr:              exec.RefIdAttr,
				ExecutableTypeAttr:     exec.ExecutableTypeAttr,
				ObjectNameAttr:         exec.ObjectNameAttr,
				ThreadHintAttr:         exec.ThreadHintAttr,
				ForEachEnumerator:      exec.ForEachEnumerator,
				Property:               exec.Property,
				Variable:               exec.Variable,
				LoggingOptions:         exec.LoggingOptions,
				PropertyExpression:     exec.PropertyExpression,
				Executable:             exec.Executable,
				PrecedenceConstraint:   exec.PrecedenceConstraint,
				ForEachVariableMapping: exec.ForEachVariableMapping,
				EventHandler:           exec.EventHandler,
				ObjectData:             exec.ObjectData,
			}
			// Set required attributes if not present
			if copiedExec.RefIdAttr == nil && copiedExec.ObjectNameAttr != nil {
				refId := fmt.Sprintf("Package\\%s", *exec.ObjectNameAttr)
				copiedExec.RefIdAttr = &refId
			}
			if copiedExec.ObjectNameAttr == nil && copiedExec.RefIdAttr != nil {
				// Extract ObjectName from refId, e.g., "Package\Extract Data" -> "Extract Data"
				if strings.HasPrefix(*exec.RefIdAttr, "Package\\") {
					objectName := (*exec.RefIdAttr)[8:] // Remove "Package\"
					copiedExec.ObjectNameAttr = &objectName
				}
			}
			if copiedExec.DTSIDAttr == nil {
				copiedExec.DTSIDAttr = stringPtr(generateGUID())
			}
			if copiedExec.ExecutableTypeAttr == "" || copiedExec.ExecutableTypeAttr == "ExecuteSQLTask" {
				copiedExec.ExecutableTypeAttr = "Microsoft.ExecuteSQLTask"
			}
			if copiedExec.ExecutableTypeAttr == "DataFlowTask" {
				copiedExec.ExecutableTypeAttr = "SSIS.Pipeline.8"
			}
			if copiedExec.CreationNameAttr == nil {
				if strings.Contains(copiedExec.ExecutableTypeAttr, "ExecuteSQLTask") {
					copiedExec.CreationNameAttr = stringPtr("Microsoft.ExecuteSQLTask")
				} else if copiedExec.ExecutableTypeAttr == "SSIS.Pipeline.8" {
					copiedExec.CreationNameAttr = stringPtr("Microsoft.Pipeline")
				}
			}
			pkg.Executable[i] = copiedExec
		}
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
	// Substitute in package-level attributes
	if pkg.ObjectNameAttr != nil && *pkg.ObjectNameAttr != "" {
		newValue, err := pt.substituteString(*pkg.ObjectNameAttr, params)
		if err != nil {
			return err
		}
		pkg.ObjectNameAttr = &newValue
	}

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

	// Substitute in executables
	if pkg.Executable != nil {
		for _, exec := range pkg.Executable {
			if exec.Property != nil {
				for _, prop := range exec.Property {
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

// LoadFromDirectory loads all templates from a directory
func (tr *TemplateRegistry) LoadFromDirectory(dirPath string) error {
	// Clean the path to ensure platform agnostic behavior
	dirPath = filepath.Clean(dirPath)

	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return fmt.Errorf("templates directory does not exist: %s", dirPath)
	}

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read templates directory: %w", err)
	}

	loadedCount := 0
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		templatePath := filepath.Join(dirPath, file.Name())
		template, err := LoadTemplateFromFile(templatePath)
		if err != nil {
			return fmt.Errorf("failed to load template %s: %w", file.Name(), err)
		}

		tr.Register(template)
		loadedCount++
	}

	if loadedCount == 0 {
		return fmt.Errorf("no template files found in directory: %s", dirPath)
	}

	return nil
}

// SaveToDirectory saves all templates in the registry to a directory
func (tr *TemplateRegistry) SaveToDirectory(dirPath string) error {
	// Clean the path to ensure platform agnostic behavior
	dirPath = filepath.Clean(dirPath)

	// Create directory if it doesn't exist
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	for name, template := range tr.templates {
		// Create platform-agnostic filename
		filename := strings.ReplaceAll(strings.ToLower(name), " ", "_") + ".json"
		filePath := filepath.Join(dirPath, filename)

		err := template.SaveToFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to save template %s: %w", name, err)
		}
	}

	return nil
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

// CreateTableCopyTemplate creates a template for copying data between identical tables
func CreateTableCopyTemplate() *PackageTemplate {
	template := &PackageTemplate{
		Name:        "Table Copy",
		Description: "Template for copying data from one database table to another with identical layout",
		Parameters: map[string]string{
			"SourceConnection":      "Source database connection string",
			"DestinationConnection": "Destination database connection string",
			"SourceTable":           "Source table name (schema.table format)",
			"DestinationTable":      "Destination table name (schema.table format)",
			"PackageName":           "Name for the generated package",
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
							ObjectNameAttr: stringPtr("SourceTable"),
							VariableValue: &schema.VariableValue{
								Value: "{{SourceTable}}",
							},
						},
						{
							NamespaceAttr:  stringPtr("User"),
							ObjectNameAttr: stringPtr("DestinationTable"),
							VariableValue: &schema.VariableValue{
								Value: "{{DestinationTable}}",
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
						RefIdAttr:          stringPtr("Package\\Extract Data"),
						ObjectNameAttr:     stringPtr("Extract Data"),
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
										Value: "SELECT * FROM {{SourceTable}}",
									},
								},
							},
						},
					},
					{
						RefIdAttr:          stringPtr("Package\\Load Data"),
						ObjectNameAttr:     stringPtr("Load Data"),
						ExecutableTypeAttr: "ExecuteSQLTask",
						Property: []*schema.Property{
							{
								NameAttr: stringPtr("Connection"),
								PropertyElementBaseType: &schema.PropertyElementBaseType{
									AnySimpleType: &schema.AnySimpleType{
										Value: "DestinationDB",
									},
								},
							},
							{
								NameAttr: stringPtr("SqlStatementSource"),
								PropertyElementBaseType: &schema.PropertyElementBaseType{
									AnySimpleType: &schema.AnySimpleType{
										Value: "INSERT INTO {{DestinationTable}} SELECT * FROM {{SourceTable}}",
									},
								},
							},
						},
						PrecedenceConstraint: []*schema.PrecedenceConstraintType{
							{
								Executable: []*schema.PrecedenceConstraintExecutableReferenceType{
									{
										IDREFAttr: stringPtr("Package\\Extract Data"),
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

	// Try to load templates from the templates directory
	templatesDir := GetTemplatesDirectory()
	err := registry.LoadFromDirectory(templatesDir)
	if err != nil {
		// If loading fails, create built-in templates and save them
		registry.Register(CreateBasicETLTemplate())
		registry.Register(CreateTableCopyTemplate())
		registry.Register(CreateFileProcessingTemplate())

		// Save the built-in templates to the directory for future use
		saveErr := registry.SaveToDirectory(templatesDir)
		if saveErr != nil {
			// Log the error but don't fail - templates are still available in memory
			fmt.Printf("Warning: Failed to save built-in templates to directory: %v\n", saveErr)
		}
	}

	return registry
}

// GetTemplatesDirectory returns the default templates directory path
func GetTemplatesDirectory() string {
	// Use a relative path that's platform agnostic
	return filepath.Clean("templates")
}

// GetAbsoluteTemplatesDirectory returns the absolute path to the templates directory
func GetAbsoluteTemplatesDirectory() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	return filepath.Join(wd, GetTemplatesDirectory()), nil
}

// LoadTemplatesFromDirectory creates a registry and loads templates from the specified directory
func LoadTemplatesFromDirectory(dirPath string) (*TemplateRegistry, error) {
	registry := NewTemplateRegistry()
	err := registry.LoadFromDirectory(filepath.Clean(dirPath))
	if err != nil {
		return nil, err
	}
	return registry, nil
}

// InitializeTemplatesDirectory creates the templates directory and saves built-in templates
func InitializeTemplatesDirectory(dirPath string) error {
	registry := NewTemplateRegistry()
	registry.Register(CreateBasicETLTemplate())
	registry.Register(CreateTableCopyTemplate())
	registry.Register(CreateFileProcessingTemplate())

	return registry.SaveToDirectory(filepath.Clean(dirPath))
}
