# DTSX Package Templates

The DTSX library provides a powerful template system for creating reusable SSIS package patterns. Templates allow you to define common ETL workflows that can be customized with parameters, making it easy to create consistent, validated packages for recurring data integration scenarios.

## Overview

Templates are predefined package structures with:

- **Parameterized configuration**: Variables, connection strings, and properties can be customized
- **Built-in validation**: Templates generate valid DTSX packages
- **Dependency management**: Automatic handling of connections and variables
- **File-based storage**: Templates are stored as JSON files in a configurable directory
- **Extensibility**: Easy to create custom templates for your organization

## Template Storage

Templates are stored as JSON files in a templates directory. By default, templates are loaded from a `templates/` directory in the current working directory. You can:

- **Load from custom directories**: Use `LoadTemplatesFromDirectory()` to load templates from any directory
- **Save templates to files**: Use `template.SaveToFile()` to save individual templates
- **Platform agnostic**: All path operations use Go's `filepath` package for cross-platform compatibility
- **Automatic directory creation**: Template directories are created automatically when needed
- **Initialize directories**: Use `InitializeTemplatesDirectory()` to create a directory with built-in templates
- **Automatic fallback**: If no template files exist, built-in templates are automatically created and saved

## Built-in Templates

The library comes with three built-in templates that are automatically saved to the templates directory:

### Basic ETL Template

A fundamental Extract-Transform-Load workflow template that includes:

- Source and destination database connections
- SQL extraction task
- Data flow task for transformation and loading
- Configurable batch size and query parameters

**Parameters:**

- `SourceConnection`: Source database connection string
- `DestinationConnection`: Destination database connection string
- `SourceQuery`: SQL query for data extraction
- `DestinationTable`: Destination table name
- `BatchSize`: Batch size for data loading
- `PackageName`: Name for the generated package

### Table Copy Template

A simple template for copying data between database tables with identical layouts:

- Source and destination database connections
- Direct table-to-table data transfer using INSERT...SELECT
- Configurable source and destination table names
- Automatic precedence constraint between extract and load tasks

**Parameters:**

- `SourceConnection`: Source database connection string
- `DestinationConnection`: Destination database connection string
- `SourceTable`: Source table name (schema.table format)
- `DestinationTable`: Destination table name (schema.table format)
- `PackageName`: Name for the generated package

### File Processing Template

A template for file-based processing workflows with:

- Input/output path configuration
- File enumeration using ForEachLoop
- Process execution for file operations
- Error handling and logging variables

**Parameters:**

- `InputPath`: Path to input files
- `OutputPath`: Path for processed output
- `ArchivePath`: Path for archiving processed files
- `ErrorThreshold`: Maximum number of errors before failing

## Using Templates

### Getting Started

```go
package main

import (
    "fmt"
    "log"
    "github.com/7045kHz/dtsx"
)

func main() {
    // Get the default template registry
    registry := dtsx.GetDefaultTemplateRegistry()

    // List available templates
    templates := registry.List()
    fmt.Printf("Available templates: %v\n", templates)
}
```

### Instantiating a Template

```go
// Get a specific template
template := registry.Get("Basic ETL")
if template == nil {
    log.Fatal("Template not found")
}

// Display template information
fmt.Printf("Template: %s\n", template.Name)
fmt.Printf("Description: %s\n", template.Description)
fmt.Println("Required parameters:")
for param, desc := range template.Parameters {
    fmt.Printf("  %s: %s\n", param, desc)
}

// Prepare parameters
params := map[string]interface{}{
    "SourceConnection":      "Server=localhost;Database=SourceDB;Trusted_Connection=True",
    "DestinationConnection": "Server=localhost;Database=DestDB;Trusted_Connection=True",
    "SourceQuery":           "SELECT * FROM Customers WHERE Status = 'Active'",
    "DestinationTable":      "Customers_Backup",
    "BatchSize":             "10000",
    "PackageName":           "CustomerMigration",
}

// Instantiate the template
pkg, err := template.Instantiate(params)
if err != nil {
    log.Fatalf("Failed to instantiate template: %v", err)
}

fmt.Printf("Package created: %s\n", pkg.Property[0].Value)
```

### File-Based Template Operations

Templates are stored as JSON files and can be loaded from custom directories:

```go
// Load templates from a custom directory
registry, err := dtsx.LoadTemplatesFromDirectory("/path/to/my/templates")
if err != nil {
    log.Fatalf("Failed to load templates: %v", err)
}

// Save a template to a file
err = template.SaveToFile("my_custom_template.json")
if err != nil {
    log.Fatalf("Failed to save template: %v", err)
}

// Initialize a directory with built-in templates
err = dtsx.InitializeTemplatesDirectory("/path/to/templates")
if err != nil {
    log.Fatalf("Failed to initialize templates directory: %v", err)
}
```

### Validating Generated Packages

```go
// Validate the generated package
validator := dtsx.NewPackageValidator(pkg)
errors := validator.Validate()

if len(errors) > 0 {
    fmt.Printf("Validation issues found:\n")
    for _, err := range errors {
        fmt.Printf("  [%s] %s: %s\n", err.Severity, err.Path, err.Message)
    }
} else {
    fmt.Println("✓ Package validation passed")
}
```

### Saving Templates to Files

```go
// Save the package to a DTSX file
outputFile := "my_etl_package.dtsx"
err = dtsx.MarshalToFile(outputFile, pkg)
if err != nil {
    log.Fatalf("Failed to save package: %v", err)
}
fmt.Printf("Package saved to: %s\n", outputFile)
```

## Parameter Substitution

Templates use `{{ParameterName}}` placeholders that get replaced during instantiation:

```xml
<!-- In template definition -->
<Property Name="ConnectionString">{{SourceConnection}}</Property>
<Variable Name="BatchSize">{{BatchSize}}</Variable>

<!-- After instantiation with params["SourceConnection"] = "Server=prod;Database=Sales" -->
<Property Name="ConnectionString">Server=prod;Database=Sales</Property>
```

## Creating Custom Templates

### Template Structure

```go
type PackageTemplate struct {
    Name        string
    Description string
    BasePackage *Package
    Parameters  map[string]string // parameter name -> description
}
```

### Example Custom Template

```go
func CreateCustomDataWarehouseTemplate() *PackageTemplate {
    template := &PackageTemplate{
        Name:        "Data Warehouse Load",
        Description: "Template for loading data into a star schema warehouse",
        Parameters: map[string]string{
            "WarehouseConnection": "Data warehouse connection string",
            "StagingConnection":   "Staging database connection string",
            "FactTable":          "Target fact table name",
            "DimensionTables":    "Comma-separated list of dimension tables",
        },
        BasePackage: &Package{
            ExecutableTypePackage: &schema.ExecutableTypePackage{
                Variables: &schema.VariablesType{
                    Variable: []*schema.VariableType{
                        {
                            NamespaceAttr:  stringPtr("User"),
                            ObjectNameAttr: stringPtr("WarehouseConn"),
                            VariableValue: &schema.VariableValue{
                                Value: "{{WarehouseConnection}}",
                            },
                        },
                    },
                },
                ConnectionManagers: &schema.ConnectionManagersType{
                    ConnectionManager: []*schema.ConnectionManagerType{
                        {
                            ObjectNameAttr: stringPtr("WarehouseDB"),
                            Property: []*schema.Property{
                                {
                                    NameAttr: stringPtr("ConnectionString"),
                                    PropertyElementBaseType: &schema.PropertyElementBaseType{
                                        AnySimpleType: &schema.AnySimpleType{
                                            Value: "{{WarehouseConnection}}",
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
                                        Value: "WarehouseDB",
                                    },
                                },
                            },
                            {
                                NameAttr: stringPtr("SqlStatementSource"),
                                PropertyElementBaseType: &schema.PropertyElementBaseType{
                                    AnySimpleType: &schema.AnySimpleType{
                                        Value: "INSERT INTO {{FactTable}} SELECT * FROM StagingTable",
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
```

### Registering Custom Templates

```go
// Create a custom registry
registry := dtsx.NewTemplateRegistry()

// Register built-in templates
registry.Register(dtsx.CreateBasicETLTemplate())
registry.Register(dtsx.CreateFileProcessingTemplate())

// Register custom templates
registry.Register(CreateCustomDataWarehouseTemplate())

// Use the registry
template := registry.Get("Data Warehouse Load")
```

## Template Registry Management

### Creating a Registry

```go
// Create a new registry
registry := dtsx.NewTemplateRegistry()

// Register templates
registry.Register(myTemplate)

// Get templates
template := registry.Get("My Template")
templates := registry.List() // Returns []string of template names
```

### Default Registry

```go
// Get the default registry with built-in templates
registry := dtsx.GetDefaultTemplateRegistry()
```

## Best Practices

### Template Design

1. **Parameter naming**: Use clear, descriptive parameter names
2. **Default values**: Provide sensible defaults where possible
3. **Validation**: Include validation logic for parameter values
4. **Documentation**: Document all parameters and their expected formats

### Parameter Handling

1. **Type safety**: Specify parameter types in documentation
2. **Required vs optional**: Clearly mark required parameters
3. **Validation**: Validate parameter values during instantiation
4. **Substitution**: Use consistent placeholder syntax `{{ParamName}}`

### Package Structure

1. **Modularity**: Break complex workflows into smaller, reusable templates
2. **Naming conventions**: Use consistent naming for connections and variables
3. **Error handling**: Include appropriate error handling in templates
4. **Logging**: Add logging configurations to templates

## Advanced Usage

### Template Composition

```go
// Create composite templates by instantiating other templates
func CreateAdvancedETLTemplate() *PackageTemplate {
    // Get base ETL template
    baseTemplate := dtsx.CreateBasicETLTemplate()

    // Instantiate with base parameters
    basePkg, _ := baseTemplate.Instantiate(map[string]interface{}{
        "SourceConnection":      "{{SourceConnection}}",
        "DestinationConnection": "{{DestinationConnection}}",
        "SourceQuery":           "{{SourceQuery}}",
        "DestinationTable":      "{{DestinationTable}}",
        "BatchSize":             "{{BatchSize}}",
        "PackageName":           "{{PackageName}}",
    })

    // Add additional components
    // ... extend basePkg with additional executables, variables, etc.

    return &PackageTemplate{
        Name:        "Advanced ETL",
        Description: "Extended ETL template with additional processing steps",
        BasePackage: basePkg,
        Parameters: map[string]string{
            "SourceConnection":      "Source database connection string",
            "DestinationConnection": "Destination database connection string",
            "SourceQuery":           "SQL query for data extraction",
            "DestinationTable":      "Destination table name",
            "BatchSize":             "Batch size for data loading",
            "PackageName":           "Name for the generated package",
            "PostProcessQuery":      "Optional post-processing SQL query",
        },
    }
}
```

### Template Validation

```go
// Add custom validation to templates
func (pt *PackageTemplate) ValidateParameters(params map[string]interface{}) []error {
    var errors []error

    // Check required parameters
    required := []string{"SourceConnection", "DestinationConnection"}
    for _, param := range required {
        if _, ok := params[param]; !ok {
            errors = append(errors, fmt.Errorf("required parameter %s is missing", param))
        }
    }

    // Validate connection strings
    if conn, ok := params["SourceConnection"].(string); ok {
        if !strings.Contains(conn, "Server=") {
            errors = append(errors, fmt.Errorf("SourceConnection must contain Server="))
        }
    }

    return errors
}
```

## Examples

See `examples/use_templates.go` for a complete working example of template usage, including:

- Listing available templates
- Instantiating templates with parameters
- Validating generated packages
- Saving packages to files
- Analyzing package dependencies

## API Reference

### PackageTemplate

- `Instantiate(params map[string]interface{}) (*Package, error)`: Create a package from the template
- `Name`: Template name
- `Description`: Template description
- `Parameters`: Map of parameter names to descriptions

### TemplateRegistry

- `NewTemplateRegistry() *TemplateRegistry`: Create a new registry
- `Register(template *PackageTemplate)`: Add a template to the registry
- `Get(name string) *PackageTemplate`: Retrieve a template by name
- `List() []string`: Get all template names

### Built-in Functions

- `GetDefaultTemplateRegistry() *TemplateRegistry`: Get registry with built-in templates
- `CreateBasicETLTemplate() *PackageTemplate`: Create basic ETL template
- `CreateFileProcessingTemplate() *PackageTemplate`: Create file processing template</content>
  <parameter name="filePath">c:\Users\U00001\source\repos\dtsx\TEMPLATES.md
