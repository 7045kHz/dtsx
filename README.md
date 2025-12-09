# DTSX - SQL Server Integration Services Package Parser

A comprehensive Go library for reading, writing, and analyzing DTSX (SQL Server Integration Services) XML files with advanced expression evaluation and package management capabilities.

## Features

- **Parse DTSX files**: Read and unmarshal DTSX XML files into Go structs
- **Generate DTSX files**: Marshal Go structs back to DTSX XML format
- **Update DTSX elements**: Modify existing variables, connections, expressions, and any properties safely
- **Advanced Expression Engine**: Full SSIS expression evaluation with variables, functions, conditionals, and type casting
- **Package Builder API**: Fluent API for programmatic package creation
- **Comprehensive Validation**: Multi-level validation with error, warning, and info severities
- **Dependency Analysis**: Graph-based analysis of package relationships and impact assessment
- **Template System**: Reusable package templates for common ETL patterns
- **Query API**: Convenient methods to analyze connections, variables, executables, and expressions
- **Connection Analysis**: Comprehensive analysis of connection managers with drivers and dynamic properties
- **Full schema support**: Generated from official SSIS XSD schemas with container element support
- **Type-safe**: Strongly typed Go structures for all DTSX elements

## Installation

```bash
go get github.com/7045kHz/dtsx
```

## Usage

### Reading a DTSX File

```go
package main

import (
    "fmt"
    "log"

    "github.com/7045kHz/dtsx"
)

func main() {
    // Load DTSX file
    pkg, err := dtsx.UnmarshalFromFile("mypackage.dtsx")
    if err != nil {
        log.Fatal(err)
    }

    // Access package properties using query methods
    connections := pkg.GetConnections()
    variables := pkg.GetVariables()

    fmt.Printf("Connection Managers: %d\n", connections.Count)
    fmt.Printf("Variables: %d\n", variables.Count)

    // Access executables directly
    if pkg.Executable != nil {
        fmt.Printf("Executables: %d\n", len(pkg.Executable))
    }
}
```

### Reading from a byte slice or io.Reader

```go
// From bytes
data := []byte(xmlContent)
pkg, err := dtsx.Unmarshal(data)

// From io.Reader
file, _ := os.Open("package.dtsx")
pkg, err := dtsx.UnmarshalFromReader(file)
```

### Writing a DTSX File

```go
// Marshal to bytes
data, err := dtsx.Marshal(pkg)

// Write to file
err := dtsx.MarshalToFile("output.dtsx", pkg)

// Write to io.Writer
err := dtsx.MarshalToWriter(writer, pkg)
```

## Advanced Features

### Expression Evaluation

Evaluate SSIS expressions with full support for variables, arithmetic, functions, conditionals, and type casting:

```go
// Basic arithmetic and variables
result, err := dtsx.EvaluateExpression("@[User::MyVar] + 1", pkg)

// Built-in functions
result, err = dtsx.EvaluateExpression("UPPER(@[User::Name])", pkg)
result, err = dtsx.EvaluateExpression("DATEADD(\"DAY\", 7, @[User::StartDate])", pkg)

// Conditional expressions
result, err = dtsx.EvaluateExpression("@[User::Count] > 10 ? \"High\" : \"Low\"", pkg)

// Type casting
result, err = dtsx.EvaluateExpression("(DT_STR) @[User::Number]", pkg)
```

### Package Builder API

Create DTSX packages programmatically:

```go
pkg := dtsx.NewPackageBuilder().
    AddVariable("User", "InputPath", "C:\\data\\input.csv").
    AddVariable("User", "OutputPath", "C:\\data\\output.csv").
    AddConnection("SourceDB", "OLEDB", "Server=myserver;Database=mydb;Trusted_Connection=True;").
    Build()
```

### Package Validation

Validate packages for common issues:

```go
errors := pkg.Validate()
for _, err := range errors {
    fmt.Printf("[%s] %s: %s\n", err.Severity, err.Path, err.Message)
}
```

### Dependency Analysis

Analyze relationships between package elements:

```go
graph := pkg.BuildDependencyGraph()
impact := graph.GetVariableImpact("User::MyVariable")
fmt.Printf("Variable used in %d locations\n", len(impact))
```

### Template System

Use reusable package templates:

```go
registry := dtsx.GetDefaultTemplateRegistry()
template := registry.Get("Basic ETL")
pkg, err := template.Instantiate(params)
```

## Querying Packages

The library provides convenient query methods for analyzing DTSX packages:

### Get Connection Managers

```go
connections := pkg.GetConnections()
fmt.Printf("Found %d connection managers\n", connections.Count)
connMgrs := connections.Results.([]*schema.ConnectionManagerType)
```

### Get Variables

```go
variables := pkg.GetVariables()
fmt.Printf("Found %d variables\n", variables.Count)
vars := variables.Results.([]*schema.VariableType)
```

### Find Specific Variable

```go
variable, err := pkg.GetVariableByName("User::MyVariable")
if err != nil {
    fmt.Println("Variable not found")
}
```

### Query Executables with Filters

```go
// Get all executables
allExecutables := pkg.QueryExecutables(func(*schema.AnyNonPackageExecutableType) bool {
    return true
})

// Find SQL tasks
sqlTasks := pkg.QueryExecutables(func(exec *schema.AnyNonPackageExecutableType) bool {
    return exec.ExecutableTypeAttr == "STOCK:SQLTask"
})

// Find tasks with expressions
tasksWithExpressions := pkg.QueryExecutables(func(exec *schema.AnyNonPackageExecutableType) bool {
    return len(exec.PropertyExpression) > 0
})
```

### Get All Expressions

```go
expressions := pkg.GetExpressions()
fmt.Printf("Found %d expressions\n", expressions.Count)
exprs := expressions.Results.([]*dtsx.ExpressionInfo)

for _, expr := range exprs {
    fmt.Printf("Expression: %s (Location: %s, Context: %s)\n",
        expr.Expression, expr.Location, expr.Context)
}
```

## API Reference

### Core Functions

- `UnmarshalFromFile(filename string) (*Package, error)` - Read DTSX from file
- `UnmarshalFromReader(r io.Reader) (*Package, error)` - Read DTSX from reader
- `Unmarshal(data []byte) (*Package, error)` - Parse DTSX from bytes
- `MarshalToFile(filename string, pkg *Package) error` - Write DTSX to file
- `MarshalToWriter(w io.Writer, pkg *Package) error` - Write DTSX to writer
- `Marshal(pkg *Package) ([]byte, error)` - Convert DTSX to bytes
- `IsDTSXPackage(filename string) (*Package, bool)` - Load and validate DTSX file

### Execution Functions

- `RunPackage(dtexecPath, dtsxPath string, opts *RunOptions) (string, error)` - Execute DTSX package with dtexec.exe

### Query Methods

- `GetConnections() *QueryResult` - Get all connection managers
- `GetVariables() *QueryResult` - Get all variables
- `GetVariableByName(name string) (*schema.VariableType, error)` - Find variable by name
- `QueryExecutables(filter func(*schema.AnyNonPackageExecutableType) bool) []*schema.AnyNonPackageExecutableType` - Filter executables
- `GetExpressions() *QueryResult` - Get all expressions with context

### Update Methods

- `UpdateVariable(namespace, name, newValue string) error` - Update variable value
- `UpdateConnectionString(connectionName, newConnectionString string) error` - Update connection string
- `UpdateExpression(targetType, targetName, propertyName, newExpression string) error` - Update property expression
- `UpdateProperty(targetType, targetName, propertyName, newValue string) error` - Update any property on package, variable, connection, or executable

### Advanced Methods

- `EvaluateExpression(expr string, pkg *Package) (interface{}, error)` - Evaluate SSIS expression
- `Validate() []ValidationError` - Validate package for issues
- `BuildDependencyGraph() *DependencyGraph` - Build dependency graph
- `GetUnusedVariables() []string` - Find unused variables
- `GetOptimizationSuggestions() []ValidationError` - Get optimization suggestions

### Builder API

- `NewPackageBuilder() *PackageBuilder` - Create new package builder
- `AddVariable(namespace, name, value string) *PackageBuilder` - Add string variable
- `AddVariableWithType(namespace, name, value, dataType string) *PackageBuilder` - Add variable with specific data type
- `AddConnection(name, connectionType, connectionString string) *PackageBuilder` - Add connection manager
- `AddConnectionExpression(connectionName, propertyName, expression string) *PackageBuilder` - Add expression to connection
- `Build() *Package` - Build the package

### Template API

- `GetDefaultTemplateRegistry() *TemplateRegistry` - Get default template registry
- `registry.Get(name string) *PackageTemplate` - Get template by name
- `registry.List() []string` - List available templates
- `template.Instantiate(params map[string]interface{}) (*Package, error)` - Instantiate template

### Package Structure

The `Package` type represents a complete DTSX package with the following main components:

- `Property` - Package properties
- `ConnectionManager` - Data source connections
- `Configuration` - Package configurations
- `LogProvider` - Logging configurations
- `Variable` - Package variables
- `Executable` - Tasks and containers
- `PrecedenceConstraint` - Control flow constraints
- `EventHandler` - Event handlers

## Examples

See the [examples](./examples) directory for complete working examples:

```bash
# Analyze package structure
go run examples/analyze_dtsx.go path/to/your/package.dtsx

# Demonstrate query methods with variables and expressions
go run examples/query_dtsx.go path/to/your/package.dtsx

# Comprehensive connection analysis with expressions and variables
go run examples/analyze_connections.go path/to/your/package.dtsx

# Validate package for issues
go run examples/validate_dtsx.go path/to/your/package.dtsx

# Use package templates
go run examples/use_templates.go

# Basic read example
go run examples/read_dtsx.go path/to/your/package.dtsx

# Run package with DTExec
go run examples/run_dtsx.go path/to/your/package.dtsx

# Run package with parameters
go run examples/run_with_params.go path/to/your/package.dtsx

# Debug output as JSON
go run examples/debug_dtsx.go path/to/your/package.dtsx
```

## Schema Generation

This library uses schemas generated from the official SSIS XSD files using [xgen](https://github.com/xuri/xgen):

```bash
# Regenerate schemas (if needed)
xgen -i ./schemas -o ./dtsx -l Go
```

## Testing

```bash
go test ./...
```

## Project Structure

```
.
├── dtsx.go                 # Main package with marshal/unmarshal functions
├── expression.go           # Advanced SSIS expression evaluator
├── templates.go            # Reusable package templates
├── dtsx_test.go            # Comprehensive tests
├── dtsx/
│   └── schemas/            # Generated schema types
├── examples/               # Example code
├── schemas/                # XSD schema files
├── SSIS_EXAMPLES/          # Sample DTSX files (for testing)
├── QUICKSTART.md           # Detailed usage guide
└── README.md               # This file
```

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Related Projects

- [xgen](https://github.com/xuri/xgen) - XSD to Go struct generator

## Acknowledgments

Schema definitions are based on Microsoft SQL Server Integration Services (SSIS) XSD schemas.
