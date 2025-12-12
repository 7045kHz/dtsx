# Generated API symbols (dtsx package)

This file is auto-generated. To regenerate, run: `go generate` in the repo root.

## Exported types

### BinaryOp

BinaryOp represents a binary operation

```go
type BinaryOp struct {
	Left	Expr
	Op	string
	Right	Expr
}
```

### Cast

Cast represents a type cast

```go
type Cast struct {
	Type	string
	Expr	Expr
}
```

### Conditional

Conditional represents a ternary conditional expression

```go
type Conditional struct {
	Condition	Expr
	TrueExpr	Expr
	FalseExpr	Expr
}
```

### DependencyGraph

DependencyGraph represents relationships between package elements

```go
type DependencyGraph struct {
	// VariableDependencies: variable name -> list of expressions/locations that use it
	VariableDependencies	map[string][]string
	// ConnectionDependencies: connection name -> list of tasks/locations that use it
	ConnectionDependencies	map[string][]string
	// TaskDependencies: task ID -> list of variables/connections it depends on
	TaskDependencies	map[string][]string
	// ExpressionDependencies: expression -> list of variables it references
	ExpressionDependencies	map[string][]string
}
```

### Expr

Expr represents an expression AST node

```go
type Expr interface {
	Eval(vars map[string]interface{}) (interface{}, error)
}
```

### ExpressionDetails

ExpressionDetails provides comprehensive information about an expression

```go
type ExpressionDetails struct {
	Expression	string
	Location	string
	Name		string
	Context		string
	EvaluatedValue	string
	EvaluationError	string
	Dependencies	[]string
}
```

### ExpressionInfo

ExpressionInfo contains information about an expression found in the package

```go
type ExpressionInfo struct {
	Expression	string
	Location	string	// e.g., "Package", "Executable", "PrecedenceConstraint", etc.
	Name		string	// property name if applicable
	Context		string	// additional context like executable type, variable name, etc.
}
```

### FunctionCall

FunctionCall represents a function call

```go
type FunctionCall struct {
	Name	string
	Args	[]Expr
}
```

### Literal

Literal represents a literal value

```go
type Literal struct {
	Value interface{}
}
```

### Package

Package represents a DTSX package structure

```go
type Package struct {
	XMLName				xml.Name	`xml:"Executable"`
	RefIdAttr			*string		`xml:"refId,attr"`
	CreationDateAttr		*string		`xml:"CreationDate,attr"`
	CreationNameAttr		*string		`xml:"CreationName,attr"`
	CreatorComputerNameAttr		*string		`xml:"CreatorComputerName,attr"`
	CreatorNameAttr			*string		`xml:"CreatorName,attr"`
	DescriptionAttr			*string		`xml:"Description,attr"`
	DTSIDAttr			*string		`xml:"DTSID,attr"`
	EnableConfigAttr		*string		`xml:"EnableConfig,attr"`
	ExecutableTypeAttr		*string		`xml:"ExecutableType,attr"`
	LastModifiedProductVersionAttr	*string		`xml:"LastModifiedProductVersion,attr"`
	LocaleIDAttr			*string		`xml:"LocaleID,attr"`
	ObjectNameAttr			*string		`xml:"ObjectName,attr"`
	PackageTypeAttr			*string		`xml:"PackageType,attr"`
	VersionBuildAttr		*string		`xml:"VersionBuild,attr"`
	VersionGUIDAttr			*string		`xml:"VersionGUID,attr"`
	*schema.ExecutableTypePackage
}
```

### PackageBuilder

PackageBuilder provides a fluent API for constructing DTSX packages

```go
type PackageBuilder struct {
	pkg *Package
}
```

### PackageParser

PackageParser provides centralized parsing and analysis functionality for DTSX packages

```go
type PackageParser struct {
	pkg		*Package
	vars		map[string]interface{}
	connMap		map[string]*schema.ConnectionManagerType
	execMap		map[string]*schema.AnyNonPackageExecutableType
	varCache	map[string]interface{}	// Cache for expensive operations
}
```

### PackageValidator

PackageValidator provides validation functions for DTSX packages

```go
type PackageValidator struct {
	pkg		*Package
	parser		*PackageParser
	analyzer	*PrecedenceAnalyzer
}
```

### PrecedenceAnalyzer

PrecedenceAnalyzer handles execution order calculation with support for complex precedence constraints

```go
type PrecedenceAnalyzer struct {
	pkg		*Package
	execMap		map[string]*schema.AnyNonPackageExecutableType
	orderCache	map[string]int
	dependencies	map[string][]string
}
```

### QueryResult

QueryResult wraps query results with metadata

```go
type QueryResult struct {
	Count	int
	Results	interface{}
}
```

### RunOptions

RunOptions contains options for executing a DTSX package with dtexec.exe

```go
type RunOptions struct {
	// Package parameters (format: "[$Package::|$Project::|$ServerOption::]ParamName[(DataType)];Value")
	Parameters	[]string

	// Environment variables (format: "Name=Value")
	EnvironmentVars	[]string

	// Connection manager overrides (format: "id_or_name;connection_string")
	Connections	[]string

	// Configuration file path
	ConfigFile	string

	// Property overrides using /Set (format: "propertyPath;value")
	PropertySets	[]string

	// Decryption password for encrypted packages
	DecryptPassword	string

	// SQL Server name (for /SQL or /DTS packages)
	Server	string

	// SQL Server username (for SQL authentication)
	User	string

	// SQL Server password (for SQL authentication)
	Password	string

	// Enable checkpointing (on/off)
	Checkpointing	string

	// Checkpoint file path
	CheckpointFile	string

	// Restart mode (deny/force/ifPossible)
	Restart	string

	// Maximum concurrent executables (-1 for auto)
	MaxConcurrent	int

	// Validate only without executing
	Validate	bool

	// Treat warnings as errors
	WarnAsError	bool

	// Verify build number (format: "major;minor;build")
	VerifyBuild	string

	// Verify package ID (GUID)
	VerifyPackageID	string

	// Verify version ID (GUID)
	VerifyVersionID	string

	// Verify digital signature
	VerifySigned	bool

	// Reporting level (N=none, E=errors, W=warnings, I=info, C=custom, P=progress, V=verbose)
	ReportingLevel	string

	// Console log options (format: "displayoptions;list_options;src_name_or_guid")
	ConsoleLog	[]string

	// Log provider configuration (format: "classid_or_progid;configstring")
	Loggers	[]string

	// Enable verbose logging to file
	VerboseLog	string

	// Dump on error codes (semicolon-separated error codes)
	DumpOnCodes	string

	// Dump on any error
	DumpOnError	bool

	// Run in 32-bit mode (x86)
	X86	bool
}
```

### SQLStatement

SQLStatement represents a SQL statement found in the package

```go
type SQLStatement struct {
	TaskName	string
	TaskType	string
	SQL		string
	RefId		string
	Connections	[]string
}
```

### Token

Token represents a lexical token

```go
type Token struct {
	Type	string
	Value	string
}
```

### UnaryOp

UnaryOp represents a unary operator

```go
type UnaryOp struct {
	Op	string
	Expr	Expr
}
```

### ValidationError

ValidationError represents a validation issue in a DTSX package

```go
type ValidationError struct {
	Severity	string	// "error", "warning", "info"
	Message		string
	Path		string	// Location in the package, e.g., "Variables.User::MyVar"
}
```

### Variable

Variable represents a variable reference

```go
type Variable struct {
	Name string
}
```

## Exported functions

### EvaluateExpression

EvaluateExpression evaluates an SSIS expression in the context of a package

```go
func EvaluateExpression(expr string, pkg *Package) (interface{}, error)
```

### GetConnectionName

GetConnectionName returns the name of a connection manager

```go
func GetConnectionName(cm *schema.ConnectionManagerType) string
```

### GetConnectionString

GetConnectionString returns the connection string of a connection manager

```go
func GetConnectionString(cm *schema.ConnectionManagerType) string
```

### GetExecutableName

GetExecutableName returns the name of an executable

```go
func GetExecutableName(exec *schema.AnyNonPackageExecutableType) string
```

### GetExpressionDetails

GetExpressionDetails returns detailed information about an expression including evaluation result and dependencies

```go
func GetExpressionDetails(exprInfo *ExpressionInfo, pkg *Package) *ExpressionDetails
```

### GetProperty

GetProperty returns the value of the specified property by name for any struct

```go
func GetProperty(s interface{}, name string) interface{}
```

### GetSqlStatementSource

GetSqlStatementSource returns the SQL statement source from SqlTaskDataType

```go
func GetSqlStatementSource(s *schema.SqlTaskDataType) string
```

### GetSqlStatementSourceFromBase

GetSqlStatementSourceFromBase returns the SQL statement source from SqlTaskBaseAttributeGroup

```go
func GetSqlStatementSourceFromBase(s *schema.SqlTaskBaseAttributeGroup) string
```

### GetVariableName

GetVariableName returns the full name (namespace::name) of a variable

```go
func GetVariableName(v *schema.VariableType) string
```

### GetVariableValue

GetVariableValue returns the value of a variable

```go
func GetVariableValue(v *schema.VariableType) string
```

### IsDTSXPackage

IsDTSXPackage validates if the given filename is a valid DTSX package.
It checks if the file exists, is readable, and contains valid DTSX XML structure.
Returns the unmarshaled Package and true if the file is a valid DTSX package,
nil and false otherwise.

```go
func IsDTSXPackage(filename string) (*Package, bool)
```

### Marshal

Marshal converts a Package to DTSX XML format

```go
func Marshal(pkg *Package) ([]byte, error)
```

### NewPackageBuilder

NewPackageBuilder creates a new package builder

```go
func NewPackageBuilder() *PackageBuilder
```

### NewPackageParser

NewPackageParser creates a new PackageParser for the given package

```go
func NewPackageParser(pkg *Package) *PackageParser
```

### NewPackageValidator

NewPackageValidator creates a new validator for the package

```go
func NewPackageValidator(pkg *Package) *PackageValidator
```

### NewPrecedenceAnalyzer

NewPrecedenceAnalyzer creates a new analyzer for the given package

```go
func NewPrecedenceAnalyzer(pkg *Package) *PrecedenceAnalyzer
```

### RunPackage

RunPackage executes a DTSX package using dtexec.exe.
It takes the path to dtexec.exe, the path to the DTSX file, and optional RunOptions.
Returns the combined stdout/stderr output and any error that occurred.

```go
func RunPackage(dtexecPath, dtsxPath string, opts *RunOptions) (string, error)
```

### Unmarshal

Unmarshal parses DTSX XML data and returns a Package

```go
func Unmarshal(data []byte) (*Package, error)
```

### UnmarshalFromFile

UnmarshalFromFile reads a DTSX file and returns a Package

```go
func UnmarshalFromFile(filename string) (*Package, error)
```

### UnmarshalFromReader

UnmarshalFromReader parses DTSX XML from an io.Reader and returns a Package

```go
func UnmarshalFromReader(r io.Reader) (*Package, error)
```

## Methods on exported types

### BinaryOp

#### Eval

```go
func (b *BinaryOp) Eval(vars map[string]interface{}) (interface{}, error) {
	left, err := b.Left.Eval(vars)
	if err != nil {
		return nil, err
	}
	right, err := b.Right.Eval(vars)
	if err != nil {
		return nil, err
	}
	switch b.Op {
	case "+":
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				return l + r, nil
			}
		}
		if l, ok := left.(string); ok {
			if r, ok := right.(string); ok {
				return l + r, nil
			}
		}
		return nil, fmt.Errorf("cannot add %T and %T", left, right)
	case "-":
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				return l - r, nil
			}
		}
		return nil, fmt.Errorf("cannot subtract %T and %T", left, right)
	case "*":
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				return l * r, nil
			}
		}
		return nil, fmt.Errorf("cannot multiply %T and %T", left, right)
	case "/":
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				if r == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				return l / r, nil
			}
		}
		return nil, fmt.Errorf("cannot divide %T and %T", left, right)
	case "==":
		return left == right, nil
	case "!=":
		return left != right, nil
	case "<":
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				return l < r, nil
			}
		}
		return nil, fmt.Errorf("cannot compare %T and %T", left, right)
	case ">":
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				return l > r, nil
			}
		}
		return nil, fmt.Errorf("cannot compare %T and %T", left, right)
	case "<=":
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				return l <= r, nil
			}
		}
		return nil, fmt.Errorf("cannot compare %T and %T", left, right)
	case ">=":
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				return l >= r, nil
			}
		}
		return nil, fmt.Errorf("cannot compare %T and %T", left, right)
	case "&&":
		lb := toBool(left)
		rb := toBool(right)
		return lb && rb, nil
	case "||":
		lb := toBool(left)
		rb := toBool(right)
		return lb || rb, nil
	}
	return nil, fmt.Errorf("unknown operator: %s", b.Op)
}
```

### Cast

#### Eval

```go
func (c *Cast) Eval(vars map[string]interface{}) (interface{}, error) {
	val, err := c.Expr.Eval(vars)
	if err != nil {
		return nil, err
	}

	return castValue(val, c.Type)
}
```

### Conditional

#### Eval

```go
func (c *Conditional) Eval(vars map[string]interface{}) (interface{}, error) {
	cond, err := c.Condition.Eval(vars)
	if err != nil {
		return nil, err
	}

	// Convert to boolean
	var condition bool
	switch v := cond.(type) {
	case bool:
		condition = v
	case float64:
		condition = v != 0
	case string:
		condition = v != ""
	default:
		return nil, fmt.Errorf("cannot convert %T to boolean", cond)
	}

	if condition {
		return c.TrueExpr.Eval(vars)
	}
	return c.FalseExpr.Eval(vars)
}
```

### DependencyGraph

#### GetConnectionImpact

GetConnectionImpact returns all tasks affected by a connection change

```go
// GetConnectionImpact returns all tasks affected by a connection change
func (dg *DependencyGraph) GetConnectionImpact(connName string) []string {
	if dg == nil {
		return nil
	}
	return dg.ConnectionDependencies[connName]
}
```

#### GetVariableImpact

GetVariableImpact returns all locations affected by a variable change

```go
// GetVariableImpact returns all locations affected by a variable change
func (dg *DependencyGraph) GetVariableImpact(varName string) []string {
	if dg == nil {
		return nil
	}
	return dg.VariableDependencies[varName]
}
```

### FunctionCall

#### Eval

```go
func (f *FunctionCall) Eval(vars map[string]interface{}) (interface{}, error) {

	args := make([]interface{}, len(f.Args))
	for i, arg := range f.Args {
		val, err := arg.Eval(vars)
		if err != nil {
			return nil, err
		}
		args[i] = val
	}

	if fn, ok := functions[f.Name]; ok {
		return fn(args)
	}
	return nil, fmt.Errorf("unknown function: %s", f.Name)
}
```

### Literal

#### Eval

```go
func (l *Literal) Eval(vars map[string]interface{}) (interface{}, error) {
	return l.Value, nil
}
```

### Package

#### BuildDependencyGraph

BuildDependencyGraph analyzes the package and builds a dependency graph

```go
// BuildDependencyGraph analyzes the package and builds a dependency graph
func (p *Package) BuildDependencyGraph() *DependencyGraph {
	graph := &DependencyGraph{
		VariableDependencies:	make(map[string][]string),
		ConnectionDependencies:	make(map[string][]string),
		TaskDependencies:	make(map[string][]string),
		ExpressionDependencies:	make(map[string][]string),
	}

	if p == nil {
		return graph
	}

	exprResult := p.GetExpressions()
	if exprResult.Count > 0 {
		expressions := exprResult.Results.([]*ExpressionInfo)
		for _, expr := range expressions {
			vars := extractVariableReferences(expr.Expression)
			graph.ExpressionDependencies[expr.Expression] = vars

			location := fmt.Sprintf("%s:%s", expr.Location, expr.Name)
			for _, v := range vars {
				graph.VariableDependencies[v] = append(graph.VariableDependencies[v], location)
			}
		}
	}

	if p.Executable != nil {
		for i, exec := range p.Executable {
			taskID := fmt.Sprintf("Executable[%d]", i)
			if exec.ExecutableTypeAttr != "" {
				taskID = fmt.Sprintf("%s (%s)", taskID, exec.ExecutableTypeAttr)
			}

			if exec.Property != nil {
				for _, prop := range exec.Property {
					if prop.NameAttr != nil && *prop.NameAttr == "Connection" && prop.Value != "" {
						connName := prop.Value
						graph.ConnectionDependencies[connName] = append(graph.ConnectionDependencies[connName], taskID)
						graph.TaskDependencies[taskID] = append(graph.TaskDependencies[taskID], "Connection:"+connName)
					}
				}
			}

			if exec.Property != nil {
				for _, prop := range exec.Property {
					if prop.Value != "" {
						vars := extractVariableReferences(prop.Value)
						for _, v := range vars {
							graph.VariableDependencies[v] = append(graph.VariableDependencies[v], taskID)
							graph.TaskDependencies[taskID] = append(graph.TaskDependencies[taskID], "Variable:"+v)
						}
					}
				}
			}
		}
	}

	return graph
}
```

#### GetConnections

GetConnections returns all connection managers in the package

```go
// GetConnections returns all connection managers in the package
func (p *Package) GetConnections() *QueryResult {
	if p == nil || p.ConnectionManagers == nil || p.ConnectionManagers.ConnectionManager == nil {
		return &QueryResult{Count: 0, Results: []*schema.ConnectionManagerType{}}
	}
	return &QueryResult{
		Count:		len(p.ConnectionManagers.ConnectionManager),
		Results:	p.ConnectionManagers.ConnectionManager,
	}
}
```

#### GetExpressions

GetExpressions returns all expressions found in the package

```go
// GetExpressions returns all expressions found in the package
func (p *Package) GetExpressions() *QueryResult {
	var expressions []*ExpressionInfo
	if p == nil {
		return &QueryResult{Count: 0, Results: expressions}
	}

	if p.PropertyExpression != nil {
		for _, expr := range p.PropertyExpression {
			if expr.AnySimpleType != nil && expr.AnySimpleType.Value != "" {
				expressions = append(expressions, &ExpressionInfo{
					Expression:	expr.AnySimpleType.Value,
					Location:	"Package",
					Name:		expr.NameAttr,
					Context:	"Package Property",
				})
			}
		}
	}

	if p.Executable != nil {
		for i, exec := range p.Executable {
			if exec.PropertyExpression != nil {
				for _, expr := range exec.PropertyExpression {
					if expr.AnySimpleType != nil && expr.AnySimpleType.Value != "" {
						context := fmt.Sprintf("Executable[%d]", i)
						if exec.ExecutableTypeAttr != "" {
							context = fmt.Sprintf("%s (%s)", context, exec.ExecutableTypeAttr)
						}
						expressions = append(expressions, &ExpressionInfo{
							Expression:	expr.AnySimpleType.Value,
							Location:	"Executable",
							Name:		expr.NameAttr,
							Context:	context,
						})
					}
				}
			}

			if exec.PrecedenceConstraint != nil {
				for j, pc := range exec.PrecedenceConstraint {
					if pc.PropertyExpression != nil {
						for _, expr := range pc.PropertyExpression {
							if expr.AnySimpleType != nil && expr.AnySimpleType.Value != "" {
								context := fmt.Sprintf("Executable[%d] PrecedenceConstraint[%d]", i, j)
								expressions = append(expressions, &ExpressionInfo{
									Expression:	expr.AnySimpleType.Value,
									Location:	"PrecedenceConstraint",
									Name:		expr.NameAttr,
									Context:	context,
								})
							}
						}
					}
				}
			}
		}
	}

	if p.PrecedenceConstraint != nil {
		for i, pc := range p.PrecedenceConstraint {
			if pc.PropertyExpression != nil {
				for _, expr := range pc.PropertyExpression {
					if expr.AnySimpleType != nil && expr.AnySimpleType.Value != "" {
						context := fmt.Sprintf("Package PrecedenceConstraint[%d]", i)
						expressions = append(expressions, &ExpressionInfo{
							Expression:	expr.AnySimpleType.Value,
							Location:	"PrecedenceConstraint",
							Name:		expr.NameAttr,
							Context:	context,
						})
					}
				}
			}
		}
	}

	if p.Variables != nil && p.Variables.Variable != nil {
		for i, v := range p.Variables.Variable {
			if v.PropertyExpression != nil {
				for _, expr := range v.PropertyExpression {
					if expr.AnySimpleType != nil && expr.AnySimpleType.Value != "" {
						context := fmt.Sprintf("Variable[%d]", i)
						// Try to get variable name from attributes
						var varName string
						if v.ObjectNameAttr != nil {
							varName = *v.ObjectNameAttr
						}
						if varName != "" {
							context = fmt.Sprintf("Variable[%d] (%s)", i, varName)
						}
						expressions = append(expressions, &ExpressionInfo{
							Expression:	expr.AnySimpleType.Value,
							Location:	"Variable",
							Name:		expr.NameAttr,
							Context:	context,
						})
					}
				}
			}
		}
	}

	if p.ConnectionManagers != nil && p.ConnectionManagers.ConnectionManager != nil {
		for i, cm := range p.ConnectionManagers.ConnectionManager {
			if cm.PropertyExpression != nil {
				for _, expr := range cm.PropertyExpression {
					if expr.AnySimpleType != nil && expr.AnySimpleType.Value != "" {
						context := fmt.Sprintf("ConnectionManager[%d]", i)
						expressions = append(expressions, &ExpressionInfo{
							Expression:	expr.AnySimpleType.Value,
							Location:	"ConnectionManager",
							Name:		expr.NameAttr,
							Context:	context,
						})
					}
				}
			}
		}
	}

	return &QueryResult{
		Count:		len(expressions),
		Results:	expressions,
	}
}
```

#### GetOptimizationSuggestions

GetOptimizationSuggestions returns performance and best practice suggestions

```go
// GetOptimizationSuggestions returns performance and best practice suggestions
func (p *Package) GetOptimizationSuggestions() []ValidationError {
	var suggestions []ValidationError

	if p == nil {
		return suggestions
	}

	graph := p.BuildDependencyGraph()

	unusedVars := p.GetUnusedVariables()
	for _, v := range unusedVars {
		suggestions = append(suggestions, ValidationError{
			Severity:	"info",
			Message:	"Variable is not used in any expressions or tasks - consider removing",
			Path:		"Variables." + v,
		})
	}

	for taskID, deps := range graph.TaskDependencies {
		if len(deps) > 10 {
			suggestions = append(suggestions, ValidationError{
				Severity:	"warning",
				Message:	fmt.Sprintf("Task has %d dependencies - consider simplifying", len(deps)),
				Path:		"Executables." + taskID,
			})
		}
	}

	for varName, usages := range graph.VariableDependencies {
		if len(usages) > 20 {
			suggestions = append(suggestions, ValidationError{
				Severity:	"info",
				Message:	fmt.Sprintf("Variable is referenced in %d locations - high impact if changed", len(usages)),
				Path:		"Variables." + varName,
			})
		}
	}

	for connName, tasks := range graph.ConnectionDependencies {
		if len(tasks) > 5 {
			suggestions = append(suggestions, ValidationError{
				Severity:	"info",
				Message:	fmt.Sprintf("Connection is used by %d tasks - consider connection pooling", len(tasks)),
				Path:		"ConnectionManagers." + connName,
			})
		}
	}

	return suggestions
}
```

#### GetUnusedVariables

GetUnusedVariables returns variables that are not referenced anywhere

```go
// GetUnusedVariables returns variables that are not referenced anywhere
func (p *Package) GetUnusedVariables() []string {
	if p == nil || p.Variables == nil || p.Variables.Variable == nil {
		return nil
	}

	graph := p.BuildDependencyGraph()
	usedVars := make(map[string]bool)
	for varName := range graph.VariableDependencies {
		usedVars[varName] = true
	}

	var unused []string
	for _, v := range p.Variables.Variable {
		if v.NamespaceAttr != nil && v.ObjectNameAttr != nil {
			fullName := *v.NamespaceAttr + "::" + *v.ObjectNameAttr
			if !usedVars[fullName] {
				unused = append(unused, fullName)
			}
		}
	}
	return unused
}
```

#### GetVariableByName

GetVariableByName finds a variable by name (ObjectName property)

```go
// GetVariableByName finds a variable by name (ObjectName property)
func (p *Package) GetVariableByName(name string) (*schema.VariableType, error) {
	if p == nil || p.Variables == nil || p.Variables.Variable == nil {
		return nil, fmt.Errorf("package or variables are nil")
	}

	// Parse name to extract namespace and object name
	var searchNamespace, searchObjectName string
	if strings.Contains(name, "::") {
		parts := strings.SplitN(name, "::", 2)
		searchNamespace = parts[0]
		searchObjectName = parts[1]
	} else {
		searchObjectName = name
	}

	for _, v := range p.Variables.Variable {
		if v.ObjectNameAttr != nil && *v.ObjectNameAttr == searchObjectName {

			if searchNamespace != "" {
				if v.NamespaceAttr != nil && *v.NamespaceAttr == searchNamespace {
					return v, nil
				}
			} else {

				return v, nil
			}
		}
	}
	return nil, fmt.Errorf("variable %s not found", name)
}
```

#### GetVariables

GetVariables returns all variables in the package

```go
// GetVariables returns all variables in the package
func (p *Package) GetVariables() *QueryResult {
	if p == nil || p.Variables == nil || p.Variables.Variable == nil {
		return &QueryResult{Count: 0, Results: []*schema.VariableType{}}
	}
	return &QueryResult{
		Count:		len(p.Variables.Variable),
		Results:	p.Variables.Variable,
	}
}
```

#### QueryExecutables

QueryExecutables finds executables matching a filter function

```go
// QueryExecutables finds executables matching a filter function
func (p *Package) QueryExecutables(filter func(*schema.AnyNonPackageExecutableType) bool) []*schema.AnyNonPackageExecutableType {
	var results []*schema.AnyNonPackageExecutableType
	if p == nil || p.Executable == nil {
		return results
	}
	for _, exec := range p.Executable {
		if filter(exec) {
			results = append(results, exec)
		}
	}
	return results
}
```

#### Validate

Validate performs comprehensive validation on the package

```go
// Validate performs comprehensive validation on the package
func (p *Package) Validate() []ValidationError {
	var errors []ValidationError

	if p == nil {
		return []ValidationError{{Severity: "error", Message: "Package is nil"}}
	}

	errors = append(errors, p.validateVariables()...)

	errors = append(errors, p.validateConnections()...)

	errors = append(errors, p.validateExpressions()...)

	errors = append(errors, p.validateStructure()...)

	return errors
}
```

### PackageBuilder

#### AddConnection

AddConnection adds a connection manager to the package

```go
// AddConnection adds a connection manager to the package
func (pb *PackageBuilder) AddConnection(name, connectionType, connectionString string) *PackageBuilder {
	if pb.pkg.ConnectionManagers == nil {
		pb.pkg.ConnectionManagers = &schema.ConnectionManagersType{}
	}
	if pb.pkg.ConnectionManagers.ConnectionManager == nil {
		pb.pkg.ConnectionManagers.ConnectionManager = []*schema.ConnectionManagerType{}
	}
	cm := &schema.ConnectionManagerType{
		ObjectNameAttr:		&name,
		CreationNameAttr:	&connectionType,
		Property: []*schema.Property{
			{
				NameAttr:	stringPtr("ConnectionString"),
				PropertyElementBaseType: &schema.PropertyElementBaseType{
					AnySimpleType: &schema.AnySimpleType{
						Value: connectionString,
					},
				},
			},
		},
	}
	pb.pkg.ConnectionManagers.ConnectionManager = append(pb.pkg.ConnectionManagers.ConnectionManager, cm)
	return pb
}
```

#### AddConnectionExpression

AddConnectionExpression adds a property expression to an existing connection manager

```go
// AddConnectionExpression adds a property expression to an existing connection manager
func (pb *PackageBuilder) AddConnectionExpression(connectionName, propertyName, expression string) *PackageBuilder {
	if pb.pkg.ConnectionManagers == nil || pb.pkg.ConnectionManagers.ConnectionManager == nil {
		return pb
	}

	for _, cm := range pb.pkg.ConnectionManagers.ConnectionManager {
		if cm.ObjectNameAttr != nil && *cm.ObjectNameAttr == connectionName {

			if cm.PropertyExpression == nil {
				cm.PropertyExpression = []*schema.PropertyExpressionElementType{}
			}

			cm.PropertyExpression = append(cm.PropertyExpression, &schema.PropertyExpressionElementType{
				NameAttr:	propertyName,
				AnySimpleType: &schema.AnySimpleType{
					Value: expression,
				},
			})
			break
		}
	}

	return pb
}
```

#### AddVariable

AddVariable adds a variable to the package

```go
// AddVariable adds a variable to the package
func (pb *PackageBuilder) AddVariable(namespace, name, value string) *PackageBuilder {
	return pb.AddVariableWithType(namespace, name, value, "String")
}
```

#### AddVariableWithType

AddVariableWithType adds a variable to the package with a specific data type

```go
// AddVariableWithType adds a variable to the package with a specific data type
func (pb *PackageBuilder) AddVariableWithType(namespace, name, value string, dataType string) *PackageBuilder {
	if pb.pkg.Variables == nil {
		pb.pkg.Variables = &schema.VariablesType{}
	}
	if pb.pkg.Variables.Variable == nil {
		pb.pkg.Variables.Variable = []*schema.VariableType{}
	}

	dataTypeCode := mapDataTypeToCode(dataType)

	v := &schema.VariableType{
		NamespaceAttr:	&namespace,
		ObjectNameAttr:	&name,
		VariableValue: &schema.VariableValue{
			DataTypeAttr:	&dataTypeCode,
			Value:		value,
		},
	}
	pb.pkg.Variables.Variable = append(pb.pkg.Variables.Variable, v)
	return pb
}
```

#### Build

Build returns the constructed package

```go
// Build returns the constructed package
func (pb *PackageBuilder) Build() *Package {
	return pb.pkg
}
```

### PackageParser

#### EvaluateExpression

EvaluateExpression evaluates an expression with caching

```go
// EvaluateExpression evaluates an expression with caching
func (p *PackageParser) EvaluateExpression(expr string) (interface{}, error) {
	if expr == "" {
		return nil, fmt.Errorf("empty expression")
	}

	if cached, exists := p.varCache["expr:"+expr]; exists {
		return cached, nil
	}

	result, err := EvaluateExpression(expr, p.pkg)
	if err != nil {
		return nil, err
	}

	p.varCache["expr:"+expr] = result
	return result, nil
}
```

#### GetConnectionManager

GetConnectionManager returns a connection manager by refId or name

```go
// GetConnectionManager returns a connection manager by refId or name
func (p *PackageParser) GetConnectionManager(id string) (*schema.ConnectionManagerType, error) {
	if cm, exists := p.connMap[id]; exists {
		return cm, nil
	}
	return nil, fmt.Errorf("connection manager %s not found", id)
}
```

#### GetExecutable

GetExecutable returns an executable by refId

```go
// GetExecutable returns an executable by refId
func (p *PackageParser) GetExecutable(refId string) (*schema.AnyNonPackageExecutableType, error) {
	if exec, exists := p.execMap[refId]; exists {
		return exec, nil
	}
	return nil, fmt.Errorf("executable %s not found", refId)
}
```

#### GetSQLStatements

GetSQLStatements extracts SQL statements from all executables

```go
// GetSQLStatements extracts SQL statements from all executables
func (p *PackageParser) GetSQLStatements() []*SQLStatement {
	var statements []*SQLStatement
	if p.pkg.Executable == nil {
		return statements
	}

	for _, exec := range p.pkg.Executable {
		taskName := "Unknown"
		if exec.ObjectNameAttr != nil {
			taskName = *exec.ObjectNameAttr
		}

		if exec.Property != nil {
			for _, prop := range exec.Property {
				if prop.NameAttr != nil && *prop.NameAttr == "SqlStatementSource" &&
					prop.PropertyElementBaseType != nil && prop.PropertyElementBaseType.AnySimpleType != nil {
					statements = append(statements, &SQLStatement{
						TaskName:	taskName,
						TaskType:	"Control Flow",
						SQL:		prop.PropertyElementBaseType.AnySimpleType.Value,
						RefId:		getRefId(exec),
						Connections:	p.getConnectionsForExecutable(exec),
					})
				}
			}
		}

		if exec.ObjectData != nil {
			p.extractTaskSpecificSQL(exec, &statements)
		}

		if exec.ExecutableTypeAttr == "Microsoft.Pipeline" && exec.ObjectData != nil {
			p.extractDataflowSQL(exec, &statements)
		}
	}

	return statements
}
```

#### GetVariableValue

GetVariableValue returns the value of a variable by name

```go
// GetVariableValue returns the value of a variable by name
func (p *PackageParser) GetVariableValue(name string) (interface{}, error) {
	if value, exists := p.vars[name]; exists {
		return value, nil
	}
	return nil, fmt.Errorf("variable %s not found", name)
}
```

### PackageValidator

#### Validate

Validate performs comprehensive validation of the package

```go
// Validate performs comprehensive validation of the package
func (v *PackageValidator) Validate() []*ValidationError {
	var errors []*ValidationError

	if varErrors := v.pkg.validateVariables(); len(varErrors) > 0 {
		for _, err := range varErrors {
			errors = append(errors, &ValidationError{
				Severity:	err.Severity,
				Message:	err.Message,
				Path:		err.Path,
			})
		}
	}

	if constraintErrors := v.analyzer.ValidateConstraints(); len(constraintErrors) > 0 {
		for _, err := range constraintErrors {
			errors = append(errors, &ValidationError{
				Severity:	"error",
				Message:	err.Error(),
				Path:		"PrecedenceConstraints",
			})
		}
	}

	if connErrors := v.validateConnections(); len(connErrors) > 0 {
		errors = append(errors, connErrors...)
	}

	if exprErrors := v.validateExpressions(); len(exprErrors) > 0 {
		errors = append(errors, exprErrors...)
	}

	return errors
}
```

### PrecedenceAnalyzer

#### GetAllExecutionOrders

GetAllExecutionOrders returns execution orders for all executables

```go
// GetAllExecutionOrders returns execution orders for all executables
func (p *PrecedenceAnalyzer) GetAllExecutionOrders() (map[string]int, error) {
	orders := make(map[string]int)
	for refId := range p.execMap {
		order, err := p.GetExecutionOrder(refId)
		if err != nil {
			return nil, err
		}
		orders[refId] = order
	}
	return orders, nil
}
```

#### GetExecutableChain

GetExecutableChain returns the execution chain for an executable (all predecessors)

```go
// GetExecutableChain returns the execution chain for an executable (all predecessors)
func (p *PrecedenceAnalyzer) GetExecutableChain(refId string) ([]string, error) {
	var chain []string
	visited := make(map[string]bool)

	var buildChain func(string) error
	buildChain = func(id string) error {
		if visited[id] {
			return fmt.Errorf("circular dependency detected at %s", id)
		}
		visited[id] = true

		for _, depId := range p.dependencies[id] {
			if err := buildChain(depId); err != nil {
				return err
			}
		}

		chain = append(chain, id)
		return nil
	}

	if err := buildChain(refId); err != nil {
		return nil, err
	}

	return chain, nil
}
```

#### GetExecutionFlowDescription

GetExecutionFlowDescription returns a textual description of the execution flow

```go
// GetExecutionFlowDescription returns a textual description of the execution flow
func (p *PrecedenceAnalyzer) GetExecutionFlowDescription() string {
	if p.pkg == nil || len(p.pkg.Executable) == 0 {
		return "No executables found in package."
	}

	var flow strings.Builder
	flow.WriteString("Execution Flow Description:\n")

	orders, err := p.GetAllExecutionOrders()
	if err != nil {
		return fmt.Sprintf("Error calculating execution order: %v", err)
	}

	// Sort by order
	type execInfo struct {
		order	int
		name	string
		typ	string
	}
	var execs []execInfo
	for refId, order := range orders {
		if exec, exists := p.execMap[refId]; exists {
			name := GetExecutableName(exec)
			typ := exec.ExecutableTypeAttr
			execs = append(execs, execInfo{order, name, typ})
		}
	}

	sort.Slice(execs, func(i, j int) bool {
		return execs[i].order < execs[j].order
	})

	for _, exec := range execs {
		flow.WriteString(fmt.Sprintf("Task %d: %s", exec.order, exec.name))
		if exec.typ != "" {
			flow.WriteString(fmt.Sprintf(" (%s)", exec.typ))
		}
		flow.WriteString("\n")
	}

	return flow.String()
}
```

#### GetExecutionOrder

GetExecutionOrder returns the execution order for an executable

```go
// GetExecutionOrder returns the execution order for an executable
func (p *PrecedenceAnalyzer) GetExecutionOrder(refId string) (int, error) {
	if order, exists := p.orderCache[refId]; exists {
		return order, nil
	}

	if len(p.dependencies[refId]) == 0 {
		order := len(p.orderCache) + 1
		p.orderCache[refId] = order
		return order, nil
	}

	maxDepOrder := 0
	for _, depId := range p.dependencies[refId] {
		depOrder, err := p.GetExecutionOrder(depId)
		if err != nil {
			return 0, err
		}
		if depOrder > maxDepOrder {
			maxDepOrder = depOrder
		}
	}

	order := maxDepOrder + 1
	p.orderCache[refId] = order
	return order, nil
}
```

#### ValidateConstraints

ValidateConstraints checks for constraint violations and circular dependencies

```go
// ValidateConstraints checks for constraint violations and circular dependencies
func (p *PrecedenceAnalyzer) ValidateConstraints() []error {
	var errors []error

	for refId := range p.execMap {
		if _, err := p.GetExecutableChain(refId); err != nil {
			errors = append(errors, fmt.Errorf("constraint validation failed for %s: %v", refId, err))
		}
	}

	return errors
}
```

### UnaryOp

#### Eval

```go
func (u *UnaryOp) Eval(vars map[string]interface{}) (interface{}, error) {
	val, err := u.Expr.Eval(vars)
	if err != nil {
		return nil, err
	}

	switch u.Op {
	case "!":
		var b bool
		switch v := val.(type) {
		case bool:
			b = v
		case float64:
			b = v != 0
		case string:
			b = v != ""
		default:
			return nil, fmt.Errorf("cannot convert %T to boolean", val)
		}
		return !b, nil
	case "-":
		if f, ok := val.(float64); ok {
			return -f, nil
		}
		return nil, fmt.Errorf("cannot negate %T", val)
	}
	return nil, fmt.Errorf("unknown unary operator: %s", u.Op)
}
```

### Variable

#### Eval

```go
func (v *Variable) Eval(vars map[string]interface{}) (interface{}, error) {
	if val, ok := vars[v.Name]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("variable not found: %s", v.Name)
}
```

