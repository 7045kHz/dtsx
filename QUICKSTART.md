# DTSX Package - Quick Start

## Installation

```bash
go get github.com/7045kHz/dtsx
```

## Basic Usage

### Read a DTSX file

```go
package main

import (
    "fmt"
    "log"
    "github.com/7045kHz/dtsx"
)

func main() {
    pkg, err := dtsx.UnmarshalFromFile("mypackage.dtsx")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Package Type: %s\n", *pkg.ExecutableTypeAttr)
    fmt.Printf("Properties: %d\n", len(pkg.Property))
}
```

### Query Package Contents

```go
// Get all connection managers
connections := pkg.GetConnections()
fmt.Printf("Found %d connections\n", connections.Count)

// Get all variables
variables := pkg.GetVariables()
fmt.Printf("Found %d variables\n", variables.Count)

// Find a specific variable
var, err := pkg.GetVariableByName("User::MyVariable")
if err == nil {
    fmt.Printf("Variable found: %s\n", *var.ObjectNameAttr)
}

// Get all expressions in the package
expressions := pkg.GetExpressions()
fmt.Printf("Found %d expressions\n", expressions.Count)
```

### Update Package Elements

Modify existing DTSX package elements safely with validation:

```go
// Update a variable value
err := pkg.UpdateVariable("User", "MyVariable", "NewValue")
if err != nil {
    log.Printf("Failed to update variable: %v", err)
}

// Update a connection string
err = pkg.UpdateConnectionString("MyConnection", "Server=newserver;Database=mydb")
if err != nil {
    log.Printf("Failed to update connection: %v", err)
}

// Update a property expression
err = pkg.UpdateExpression("Variable", "User::MyVar", "Value", "@[System::StartTime]")
if err != nil {
    log.Printf("Failed to update expression: %v", err)
}

// Update any property on any element (package, variable, connection, executable)
err = pkg.UpdateProperty("executable", "MyTask", "Description", "Updated task description")
if err != nil {
    log.Printf("Failed to update property: %v", err)
}

// Save the updated package
err = dtsx.MarshalToFile("updated_package.dtsx", pkg)
if err != nil {
    log.Fatal(err)
}
```

## Advanced Features

### Expression Evaluation

Evaluate SSIS expressions with full support for variables, arithmetic, string operations, functions, conditionals, and type casting:

```go
// Basic arithmetic and variables
result, err := dtsx.EvaluateExpression("@[User::MyVar] + 1", pkg)
fmt.Printf("Result: %v\n", result) // Output: 43 (if MyVar = 42)

// String operations and concatenation
result, err = dtsx.EvaluateExpression("@[User::Name] + \"_suffix\"", pkg)

// Built-in functions
result, err = dtsx.EvaluateExpression("UPPER(@[User::Name])", pkg)
result, err = dtsx.EvaluateExpression("SUBSTRING(@[User::Path], 4, 10)", pkg)
result, err = dtsx.EvaluateExpression("GETDATE()", pkg)

// Date functions
result, err = dtsx.EvaluateExpression("DATEADD(\"DAY\", 7, @[User::StartDate])", pkg)
result, err = dtsx.EvaluateExpression("DATEDIFF(\"DAY\", @[User::StartDate], GETDATE())", pkg)

// Conditional expressions
result, err = dtsx.EvaluateExpression("@[User::Count] > 10 ? \"High\" : \"Low\"", pkg)

// Type casting
result, err = dtsx.EvaluateExpression("(DT_STR, 10, 1252) @[User::Number]", pkg)

// Comparison and logical operators
result, err = dtsx.EvaluateExpression("@[User::Value] == 100 && @[User::Status] == \"Active\"", pkg)
```

**Supported Functions:**

- String: `UPPER`, `LOWER`, `SUBSTRING`, `REPLACE`, `LEN`
- Math: `ABS`, `CEILING`, `FLOOR`
- Date: `GETDATE`, `YEAR`, `MONTH`, `DAY`, `DATEADD`, `DATEDIFF`

**Supported Operators:**

- Arithmetic: `+`, `-`, `*`, `/`
- Comparison: `==`, `!=`, `<`, `>`, `<=`, `>=`
- Logical: `&&`, `||`, `!`
- Conditional: `? :`
- Type casting: `(DT_STR)`, `(DT_INT)`, `(DT_DECIMAL)`, `(DT_BOOL)`

### Package Builder API

Create DTSX packages programmatically using the fluent builder pattern:

```go
// Create a new package with variables of different types and connections
pkg := dtsx.NewPackageBuilder().
    AddVariableWithType("User", "InputPath", "C:\\data\\input.csv", "String").
    AddVariableWithType("User", "BatchSize", "1000", "Int32").
    AddVariableWithType("User", "ProcessEnabled", "true", "Boolean").
    AddVariableWithType("User", "ProcessDate", "2025-01-01", "DateTime").
    AddConnection("SourceDB", "OLEDB", "Server=myserver;Database=mydb;Trusted_Connection=True;").
    AddConnectionExpression("SourceDB", "ConnectionString", "@[User::ConnectionStringVar]").
    Build()

// Save the package
err := dtsx.MarshalToFile("newpackage.dtsx", pkg)
```

### Package Validation

Validate packages for common issues, orphaned variables, and best practices:

```go
// Validate the package
errors := pkg.Validate()
for _, err := range errors {
    fmt.Printf("[%s] %s: %s\n", err.Severity, err.Path, err.Message)
}

// Severities: "error", "warning", "info"
// Checks for:
// - Duplicate names
// - Missing values
// - Orphaned variables (not referenced in expressions)
// - Undefined variable references
// - Structural issues
```

### Dependency Analysis

Analyze relationships between package elements for impact analysis and optimization:

```go
// Build dependency graph
graph := pkg.BuildDependencyGraph()

// Get all locations affected by changing a variable
impact := graph.GetVariableImpact("User::MyVariable")
fmt.Printf("Variable used in %d locations\n", len(impact))

// Get tasks affected by a connection change
connImpact := graph.GetConnectionImpact("MyConnection")
fmt.Printf("Connection used by %d tasks\n", len(connImpact))

// Find unused variables
unused := pkg.GetUnusedVariables()
for _, v := range unused {
    fmt.Printf("Unused variable: %s\n", v)
}

// Get optimization suggestions
suggestions := pkg.GetOptimizationSuggestions()
for _, s := range suggestions {
    fmt.Printf("[%s] %s\n", s.Severity, s.Message)
}
```

### Template System

Use reusable package templates for common ETL patterns:

```go
// Get the default template registry
registry := dtsx.GetDefaultTemplateRegistry()

// List available templates
templates := registry.List()
fmt.Printf("Available templates: %v\n", templates)

// Get a specific template
template := registry.Get("Basic ETL")
if template != nil {
    fmt.Printf("Template: %s - %s\n", template.Name, template.Description)
    fmt.Printf("Parameters: %v\n", template.Parameters)
}

// Instantiate a template with custom parameters
params := map[string]interface{}{
    "SourceConnection":     "Server=myserver;Database=sourcedb",
    "DestinationConnection": "Server=myserver;Database=destdb",
    "SourceQuery":          "SELECT id, name FROM customers",
    "DestinationTable":     "customers_backup",
    "BatchSize":           "5000",
}

pkg, err := template.Instantiate(params)
if err != nil {
    log.Fatal(err)
}

// The template is now customized and ready to use
err = dtsx.MarshalToFile("custom_etl.dtsx", pkg)
```

Built-in templates include:

- **Basic ETL**: Extract-Transform-Load workflow
- **File Processing**: File operations with error handling

### Write a DTSX file

```go
err := dtsx.MarshalToFile("output.dtsx", pkg)
if err != nil {
    log.Fatal(err)
}
```

## Running Examples

```bash
# Analyze a DTSX package structure
go run examples/analyze_dtsx.go SSIS_EXAMPLES/ConfigFile.dtsx

# Demonstrate query methods with variable values and expressions
go run examples/query_dtsx.go SSIS_EXAMPLES/Expressions.dtsx

# Comprehensive connection analysis with expressions and evaluated values
go run examples/analyze_connections.go SSIS_EXAMPLES/Expressions.dtsx

# Advanced expression evaluation with all SSIS features
go run examples/evaluate_expressions.go SSIS_EXAMPLES/Expressions.dtsx

# Comprehensive package validation and analysis
go run examples/validate_dtsx.go SSIS_EXAMPLES/ConfigFile.dtsx

# Programmatic package creation with builder API
go run examples/build_package.go

# Use reusable package templates
go run examples/use_templates.go

# Run package with DTExec
go run examples/run_dtsx.go SSIS_EXAMPLES/ConfigFile.dtsx

# Run package with parameters
go run examples/run_with_params.go SSIS_EXAMPLES/ConfigFile.dtsx

# Debug output as JSON
go run examples/debug_dtsx.go SSIS_EXAMPLES/ConfigFile.dtsx
```

## Running Tests

```bash
go test ./...
```

## Schema Support

The library includes full support for DTSX package structures including:

- **Package Properties**: All package-level configuration
- **Connection Managers**: OLE DB, Flat File, and other connection types with expressions
- **Variables**: User and system variables with default values and expressions
- **Executables**: Tasks, containers, and control flow elements
- **Expressions**: SSIS expressions that reference variables and properties
- **Container Elements**: Proper handling of `<Variables>`, `<ConnectionManagers>`, etc.

The schema types are generated from official Microsoft SSIS XSD files and support the complete DTSX structure for reading and writing packages.

## Package Structure

- `dtsx.go` - Main package API (Unmarshal/Marshal functions)
- `dtsx/schemas/` - Generated Go types from XSD schemas
- `examples/` - Example programs
- `SSIS_EXAMPLES/` - Sample DTSX files for testing
