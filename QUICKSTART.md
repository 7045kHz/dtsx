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

// Get detailed expression analysis
exprs := expressions.Results.([]*dtsx.ExpressionInfo)
for _, expr := range exprs {
    details := dtsx.GetExpressionDetails(expr, pkg)
    fmt.Printf("Expression: %s (Location: %s)\n", details.Expression, details.Location)
    if details.EvaluatedValue != "" {
        fmt.Printf("  Evaluated: %s\n", details.EvaluatedValue)
    }
    if len(details.Dependencies) > 0 {
        fmt.Printf("  Dependencies: %v\n", details.Dependencies)
    }
}
```

### Using Utility Functions

Convenient getter functions for common DTSX element properties:

```go
// Get connection manager details
connName := dtsx.GetConnectionName(connectionManager)
connString := dtsx.GetConnectionString(connectionManager)

// Get variable details
varName := dtsx.GetVariableName(variable)
varValue := dtsx.GetVariableValue(variable)

// Get executable details
execName := dtsx.GetExecutableName(executable)
```

### Update Package Elements

Modify existing DTSX package elements safely with validation:

```go
// Update a variable value by mutating the package structs directly
if pkg.Variables != nil && len(pkg.Variables.Variable) > 0 {
    v := pkg.Variables.Variable[0]
    if v.VariableValue != nil {
        v.VariableValue.Value = "NewValue"
    } else {
        v.VariableValue = &schema.VariableValue{Value: "NewValue"}
    }
}

// Update a connection string by mutating connection manager properties directly
if pkg.ConnectionManagers != nil && len(pkg.ConnectionManagers.ConnectionManager) > 0 {
    for _, cm := range pkg.ConnectionManagers.ConnectionManager {
        var name string
        for _, prop := range cm.Property {
            if prop.NameAttr != nil && *prop.NameAttr == "ObjectName" && prop.PropertyElementBaseType != nil && prop.PropertyElementBaseType.AnySimpleType != nil {
                name = prop.PropertyElementBaseType.AnySimpleType.Value
            }
            if prop.NameAttr != nil && *prop.NameAttr == "ConnectionString" && prop.PropertyElementBaseType != nil && prop.PropertyElementBaseType.AnySimpleType != nil && name == "MyConnection" {
                prop.PropertyElementBaseType.AnySimpleType.Value = "Server=newserver;Database=mydb"
            }
        }
    }
}

// Update a property expression by modifying the package struct (for example variable value expression)
// This is an illustrative example; application logic should find and update the correct target
if pkg.Variables != nil && len(pkg.Variables.Variable) > 0 {
    for _, v := range pkg.Variables.Variable {
        if v.NamespaceAttr != nil && v.ObjectNameAttr != nil && *v.NamespaceAttr+"::"+*v.ObjectNameAttr == "User::MyVar" {
            if v.PropertyExpression == nil {
                v.PropertyExpression = []*schema.PropertyExpressionElementType{}
            }
            v.PropertyExpression = append(v.PropertyExpression, &schema.PropertyExpressionElementType{
                NameAttr: "Value",
                AnySimpleType: &schema.AnySimpleType{Value: "@[System::StartTime]"},
            })
            break
        }
    }
}

// Update any property on any element by mutating the struct directly (e.g., executable's description)
if len(pkg.Executable) > 0 {
    for _, exec := range pkg.Executable {
        for _, prop := range exec.Property {
            if prop.NameAttr != nil && *prop.NameAttr == "Description" {
                prop.Value = "Updated task description"
            }
        }
    }
}

// Serialize then write the updated package (write helpers are internalized)
data, err := dtsx.Marshal(pkg)
if err != nil {
    log.Fatal(err)
}
// Use the stdlib to write files
// err = os.WriteFile("updated_package.dtsx", data, 0644)
// if err != nil {
//     log.Fatal(err)
// }
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

// Serialize and write the package (manual write)
data, err := dtsx.Marshal(pkg)
if err != nil {
    log.Fatal(err)
}
// err := os.WriteFile("newpackage.dtsx", data, 0644)
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

### Package Parser

Use the centralized PackageParser for efficient analysis with built-in caching:

```go
// Create parser for advanced analysis
parser := dtsx.NewPackageParser(pkg)

// Extract all SQL statements from control flow and dataflow tasks
sqlStatements := parser.GetSQLStatements()
for _, stmt := range sqlStatements {
    fmt.Printf("Task: %s (%s)\n", stmt.TaskName, stmt.TaskType)
    fmt.Printf("  SQL: %s\n", stmt.SQL)
    if len(stmt.Connections) > 0 {
        fmt.Printf("  Connections: %v\n", stmt.Connections)
    }
}

// Evaluate expressions with automatic caching for performance
result, err := parser.EvaluateExpression("@[User::MyVar] + @[User::Increment]")
if err == nil {
    fmt.Printf("Result: %v\n", result)
}

// Access variables and connections efficiently
varValue, err := parser.GetVariableValue("User::MyVariable")
connMgr, err := parser.GetConnectionManager("MyConnection")
```

### Execution Order Analysis

Analyze task execution order and precedence constraints:

```go
// Create precedence analyzer
analyzer := dtsx.NewPrecedenceAnalyzer(pkg)

// Get execution order for specific tasks
order, err := analyzer.GetExecutionOrder("Package\\DataFlowTask")
if err == nil {
    fmt.Printf("Task executes at order: %d\n", order)
}

// Get execution orders for all tasks
allOrders, err := analyzer.GetAllExecutionOrders()
if err == nil {
    for refId, order := range allOrders {
        fmt.Printf("%s: Order %d\n", refId, order)
    }
}

// Get textual execution flow description
flowDesc := analyzer.GetExecutionFlowDescription()
fmt.Print(flowDesc)

// Get the execution chain (all predecessors) for a task
chain, err := analyzer.GetExecutableChain("Package\\FinalTask")
if err == nil {
    fmt.Printf("Execution chain: %v\n", chain)
}

// Validate precedence constraints for circular dependencies
constraintErrors := analyzer.ValidateConstraints()
for _, err := range constraintErrors {
    fmt.Printf("Constraint error: %v\n", err)
}
```

### Enhanced Package Validation

Use the comprehensive PackageValidator for detailed package analysis:

```go
// Create validator
validator := dtsx.NewPackageValidator(pkg)

// Perform comprehensive validation
validationErrors := validator.Validate()

// Display results by severity
for _, err := range validationErrors {
    fmt.Printf("[%s] %s: %s\n", err.Severity, err.Path, err.Message)
    if err.Severity == "error" {
        // Handle critical errors
    }
}
```

The validator checks for:

- Expression evaluation errors
- Missing or invalid connection properties
- Precedence constraint violations
- Variable scoping issues
- Structural problems

### Write a DTSX file

```go
data, err := dtsx.Marshal(pkg)
if err != nil {
    log.Fatal(err)
}
// err = os.WriteFile("output.dtsx", data, 0644)
// if err != nil {
//     log.Fatal(err)
// }
```

## Running Examples

```bash
# Analyze a DTSX package structure
go run examples/analyze_dtsx.go SSIS_EXAMPLES/ConfigFile.dtsx

# Demonstrate query methods with variable values and expressions
go run examples/query_dtsx.go SSIS_EXAMPLES/Expressions.dtsx

# Comprehensive connection analysis with expressions and evaluated values
go run examples/analyze_connections.go SSIS_EXAMPLES/Expressions.dtsx

# Advanced package analysis with parser, validator, and analyzer
go run examples/package_analysis.go SSIS_EXAMPLES/Expressions.dtsx

# Advanced expression evaluation with all SSIS features
go run examples/evaluate_expressions.go SSIS_EXAMPLES/Expressions.dtsx

# Comprehensive package validation and analysis
go run examples/validate_dtsx.go SSIS_EXAMPLES/ConfigFile.dtsx

# Programmatic package creation with builder API
go run examples/build_package.go

# Use reusable package templates
go run examples/use_templates.go

# Custom template directory management
go run examples/custom_templates.go

# Use reusable package templates
go run examples/use_templates.go

# Custom template directory management
go run examples/custom_templates.go

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

- `dtsx.go` - Main package API, PackageParser, PrecedenceAnalyzer, PackageValidator
- `expression.go` - Advanced SSIS expression evaluator with caching
- `dtsx/schemas/` - Generated Go types from XSD schemas
- `examples/` - Example programs including package_analysis.go
- `SSIS_EXAMPLES/` - Sample DTSX files for testing
