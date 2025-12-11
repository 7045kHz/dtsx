# DTSX - SQL Server Integration Services Package Parser

A comprehensive Go library for reading, writing, and analyzing DTSX (SQL Server Integration Services) XML files with advanced expression evaluation and package management capabilities.

## Important Notice:

- This project is under active development. While many features are implemented and tested, some advanced SSIS functionalities may not be fully supported yet.
- This project is not affiliated with or endorsed by Microsoft. It is an independent implementation for working with SSIS DTSX files in Go.
- DTSX files can be complex and may contain proprietary elements. This library aims to provide robust support for common DTSX features but may not cover all edge cases or custom components.
- Use this library at your own risk, especially in production environments. Always validate and test packages thoroughly.
- Contributions and feedback are welcome to improve compatibility and functionality.

## Features

- **Parse DTSX files**: Read and unmarshal DTSX XML files into Go structs
- **Generate DTSX files**: Marshal Go structs back to DTSX XML format
- **Update DTSX elements**: Modify existing variables, connections, expressions, and any properties safely
- **Advanced Expression Engine**: Full SSIS expression evaluation with variables, functions, conditionals, and type casting
- **Package Builder API**: Fluent API for programmatic package creation
- **Comprehensive Validation**: Multi-level validation with error, warning, and info severities
- **Dependency Analysis**: Graph-based analysis of package relationships and impact assessment
- **Query API**: Convenient methods to analyze connections, variables, executables, and expressions
- **Connection Analysis**: Comprehensive analysis of connection managers with drivers and dynamic properties
- **SQL Extraction**: Extract SQL statements from control flow and dataflow tasks
- **Execution Order Analysis**: Topological sorting and precedence constraint analysis with execution flow descriptions
- **Expression Details**: Detailed expression analysis with evaluation results and dependency tracking
- **Package Parser**: Centralized parsing with caching for performance
- **Utility Functions**: Convenient getters for connection names, variable values, executable names, and more
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

// Write to file using the standard library (the package intentionally
// no longer exposes direct file-write helpers; write manually):
// err := os.WriteFile("output.dtsx", data, 0644)

// Or write to an io.Writer by writing `data` to the writer
// (e.g., writer.Write(data))
```

## Advanced Features

### Expression Evaluation

Evaluate SSIS expressions with full support for variables, arithmetic, functions, conditionals, and type casting:

```go
// Create parser for efficient expression evaluation with caching
parser := dtsx.NewPackageParser(pkg)

// Basic arithmetic and variables
result, err := parser.EvaluateExpression("@[User::MyVar] + 1")

// Built-in functions
result, err = parser.EvaluateExpression("UPPER(@[User::Name])")
result, err = parser.EvaluateExpression("DATEADD(\"DAY\", 7, @[User::StartDate])")

// Conditional expressions
result, err = parser.EvaluateExpression("@[User::Count] > 10 ? \"High\" : \"Low\"")

// Type casting
result, err = parser.EvaluateExpression("(DT_STR) @[User::Number]")
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

Validate packages for common issues using the comprehensive PackageValidator:

```go
validator := dtsx.NewPackageValidator(pkg)
errors := validator.Validate()
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

### Package Parser

Use the centralized PackageParser for efficient analysis with caching:

```go
parser := dtsx.NewPackageParser(pkg)

// Get SQL statements from all tasks
sqlStatements := parser.GetSQLStatements()
for _, stmt := range sqlStatements {
    fmt.Printf("Task: %s (%s) - SQL: %s\n", stmt.TaskName, stmt.TaskType, stmt.SQL)
}

// Evaluate expressions with caching
result, err := parser.EvaluateExpression("@[User::MyVar] + 1")
```

### Execution Order Analysis

Analyze task execution order and precedence constraints:

```go
analyzer := dtsx.NewPrecedenceAnalyzer(pkg)

// Get execution order for a specific task
order, err := analyzer.GetExecutionOrder("Package\\MyTask")
fmt.Printf("Task executes at order: %d\n", order)

// Get all execution orders
orders, err := analyzer.GetAllExecutionOrders()
for refId, order := range orders {
    fmt.Printf("%s: Order %d\n", refId, order)
}

// Get textual execution flow description
flowDesc := analyzer.GetExecutionFlowDescription()
fmt.Print(flowDesc)

// Validate precedence constraints
errors := analyzer.ValidateConstraints()
```

### Expression Details Analysis

Get comprehensive information about expressions including evaluation results and dependencies:

```go
expressions := pkg.GetExpressions()
exprs := expressions.Results.([]*dtsx.ExpressionInfo)

for _, expr := range exprs {
    details := dtsx.GetExpressionDetails(expr, pkg)
    fmt.Printf("Expression: %s\n", details.Expression)
    fmt.Printf("  Location: %s\n", details.Location)
    if details.EvaluatedValue != "" {
        fmt.Printf("  Evaluated: %s\n", details.EvaluatedValue)
    }
    if len(details.Dependencies) > 0 {
        fmt.Printf("  Dependencies: %v\n", details.Dependencies)
    }
}
```

### Utility Functions

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

// Get detailed expression analysis
details := dtsx.GetExpressionDetails(exprInfo, pkg)
fmt.Printf("Evaluated: %s, Dependencies: %v\n", details.EvaluatedValue, details.Dependencies)
```

### Enhanced Package Validation

Use the comprehensive PackageValidator for detailed analysis:

```go
validator := dtsx.NewPackageValidator(pkg)
errors := validator.Validate()

for _, err := range errors {
    fmt.Printf("[%s] %s: %s\n", err.Severity, err.Path, err.Message)
}
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
    return exec.ExecutableTypeAttr == "ExecuteSQLTask"
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

// Note: package-provided file-write helpers have been internalized to
// disable direct package-managed file writes. Use `Marshal` and the
// standard library (e.g., `os.WriteFile`) to persist packages.
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

Note: Mutating helpers were removed from the public API. To programmatically
modify packages, mutate the package structs directly (e.g., `pkg.Variables`,
`pkg.ConnectionManagers`) or use internal functions within the package.

### Advanced Methods

- `EvaluateExpression(expr string, pkg *Package) (interface{}, error)` - Evaluate SSIS expression (use PackageParser for better performance)
- `Validate() []ValidationError` - Validate package for issues (use PackageValidator for comprehensive validation)
- `BuildDependencyGraph() *DependencyGraph` - Build dependency graph
- `GetUnusedVariables() []string` - Find unused variables
- `GetOptimizationSuggestions() []ValidationError` - Get optimization suggestions
- `NewPackageParser(pkg *Package) *PackageParser` - Create centralized parser with caching
- `NewPrecedenceAnalyzer(pkg *Package) *PrecedenceAnalyzer` - Create execution order analyzer
- `NewPackageValidator(pkg *Package) *PackageValidator` - Create comprehensive validator

### PackageParser Methods

- `GetVariableValue(name string) (interface{}, error)` - Get variable value by name
- `GetConnectionManager(id string) (*schema.ConnectionManagerType, error)` - Get connection manager
- `GetExecutable(refId string) (*schema.AnyNonPackageExecutableType, error)` - Get executable
- `EvaluateExpression(expr string) (interface{}, error)` - Evaluate expression with caching
- `GetSQLStatements() []*SQLStatement` - Extract all SQL statements

### PrecedenceAnalyzer Methods

- `GetExecutionOrder(refId string) (int, error)` - Get execution order for task
- `GetAllExecutionOrders() (map[string]int, error)` - Get all execution orders
- `GetExecutableChain(refId string) ([]string, error)` - Get execution chain
- `GetExecutionFlowDescription() string` - Get textual execution flow description
- `ValidateConstraints() []error` - Validate precedence constraints

### PackageValidator Methods

- `Validate() []*ValidationError` - Comprehensive package validation

### Builder API

- `NewPackageBuilder() *PackageBuilder` - Create new package builder
- `AddVariable(namespace, name, value string) *PackageBuilder` - Add string variable
- `AddVariableWithType(namespace, name, value, dataType string) *PackageBuilder` - Add variable with specific data type
- `AddConnection(name, connectionType, connectionString string) *PackageBuilder` - Add connection manager
- `AddConnectionExpression(connectionName, propertyName, expression string) *PackageBuilder` - Add expression to connection
- `Build() *Package` - Build the package

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

# Advanced package analysis with parser, validator, and analyzer
go run examples/package_analysis.go path/to/your/package.dtsx

# Validate package for issues
go run examples/validate_dtsx.go path/to/your/package.dtsx

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
├── dtsx.go                 # Main package with marshal/unmarshal functions, PackageParser, PrecedenceAnalyzer, PackageValidator
├── expression.go           # Advanced SSIS expression evaluator with caching
├── dtsx_test.go            # Comprehensive tests
├── dtsx/
│   └── schemas/            # Generated schema types
├── examples/               # Example code including package_analysis.go
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
