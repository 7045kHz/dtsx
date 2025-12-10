//go:build ignore

// analyze_connections.go - Comprehensive DTSX connection analysis
//
// This example demonstrates how to analyze all connection managers in a DTSX package,
// including their drivers, associated variables, expressions, and evaluated values.
// Also includes SQL extraction, execution order analysis, and textual execution flow.
//
// Usage: go run examples/analyze_connections.go <dtsx_file>
//
// Features:
// - Shows connection details (name, type, driver)
// - Lists property expressions and referenced variables
// - Displays variable default values
// - Attempts to evaluate simple expressions by substituting variables
// - Extracts SQL statements from control flow and dataflow tasks
// - Calculates execution order based on precedence constraints
// - Provides textual execution flow description
// - Uses PackageParser for efficient analysis
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/7045kHz/dtsx"
	schema "github.com/7045kHz/dtsx/schemas"
)

type ConnectionAnalysis struct {
	Name        string
	Type        string
	Driver      string
	Properties  map[string]string
	Expressions map[string]string
	Variables   []string
	Evaluated   map[string]string
}

type SQLInfo struct {
	SQL              string
	ConnectionName   string
	ConnectionType   string
	Driver           string
	ConnectionString string
}

type ComponentInfo struct {
	Name                string
	Type                string
	SQL                 string
	ConnectionName      string
	ConnectionType      string
	Driver              string
	ConnectionString    string
	Order               int
	DataflowConnections string
	DataflowFlow        string
}

type CSVRow struct {
	File                string
	ConnectionIndex     int
	ConnectionName      string
	ConnectionType      string
	Driver              string
	PropertyName        string
	PropertyValue       string
	ExpressionProperty  string
	Expression          string
	Variable            string
	VariableValue       string
	EvaluatedProperty   string
	EvaluatedValue      string
	TaskName            string
	SQLStatement        string
	ConnectionString    string
	ExecutionOrder      int
	TaskType            string
	DataflowName        string
	ComponentName       string
	ComponentType       string
	ComponentOrder      int
	DataflowConnections string
	DataflowFlow        string
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run analyze_connections.go <dtsx_file>")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Load the DTSX package
	pkg, err := dtsx.UnmarshalFromFile(filename)
	if err != nil {
		log.Fatalf("Failed to load DTSX file: %v", err)
	}

	// Create parser for advanced analysis
	parser := dtsx.NewPackageParser(pkg)

	// Get all variables for reference
	variables := make(map[string]string)
	if pkg.Variables != nil && pkg.Variables.Variable != nil {
		for _, v := range pkg.Variables.Variable {
			varName := "unnamed"
			if v.ObjectNameAttr != nil {
				varName = *v.ObjectNameAttr
			}
			var namespace string
			if v.NamespaceAttr != nil {
				namespace = *v.NamespaceAttr
			}
			fullName := fmt.Sprintf("%s::%s", namespace, varName)
			if v.VariableValue != nil {
				variables[fullName] = v.VariableValue.Value
			}
		}
	}

	// Analyze connections and collect CSV data
	connections := pkg.GetConnections()
	if connections.Count == 0 {
		fmt.Println("No connection managers found in package.")
		return
	}

	var csvRows []CSVRow
	connMgrs := connections.Results.([]*schema.ConnectionManagerType)

	// Create a map of connection manager refIds to their details
	connMap := make(map[string]*ConnectionAnalysis)
	for i, cm := range connMgrs {
		analysis := analyzeConnection(cm, variables, pkg)
		// Assume refId is something like "Package.ConnectionManagers[NAME]"
		connName := analysis.Name
		refId := fmt.Sprintf("Package.ConnectionManagers[%s]", connName)
		connMap[refId] = analysis
		connMap[connName] = analysis // Also map by name for simplicity

		// Add rows for properties
		for k, v := range analysis.Properties {
			csvRows = append(csvRows, CSVRow{
				File:            filename,
				ConnectionIndex: i + 1,
				ConnectionName:  analysis.Name,
				ConnectionType:  analysis.Type,
				Driver:          analysis.Driver,
				PropertyName:    k,
				PropertyValue:   v,
			})
		}

		// Add rows for expressions
		for prop, expr := range analysis.Expressions {
			csvRows = append(csvRows, CSVRow{
				File:               filename,
				ConnectionIndex:    i + 1,
				ConnectionName:     analysis.Name,
				ConnectionType:     analysis.Type,
				Driver:             analysis.Driver,
				ExpressionProperty: prop,
				Expression:         expr,
			})
		}

		// Add rows for variables
		for _, v := range analysis.Variables {
			varValue := ""
			if value, exists := variables[v]; exists {
				varValue = value
			}
			csvRows = append(csvRows, CSVRow{
				File:            filename,
				ConnectionIndex: i + 1,
				ConnectionName:  analysis.Name,
				ConnectionType:  analysis.Type,
				Driver:          analysis.Driver,
				Variable:        v,
				VariableValue:   varValue,
			})
		}

		// Add rows for evaluated values
		for prop, value := range analysis.Evaluated {
			csvRows = append(csvRows, CSVRow{
				File:              filename,
				ConnectionIndex:   i + 1,
				ConnectionName:    analysis.Name,
				ConnectionType:    analysis.Type,
				Driver:            analysis.Driver,
				EvaluatedProperty: prop,
				EvaluatedValue:    value,
			})
		}

		// If no details, add a row with just connection info
		if len(analysis.Properties) == 0 && len(analysis.Expressions) == 0 && len(analysis.Variables) == 0 && len(analysis.Evaluated) == 0 {
			csvRows = append(csvRows, CSVRow{
				File:            filename,
				ConnectionIndex: i + 1,
				ConnectionName:  analysis.Name,
				ConnectionType:  analysis.Type,
				Driver:          analysis.Driver,
			})
		}
	}

	// Calculate execution order using PrecedenceAnalyzer
	analyzer := dtsx.NewPrecedenceAnalyzer(pkg)
	executionOrder, err := analyzer.GetAllExecutionOrders()
	if err != nil {
		fmt.Printf("Warning: Could not calculate execution order: %v\n", err)
		executionOrder = make(map[string]int)
	}

	// Use parser to get SQL statements
	sqlStatements := parser.GetSQLStatements()

	// Add control flow SQL statements
	for _, stmt := range sqlStatements {
		if stmt.TaskType == "Control Flow" {
			order := 0
			// Try to find execution order for this task
			for _, exec := range pkg.Executable {
				taskName := "Unknown"
				if exec.ObjectNameAttr != nil {
					taskName = *exec.ObjectNameAttr
				}
				if taskName == stmt.TaskName && exec.RefIdAttr != nil {
					if o, exists := executionOrder[*exec.RefIdAttr]; exists {
						order = o
					}
					break
				}
			}

			csvRows = append(csvRows, CSVRow{
				File:           filename,
				TaskName:       stmt.TaskName,
				SQLStatement:   stmt.SQL,
				ExecutionOrder: order,
				TaskType:       stmt.TaskType,
				DataflowName:   "",
			})
		}
	}

	// Extract SQL from dataflow components (for detailed component analysis)
	if pkg.Executable != nil {
		for _, exec := range pkg.Executable {
			taskName := "Unknown"
			if exec.ObjectNameAttr != nil {
				taskName = *exec.ObjectNameAttr
			}

			order := 0
			if exec.RefIdAttr != nil {
				if o, exists := executionOrder[*exec.RefIdAttr]; exists {
					order = o
				}
			}

			// Extract SQL from dataflow components if this is a pipeline
			if exec.ExecutableTypeAttr == "Microsoft.Pipeline" && exec.ObjectData != nil {
				dataflowComponents := extractDataflowComponents(exec.ObjectData, connMap)
				for _, compInfo := range dataflowComponents {
					csvRows = append(csvRows, CSVRow{
						File:                filename,
						ConnectionIndex:     0, // Not applicable for component SQL
						ConnectionName:      compInfo.ConnectionName,
						ConnectionType:      compInfo.ConnectionType,
						Driver:              compInfo.Driver,
						TaskName:            taskName,
						SQLStatement:        compInfo.SQL,
						ConnectionString:    compInfo.ConnectionString,
						ExecutionOrder:      order,
						TaskType:            "Dataflow",
						DataflowName:        taskName,
						ComponentName:       compInfo.Name,
						ComponentType:       compInfo.Type,
						ComponentOrder:      compInfo.Order,
						DataflowConnections: compInfo.DataflowConnections,
						DataflowFlow:        compInfo.DataflowFlow,
					})
				}
			}
		}
	}

	// Write CSV
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	writer.Write([]string{
		"File",
		"ConnectionIndex",
		"ConnectionName",
		"ConnectionType",
		"Driver",
		"PropertyName",
		"PropertyValue",
		"ExpressionProperty",
		"Expression",
		"Variable",
		"VariableValue",
		"EvaluatedProperty",
		"EvaluatedValue",
		"TaskName",
		"SQLStatement",
		"ConnectionString",
		"ExecutionOrder",
		"TaskType",
		"DataflowName",
		"ComponentName",
		"ComponentType",
		"ComponentOrder",
		"DataflowConnections",
		"DataflowFlow",
	})

	// Write rows
	for _, row := range csvRows {
		writer.Write([]string{
			row.File,
			fmt.Sprintf("%d", row.ConnectionIndex),
			row.ConnectionName,
			row.ConnectionType,
			row.Driver,
			row.PropertyName,
			row.PropertyValue,
			row.ExpressionProperty,
			row.Expression,
			row.Variable,
			row.VariableValue,
			row.EvaluatedProperty,
			row.EvaluatedValue,
			row.TaskName,
			row.SQLStatement,
			row.ConnectionString,
			fmt.Sprintf("%d", row.ExecutionOrder),
			row.TaskType,
			row.DataflowName,
			row.ComponentName,
			row.ComponentType,
			fmt.Sprintf("%d", row.ComponentOrder),
			row.DataflowConnections,
			row.DataflowFlow,
		})
	}

	// Print execution flow description
	printExecutionFlow(pkg, parser)
}

func analyzeConnection(cm *schema.ConnectionManagerType, variables map[string]string, pkg *dtsx.Package) *ConnectionAnalysis {
	analysis := &ConnectionAnalysis{
		Name:        "Unknown",
		Type:        "Unknown",
		Driver:      "Unknown",
		Properties:  make(map[string]string),
		Expressions: make(map[string]string),
		Evaluated:   make(map[string]string),
	}

	// Extract attributes
	if cm.ObjectNameAttr != nil {
		analysis.Name = *cm.ObjectNameAttr
	}
	if cm.CreationNameAttr != nil {
		analysis.Driver = *cm.CreationNameAttr
		analysis.Type = getConnectionType(*cm.CreationNameAttr)
	}

	// Extract properties
	if cm.Property != nil {
		for _, prop := range cm.Property {
			if prop.NameAttr != nil && prop.PropertyElementBaseType != nil &&
				prop.PropertyElementBaseType.AnySimpleType != nil {
				name := *prop.NameAttr
				value := prop.PropertyElementBaseType.AnySimpleType.Value
				analysis.Properties[name] = value
			}
		}
	}

	// Extract expressions
	if cm.PropertyExpression != nil {
		for _, expr := range cm.PropertyExpression {
			if expr.NameAttr != "" && expr.AnySimpleType != nil {
				propName := expr.NameAttr
				expression := expr.AnySimpleType.Value
				analysis.Expressions[propName] = expression

				// Extract variable references
				vars := extractVariables(expression)
				analysis.Variables = append(analysis.Variables, vars...)

				// Try to evaluate expressions using the advanced evaluator
				if evaluated := evaluateExpressionAdvanced(expression, pkg); evaluated != "" {
					analysis.Evaluated[propName] = evaluated
				}
			}
		}
	}

	return analysis
}

func getConnectionType(creationName string) string {
	switch strings.ToUpper(creationName) {
	case "OLEDB":
		return "OLE DB Database"
	case "FLATFILE":
		return "Flat File"
	case "ADO.NET":
		return "ADO.NET Database"
	case "EXCEL":
		return "Excel File"
	case "HTTP":
		return "HTTP Connection"
	case "FTP":
		return "FTP Connection"
	case "SMTP":
		return "SMTP Connection"
	case "MSMQ":
		return "MSMQ Connection"
	case "WMI":
		return "WMI Connection"
	default:
		return creationName
	}
}

func extractVariables(expression string) []string {
	// Simple regex to find @[Namespace::Variable] patterns
	re := regexp.MustCompile(`@\[([^\]]+)\]`)
	matches := re.FindAllStringSubmatch(expression, -1)

	var vars []string
	for _, match := range matches {
		if len(match) > 1 {
			vars = append(vars, match[1])
		}
	}
	return vars
}

func evaluateExpressionAdvanced(expression string, pkg *dtsx.Package) string {
	// Use the advanced SSIS expression evaluator
	result, err := dtsx.EvaluateExpression(expression, pkg)
	if err != nil {
		// Return empty string if evaluation fails
		return ""
	}

	// Convert result to string
	switch v := result.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v) // Remove decimal for integers
	default:
		return fmt.Sprintf("%v", v)
	}
}

func extractDataflowComponents(objectData *schema.ExecutableObjectDataType, connMap map[string]*ConnectionAnalysis) []ComponentInfo {
	var compInfos []ComponentInfo

	if objectData.Pipeline == nil || objectData.Pipeline.Components == nil {
		return compInfos
	}

	// Collect unique connections used in the dataflow
	dataflowConnSet := make(map[string]bool)

	// Build component map by name (assuming ids are names)
	compMap := make(map[string]*schema.PipelineComponentType)
	for _, comp := range objectData.Pipeline.Components.Component {
		if comp.NameAttr != nil {
			compMap[*comp.NameAttr] = comp
		}
	}

	// Build adjacency list for paths (precedence in dataflow)
	graph := make(map[string][]string)
	if objectData.Pipeline.Paths != nil {
		for _, path := range objectData.Pipeline.Paths.Path {
			if path.StartIdAttr != nil && path.EndIdAttr != nil {
				graph[*path.StartIdAttr] = append(graph[*path.StartIdAttr], *path.EndIdAttr)
			}
		}
	}

	// Topological sort to determine execution order
	order := topologicalSort(graph, compMap)
	// Sort order by component type priority, then name for deterministic output
	typePriority := map[string]int{
		"Microsoft.OLEDBSource":         1,
		"Microsoft.DataConvert":         2,
		"Microsoft.FlatFileDestination": 3,
	}
	sort.Slice(order, func(i, j int) bool {
		compI, existsI := compMap[order[i]]
		compJ, existsJ := compMap[order[j]]
		priI := 999
		if existsI && compI.ComponentClassIDAttr != nil {
			if p, ok := typePriority[*compI.ComponentClassIDAttr]; ok {
				priI = p
			}
		}
		priJ := 999
		if existsJ && compJ.ComponentClassIDAttr != nil {
			if p, ok := typePriority[*compJ.ComponentClassIDAttr]; ok {
				priJ = p
			}
		}
		if priI != priJ {
			return priI < priJ
		}
		return order[i] < order[j]
	})
	orderMap := make(map[string]int)
	for i, id := range order {
		orderMap[id] = i + 1
	}

	// Build dataflow flow string
	var flowNames []string
	for _, id := range order {
		if comp, exists := compMap[id]; exists {
			if comp.NameAttr != nil {
				flowNames = append(flowNames, *comp.NameAttr)
			}
		}
	}
	dataflowFlowStr := strings.Join(flowNames, " -> ")

	// Process each component
	for _, comp := range objectData.Pipeline.Components.Component {
		compID := "Unknown"
		if comp.NameAttr != nil {
			compID = *comp.NameAttr
		}
		compName := "Unknown"
		if comp.NameAttr != nil {
			compName = *comp.NameAttr
		}
		compType := "Unknown"
		if comp.ComponentClassIDAttr != nil {
			compType = *comp.ComponentClassIDAttr
		}

		compOrder := 0
		if o, exists := orderMap[compID]; exists {
			compOrder = o
		}

		compInfo := ComponentInfo{
			Name:  compName,
			Type:  compType,
			Order: compOrder,
		}

		// Find connection
		if comp.Connections != nil {
			for _, conn := range comp.Connections.Connection {
				if conn.ConnectionManagerIDAttr != nil {
					if ca, exists := connMap[*conn.ConnectionManagerIDAttr]; exists {
						compInfo.ConnectionName = ca.Name
						compInfo.ConnectionType = ca.Type
						compInfo.Driver = ca.Driver
						dataflowConnSet[ca.Name] = true // Add to dataflow connections
						if connStr, exists := ca.Evaluated["ConnectionString"]; exists && connStr != "" {
							compInfo.ConnectionString = connStr
						} else if connStr, exists := ca.Properties["ConnectionString"]; exists {
							compInfo.ConnectionString = connStr
						}
						break
					}
				}
			}
		}

		// Extract SQL if present
		if comp.Properties != nil {
			for _, prop := range comp.Properties.Property {
				if prop.NameAttr == nil {
					continue
				}
				propName := *prop.NameAttr
				if propName == "SqlCommand" || propName == "SqlStatement" || propName == "CommandText" ||
					propName == "Query" || propName == "SelectQuery" || propName == "InsertQuery" ||
					propName == "UpdateQuery" || propName == "DeleteQuery" || propName == "OpenRowset" {
					sql := strings.TrimSpace(prop.Value)
					if sql != "" {
						if propName == "OpenRowset" {
							sql = "SELECT * FROM " + sql
						}
						compInfo.SQL = sql
						break // Assume one SQL per component
					}
				}
			}
		}

		compInfos = append(compInfos, compInfo)
	}

	// Build dataflow connections string
	var dataflowConns []string
	for conn := range dataflowConnSet {
		dataflowConns = append(dataflowConns, conn)
	}
	dataflowConnStr := strings.Join(dataflowConns, ";")

	// Set DataflowConnections and DataflowFlow for each component
	for i := range compInfos {
		compInfos[i].DataflowConnections = dataflowConnStr
		compInfos[i].DataflowFlow = dataflowFlowStr
	}

	return compInfos
}

func topologicalSort(graph map[string][]string, compMap map[string]*schema.PipelineComponentType) []string {
	// Kahn's algorithm
	inDegree := make(map[string]int)
	for _, neighbors := range graph {
		for _, n := range neighbors {
			inDegree[n]++
		}
	}
	// All components with inDegree 0
	var queue []string
	for id := range compMap {
		if inDegree[id] == 0 {
			queue = append(queue, id)
		}
	}
	// Sort queue by component type priority, then name for deterministic order
	typePriority := map[string]int{
		"Microsoft.OLEDBSource":         1,
		"Microsoft.DataConvert":         2,
		"Microsoft.FlatFileDestination": 3,
	}
	sort.Slice(queue, func(i, j int) bool {
		compI, existsI := compMap[queue[i]]
		compJ, existsJ := compMap[queue[j]]
		priI := 999
		if existsI && compI.ComponentClassIDAttr != nil {
			if p, ok := typePriority[*compI.ComponentClassIDAttr]; ok {
				priI = p
			}
		}
		priJ := 999
		if existsJ && compJ.ComponentClassIDAttr != nil {
			if p, ok := typePriority[*compJ.ComponentClassIDAttr]; ok {
				priJ = p
			}
		}
		if priI != priJ {
			return priI < priJ
		}
		return queue[i] < queue[j]
	})
	var order []string
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		order = append(order, current)
		for _, neighbor := range graph[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}
	// If not all components are in order, there might be cycles, but for now, return order
	return order
}

func printExecutionFlow(pkg *dtsx.Package, parser *dtsx.PackageParser) {
	fmt.Println("\n=== Execution Flow Description ===")

	if len(pkg.Executable) == 0 {
		fmt.Println("No executables found in package.")
		return
	}

	// Create precedence analyzer for execution order
	analyzer := dtsx.NewPrecedenceAnalyzer(pkg)

	for _, exec := range pkg.Executable {
		taskName := "Unknown"
		if exec.ObjectNameAttr != nil {
			taskName = *exec.ObjectNameAttr
		}

		order := 1
		if exec.RefIdAttr != nil {
			if o, err := analyzer.GetExecutionOrder(*exec.RefIdAttr); err == nil {
				order = o
			}
		}

		fmt.Printf("Task %d: %s", order, taskName)

		if exec.ExecutableTypeAttr != "" {
			fmt.Printf(" (%s)", exec.ExecutableTypeAttr)
		}
		fmt.Println()

		// If it's a dataflow, show component execution order
		if exec.ExecutableTypeAttr == "Microsoft.Pipeline" && exec.ObjectData != nil {
			printDataflowExecutionFlow(exec.ObjectData, taskName)
		}
	}
}

func printDataflowExecutionFlow(objectData *schema.ExecutableObjectDataType, dataflowName string) {
	if objectData.Pipeline == nil || objectData.Pipeline.Components == nil {
		return
	}

	// Build component map
	compMap := make(map[string]*schema.PipelineComponentType)
	for _, comp := range objectData.Pipeline.Components.Component {
		if comp.NameAttr != nil {
			compMap[*comp.NameAttr] = comp
		}
	}

	// Build graph for topological sort
	graph := make(map[string][]string)
	if objectData.Pipeline.Paths != nil {
		for _, path := range objectData.Pipeline.Paths.Path {
			if path.StartIdAttr != nil && path.EndIdAttr != nil {
				graph[*path.StartIdAttr] = append(graph[*path.StartIdAttr], *path.EndIdAttr)
			}
		}
	}

	// Get execution order
	order := topologicalSort(graph, compMap)

	fmt.Printf("  Dataflow '%s' execution order:\n", dataflowName)
	for i, compID := range order {
		if comp, exists := compMap[compID]; exists {
			compName := compID
			compType := "Unknown"
			if comp.ComponentClassIDAttr != nil {
				compType = *comp.ComponentClassIDAttr
			}

			// Extract SQL if present
			var sql string
			if comp.Properties != nil {
				for _, prop := range comp.Properties.Property {
					if prop.NameAttr != nil {
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
			}

			fmt.Printf("    Component %d: %s (%s)", i+1, compName, compType)
			if sql != "" {
				fmt.Printf(" - SQL: %s", sql)
			}
			fmt.Println()
		}
	}
}
