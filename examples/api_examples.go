//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/7045kHz/dtsx"
	schema "github.com/7045kHz/dtsx/schemas"
)

func main() {
	// Build a minimal package using the builder
	pkg := dtsx.NewPackageBuilder().
		AddVariable("User", "Count", "5").
		AddConnection("SourceDB", "OLEDB", "Server=localhost;Database=Example;").
		AddConnectionExpression("SourceDB", "ConnectionString", "@[User::Count]").
		Build()

	fmt.Println("Package built")

	// Marshal and Unmarshal (parse round-trip)
	b, err := dtsx.Marshal(pkg)
	if err != nil {
		log.Fatalf("Marshal failed: %v", err)
	}
	fmt.Printf("Marshaled %d bytes\n", len(b))

	pkg2, err := dtsx.Unmarshal(b)
	if err != nil {
		log.Fatalf("Unmarshal failed: %v", err)
	}
	fmt.Println("Unmarshal round-trip OK")

	// Create a PackageParser for evaluation & extraction
	pp := dtsx.NewPackageParser(pkg2)

	// Get variable value via PackageParser
	v, err := pp.GetVariableValue("User::Count")
	if err != nil {
		fmt.Printf("GetVariableValue error: %v\n", err)
	} else {
		fmt.Printf("Variable User::Count = %v\n", v)
	}

	// Evaluate expression using PackageParser (cached)
	res, err := pp.EvaluateExpression("@[User::Count] + 1")
	if err != nil {
		fmt.Printf("EvaluateExpression error: %v\n", err)
	} else {
		fmt.Printf("@[User::Count] + 1 => %v\n", res)
	}

	// Evaluate expression using package-level helper
	res2, err := dtsx.EvaluateExpression("@[User::Count] + 2", pkg2)
	if err != nil {
		fmt.Printf("EvaluateExpression (pkg-level) error: %v\n", err)
	} else {
		fmt.Printf("(pkg-level) @[User::Count] + 2 => %v\n", res2)
	}

	// Query connections/variables
	conns := pkg2.GetConnections()
	fmt.Printf("Connections: %d\n", conns.Count)
	if conns.Count > 0 {
		cm := conns.Results.([]*schema.ConnectionManagerType)[0]
		fmt.Printf("Connection name: %s\n", dtsx.GetConnectionName(cm))
		fmt.Printf("Connection string: %s\n", dtsx.GetConnectionString(cm))
	}

	vars := pkg2.GetVariables()
	fmt.Printf("Variables: %d\n", vars.Count)
	if vars.Count > 0 {
		vv := vars.Results.([]*schema.VariableType)[0]
		fmt.Printf("Variable name: %s value: %s\n", dtsx.GetVariableName(vv), dtsx.GetVariableValue(vv))
	}

	// GetExpressions: we added a connection expression above, so it should appear
	exprs := pkg2.GetExpressions()
	fmt.Printf("Expressions found: %d\n", exprs.Count)
	if exprs.Count > 0 {
		exprList := exprs.Results.([]*dtsx.ExpressionInfo)
		for _, e := range exprList {
			fmt.Printf("Expression: %s (Location=%s)\n", e.Expression, e.Location)
		}
	}

	// Get SQL statements (none expected in this minimal package)
	stmts := pp.GetSQLStatements()
	fmt.Printf("SQL statements found: %d\n", len(stmts))

	// Expression AST usage
	lit := &dtsx.Literal{Value: 42.0}
	if r, err := lit.Eval(nil); err == nil {
		fmt.Printf("Literal Eval => %v\n", r)
	}
	varNode := &dtsx.Variable{Name: "User::Count"}
	varsMap := map[string]interface{}{"User::Count": 10.0}
	if r, err := varNode.Eval(varsMap); err == nil {
		fmt.Printf("Variable Eval => %v\n", r)
	}

	// Build dependency graph
	dg := pkg2.BuildDependencyGraph()
	impact := dg.GetVariableImpact("User::Count")
	fmt.Printf("Variable impact for User::Count: %v\n", impact)

	// Get unused variables
	unused := pkg2.GetUnusedVariables()
	fmt.Printf("Unused variables: %v\n", unused)

	// Optimization suggestions
	suggestions := pkg2.GetOptimizationSuggestions()
	fmt.Printf("Optimization suggestions: %d\n", len(suggestions))

	// Precedence analyzer
	pa := dtsx.NewPrecedenceAnalyzer(pkg2)
	flow := pa.GetExecutionFlowDescription()
	fmt.Println(flow)

	// Validator
	validator := dtsx.NewPackageValidator(pkg2)
	issues := validator.Validate()
	fmt.Printf("Validation issues: %d\n", len(issues))

	// Show use of GetProperty helper
	if conns.Count > 0 {
		cm := conns.Results.([]*schema.ConnectionManagerType)[0]
		objName := dtsx.GetProperty(cm, "ObjectNameAttr")
		fmt.Printf("GetProperty(ObjectNameAttr) => %v\n", objName)
	}

	// Demonstrate calling RunPackage (commented out to avoid external execution)
	if false {
		out, err := dtsx.RunPackage("C:\\Program Files\\Microsoft SQL Server\\130\\DTS\\Binn\\DTExec.exe", "mypackage.dtsx", &dtsx.RunOptions{Validate: true})
		fmt.Printf("RunPackage output: %s err: %v\n", out, err)
	}

	// Write a packed DTSX to a temp file as demonstration (write in current directory)
	tmpFile := "./output_sample.dtsx"
	if err := os.WriteFile(tmpFile, b, 0644); err == nil {
		fmt.Printf("Wrote sample DTSX to %s\n", tmpFile)
	} else {
		fmt.Printf("Failed to write sample DTSX: %v\n", err)
	}
}
