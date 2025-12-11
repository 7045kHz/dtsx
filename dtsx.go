// Package dtsx provides functionality for reading and writing DTSX (SQL Server Integration Services) XML files.
package dtsx

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	schema "github.com/7045kHz/dtsx/schemas"
)

// generateGUID creates a simple GUID-like string for DTSID
func generateGUID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("{%s-%s-%s-%s-%s}",
		hex.EncodeToString(bytes[0:4]),
		hex.EncodeToString(bytes[4:6]),
		hex.EncodeToString(bytes[6:8]),
		hex.EncodeToString(bytes[8:10]),
		hex.EncodeToString(bytes[10:16]))
}

// Package represents a DTSX package structure
type Package struct {
	XMLName                        xml.Name `xml:"Executable"`
	RefIdAttr                      *string  `xml:"refId,attr"`
	CreationDateAttr               *string  `xml:"CreationDate,attr"`
	CreationNameAttr               *string  `xml:"CreationName,attr"`
	CreatorComputerNameAttr        *string  `xml:"CreatorComputerName,attr"`
	CreatorNameAttr                *string  `xml:"CreatorName,attr"`
	DescriptionAttr                *string  `xml:"Description,attr"`
	DTSIDAttr                      *string  `xml:"DTSID,attr"`
	EnableConfigAttr               *string  `xml:"EnableConfig,attr"`
	ExecutableTypeAttr             *string  `xml:"ExecutableType,attr"`
	LastModifiedProductVersionAttr *string  `xml:"LastModifiedProductVersion,attr"`
	LocaleIDAttr                   *string  `xml:"LocaleID,attr"`
	ObjectNameAttr                 *string  `xml:"ObjectName,attr"`
	PackageTypeAttr                *string  `xml:"PackageType,attr"`
	VersionBuildAttr               *string  `xml:"VersionBuild,attr"`
	VersionGUIDAttr                *string  `xml:"VersionGUID,attr"`
	*schema.ExecutableTypePackage
}

// PackageParser provides centralized parsing and analysis functionality for DTSX packages
type PackageParser struct {
	pkg      *Package
	vars     map[string]interface{}
	connMap  map[string]*schema.ConnectionManagerType
	execMap  map[string]*schema.AnyNonPackageExecutableType
	varCache map[string]interface{} // Cache for expensive operations
}

// NewPackageParser creates a new PackageParser for the given package
func NewPackageParser(pkg *Package) *PackageParser {
	parser := &PackageParser{
		pkg:      pkg,
		varCache: make(map[string]interface{}),
	}
	parser.initialize()
	return parser
}

// initialize builds internal maps and caches
func (p *PackageParser) initialize() {
	p.buildVariableMap()
	p.buildConnectionMap()
	p.buildExecutableMap()
}

// buildVariableMap creates a map of all variables with their values
func (p *PackageParser) buildVariableMap() {
	p.vars = make(map[string]interface{})
	if p.pkg.Variables == nil || p.pkg.Variables.Variable == nil {
		return
	}
	for _, v := range p.pkg.Variables.Variable {
		if v.NamespaceAttr == nil || v.ObjectNameAttr == nil {
			continue
		}
		fullName := *v.NamespaceAttr + "::" + *v.ObjectNameAttr
		var value interface{}
		if v.VariableValue != nil {
			if num, err := strconv.ParseFloat(v.VariableValue.Value, 64); err == nil {
				value = num
			} else {
				value = v.VariableValue.Value
			}
		} else {
			// From properties
			for _, prop := range v.Property {
				if prop.NameAttr != nil && *prop.NameAttr == "Value" {
					if num, err := strconv.ParseFloat(prop.Value, 64); err == nil {
						value = num
					} else {
						value = prop.Value
					}
					break
				}
			}
		}
		if value != nil {
			p.vars[fullName] = value
		}
	}
}

// buildConnectionMap creates a map of connection managers by refId and name
func (p *PackageParser) buildConnectionMap() {
	p.connMap = make(map[string]*schema.ConnectionManagerType)
	if p.pkg.ConnectionManagers == nil || p.pkg.ConnectionManagers.ConnectionManager == nil {
		return
	}
	for _, cm := range p.pkg.ConnectionManagers.ConnectionManager {
		if cm.RefIdAttr != nil {
			p.connMap[*cm.RefIdAttr] = cm
		}
		if cm.ObjectNameAttr != nil {
			p.connMap[*cm.ObjectNameAttr] = cm
		}
	}
}

// buildExecutableMap creates a map of executables by refId
func (p *PackageParser) buildExecutableMap() {
	p.execMap = make(map[string]*schema.AnyNonPackageExecutableType)
	if p.pkg.Executable == nil {
		return
	}
	for _, exec := range p.pkg.Executable {
		if exec.RefIdAttr != nil {
			p.execMap[*exec.RefIdAttr] = exec
		}
	}
}

// GetVariableValue returns the value of a variable by name
func (p *PackageParser) GetVariableValue(name string) (interface{}, error) {
	if value, exists := p.vars[name]; exists {
		return value, nil
	}
	return nil, fmt.Errorf("variable %s not found", name)
}

// GetConnectionManager returns a connection manager by refId or name
func (p *PackageParser) GetConnectionManager(id string) (*schema.ConnectionManagerType, error) {
	if cm, exists := p.connMap[id]; exists {
		return cm, nil
	}
	return nil, fmt.Errorf("connection manager %s not found", id)
}

// GetExecutable returns an executable by refId
func (p *PackageParser) GetExecutable(refId string) (*schema.AnyNonPackageExecutableType, error) {
	if exec, exists := p.execMap[refId]; exists {
		return exec, nil
	}
	return nil, fmt.Errorf("executable %s not found", refId)
}

// EvaluateExpression evaluates an expression with caching
func (p *PackageParser) EvaluateExpression(expr string) (interface{}, error) {
	if expr == "" {
		return nil, fmt.Errorf("empty expression")
	}

	// Check cache first
	if cached, exists := p.varCache["expr:"+expr]; exists {
		return cached, nil
	}

	// Evaluate using the package's EvaluateExpression
	result, err := EvaluateExpression(expr, p.pkg)
	if err != nil {
		return nil, err
	}

	// Cache the result
	p.varCache["expr:"+expr] = result
	return result, nil
}

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

		// Extract from properties (control flow tasks)
		if exec.Property != nil {
			for _, prop := range exec.Property {
				if prop.NameAttr != nil && *prop.NameAttr == "SqlStatementSource" &&
					prop.PropertyElementBaseType != nil && prop.PropertyElementBaseType.AnySimpleType != nil {
					statements = append(statements, &SQLStatement{
						TaskName:    taskName,
						TaskType:    "Control Flow",
						SQL:         prop.PropertyElementBaseType.AnySimpleType.Value,
						RefId:       getRefId(exec),
						Connections: p.getConnectionsForExecutable(exec),
					})
				}
			}
		}

		// Extract from task-specific ObjectData
		if exec.ObjectData != nil {
			p.extractTaskSpecificSQL(exec, &statements)
		}

		// Extract from dataflow components
		if exec.ExecutableTypeAttr == "Microsoft.Pipeline" && exec.ObjectData != nil {
			p.extractDataflowSQL(exec, &statements)
		}
	}

	return statements
}

// SQLStatement represents a SQL statement found in the package
type SQLStatement struct {
	TaskName    string
	TaskType    string
	SQL         string
	RefId       string
	Connections []string
}

// getRefId safely gets the refId from an executable
func getRefId(exec *schema.AnyNonPackageExecutableType) string {
	if exec.RefIdAttr != nil {
		return *exec.RefIdAttr
	}
	return ""
}

// getConnectionsForExecutable finds connection managers used by an executable
func (p *PackageParser) getConnectionsForExecutable(exec *schema.AnyNonPackageExecutableType) []string {
	var connections []string

	// Check property expressions for connection references
	if exec.PropertyExpression != nil {
		for _, expr := range exec.PropertyExpression {
			if expr.AnySimpleType != nil {
				conns := p.extractConnectionRefs(expr.AnySimpleType.Value)
				connections = append(connections, conns...)
			}
		}
	}

	// For dataflows, check component connections
	if exec.ExecutableTypeAttr == "Microsoft.Pipeline" && exec.ObjectData != nil {
		if exec.ObjectData.Pipeline != nil && exec.ObjectData.Pipeline.Components != nil {
			for _, comp := range exec.ObjectData.Pipeline.Components.Component {
				if comp.Connections != nil {
					for _, conn := range comp.Connections.Connection {
						if conn.ConnectionManagerIDAttr != nil {
							if cm, exists := p.connMap[*conn.ConnectionManagerIDAttr]; exists {
								if cm.ObjectNameAttr != nil {
									connections = append(connections, *cm.ObjectNameAttr)
								}
							}
						}
					}
				}
			}
		}
	}

	return connections
}

// extractConnectionRefs finds connection manager references in expressions
func (p *PackageParser) extractConnectionRefs(expr string) []string {
	var connections []string
	// Look for patterns like @[ConnectionManager::Name]
	re := regexp.MustCompile(`@\[ConnectionManager::([^\]]+)\]`)
	matches := re.FindAllStringSubmatch(expr, -1)
	for _, match := range matches {
		if len(match) > 1 {
			connections = append(connections, match[1])
		}
	}
	return connections
}

// extractDataflowSQL extracts SQL from dataflow components
func (p *PackageParser) extractDataflowSQL(exec *schema.AnyNonPackageExecutableType, statements *[]*SQLStatement) {
	taskName := "Unknown"
	if exec.ObjectNameAttr != nil {
		taskName = *exec.ObjectNameAttr
	}

	if exec.ObjectData.Pipeline == nil || exec.ObjectData.Pipeline.Components == nil {
		return
	}

	for _, comp := range exec.ObjectData.Pipeline.Components.Component {
		var sql string
		if comp.Properties != nil {
			for _, prop := range comp.Properties.Property {
				if prop.NameAttr == nil {
					continue
				}
				propName := *prop.NameAttr
				if propName == "SqlCommand" || propName == "SqlStatement" || propName == "CommandText" ||
					propName == "Query" || propName == "SelectQuery" || propName == "InsertQuery" ||
					propName == "UpdateQuery" || propName == "DeleteQuery" || propName == "OpenRowset" {
					sql = strings.TrimSpace(prop.Value)
					if propName == "OpenRowset" && sql != "" {
						sql = "SELECT * FROM " + sql
					}
					break
				}
			}
		}

		if sql != "" {
			connections := p.getConnectionsForComponent(comp)
			*statements = append(*statements, &SQLStatement{
				TaskName:    taskName,
				TaskType:    "Dataflow",
				SQL:         sql,
				RefId:       getRefId(exec),
				Connections: connections,
			})
		}
	}
}

// getConnectionsForComponent finds connections used by a component
func (p *PackageParser) getConnectionsForComponent(comp *schema.PipelineComponentType) []string {
	var connections []string
	if comp.Connections != nil {
		for _, conn := range comp.Connections.Connection {
			if conn.ConnectionManagerIDAttr != nil {
				if cm, exists := p.connMap[*conn.ConnectionManagerIDAttr]; exists {
					if cm.ObjectNameAttr != nil {
						connections = append(connections, *cm.ObjectNameAttr)
					}
				}
			}
		}
	}
	return connections
}

// PrecedenceAnalyzer handles execution order calculation with support for complex precedence constraints
type PrecedenceAnalyzer struct {
	pkg          *Package
	execMap      map[string]*schema.AnyNonPackageExecutableType
	orderCache   map[string]int
	dependencies map[string][]string
}

// NewPrecedenceAnalyzer creates a new analyzer for the given package
func NewPrecedenceAnalyzer(pkg *Package) *PrecedenceAnalyzer {
	analyzer := &PrecedenceAnalyzer{
		pkg:          pkg,
		execMap:      make(map[string]*schema.AnyNonPackageExecutableType),
		orderCache:   make(map[string]int),
		dependencies: make(map[string][]string),
	}
	analyzer.buildExecutableMap()
	analyzer.buildDependencies()
	return analyzer
}

// buildExecutableMap creates a map of executables by refId
func (p *PrecedenceAnalyzer) buildExecutableMap() {
	if p.pkg.Executable == nil {
		return
	}
	for _, exec := range p.pkg.Executable {
		if exec.RefIdAttr != nil {
			p.execMap[*exec.RefIdAttr] = exec
		}
	}
}

// buildDependencies analyzes precedence constraints to build dependency graph
func (p *PrecedenceAnalyzer) buildDependencies() {
	if p.pkg.Executable == nil {
		return
	}

	// Build dependency graph from precedence constraints
	for _, exec := range p.pkg.Executable {
		if exec.RefIdAttr == nil {
			continue
		}
		refId := *exec.RefIdAttr

		if exec.PrecedenceConstraint != nil {
			for _, pc := range exec.PrecedenceConstraint {
				if pc.Executable != nil {
					for _, pcExec := range pc.Executable {
						if pcExec.IDREFAttr != nil {
							// This executable depends on the referenced executable
							p.dependencies[refId] = append(p.dependencies[refId], *pcExec.IDREFAttr)
						}
					}
				}
			}
		}
	}
}

// GetExecutionOrder returns the execution order for an executable
func (p *PrecedenceAnalyzer) GetExecutionOrder(refId string) (int, error) {
	if order, exists := p.orderCache[refId]; exists {
		return order, nil
	}

	// If no dependencies, assign sequential order
	if len(p.dependencies[refId]) == 0 {
		order := len(p.orderCache) + 1
		p.orderCache[refId] = order
		return order, nil
	}

	// Find maximum order among dependencies
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

// ValidateConstraints checks for constraint violations and circular dependencies
func (p *PrecedenceAnalyzer) ValidateConstraints() []error {
	var errors []error

	// Check for circular dependencies
	for refId := range p.execMap {
		if _, err := p.GetExecutableChain(refId); err != nil {
			errors = append(errors, fmt.Errorf("constraint validation failed for %s: %v", refId, err))
		}
	}

	return errors
}

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
		order int
		name  string
		typ   string
	}
	var execs []execInfo
	for refId, order := range orders {
		if exec, exists := p.execMap[refId]; exists {
			name := GetExecutableName(exec)
			typ := exec.ExecutableTypeAttr
			execs = append(execs, execInfo{order, name, typ})
		}
	}

	// Sort by order
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

// PackageValidator provides validation functions for DTSX packages
type PackageValidator struct {
	pkg      *Package
	parser   *PackageParser
	analyzer *PrecedenceAnalyzer
}

// NewPackageValidator creates a new validator for the package
func NewPackageValidator(pkg *Package) *PackageValidator {
	return &PackageValidator{
		pkg:      pkg,
		parser:   NewPackageParser(pkg),
		analyzer: NewPrecedenceAnalyzer(pkg),
	}
}

// Validate performs comprehensive validation of the package
func (v *PackageValidator) Validate() []*ValidationError {
	var errors []*ValidationError

	// Validate variables
	if varErrors := v.pkg.validateVariables(); len(varErrors) > 0 {
		for _, err := range varErrors {
			errors = append(errors, &ValidationError{
				Severity: err.Severity,
				Message:  err.Message,
				Path:     err.Path,
			})
		}
	}

	// Validate precedence constraints
	if constraintErrors := v.analyzer.ValidateConstraints(); len(constraintErrors) > 0 {
		for _, err := range constraintErrors {
			errors = append(errors, &ValidationError{
				Severity: "error",
				Message:  err.Error(),
				Path:     "PrecedenceConstraints",
			})
		}
	}

	// Validate connections
	if connErrors := v.validateConnections(); len(connErrors) > 0 {
		errors = append(errors, connErrors...)
	}

	// Validate expressions
	if exprErrors := v.validateExpressions(); len(exprErrors) > 0 {
		errors = append(errors, exprErrors...)
	}

	return errors
}

// validateConnections checks connection managers for issues
func (v *PackageValidator) validateConnections() []*ValidationError {
	var errors []*ValidationError

	if v.pkg.ConnectionManagers == nil || v.pkg.ConnectionManagers.ConnectionManager == nil {
		return errors
	}

	// Check for duplicate connection names
	nameMap := make(map[string]bool)
	for _, cm := range v.pkg.ConnectionManagers.ConnectionManager {
		if cm.ObjectNameAttr == nil {
			errors = append(errors, &ValidationError{
				Severity: "error",
				Message:  "Connection manager missing name",
				Path:     "ConnectionManagers",
			})
			continue
		}
		name := *cm.ObjectNameAttr
		if nameMap[name] {
			errors = append(errors, &ValidationError{
				Severity: "error",
				Message:  "Duplicate connection manager name: " + name,
				Path:     "ConnectionManagers." + name,
			})
		}
		nameMap[name] = true

		// Check for connection string
		hasConnStr := false
		for _, prop := range cm.Property {
			if prop.NameAttr != nil && *prop.NameAttr == "ConnectionString" && prop.Value != "" {
				hasConnStr = true
				break
			}
		}
		if !hasConnStr {
			errors = append(errors, &ValidationError{
				Severity: "warning",
				Message:  "Connection manager has no connection string",
				Path:     "ConnectionManagers." + name,
			})
		}
	}

	return errors
}

// validateExpressions checks all expressions in the package
func (v *PackageValidator) validateExpressions() []*ValidationError {
	var errors []*ValidationError

	expressions := v.pkg.GetExpressions()
	if expressions.Count == 0 {
		return errors
	}

	exprInfos := expressions.Results.([]*ExpressionInfo)
	for _, expr := range exprInfos {
		_, err := v.parser.EvaluateExpression(expr.Expression)
		if err != nil {
			errors = append(errors, &ValidationError{
				Severity: "error",
				Message:  fmt.Sprintf("Expression evaluation failed: %v", err),
				Path:     expr.Location + "." + expr.Context,
			})
		}
	}

	return errors
}

// QueryResult wraps query results with metadata
type QueryResult struct {
	Count   int
	Results interface{}
}

// GetConnectionString returns the connection string of a connection manager
func GetConnectionString(cm *schema.ConnectionManagerType) string {
	if cm == nil {
		return ""
	}
	for _, prop := range cm.Property {
		if prop.NameAttr != nil && *prop.NameAttr == "ConnectionString" &&
			prop.PropertyElementBaseType != nil && prop.PropertyElementBaseType.AnySimpleType != nil {
			return prop.PropertyElementBaseType.AnySimpleType.Value
		}
	}
	return ""
}

// GetVariableValue returns the value of a variable
func GetVariableValue(v *schema.VariableType) string {
	if v == nil {
		return ""
	}
	if v.VariableValue != nil && v.VariableValue.Value != "" {
		return v.VariableValue.Value
	}
	// Fallback to property
	for _, prop := range v.Property {
		if prop.NameAttr != nil && *prop.NameAttr == "Value" && prop.Value != "" {
			return prop.Value
		}
	}
	return ""
}

// GetConnectionName returns the name of a connection manager
func GetConnectionName(cm *schema.ConnectionManagerType) string {
	if cm == nil {
		return "unnamed"
	}
	if cm.ObjectNameAttr != nil {
		return *cm.ObjectNameAttr
	}
	// Fallback to property
	for _, prop := range cm.Property {
		if prop.NameAttr != nil && *prop.NameAttr == "ObjectName" &&
			prop.PropertyElementBaseType != nil && prop.PropertyElementBaseType.AnySimpleType != nil {
			return prop.PropertyElementBaseType.AnySimpleType.Value
		}
	}
	return "unnamed"
}

// GetVariableName returns the full name (namespace::name) of a variable
func GetVariableName(v *schema.VariableType) string {
	if v == nil {
		return "unnamed"
	}
	namespace := "User"
	if v.NamespaceAttr != nil {
		namespace = *v.NamespaceAttr
	}
	name := "unnamed"
	if v.ObjectNameAttr != nil {
		name = *v.ObjectNameAttr
	}
	return namespace + "::" + name
}

// GetExecutableName returns the name of an executable
func GetExecutableName(exec *schema.AnyNonPackageExecutableType) string {
	if exec == nil {
		return "unnamed"
	}
	if exec.ObjectNameAttr != nil {
		return *exec.ObjectNameAttr
	}
	// Fallback to property
	for _, prop := range exec.Property {
		if prop.NameAttr != nil && *prop.NameAttr == "ObjectName" &&
			prop.PropertyElementBaseType != nil && prop.PropertyElementBaseType.AnySimpleType != nil {
			return prop.PropertyElementBaseType.AnySimpleType.Value
		}
	}
	return "unnamed"
}

// GetExpressionDetails returns detailed information about an expression including evaluation result and dependencies
func GetExpressionDetails(exprInfo *ExpressionInfo, pkg *Package) *ExpressionDetails {
	if exprInfo == nil {
		return nil
	}

	details := &ExpressionDetails{
		Expression: exprInfo.Expression,
		Location:   exprInfo.Location,
		Name:       exprInfo.Name,
		Context:    exprInfo.Context,
	}

	// Try to evaluate the expression
	if pkg != nil {
		parser := NewPackageParser(pkg)
		if result, err := parser.EvaluateExpression(exprInfo.Expression); err == nil {
			details.EvaluatedValue = fmt.Sprintf("%v", result)
		} else {
			details.EvaluationError = err.Error()
		}

		// Extract dependencies (variables, parameters, etc.)
		details.Dependencies = extractExpressionDependencies(exprInfo.Expression, pkg)
	}

	return details
}

// ExpressionDetails provides comprehensive information about an expression
type ExpressionDetails struct {
	Expression      string
	Location        string
	Name            string
	Context         string
	EvaluatedValue  string
	EvaluationError string
	Dependencies    []string
}

// extractExpressionDependencies extracts variable and parameter references from an expression
func extractExpressionDependencies(expr string, pkg *Package) []string {
	var deps []string

	// Simple regex patterns for common SSIS expression syntax
	// Variables: @[User::VarName] or @[System::VarName]
	varRegex := regexp.MustCompile(`@\[([^]]+)\]`)
	matches := varRegex.FindAllStringSubmatch(expr, -1)
	for _, match := range matches {
		if len(match) > 1 {
			deps = append(deps, match[1])
		}
	}

	// Parameters: $Project::ParamName or $Package::ParamName
	paramRegex := regexp.MustCompile(`\$([^:]+)::([^\s\)]+)`)
	paramMatches := paramRegex.FindAllStringSubmatch(expr, -1)
	for _, match := range paramMatches {
		if len(match) > 2 {
			deps = append(deps, match[1]+"::"+match[2])
		}
	}

	return deps
}

// GetProperty returns the value of the specified property by name for any struct
func GetProperty(s interface{}, name string) interface{} {
	t := reflect.TypeOf(s)
	v := reflect.ValueOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name == name {
			return v.Field(i).Interface()
		}
	}
	return nil
}

// GetConnections returns all connection managers in the package
func (p *Package) GetConnections() *QueryResult {
	if p == nil || p.ConnectionManagers == nil || p.ConnectionManagers.ConnectionManager == nil {
		return &QueryResult{Count: 0, Results: []*schema.ConnectionManagerType{}}
	}
	return &QueryResult{
		Count:   len(p.ConnectionManagers.ConnectionManager),
		Results: p.ConnectionManagers.ConnectionManager,
	}
}

// GetVariables returns all variables in the package
func (p *Package) GetVariables() *QueryResult {
	if p == nil || p.Variables == nil || p.Variables.Variable == nil {
		return &QueryResult{Count: 0, Results: []*schema.VariableType{}}
	}
	return &QueryResult{
		Count:   len(p.Variables.Variable),
		Results: p.Variables.Variable,
	}
}

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
			// If namespace was specified, check it matches
			if searchNamespace != "" {
				if v.NamespaceAttr != nil && *v.NamespaceAttr == searchNamespace {
					return v, nil
				}
			} else {
				// No namespace specified, return the first match
				return v, nil
			}
		}
	}
	return nil, fmt.Errorf("variable %s not found", name)
}

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

// ExpressionInfo contains information about an expression found in the package
type ExpressionInfo struct {
	Expression string
	Location   string // e.g., "Package", "Executable", "PrecedenceConstraint", etc.
	Name       string // property name if applicable
	Context    string // additional context like executable type, variable name, etc.
}

// GetExpressions returns all expressions found in the package
func (p *Package) GetExpressions() *QueryResult {
	var expressions []*ExpressionInfo
	if p == nil {
		return &QueryResult{Count: 0, Results: expressions}
	}

	// Package-level expressions
	if p.PropertyExpression != nil {
		for _, expr := range p.PropertyExpression {
			if expr.AnySimpleType != nil && expr.AnySimpleType.Value != "" {
				expressions = append(expressions, &ExpressionInfo{
					Expression: expr.AnySimpleType.Value,
					Location:   "Package",
					Name:       expr.NameAttr,
					Context:    "Package Property",
				})
			}
		}
	}

	// Executable expressions
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
							Expression: expr.AnySimpleType.Value,
							Location:   "Executable",
							Name:       expr.NameAttr,
							Context:    context,
						})
					}
				}
			}

			// Precedence constraints within executables
			if exec.PrecedenceConstraint != nil {
				for j, pc := range exec.PrecedenceConstraint {
					if pc.PropertyExpression != nil {
						for _, expr := range pc.PropertyExpression {
							if expr.AnySimpleType != nil && expr.AnySimpleType.Value != "" {
								context := fmt.Sprintf("Executable[%d] PrecedenceConstraint[%d]", i, j)
								expressions = append(expressions, &ExpressionInfo{
									Expression: expr.AnySimpleType.Value,
									Location:   "PrecedenceConstraint",
									Name:       expr.NameAttr,
									Context:    context,
								})
							}
						}
					}
				}
			}
		}
	}

	// Package-level precedence constraints
	if p.PrecedenceConstraint != nil {
		for i, pc := range p.PrecedenceConstraint {
			if pc.PropertyExpression != nil {
				for _, expr := range pc.PropertyExpression {
					if expr.AnySimpleType != nil && expr.AnySimpleType.Value != "" {
						context := fmt.Sprintf("Package PrecedenceConstraint[%d]", i)
						expressions = append(expressions, &ExpressionInfo{
							Expression: expr.AnySimpleType.Value,
							Location:   "PrecedenceConstraint",
							Name:       expr.NameAttr,
							Context:    context,
						})
					}
				}
			}
		}
	}

	// Variable expressions
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
							Expression: expr.AnySimpleType.Value,
							Location:   "Variable",
							Name:       expr.NameAttr,
							Context:    context,
						})
					}
				}
			}
		}
	}

	// Connection manager expressions
	if p.ConnectionManagers != nil && p.ConnectionManagers.ConnectionManager != nil {
		for i, cm := range p.ConnectionManagers.ConnectionManager {
			if cm.PropertyExpression != nil {
				for _, expr := range cm.PropertyExpression {
					if expr.AnySimpleType != nil && expr.AnySimpleType.Value != "" {
						context := fmt.Sprintf("ConnectionManager[%d]", i)
						expressions = append(expressions, &ExpressionInfo{
							Expression: expr.AnySimpleType.Value,
							Location:   "ConnectionManager",
							Name:       expr.NameAttr,
							Context:    context,
						})
					}
				}
			}
		}
	}

	return &QueryResult{
		Count:   len(expressions),
		Results: expressions,
	}
}

// Unmarshal parses DTSX XML data and returns a Package
func Unmarshal(data []byte) (*Package, error) {
	// Preprocess XML to ensure DTS namespace compatibility
	xmlStr := string(data)

	// Remove DTS prefixes to match the schema structs
	xmlStr = strings.ReplaceAll(xmlStr, `<DTS:`, `<`)
	xmlStr = strings.ReplaceAll(xmlStr, `</DTS:`, `</`)
	xmlStr = strings.ReplaceAll(xmlStr, ` DTS:`, ` `)
	xmlStr = strings.ReplaceAll(xmlStr, `xmlns:DTS="www.microsoft.com/SqlServer/Dts"`, ``)

	data = []byte(xmlStr)

	var pkg Package
	err := xml.Unmarshal(data, &pkg)
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}

// UnmarshalFromReader parses DTSX XML from an io.Reader and returns a Package
func UnmarshalFromReader(r io.Reader) (*Package, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return Unmarshal(data)
}

// UnmarshalFromFile reads a DTSX file and returns a Package
func UnmarshalFromFile(filename string) (*Package, error) {
	filename = filepath.Clean(filename)
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	filename = absPath

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return UnmarshalFromReader(file)
}

// Marshal converts a Package to DTSX XML format
func Marshal(pkg *Package) ([]byte, error) {
	data, err := xml.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return nil, err
	}
	// Add XML declaration and fix namespace
	xmlStr := string(data)
	// Replace the root Executable element with DTS prefix
	xmlStr = strings.Replace(xmlStr, `<Executable `, `<DTS:Executable `, 1)

	// Add DTS prefix to all attributes except those in xmlns
	xmlStr = regexp.MustCompile(`(\w+)="([^"]*)"`).ReplaceAllStringFunc(xmlStr, func(match string) string {
		// Skip if this is part of xmlns declaration
		if strings.Contains(match, `xmlns`) || strings.HasPrefix(match, `DTS:`) {
			return match
		}
		parts := strings.SplitN(match, `="`, 2)
		if len(parts) == 2 {
			return `DTS:` + parts[0] + `="` + parts[1]
		}
		return match
	})

	// Add xmlns declaration to the root element
	xmlStr = strings.Replace(xmlStr, `<DTS:Executable `, `<DTS:Executable xmlns:DTS="www.microsoft.com/SqlServer/Dts" `, 1)

	// Add DTS prefix to all inner elements
	xmlStr = regexp.MustCompile(`<Property`).ReplaceAllString(xmlStr, `<DTS:Property`)
	xmlStr = regexp.MustCompile(`</Property>`).ReplaceAllString(xmlStr, `</DTS:Property>`)
	xmlStr = regexp.MustCompile(`<ConnectionManagers`).ReplaceAllString(xmlStr, `<DTS:ConnectionManagers`)
	xmlStr = regexp.MustCompile(`</ConnectionManagers>`).ReplaceAllString(xmlStr, `</DTS:ConnectionManagers>`)
	xmlStr = regexp.MustCompile(`<ConnectionManager`).ReplaceAllString(xmlStr, `<DTS:ConnectionManager`)
	xmlStr = regexp.MustCompile(`</ConnectionManager>`).ReplaceAllString(xmlStr, `</DTS:ConnectionManager>`)
	xmlStr = regexp.MustCompile(`<Variables`).ReplaceAllString(xmlStr, `<DTS:Variables`)
	xmlStr = regexp.MustCompile(`</Variables>`).ReplaceAllString(xmlStr, `</DTS:Variables>`)
	xmlStr = regexp.MustCompile(`<Variable`).ReplaceAllString(xmlStr, `<DTS:Variable`)
	xmlStr = regexp.MustCompile(`</Variable>`).ReplaceAllString(xmlStr, `</DTS:Variable>`)
	xmlStr = regexp.MustCompile(`<VariableValue`).ReplaceAllString(xmlStr, `<DTS:VariableValue`)
	xmlStr = regexp.MustCompile(`</VariableValue>`).ReplaceAllString(xmlStr, `</DTS:VariableValue>`)
	xmlStr = regexp.MustCompile(`<Executables`).ReplaceAllString(xmlStr, `<DTS:Executables`)
	xmlStr = regexp.MustCompile(`</Executables>`).ReplaceAllString(xmlStr, `</DTS:Executables>`)
	xmlStr = regexp.MustCompile(`<Executable`).ReplaceAllString(xmlStr, `<DTS:Executable`)
	xmlStr = regexp.MustCompile(`</Executable>`).ReplaceAllString(xmlStr, `</DTS:Executable>`)
	xmlStr = regexp.MustCompile(`<ObjectData`).ReplaceAllString(xmlStr, `<DTS:ObjectData`)
	xmlStr = regexp.MustCompile(`</ObjectData>`).ReplaceAllString(xmlStr, `</DTS:ObjectData>`)
	return []byte(xml.Header + xmlStr), nil
}

// MarshalToWriter writes a Package as DTSX XML to an io.Writer
func MarshalToWriter(w io.Writer, pkg *Package) error {
	data, err := Marshal(pkg)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// MarshalToFile writes a Package as DTSX XML to a file
func MarshalToFile(filename string, pkg *Package) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return MarshalToWriter(file, pkg)
}

// IsDTSXPackage validates if the given filename is a valid DTSX package.
// It checks if the file exists, is readable, and contains valid DTSX XML structure.
// Returns the unmarshaled Package and true if the file is a valid DTSX package,
// nil and false otherwise.
func IsDTSXPackage(filename string) (*Package, bool) {
	pkg, err := UnmarshalFromFile(filename)
	if err != nil {
		return nil, false
	}
	return pkg, true
}

// RunOptions contains options for executing a DTSX package with dtexec.exe
type RunOptions struct {
	// Package parameters (format: "[$Package::|$Project::|$ServerOption::]ParamName[(DataType)];Value")
	Parameters []string

	// Environment variables (format: "Name=Value")
	EnvironmentVars []string

	// Connection manager overrides (format: "id_or_name;connection_string")
	Connections []string

	// Configuration file path
	ConfigFile string

	// Property overrides using /Set (format: "propertyPath;value")
	PropertySets []string

	// Decryption password for encrypted packages
	DecryptPassword string

	// SQL Server name (for /SQL or /DTS packages)
	Server string

	// SQL Server username (for SQL authentication)
	User string

	// SQL Server password (for SQL authentication)
	Password string

	// Enable checkpointing (on/off)
	Checkpointing string

	// Checkpoint file path
	CheckpointFile string

	// Restart mode (deny/force/ifPossible)
	Restart string

	// Maximum concurrent executables (-1 for auto)
	MaxConcurrent int

	// Validate only without executing
	Validate bool

	// Treat warnings as errors
	WarnAsError bool

	// Verify build number (format: "major;minor;build")
	VerifyBuild string

	// Verify package ID (GUID)
	VerifyPackageID string

	// Verify version ID (GUID)
	VerifyVersionID string

	// Verify digital signature
	VerifySigned bool

	// Reporting level (N=none, E=errors, W=warnings, I=info, C=custom, P=progress, V=verbose)
	ReportingLevel string

	// Console log options (format: "displayoptions;list_options;src_name_or_guid")
	ConsoleLog []string

	// Log provider configuration (format: "classid_or_progid;configstring")
	Loggers []string

	// Enable verbose logging to file
	VerboseLog string

	// Dump on error codes (semicolon-separated error codes)
	DumpOnCodes string

	// Dump on any error
	DumpOnError bool

	// Run in 32-bit mode (x86)
	X86 bool
}

// RunPackage executes a DTSX package using dtexec.exe.
// It takes the path to dtexec.exe, the path to the DTSX file, and optional RunOptions.
// Returns the combined stdout/stderr output and any error that occurred.
func RunPackage(dtexecPath, dtsxPath string, opts *RunOptions) (string, error) {
	args := []string{"/File", dtsxPath}

	if opts != nil {
		// Add parameters
		for _, param := range opts.Parameters {
			args = append(args, "/Par", param)
		}

		// Add environment variables
		for _, env := range opts.EnvironmentVars {
			args = append(args, "/Env", env)
		}

		// Add connection overrides
		for _, conn := range opts.Connections {
			args = append(args, "/Conn", conn)
		}

		// Add configuration file
		if opts.ConfigFile != "" {
			args = append(args, "/ConfigFile", opts.ConfigFile)
		}

		// Add property sets
		for _, set := range opts.PropertySets {
			args = append(args, "/Set", set)
		}

		// Add decryption password
		if opts.DecryptPassword != "" {
			args = append(args, "/Decrypt", opts.DecryptPassword)
		}

		// Add server
		if opts.Server != "" {
			args = append(args, "/Server", opts.Server)
		}

		// Add SQL authentication
		if opts.User != "" {
			args = append(args, "/User", opts.User)
		}
		if opts.Password != "" {
			args = append(args, "/Password", opts.Password)
		}

		// Add checkpointing
		if opts.Checkpointing != "" {
			args = append(args, "/CheckPointing", opts.Checkpointing)
		}
		if opts.CheckpointFile != "" {
			args = append(args, "/CheckFile", opts.CheckpointFile)
		}

		// Add restart mode
		if opts.Restart != "" {
			args = append(args, "/Restart", opts.Restart)
		}

		// Add max concurrent executables
		if opts.MaxConcurrent != 0 {
			args = append(args, "/MaxConcurrent", fmt.Sprintf("%d", opts.MaxConcurrent))
		}

		// Add validation flag
		if opts.Validate {
			args = append(args, "/Validate")
		}

		// Add warn as error flag
		if opts.WarnAsError {
			args = append(args, "/WarnAsError")
		}

		// Add verify build
		if opts.VerifyBuild != "" {
			args = append(args, "/VerifyBuild", opts.VerifyBuild)
		}

		// Add verify package ID
		if opts.VerifyPackageID != "" {
			args = append(args, "/VerifyPackageID", opts.VerifyPackageID)
		}

		// Add verify version ID
		if opts.VerifyVersionID != "" {
			args = append(args, "/VerifyVersionID", opts.VerifyVersionID)
		}

		// Add verify signed
		if opts.VerifySigned {
			args = append(args, "/VerifySigned")
		}

		// Add reporting level
		if opts.ReportingLevel != "" {
			args = append(args, "/Reporting", opts.ReportingLevel)
		}

		// Add console log options
		for _, consoleLog := range opts.ConsoleLog {
			args = append(args, "/ConsoleLog", consoleLog)
		}

		// Add loggers
		for _, logger := range opts.Loggers {
			args = append(args, "/Logger", logger)
		}

		// Add verbose log
		if opts.VerboseLog != "" {
			args = append(args, "/VLog", opts.VerboseLog)
		}

		// Add dump on codes
		if opts.DumpOnCodes != "" {
			args = append(args, "/Dump", opts.DumpOnCodes)
		}

		// Add dump on error
		if opts.DumpOnError {
			args = append(args, "/DumpOnError")
		}

		// Add x86 flag
		if opts.X86 {
			args = append(args, "/X86")
		}
	}

	cmd := exec.Command(dtexecPath, args...)
	output, err := cmd.CombinedOutput()

	return strings.TrimSpace(string(output)), err
}

// PackageBuilder provides a fluent API for constructing DTSX packages
type PackageBuilder struct {
	pkg *Package
}

// NewPackageBuilder creates a new package builder
func NewPackageBuilder() *PackageBuilder {
	return &PackageBuilder{
		pkg: &Package{
			ExecutableTypePackage: &schema.ExecutableTypePackage{},
		},
	}
}

// AddVariable adds a variable to the package
func (pb *PackageBuilder) AddVariable(namespace, name, value string) *PackageBuilder {
	return pb.AddVariableWithType(namespace, name, value, "String")
}

// AddVariableWithType adds a variable to the package with a specific data type
func (pb *PackageBuilder) AddVariableWithType(namespace, name, value string, dataType string) *PackageBuilder {
	if pb.pkg.Variables == nil {
		pb.pkg.Variables = &schema.VariablesType{}
	}
	if pb.pkg.Variables.Variable == nil {
		pb.pkg.Variables.Variable = []*schema.VariableType{}
	}

	// Map common data type names to SSIS data type codes
	dataTypeCode := mapDataTypeToCode(dataType)

	v := &schema.VariableType{
		NamespaceAttr:  &namespace,
		ObjectNameAttr: &name,
		VariableValue: &schema.VariableValue{
			DataTypeAttr: &dataTypeCode,
			Value:        value,
		},
	}
	pb.pkg.Variables.Variable = append(pb.pkg.Variables.Variable, v)
	return pb
}

// mapDataTypeToCode maps common data type names to SSIS data type codes
func mapDataTypeToCode(dataType string) int {
	switch strings.ToLower(dataType) {
	case "string", "str", "dt_str":
		return 8 // DT_WSTR (Unicode string)
	case "int", "int32", "i4", "dt_i4":
		return 3 // DT_I4
	case "int64", "i8", "dt_i8":
		return 20 // DT_I8
	case "bool", "boolean", "dt_bool":
		return 11 // DT_BOOL
	case "datetime", "dt_dbtimestamp":
		return 135 // DT_DBTIMESTAMP
	case "decimal", "dt_decimal":
		return 25 // DT_DECIMAL
	case "double", "float", "r8", "dt_r8":
		return 5 // DT_R8
	case "guid", "dt_guid":
		return 72 // DT_GUID
	case "object", "dt_object":
		return 301 // DT_OBJECT
	default:
		return 8 // Default to string (DT_WSTR)
	}
}

// AddConnection adds a connection manager to the package
func (pb *PackageBuilder) AddConnection(name, connectionType, connectionString string) *PackageBuilder {
	if pb.pkg.ConnectionManagers == nil {
		pb.pkg.ConnectionManagers = &schema.ConnectionManagersType{}
	}
	if pb.pkg.ConnectionManagers.ConnectionManager == nil {
		pb.pkg.ConnectionManagers.ConnectionManager = []*schema.ConnectionManagerType{}
	}
	cm := &schema.ConnectionManagerType{
		ObjectNameAttr:   &name,
		CreationNameAttr: &connectionType,
		Property: []*schema.Property{
			{
				NameAttr: stringPtr("ConnectionString"),
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

// AddConnectionExpression adds a property expression to an existing connection manager
func (pb *PackageBuilder) AddConnectionExpression(connectionName, propertyName, expression string) *PackageBuilder {
	if pb.pkg.ConnectionManagers == nil || pb.pkg.ConnectionManagers.ConnectionManager == nil {
		return pb // No connections to add expression to
	}

	// Find the connection manager by name
	for _, cm := range pb.pkg.ConnectionManagers.ConnectionManager {
		if cm.ObjectNameAttr != nil && *cm.ObjectNameAttr == connectionName {
			// Initialize PropertyExpression slice if needed
			if cm.PropertyExpression == nil {
				cm.PropertyExpression = []*schema.PropertyExpressionElementType{}
			}

			// Add the expression
			cm.PropertyExpression = append(cm.PropertyExpression, &schema.PropertyExpressionElementType{
				NameAttr: propertyName,
				AnySimpleType: &schema.AnySimpleType{
					Value: expression,
				},
			})
			break
		}
	}

	return pb
}

// Build returns the constructed package
func (pb *PackageBuilder) Build() *Package {
	return pb.pkg
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// ValidationError represents a validation issue in a DTSX package
type ValidationError struct {
	Severity string // "error", "warning", "info"
	Message  string
	Path     string // Location in the package, e.g., "Variables.User::MyVar"
}

// Validate performs comprehensive validation on the package
func (p *Package) Validate() []ValidationError {
	var errors []ValidationError

	if p == nil {
		return []ValidationError{{Severity: "error", Message: "Package is nil"}}
	}

	// Validate variables
	errors = append(errors, p.validateVariables()...)

	// Validate connections
	errors = append(errors, p.validateConnections()...)

	// Validate expressions
	errors = append(errors, p.validateExpressions()...)

	// Validate structure
	errors = append(errors, p.validateStructure()...)

	return errors
}

// validateVariables checks for variable-related issues
func (p *Package) validateVariables() []ValidationError {
	var errors []ValidationError

	if p.Variables == nil || p.Variables.Variable == nil {
		return errors
	}

	// Check for duplicate variable names
	nameMap := make(map[string]bool)
	for _, v := range p.Variables.Variable {
		if v.NamespaceAttr == nil || v.ObjectNameAttr == nil {
			errors = append(errors, ValidationError{
				Severity: "error",
				Message:  "Variable missing namespace or name",
				Path:     "Variables",
			})
			continue
		}
		fullName := *v.NamespaceAttr + "::" + *v.ObjectNameAttr
		if nameMap[fullName] {
			errors = append(errors, ValidationError{
				Severity: "error",
				Message:  "Duplicate variable name: " + fullName,
				Path:     "Variables." + fullName,
			})
		}
		nameMap[fullName] = true

		// Check for empty values
		hasValue := false
		if v.VariableValue != nil && v.VariableValue.Value != "" {
			hasValue = true
		}
		if !hasValue {
			for _, prop := range v.Property {
				if prop.NameAttr != nil && *prop.NameAttr == "Value" && prop.Value != "" {
					hasValue = true
					break
				}
			}
		}
		if !hasValue {
			errors = append(errors, ValidationError{
				Severity: "warning",
				Message:  "Variable has no value",
				Path:     "Variables." + fullName,
			})
		}
	}

	return errors
}

// validateConnections checks for connection-related issues
func (p *Package) validateConnections() []ValidationError {
	var errors []ValidationError

	if p.ConnectionManagers == nil || p.ConnectionManagers.ConnectionManager == nil {
		return errors
	}

	// Check for duplicate connection names
	nameMap := make(map[string]bool)
	for _, cm := range p.ConnectionManagers.ConnectionManager {
		if cm.ObjectNameAttr == nil {
			errors = append(errors, ValidationError{
				Severity: "error",
				Message:  "Connection manager missing name",
				Path:     "ConnectionManagers",
			})
			continue
		}
		name := *cm.ObjectNameAttr
		if nameMap[name] {
			errors = append(errors, ValidationError{
				Severity: "error",
				Message:  "Duplicate connection manager name: " + name,
				Path:     "ConnectionManagers." + name,
			})
		}
		nameMap[name] = true

		// Check for connection string
		hasConnStr := false
		for _, prop := range cm.Property {
			if prop.NameAttr != nil && *prop.NameAttr == "ConnectionString" && prop.Value != "" {
				hasConnStr = true
				break
			}
		}
		if !hasConnStr {
			errors = append(errors, ValidationError{
				Severity: "warning",
				Message:  "Connection manager has no connection string",
				Path:     "ConnectionManagers." + name,
			})
		}
	}

	return errors
}

// validateExpressions checks for expression-related issues
func (p *Package) validateExpressions() []ValidationError {
	var errors []ValidationError

	// Get all expressions in the package
	exprResult := p.GetExpressions()
	if exprResult == nil || exprResult.Count == 0 {
		return errors
	}

	expressions := exprResult.Results.([]*ExpressionInfo)

	// Collect all variable references
	varRefs := make(map[string]bool)
	for _, expr := range expressions {
		refs := extractVariableReferences(expr.Expression)
		for _, ref := range refs {
			varRefs[ref] = true
		}
	}

	// Check for orphaned variables (variables not referenced in any expression)
	if p.Variables != nil && p.Variables.Variable != nil {
		for _, v := range p.Variables.Variable {
			if v.NamespaceAttr != nil && v.ObjectNameAttr != nil {
				fullName := *v.NamespaceAttr + "::" + *v.ObjectNameAttr
				if !varRefs[fullName] {
					errors = append(errors, ValidationError{
						Severity: "info",
						Message:  "Variable is not referenced in any expression",
						Path:     "Variables." + fullName,
					})
				}
			}
		}
	}

	// Check for undefined variables in expressions
	definedVars := make(map[string]bool)
	if p.Variables != nil && p.Variables.Variable != nil {
		for _, v := range p.Variables.Variable {
			if v.NamespaceAttr != nil && v.ObjectNameAttr != nil {
				definedVars[*v.NamespaceAttr+"::"+*v.ObjectNameAttr] = true
			}
		}
	}

	for _, expr := range expressions {
		refs := extractVariableReferences(expr.Expression)
		for _, ref := range refs {
			if !definedVars[ref] {
				errors = append(errors, ValidationError{
					Severity: "error",
					Message:  "Expression references undefined variable: " + ref,
					Path:     expr.Location,
				})
			}
		}
	}

	return errors
}

// validateStructure checks for structural issues
func (p *Package) validateStructure() []ValidationError {
	var errors []ValidationError

	// Check for empty package
	if len(p.Executable) == 0 {
		errors = append(errors, ValidationError{
			Severity: "info",
			Message:  "Package has no executable tasks",
			Path:     "Executables",
		})
	}

	// Check for missing package properties
	if len(p.Property) == 0 {
		errors = append(errors, ValidationError{
			Severity: "info",
			Message:  "Package has no properties",
			Path:     "Properties",
		})
	}

	return errors
}

// extractVariableReferences extracts @[Namespace::Name] patterns from an expression
func extractVariableReferences(expr string) []string {
	re := regexp.MustCompile(`@\[([^\]]+)\]`)
	matches := re.FindAllStringSubmatch(expr, -1)
	var refs []string
	for _, match := range matches {
		if len(match) > 1 {
			refs = append(refs, match[1])
		}
	}
	return refs
}

// DependencyGraph represents relationships between package elements
type DependencyGraph struct {
	// VariableDependencies: variable name -> list of expressions/locations that use it
	VariableDependencies map[string][]string
	// ConnectionDependencies: connection name -> list of tasks/locations that use it
	ConnectionDependencies map[string][]string
	// TaskDependencies: task ID -> list of variables/connections it depends on
	TaskDependencies map[string][]string
	// ExpressionDependencies: expression -> list of variables it references
	ExpressionDependencies map[string][]string
}

// BuildDependencyGraph analyzes the package and builds a dependency graph
func (p *Package) BuildDependencyGraph() *DependencyGraph {
	graph := &DependencyGraph{
		VariableDependencies:   make(map[string][]string),
		ConnectionDependencies: make(map[string][]string),
		TaskDependencies:       make(map[string][]string),
		ExpressionDependencies: make(map[string][]string),
	}

	if p == nil {
		return graph
	}

	// Analyze expressions for variable dependencies
	exprResult := p.GetExpressions()
	if exprResult.Count > 0 {
		expressions := exprResult.Results.([]*ExpressionInfo)
		for _, expr := range expressions {
			vars := extractVariableReferences(expr.Expression)
			graph.ExpressionDependencies[expr.Expression] = vars

			// Record which expressions use each variable
			location := fmt.Sprintf("%s:%s", expr.Location, expr.Name)
			for _, v := range vars {
				graph.VariableDependencies[v] = append(graph.VariableDependencies[v], location)
			}
		}
	}

	// Analyze tasks for connection dependencies
	if p.Executable != nil {
		for i, exec := range p.Executable {
			taskID := fmt.Sprintf("Executable[%d]", i)
			if exec.ExecutableTypeAttr != "" {
				taskID = fmt.Sprintf("%s (%s)", taskID, exec.ExecutableTypeAttr)
			}

			// Check properties for connection references
			if exec.Property != nil {
				for _, prop := range exec.Property {
					if prop.NameAttr != nil && *prop.NameAttr == "Connection" && prop.Value != "" {
						connName := prop.Value
						graph.ConnectionDependencies[connName] = append(graph.ConnectionDependencies[connName], taskID)
						graph.TaskDependencies[taskID] = append(graph.TaskDependencies[taskID], "Connection:"+connName)
					}
				}
			}

			// Check for variable references in task properties
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

// GetVariableImpact returns all locations affected by a variable change
func (dg *DependencyGraph) GetVariableImpact(varName string) []string {
	if dg == nil {
		return nil
	}
	return dg.VariableDependencies[varName]
}

// GetConnectionImpact returns all tasks affected by a connection change
func (dg *DependencyGraph) GetConnectionImpact(connName string) []string {
	if dg == nil {
		return nil
	}
	return dg.ConnectionDependencies[connName]
}

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

// GetOptimizationSuggestions returns performance and best practice suggestions
func (p *Package) GetOptimizationSuggestions() []ValidationError {
	var suggestions []ValidationError

	if p == nil {
		return suggestions
	}

	// Build dependency graph for analysis
	graph := p.BuildDependencyGraph()

	// Check for unused variables
	unusedVars := p.GetUnusedVariables()
	for _, v := range unusedVars {
		suggestions = append(suggestions, ValidationError{
			Severity: "info",
			Message:  "Variable is not used in any expressions or tasks - consider removing",
			Path:     "Variables." + v,
		})
	}

	// Check for tasks with many dependencies (potential performance bottleneck)
	for taskID, deps := range graph.TaskDependencies {
		if len(deps) > 10 {
			suggestions = append(suggestions, ValidationError{
				Severity: "warning",
				Message:  fmt.Sprintf("Task has %d dependencies - consider simplifying", len(deps)),
				Path:     "Executables." + taskID,
			})
		}
	}

	// Check for variables used in many places (high impact)
	for varName, usages := range graph.VariableDependencies {
		if len(usages) > 20 {
			suggestions = append(suggestions, ValidationError{
				Severity: "info",
				Message:  fmt.Sprintf("Variable is referenced in %d locations - high impact if changed", len(usages)),
				Path:     "Variables." + varName,
			})
		}
	}

	// Check for connections used by many tasks
	for connName, tasks := range graph.ConnectionDependencies {
		if len(tasks) > 5 {
			suggestions = append(suggestions, ValidationError{
				Severity: "info",
				Message:  fmt.Sprintf("Connection is used by %d tasks - consider connection pooling", len(tasks)),
				Path:     "ConnectionManagers." + connName,
			})
		}
	}

	return suggestions
}

// UpdateVariable updates the value of an existing variable
func (p *Package) UpdateVariable(namespace string, name, newValue string) error {
	if p == nil || p.Variables == nil || p.Variables.Variable == nil {
		return fmt.Errorf("package has no variables")
	}

	for _, v := range p.Variables.Variable {
		if v.NamespaceAttr != nil && v.ObjectNameAttr != nil &&
			*v.NamespaceAttr == namespace && *v.ObjectNameAttr == name {
			if v.VariableValue != nil {
				v.VariableValue.Value = newValue
				return nil
			}
			// If no VariableValue, create one
			v.VariableValue = &schema.VariableValue{Value: newValue}
			return nil
		}
	}

	return fmt.Errorf("variable %s::%s not found", namespace, name)
}

// UpdateConnectionString updates the connection string of an existing connection manager
func (p *Package) UpdateConnectionString(connectionName, newConnectionString string) error {
	if p == nil || p.ConnectionManagers == nil || p.ConnectionManagers.ConnectionManager == nil {
		return fmt.Errorf("package has no connection managers")
	}

	for _, cm := range p.ConnectionManagers.ConnectionManager {
		// Find the connection by name
		var connName string
		for _, prop := range cm.Property {
			if prop.NameAttr != nil && *prop.NameAttr == "ObjectName" &&
				prop.PropertyElementBaseType != nil && prop.PropertyElementBaseType.AnySimpleType != nil {
				connName = prop.PropertyElementBaseType.AnySimpleType.Value
				break
			}
		}

		if connName == connectionName {
			// Find and update the connection string property
			for _, prop := range cm.Property {
				if prop.NameAttr != nil && *prop.NameAttr == "ConnectionString" &&
					prop.PropertyElementBaseType != nil && prop.PropertyElementBaseType.AnySimpleType != nil {
					prop.PropertyElementBaseType.AnySimpleType.Value = newConnectionString
					return nil
				}
			}
			return fmt.Errorf("connection string property not found for connection %s", connectionName)
		}
	}

	return fmt.Errorf("connection manager %s not found", connectionName)
}

// UpdateExpression updates an expression for a specific property
func (p *Package) UpdateExpression(targetType, targetName, propertyName, newExpression string) error {
	if p == nil {
		return fmt.Errorf("package is nil")
	}

	switch targetType {
	case "variable":
		return p.updateVariableExpression(targetName, propertyName, newExpression)
	case "connection":
		return p.updateConnectionExpression(targetName, propertyName, newExpression)
	case "executable":
		return p.updateExecutableExpression(targetName, propertyName, newExpression)
	default:
		return fmt.Errorf("unsupported target type: %s (supported: variable, connection, executable)", targetType)
	}
}

// updateVariableExpression updates an expression on a variable
func (p *Package) updateVariableExpression(varName, propertyName, newExpression string) error {
	if p.Variables == nil || p.Variables.Variable == nil {
		return fmt.Errorf("package has no variables")
	}

	// Parse namespace::name format
	parts := strings.Split(varName, "::")
	if len(parts) != 2 {
		return fmt.Errorf("variable name must be in format namespace::name")
	}
	namespace, name := parts[0], parts[1]

	for _, v := range p.Variables.Variable {
		if v.NamespaceAttr != nil && v.ObjectNameAttr != nil &&
			*v.NamespaceAttr == namespace && *v.ObjectNameAttr == name {

			// For variables, expressions are typically on the VariableValue or as PropertyExpression
			if propertyName == "Value" {
				// Create or update PropertyExpression for the variable value
				if v.PropertyExpression == nil {
					v.PropertyExpression = []*schema.PropertyExpressionElementType{}
				}

				// Find existing expression or create new one
				found := false
				for _, expr := range v.PropertyExpression {
					if expr.NameAttr == propertyName {
						if expr.AnySimpleType == nil {
							expr.AnySimpleType = &schema.AnySimpleType{}
						}
						expr.AnySimpleType.Value = newExpression
						found = true
						break
					}
				}

				if !found {
					v.PropertyExpression = append(v.PropertyExpression, &schema.PropertyExpressionElementType{
						NameAttr: propertyName,
						AnySimpleType: &schema.AnySimpleType{
							Value: newExpression,
						},
					})
				}
				return nil
			}
		}
	}

	return fmt.Errorf("variable %s not found", varName)
}

// updateConnectionExpression updates an expression on a connection manager
func (p *Package) updateConnectionExpression(connName, propertyName, newExpression string) error {
	if p.ConnectionManagers == nil || p.ConnectionManagers.ConnectionManager == nil {
		return fmt.Errorf("package has no connection managers")
	}

	for _, cm := range p.ConnectionManagers.ConnectionManager {
		// Find connection by name
		var currentName string
		for _, prop := range cm.Property {
			if prop.NameAttr != nil && *prop.NameAttr == "ObjectName" &&
				prop.PropertyElementBaseType != nil && prop.PropertyElementBaseType.AnySimpleType != nil {
				currentName = prop.PropertyElementBaseType.AnySimpleType.Value
				break
			}
		}

		if currentName == connName {
			// Update or create expression
			if cm.PropertyExpression == nil {
				cm.PropertyExpression = []*schema.PropertyExpressionElementType{}
			}

			found := false
			for _, expr := range cm.PropertyExpression {
				if expr.NameAttr == propertyName {
					if expr.AnySimpleType == nil {
						expr.AnySimpleType = &schema.AnySimpleType{}
					}
					expr.AnySimpleType.Value = newExpression
					found = true
					break
				}
			}

			if !found {
				cm.PropertyExpression = append(cm.PropertyExpression, &schema.PropertyExpressionElementType{
					NameAttr: propertyName,
					AnySimpleType: &schema.AnySimpleType{
						Value: newExpression,
					},
				})
			}
			return nil
		}
	}

	return fmt.Errorf("connection manager %s not found", connName)
}

// updateExecutableExpression updates an expression on an executable/task
func (p *Package) updateExecutableExpression(execName, propertyName, newExpression string) error {
	if p.Executable == nil {
		return fmt.Errorf("package has no executables")
	}

	for _, exec := range p.Executable {
		// Find executable by name
		var currentName string
		for _, prop := range exec.Property {
			if prop.NameAttr != nil && *prop.NameAttr == "ObjectName" &&
				prop.PropertyElementBaseType != nil && prop.PropertyElementBaseType.AnySimpleType != nil {
				currentName = prop.PropertyElementBaseType.AnySimpleType.Value
				break
			}
		}

		if currentName == execName {
			// Update or create expression
			if exec.PropertyExpression == nil {
				exec.PropertyExpression = []*schema.PropertyExpressionElementType{}
			}

			found := false
			for _, expr := range exec.PropertyExpression {
				if expr.NameAttr == propertyName {
					if expr.AnySimpleType == nil {
						expr.AnySimpleType = &schema.AnySimpleType{}
					}
					expr.AnySimpleType.Value = newExpression
					found = true
					break
				}
			}

			if !found {
				exec.PropertyExpression = append(exec.PropertyExpression, &schema.PropertyExpressionElementType{
					NameAttr: propertyName,
					AnySimpleType: &schema.AnySimpleType{
						Value: newExpression,
					},
				})
			}
			return nil
		}
	}

	return fmt.Errorf("executable %s not found", execName)
}

// UpdateProperty updates any property on any element (package, variable, connection, executable)
func (p *Package) UpdateProperty(targetType, targetName, propertyName, newValue string) error {
	if p == nil {
		return fmt.Errorf("package is nil")
	}

	switch targetType {
	case "package":
		return p.updatePackageProperty(propertyName, newValue)
	case "variable":
		return p.updateVariableProperty(targetName, propertyName, newValue)
	case "connection":
		return p.updateConnectionProperty(targetName, propertyName, newValue)
	case "executable":
		return p.updateExecutableProperty(targetName, propertyName, newValue)
	default:
		return fmt.Errorf("unsupported target type: %s (supported: package, variable, connection, executable)", targetType)
	}
}

// updatePackageProperty updates a property on the package itself
func (p *Package) updatePackageProperty(propertyName, newValue string) error {
	if p.Property == nil {
		p.Property = []*schema.Property{}
	}

	// Find existing property or create new one
	found := false
	for _, prop := range p.Property {
		if prop.NameAttr != nil && *prop.NameAttr == propertyName {
			if prop.PropertyElementBaseType == nil {
				prop.PropertyElementBaseType = &schema.PropertyElementBaseType{}
			}
			if prop.PropertyElementBaseType.AnySimpleType == nil {
				prop.PropertyElementBaseType.AnySimpleType = &schema.AnySimpleType{}
			}
			prop.PropertyElementBaseType.AnySimpleType.Value = newValue
			found = true
			break
		}
	}

	if !found {
		p.Property = append(p.Property, &schema.Property{
			NameAttr: &propertyName,
			PropertyElementBaseType: &schema.PropertyElementBaseType{
				AnySimpleType: &schema.AnySimpleType{
					Value: newValue,
				},
			},
		})
	}

	return nil
}

// updateVariableProperty updates a property on a variable
func (p *Package) updateVariableProperty(varName, propertyName, newValue string) error {
	if p.Variables == nil || p.Variables.Variable == nil {
		return fmt.Errorf("package has no variables")
	}

	// Parse namespace::name format
	parts := strings.Split(varName, "::")
	if len(parts) != 2 {
		return fmt.Errorf("variable name must be in format namespace::name")
	}
	namespace, name := parts[0], parts[1]

	for _, v := range p.Variables.Variable {
		if v.NamespaceAttr != nil && v.ObjectNameAttr != nil &&
			*v.NamespaceAttr == namespace && *v.ObjectNameAttr == name {

			// Special handling for VariableValue
			if propertyName == "Value" {
				if v.VariableValue == nil {
					v.VariableValue = &schema.VariableValue{}
				}
				v.VariableValue.Value = newValue
				return nil
			}

			// Handle other properties
			if v.Property == nil {
				v.Property = []*schema.Property{}
			}

			found := false
			for _, prop := range v.Property {
				if prop.NameAttr != nil && *prop.NameAttr == propertyName {
					if prop.PropertyElementBaseType == nil {
						prop.PropertyElementBaseType = &schema.PropertyElementBaseType{}
					}
					if prop.PropertyElementBaseType.AnySimpleType == nil {
						prop.PropertyElementBaseType.AnySimpleType = &schema.AnySimpleType{}
					}
					prop.PropertyElementBaseType.AnySimpleType.Value = newValue
					found = true
					break
				}
			}

			if !found {
				v.Property = append(v.Property, &schema.Property{
					NameAttr: &propertyName,
					PropertyElementBaseType: &schema.PropertyElementBaseType{
						AnySimpleType: &schema.AnySimpleType{
							Value: newValue,
						},
					},
				})
			}
			return nil
		}
	}

	return fmt.Errorf("variable %s not found", varName)
}

// updateConnectionProperty updates a property on a connection manager
func (p *Package) updateConnectionProperty(connName, propertyName, newValue string) error {
	if p.ConnectionManagers == nil || p.ConnectionManagers.ConnectionManager == nil {
		return fmt.Errorf("package has no connection managers")
	}

	for _, cm := range p.ConnectionManagers.ConnectionManager {
		// Find connection by name
		var currentName string
		for _, prop := range cm.Property {
			if prop.NameAttr != nil && *prop.NameAttr == "ObjectName" &&
				prop.PropertyElementBaseType != nil && prop.PropertyElementBaseType.AnySimpleType != nil {
				currentName = prop.PropertyElementBaseType.AnySimpleType.Value
				break
			}
		}

		if currentName == connName {
			// Handle properties
			if cm.Property == nil {
				cm.Property = []*schema.Property{}
			}

			found := false
			for _, prop := range cm.Property {
				if prop.NameAttr != nil && *prop.NameAttr == propertyName {
					if prop.PropertyElementBaseType == nil {
						prop.PropertyElementBaseType = &schema.PropertyElementBaseType{}
					}
					if prop.PropertyElementBaseType.AnySimpleType == nil {
						prop.PropertyElementBaseType.AnySimpleType = &schema.AnySimpleType{}
					}
					prop.PropertyElementBaseType.AnySimpleType.Value = newValue
					found = true
					break
				}
			}

			if !found {
				cm.Property = append(cm.Property, &schema.Property{
					NameAttr: &propertyName,
					PropertyElementBaseType: &schema.PropertyElementBaseType{
						AnySimpleType: &schema.AnySimpleType{
							Value: newValue,
						},
					},
				})
			}
			return nil
		}
	}

	return fmt.Errorf("connection manager %s not found", connName)
}

// updateExecutableProperty updates a property on an executable/task
func (p *Package) updateExecutableProperty(execName, propertyName, newValue string) error {
	if p.Executable == nil {
		return fmt.Errorf("package has no executables")
	}

	for _, exec := range p.Executable {
		// Find executable by name
		var currentName string
		for _, prop := range exec.Property {
			if prop.NameAttr != nil && *prop.NameAttr == "ObjectName" &&
				prop.PropertyElementBaseType != nil && prop.PropertyElementBaseType.AnySimpleType != nil {
				currentName = prop.PropertyElementBaseType.AnySimpleType.Value
				break
			}
		}

		if currentName == execName {
			// Handle properties
			if exec.Property == nil {
				exec.Property = []*schema.Property{}
			}

			found := false
			for _, prop := range exec.Property {
				if prop.NameAttr != nil && *prop.NameAttr == propertyName {
					if prop.PropertyElementBaseType == nil {
						prop.PropertyElementBaseType = &schema.PropertyElementBaseType{}
					}
					if prop.PropertyElementBaseType.AnySimpleType == nil {
						prop.PropertyElementBaseType.AnySimpleType = &schema.AnySimpleType{}
					}
					prop.PropertyElementBaseType.AnySimpleType.Value = newValue
					found = true
					break
				}
			}

			if !found {
				exec.Property = append(exec.Property, &schema.Property{
					NameAttr: &propertyName,
					PropertyElementBaseType: &schema.PropertyElementBaseType{
						AnySimpleType: &schema.AnySimpleType{
							Value: newValue,
						},
					},
				})
			}
			return nil
		}
	}

	return fmt.Errorf("executable %s not found", execName)
}

// GetSqlStatementSource returns the SQL statement source from SqlTaskDataType
func GetSqlStatementSource(s *schema.SqlTaskDataType) string {
	if s == nil || s.SQLTaskSqlTaskBaseAttributeGroup == nil {
		return ""
	}
	return s.SQLTaskSqlTaskBaseAttributeGroup.SqlStatementSourceAttr
}

// GetSqlStatementSourceFromBase returns the SQL statement source from SqlTaskBaseAttributeGroup
func GetSqlStatementSourceFromBase(s *schema.SqlTaskBaseAttributeGroup) string {
	if s == nil {
		return ""
	}
	return s.SqlStatementSourceAttr
}

// extractTaskSpecificSQL extracts SQL from task-specific ObjectData (like Execute SQL Task)
func (p *PackageParser) extractTaskSpecificSQL(exec *schema.AnyNonPackageExecutableType, statements *[]*SQLStatement) {
	taskName := "Unknown"
	if exec.ObjectNameAttr != nil {
		taskName = *exec.ObjectNameAttr
	}

	if exec.ObjectData == nil {
		return
	}

	// Special handling for Execute SQL Task due to namespace parsing issues
	if exec.ExecutableTypeAttr == "Microsoft.ExecuteSQLTask" {
		// First try the normal schema parsing
		if exec.ObjectData.SQLTaskSqlTaskData != nil {
			sql := GetSqlStatementSource(exec.ObjectData.SQLTaskSqlTaskData)
			if sql != "" {
				*statements = append(*statements, &SQLStatement{
					TaskName:    taskName,
					TaskType:    "Control Flow",
					SQL:         sql,
					RefId:       getRefId(exec),
					Connections: p.getConnectionsForExecutable(exec),
				})
				return
			}
		}
		// Fallback to raw XML parsing
		sql := p.extractSQLFromExecuteSQLTask(exec)
		if sql != "" {
			*statements = append(*statements, &SQLStatement{
				TaskName:    taskName,
				TaskType:    "Control Flow",
				SQL:         sql,
				RefId:       getRefId(exec),
				Connections: p.getConnectionsForExecutable(exec),
			})
		}
		return
	}

	// Check for SQL Task data
	if exec.ObjectData.SQLTaskSqlTaskData != nil {
		sqlTaskData := exec.ObjectData.SQLTaskSqlTaskData
		if sqlTaskData.SQLTaskSqlTaskBaseAttributeGroup != nil &&
			sqlTaskData.SQLTaskSqlTaskBaseAttributeGroup.SqlStatementSourceAttr != "" {
			*statements = append(*statements, &SQLStatement{
				TaskName:    taskName,
				TaskType:    "Control Flow",
				SQL:         sqlTaskData.SQLTaskSqlTaskBaseAttributeGroup.SqlStatementSourceAttr,
				RefId:       getRefId(exec),
				Connections: p.getConnectionsForExecutable(exec),
			})
		}
	}

	// Check for other task types that might have SQL
	// Add more task types here as needed
}

// extractSQLFromExecuteSQLTask extracts SQL from Execute SQL Task by parsing the raw XML
func (p *PackageParser) extractSQLFromExecuteSQLTask(exec *schema.AnyNonPackageExecutableType) string {
	if exec.ObjectData == nil {
		return ""
	}

	// Use the InnerXML field that contains the raw XML
	xmlStr := exec.ObjectData.InnerXML

	// Find the SqlStatementSource attribute
	// The XML contains: SQLTask:SqlStatementSource="EXEC [ETC].[GetUtcDate]"
	start := strings.Index(xmlStr, `SqlStatementSource="`)
	if start == -1 {
		return ""
	}

	start += len(`SqlStatementSource="`)
	end := strings.Index(xmlStr[start:], `"`)
	if end == -1 {
		return ""
	}

	sql := xmlStr[start : start+end]
	// Unescape XML entities if any
	sql = strings.ReplaceAll(sql, "&lt;", "<")
	sql = strings.ReplaceAll(sql, "&gt;", ">")
	sql = strings.ReplaceAll(sql, "&amp;", "&")
	sql = strings.ReplaceAll(sql, "&quot;", `"`)
	sql = strings.ReplaceAll(sql, "&apos;", "'")

	return sql
}
