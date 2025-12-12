# API Reference

This document lists the public types, functions, and methods provided by the `github.com/7045kHz/dtsx` package with short descriptions and minimal usage examples. For full examples, see the `examples/` directory and the README.

Note: An exhaustive, generated listing of exported symbols with signatures is available at [docs/API_symbols.md](docs/API_symbols.md). Regenerate it with `go generate ./...` from the repository root when the public API changes.


**Overview**

- Purpose: Read, write, analyze, validate, and programmatically construct SSIS DTSX packages.
- Scope: Focused on the main `dtsx` package APIs (includes expression evaluation, package analysis, builder, and execution helpers). This document intentionally excludes exhaustive `schemas/` package details, which are auto-generated from XSD files and can be referenced directly when needed.

---

## Top-level convenience functions

- `Unmarshal(data []byte) (*Package, error)` — Parse DTSX XML from bytes.

Example:

```go
pkg, err := dtsx.Unmarshal(data)
```

- `UnmarshalFromReader(r io.Reader) (*Package, error)` — Parse DTSX XML from an `io.Reader`.
- `UnmarshalFromFile(filename string) (*Package, error)` — Read and parse a DTSX file from disk.

Example:

```go
pkg, err := dtsx.UnmarshalFromFile("mypackage.dtsx")
if err != nil {
    log.Fatal(err)
}
```

- `Marshal(pkg *Package) ([]byte, error)` — Convert `Package` back to DTSX XML bytes.
- `IsDTSXPackage(filename string) (*Package, bool)` — Validate a file is a DTSX package and return the parsed `Package`.

---

## Core types & builders

- `Package` — Root representation of a DTSX package. (See also `pkg.GetConnections()`, `pkg.GetVariables()`, and other query helpers.)

- `NewPackageBuilder() *PackageBuilder` — Fluent builder for creating packages programmatically.

Example:

```go
pkg := dtsx.NewPackageBuilder().
    AddVariable("User", "InputPath", "C:\\data\\input.csv").
    AddVariable("User", "OutputPath", "C:\\data\\output.csv").
    AddConnection("SourceDB", "OLEDB", "Server=myserver;Database=mydb;").
    Build()
```

---

## Parser & evaluation helpers (PackageParser)

- `NewPackageParser(pkg *Package) *PackageParser` — Create a parser with caching and utility methods.

- `(*PackageParser) EvaluateExpression(expr string) (interface{}, error)` — Evaluate an SSIS expression with caching.

Example: (See [examples/evaluate_expressions.go](examples/evaluate_expressions.go#L120-L128))

```go
parser := dtsx.NewPackageParser(pkg)
val, err := parser.EvaluateExpression("@[User::Count] + 1")
```

- `(*PackageParser) GetVariableValue(name string) (interface{}, error)` — Retrieve a variable value by name.
- `(*PackageParser) GetSQLStatements() []*SQLStatement` — Extract SQL statements from control-flow and dataflow tasks.

---

## Package query helpers

- `(*Package) GetConnections() *QueryResult` — Returns connection managers.
- `(*Package) GetVariables() *QueryResult` — Returns variables.
- `(*Package) GetVariableByName(name string) (*schema.VariableType, error)` — Find variable by name.
- `(*Package) GetExpressions() *QueryResult` — Returns all expressions found in the package (with locations).
- `(*Package) QueryExecutables(filter func(*schema.AnyNonPackageExecutableType) bool) []*schema.AnyNonPackageExecutableType` — Filter executables using a predicate.

Example:

```go
connections := pkg.GetConnections()
fmt.Printf("Connection Managers: %d\n", connections.Count)
vars := pkg.GetVariables()
fmt.Printf("Variables: %d\n", vars.Count)
```

---

## Analysis & Validation

- `NewPrecedenceAnalyzer(pkg *Package) *PrecedenceAnalyzer` — Analyze execution order and precedence constraints.
- `(*PrecedenceAnalyzer) GetExecutionOrder(refId string) (int, error)` — Get execution order number for a task.
- `(*PrecedenceAnalyzer) GetAllExecutionOrders() (map[string]int, error)` — Get execution order for all executables.
- `(*PrecedenceAnalyzer) GetExecutionFlowDescription() string` — Get a textual flow description.
- `(*PrecedenceAnalyzer) ValidateConstraints() []error` — Check for constraint problems.

- `NewPackageValidator(pkg *Package) *PackageValidator` — Create a validator.
- `(*PackageValidator) Validate() []*ValidationError` — Run validation.

Example: (See [examples/validate_dtsx.go](examples/validate_dtsx.go#L18-L38))

```go
validator := dtsx.NewPackageValidator(pkg)
errList := validator.Validate()
for _, e := range errList {
    fmt.Printf("[%s] %s: %s\n", e.Severity, e.Path, e.Message)
}
```

---

## Utility helpers

- `GetConnectionName(cm *schema.ConnectionManagerType) string`
- `GetConnectionString(cm *schema.ConnectionManagerType) string`
- `GetVariableName(v *schema.VariableType) string`
- `GetVariableValue(v *schema.VariableType) string` (package-level helper)
- `GetExecutableName(exec *schema.AnyNonPackageExecutableType) string`
- `GetExpressionDetails(exprInfo *ExpressionInfo, pkg *Package) *ExpressionDetails`
  - Retrieve detailed expression info including evaluated value and dependencies (see [examples/query_dtsx.go](examples/query_dtsx.go#L136-L148)).

Example:

```go
connName := dtsx.GetConnectionName(cm)
varName := dtsx.GetVariableName(variable)
details := dtsx.GetExpressionDetails(exprInfo, pkg)
fmt.Printf("Evaluated: %s, Dependencies: %v\n", details.EvaluatedValue, details.Dependencies)
```

---

## Execution helper

- `RunPackage(dtexecPath, dtsxPath string, opts *RunOptions) (string, error)` — Execute DTSX package using `dtexec.exe`.

Example: (See [examples/run_with_params.go](examples/run_with_params.go#L12-L82))

```go
out, err := dtsx.RunPackage("C:\\Program Files\\Microsoft SQL Server\\130\\DTS\\Binn\\DTExec.exe", "myPackage.dtsx", &dtsx.RunOptions{Parameters: opts})
```

---

## Dependency & Optimization analysis

- `(*Package) BuildDependencyGraph() *DependencyGraph`
- `(*DependencyGraph) GetVariableImpact(varName string) []string`
- `(*DependencyGraph) GetConnectionImpact(connName string) []string`
- `(*Package) GetUnusedVariables() []string`
- `(*Package) GetOptimizationSuggestions() []ValidationError`

Example: (see [examples/package_analysis.go](examples/package_analysis.go#L28-L44))

```go
graph := pkg.BuildDependencyGraph()
impact := graph.GetVariableImpact("User::MyVariable")
```

---

## Examples & Common Workflows

- Parse & analyze package: [examples/analyze_dtsx.go](examples/analyze_dtsx.go#L1-L140)
- Inspect connections: [examples/analyze_connections.go](examples/analyze_connections.go#L96-L140)
- Evaluate expressions: [examples/evaluate_expressions.go](examples/evaluate_expressions.go#L120-L128)
- Build packages programmatically: [examples/build_package.go](examples/build_package.go#L24-L40)
- Validate packages: [examples/validate_dtsx.go](examples/validate_dtsx.go#L18-L38)
- Run packages via `dtexec.exe`: [examples/run_dtsx.go](examples/run_dtsx.go#L48-L64)
- Per-symbol, consolidated examples: [examples/api_examples.go](examples/api_examples.go#L1-L220)

How to run the per-symbol example:

```bash
# Run from the repository root
go run examples/api_examples.go
```

Notes:

- The per-symbol example demonstrates the most commonly used public `dtsx` package APIs and is intended for learning and exploration.

- It writes a sample DTSX file named `output_sample.dtsx` to the current working directory; remove it after running if not needed.

- The example is marked with a `//go:build ignore` tag to prevent it being included in normal library builds. Use `go run` to execute it explicitly.

### Example symbol references (api_examples.go)

- `NewPackageBuilder()` — [examples/api_examples.go](examples/api_examples.go#L16-L19)
- `Marshal()` — [examples/api_examples.go](examples/api_examples.go#L25)
- `Unmarshal()` — [examples/api_examples.go](examples/api_examples.go#L31)
- `NewPackageParser()` — [examples/api_examples.go](examples/api_examples.go#L38)
- `(p *PackageParser) GetVariableValue` — [examples/api_examples.go](examples/api_examples.go#L41)
- `(p *PackageParser) EvaluateExpression` — [examples/api_examples.go](examples/api_examples.go#L49)
- `EvaluateExpression()` (package-level) — [examples/api_examples.go](examples/api_examples.go#L57)
- `(p *Package) GetConnections()` — [examples/api_examples.go](examples/api_examples.go#L65)
- `(p *Package) GetVariables()` — [examples/api_examples.go](examples/api_examples.go#L73)
- `GetVariableName/GetVariableValue` helpers — [examples/api_examples.go](examples/api_examples.go#L77)
- `(p *Package) GetExpressions()` — [examples/api_examples.go](examples/api_examples.go#L81)
- `(p *PackageParser) GetSQLStatements()` — [examples/api_examples.go](examples/api_examples.go#L91)
- `Expr` AST usage (`Literal`, `Variable`) — [examples/api_examples.go](examples/api_examples.go#L95-L102)
- `BuildDependencyGraph()` & `GetVariableImpact()` — [examples/api_examples.go](examples/api_examples.go#L106-L107)
- `GetUnusedVariables()` — [examples/api_examples.go](examples/api_examples.go#L111)
- `GetOptimizationSuggestions()` — [examples/api_examples.go](examples/api_examples.go#L115)
- `NewPrecedenceAnalyzer()` — [examples/api_examples.go](examples/api_examples.go#L119)
- `NewPackageValidator()` & `Validate()` — [examples/api_examples.go](examples/api_examples.go#L125-L128)
- `GetProperty()` — [examples/api_examples.go](examples/api_examples.go#L131)

---

## Appendix & Notes

- The `schemas/` package contains many exported types generated from XSD schemas. If you want exhaustive schema docs, we can add a generated section for them.
- Optionally, I can produce short, self-contained code snippets for every exported function/type for offline reading.

---

## Maintaining this API doc

When adding or changing exported symbols in the `dtsx` package:

- Update `docs/API.md` with the new symbol or change.
- Add a small runnable snippet to `examples/api_examples.go` demonstrating the new symbol where appropriate.
- Update the 'Example symbol references' mapping in `docs/API.md` if you added new snippets.

## Exhaustive Public API (dtsx package)

This section lists all exported types, functions, and methods from the `dtsx` package (excluding the generated `schemas/` package). Each entry includes a short signature, one-line description, and a minimal runnable snippet showing basic usage.

### Parsing & Marshalling

- `Unmarshal(data []byte) (*Package, error)`
  - Parse DTSX XML from bytes.

```go
pkg, err := dtsx.Unmarshal(data)
```

- `UnmarshalFromReader(r io.Reader) (*Package, error)`
  - Parse DTSX XML from an `io.Reader`.

```go
f, _ := os.Open("package.dtsx")
pkg, _ := dtsx.UnmarshalFromReader(f)
```

- `UnmarshalFromFile(filename string) (*Package, error)`
  - Read and parse a DTSX file from disk.

```go
pkg, _ := dtsx.UnmarshalFromFile("mypackage.dtsx")
```

- `Marshal(pkg *Package) ([]byte, error)`
  - Convert `Package` back to DTSX XML bytes.

```go
b, _ := dtsx.Marshal(pkg)
_ = os.WriteFile("out.dtsx", b, 0644)
```

- `IsDTSXPackage(filename string) (*Package, bool)`
  - Quickly validate and parse a DTSX file.

```go
if pkg, ok := dtsx.IsDTSXPackage("mypackage.dtsx"); ok {
    fmt.Println("Valid DTSX package")
    _ = pkg
}
```

### PackageBuilder (construct packages programmatically)

- `NewPackageBuilder() *PackageBuilder` — Create a new builder.

```go
pb := dtsx.NewPackageBuilder()
```

- `(pb *PackageBuilder) AddVariable(namespace, name, value string) *PackageBuilder` — Add a string variable.

```go
pkg := dtsx.NewPackageBuilder().AddVariable("User","InputPath","C:\\data\\in.csv").Build()
```

- `(pb *PackageBuilder) AddVariableWithType(namespace, name, value string, dataType string) *PackageBuilder` — Add variable with explicit data type.

- `(pb *PackageBuilder) AddConnection(name, connectionType, connectionString string) *PackageBuilder` — Add a connection manager.

```go
pkg := dtsx.NewPackageBuilder().AddConnection("SourceDB","OLEDB","Server=.;Database=db;").Build()
```

- `(pb *PackageBuilder) AddConnectionExpression(connectionName, propertyName, expression string) *PackageBuilder` — Add a property expression to a connection.

- `(pb *PackageBuilder) Build() *Package` — Finalize builder and return `*Package`.

### PackageParser (parsing, caching, evaluation)

- `NewPackageParser(pkg *Package) *PackageParser` — Create a parser with caching.

```go
parser := dtsx.NewPackageParser(pkg)
```

- `(p *PackageParser) GetVariableValue(name string) (interface{}, error)`
 — Get variable value by full name (e.g., `User::Var`).

```go
v, _ := parser.GetVariableValue("User::Count")
```

- `(p *PackageParser) GetExecutable(refId string) (*schema.AnyNonPackageExecutableType, error)` — Get executable by reference id.

```go
ex, _ := parser.GetExecutable("Package\\MyTask")
fmt.Println(dtsx.GetExecutableName(ex))
```

- `(p *PackageParser) EvaluateExpression(expr string) (interface{}, error)` — Evaluate expression with caching.

```go
val, _ := parser.EvaluateExpression("@[User::Count] + 1")
```

- `(p *PackageParser) GetSQLStatements() []*SQLStatement` — Extract SQL statements from control flow and dataflow tasks.

```go
stmts := parser.GetSQLStatements()
for _, s := range stmts { fmt.Println(s.TaskName, s.SQL) }
```

- `(p *PackageParser) GetConnectionManager(id string) (*schema.ConnectionManagerType, error)`


```go
cm, _ := parser.GetConnectionManager("SourceDB")
```

- `(p *PackageParser) GetExecutable(refId string) (*schema.AnyNonPackageExecutableType, error)`

```go
ex, _ := parser.GetExecutable("Package\\MyTask")
fmt.Println(dtsx.GetExecutableName(ex))
```

- `(p *PackageParser) EvaluateExpression(expr string) (interface{}, error)` — Evaluate expression with caching.

```go
val, _ := parser.EvaluateExpression("@[User::Count] + 1")
```

- `(p *PackageParser) GetSQLStatements() []*SQLStatement` — Extract SQL statements from control flow and dataflow tasks.

```go
stmts := parser.GetSQLStatements()
for _, s := range stmts { fmt.Println(s.TaskName, s.SQL) }
```

### Execution analysis (PrecedenceAnalyzer)

- `NewPrecedenceAnalyzer(pkg *Package) *PrecedenceAnalyzer`

```go
pa := dtsx.NewPrecedenceAnalyzer(pkg)
```

- `(p *PrecedenceAnalyzer) GetExecutionOrder(refId string) (int, error)`

```go
order, _ := pa.GetExecutionOrder("\"Package\\MyTask\"")
```

- `(p *PrecedenceAnalyzer) GetAllExecutionOrders() (map[string]int, error)`

- `(p *PrecedenceAnalyzer) GetExecutableChain(refId string) ([]string, error)`

- `(p *PrecedenceAnalyzer) ValidateConstraints() []error`

```go
errs := pa.ValidateConstraints()
```

- `(p *PrecedenceAnalyzer) GetExecutionFlowDescription() string`

```go
fmt.Println(pa.GetExecutionFlowDescription())
```

### Validation (PackageValidator)

- `NewPackageValidator(pkg *Package) *PackageValidator`

```go
v := dtsx.NewPackageValidator(pkg)
```

- `(v *PackageValidator) Validate() []*ValidationError`

```go
issues := v.Validate()
```

### Package helpers & queries

- `(p *Package) GetConnections() *QueryResult` — Returns connections as `QueryResult`.

```go
conns := pkg.GetConnections()
cm := conns.Results.([]*schema.ConnectionManagerType)[0]
```

- `(p *Package) GetVariables() *QueryResult`

- `(p *Package) GetVariableByName(name string) (*schema.VariableType, error)`

```go
v, _ := pkg.GetVariableByName("User::MyVar")
```

- `(p *Package) QueryExecutables(filter func(*schema.AnyNonPackageExecutableType) bool) []*schema.AnyNonPackageExecutableType`

```go
execs := pkg.QueryExecutables(func(e *schema.AnyNonPackageExecutableType) bool { return e.ExecutableTypeAttr == "STOCK:SQLTask" })
```

- `(p *Package) GetExpressions() *QueryResult` — Returns `ExpressionInfo` entries.

```go
exprs := pkg.GetExpressions()
for _, e := range exprs.Results.([]*dtsx.ExpressionInfo) { fmt.Println(e.Expression) }
```

- `(p *Package) Validate() []ValidationError` — Package-level convenience validation.

```go
v := pkg.Validate()
```

### Expression engine (package-level and AST)

- `EvaluateExpression(expr string, pkg *Package) (interface{}, error)` — Evaluate SSIS expression.

```go
val, _ := dtsx.EvaluateExpression("SUBSTRING('hello',1,2)", pkg)
```

- AST types: `Expr`, `Literal`, `Variable`, `BinaryOp`, `FunctionCall`, `Conditional`, `Cast`, `UnaryOp`, `Token`.

```go
lit := &dtsx.Literal{Value: 42.0}
res, _ := lit.Eval(nil) // 42
vars := map[string]interface{}{"User::X": 10.0}
varNode := &dtsx.Variable{Name: "User::X"}
res2, _ := varNode.Eval(vars) // 10
```

### Utilities & helpers

- `GetConnectionString(cm *schema.ConnectionManagerType) string`

```go
connStr := dtsx.GetConnectionString(cm)
```

- `GetConnectionName(cm *schema.ConnectionManagerType) string`
- `GetVariableName(v *schema.VariableType) string`
- `GetVariableValue(v *schema.VariableType) string` (package-level helper)

```go
name := dtsx.GetVariableName(v)
val := dtsx.GetVariableValue(v)
```

- `GetExecutableName(exec *schema.AnyNonPackageExecutableType) string`
- `GetExpressionDetails(exprInfo *ExpressionInfo, pkg *Package) *ExpressionDetails`

```go
details := dtsx.GetExpressionDetails(exprInfo, pkg)
```

- `GetProperty(s interface{}, name string) interface{}` — Reflection helper to get a field value by name.

```go
obj := dtsx.GetProperty(cm, "ObjectNameAttr")
```

- `GetSqlStatementSource(s *schema.SqlTaskDataType) string`
- `GetSqlStatementSourceFromBase(s *schema.SqlTaskBaseAttributeGroup) string`

### Dependency & optimization

- `(p *Package) BuildDependencyGraph() *DependencyGraph`

```go
g := pkg.BuildDependencyGraph()
```

- `(dg *DependencyGraph) GetVariableImpact(varName string) []string`
- `(dg *DependencyGraph) GetConnectionImpact(connName string) []string`
- `(p *Package) GetUnusedVariables() []string`
- `(p *Package) GetOptimizationSuggestions() []ValidationError`

### Execution (RunPackage)

- `RunOptions` — Options struct for `RunPackage`.

- `RunPackage(dtexecPath, dtsxPath string, opts *RunOptions) (string, error)`

```go
out, err := dtsx.RunPackage("C:\\Program Files\\Microsoft SQL Server\\130\\DTS\\Binn\\DTExec.exe", "pkg.dtsx", &dtsx.RunOptions{Validate:true})
```

---

*Generated by the repo documentation task.*

