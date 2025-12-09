package dtsx_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/7045kHz/dtsx"
	schema "github.com/7045kHz/dtsx/schemas"
)

const dtexecPath = `C:\Program Files\Microsoft SQL Server\160\DTS\Binn\DTExec.exe`

func TestUnmarshalFromFile(t *testing.T) {
	// This test requires a sample DTSX file
	// Skip if SSIS_EXAMPLES directory doesn't have files
	files, err := os.ReadDir("SSIS_EXAMPLES")
	if err != nil || len(files) == 0 {
		t.Skip("No DTSX example files found in SSIS_EXAMPLES directory")
	}

	// Find first .dtsx file
	var dtsxFile string
	for _, file := range files {
		if !file.IsDir() && len(file.Name()) > 5 && file.Name()[len(file.Name())-5:] == ".dtsx" {
			dtsxFile = "SSIS_EXAMPLES/" + file.Name()
			break
		}
	}

	if dtsxFile == "" {
		t.Skip("No .dtsx files found in SSIS_EXAMPLES directory")
	}

	pkg, err := dtsx.UnmarshalFromFile(dtsxFile)
	if err != nil {
		t.Fatalf("Failed to unmarshal DTSX file %s: %v", dtsxFile, err)
	}

	if pkg == nil {
		t.Fatal("Expected package, got nil")
	}

	t.Logf("Successfully parsed DTSX file: %s", dtsxFile)
}

func TestMarshalUnmarshal(t *testing.T) {
	// Create a simple package
	pkg := &dtsx.Package{
		ExecutableTypePackage: &schema.ExecutableTypePackage{},
	}

	// Marshal to XML
	data, err := dtsx.Marshal(pkg)
	if err != nil {
		t.Fatalf("Failed to marshal package: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Expected XML data, got empty bytes")
	}

	// Unmarshal back
	pkg2, err := dtsx.Unmarshal(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal package: %v", err)
	}

	if pkg2 == nil {
		t.Fatal("Expected package, got nil")
	}

	t.Logf("Successfully marshaled and unmarshaled package")
}

func TestIsDTSXPackage(t *testing.T) {
	// Test with a valid DTSX file
	files, err := os.ReadDir("SSIS_EXAMPLES")
	if err != nil || len(files) == 0 {
		t.Skip("No DTSX example files found in SSIS_EXAMPLES directory")
	}

	// Find first .dtsx file
	var dtsxFile string
	for _, file := range files {
		if !file.IsDir() && len(file.Name()) > 5 && file.Name()[len(file.Name())-5:] == ".dtsx" {
			dtsxFile = "SSIS_EXAMPLES/" + file.Name()
			break
		}
	}

	if dtsxFile == "" {
		t.Skip("No .dtsx files found in SSIS_EXAMPLES directory")
	}

	pkg, ok := dtsx.IsDTSXPackage(dtsxFile)
	if !ok {
		t.Fatalf("Expected %s to be a valid DTSX package", dtsxFile)
	}
	if pkg == nil {
		t.Fatal("Expected package to be non-nil when ok is true")
	}

	// Test with invalid file
	pkg, ok = dtsx.IsDTSXPackage("nonexistent.dtsx")
	if ok {
		t.Fatal("Expected nonexistent file to return false")
	}
	if pkg != nil {
		t.Fatal("Expected nil package when ok is false")
	}

	t.Logf("IsDTSXPackage validation works correctly")
}

func TestRunPackage(t *testing.T) {
	// Check if dtexec exists
	if _, err := os.Stat(dtexecPath); os.IsNotExist(err) {
		t.Skipf("DTExec.exe not found at %s", dtexecPath)
	}

	// Find a test DTSX file
	files, err := os.ReadDir("SSIS_EXAMPLES")
	if err != nil || len(files) == 0 {
		t.Skip("No DTSX example files found in SSIS_EXAMPLES directory")
	}

	var dtsxFile string
	for _, file := range files {
		if !file.IsDir() && len(file.Name()) > 5 && file.Name()[len(file.Name())-5:] == ".dtsx" {
			dtsxFile = filepath.Join("SSIS_EXAMPLES", file.Name())
			break
		}
	}

	if dtsxFile == "" {
		t.Skip("No .dtsx files found in SSIS_EXAMPLES directory")
	}

	// Get absolute path
	absPath, err := filepath.Abs(dtsxFile)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	t.Run("ValidateOnly", func(t *testing.T) {
		opts := &dtsx.RunOptions{
			Validate:       true,
			ReportingLevel: "V",
		}

		output, err := dtsx.RunPackage(dtexecPath, absPath, opts)
		// Validation may succeed or fail depending on package dependencies
		// Just ensure we get some output
		if output == "" && err == nil {
			t.Log("Package validation completed with no output")
		} else if err != nil {
			t.Logf("Validation result (may fail due to missing connections): %v", err)
			t.Logf("Output: %s", output)
		} else {
			t.Logf("Validation output: %s", output)
		}
	})

	t.Run("WithOptions", func(t *testing.T) {
		opts := &dtsx.RunOptions{
			Validate:       true,
			WarnAsError:    false,
			ReportingLevel: "E",
		}

		output, err := dtsx.RunPackage(dtexecPath, absPath, opts)
		if err != nil {
			t.Logf("Execution with options (expected to potentially fail): %v", err)
		}
		t.Logf("Output: %s", output)
	})

	t.Run("NilOptions", func(t *testing.T) {
		// Test with nil options
		output, err := dtsx.RunPackage(dtexecPath, absPath, nil)
		if err != nil {
			t.Logf("Execution with nil options (expected to potentially fail): %v", err)
		}
		t.Logf("Output: %s", output)
	})
}

func TestGetConnections(t *testing.T) {
	// Test with a valid DTSX file
	files, err := os.ReadDir("SSIS_EXAMPLES")
	if err != nil || len(files) == 0 {
		t.Skip("No DTSX example files found in SSIS_EXAMPLES directory")
	}

	var dtsxFile string
	for _, file := range files {
		if !file.IsDir() && len(file.Name()) > 5 && file.Name()[len(file.Name())-5:] == ".dtsx" {
			dtsxFile = "SSIS_EXAMPLES/" + file.Name()
			break
		}
	}

	if dtsxFile == "" {
		t.Skip("No .dtsx files found in SSIS_EXAMPLES directory")
	}

	pkg, err := dtsx.UnmarshalFromFile(dtsxFile)
	if err != nil {
		t.Fatalf("Failed to unmarshal DTSX file: %v", err)
	}

	result := pkg.GetConnections()
	if result == nil {
		t.Fatal("Expected QueryResult, got nil")
	}

	connections, ok := result.Results.([]*schema.ConnectionManagerType)
	if !ok {
		t.Fatal("Expected []*schema.ConnectionManagerType, got different type")
	}

	if result.Count != len(connections) {
		t.Fatalf("Count mismatch: expected %d, got %d", len(connections), result.Count)
	}

	t.Logf("Found %d connections in package", result.Count)
}

func TestGetVariables(t *testing.T) {
	// Test with a valid DTSX file
	files, err := os.ReadDir("SSIS_EXAMPLES")
	if err != nil || len(files) == 0 {
		t.Skip("No DTSX example files found in SSIS_EXAMPLES directory")
	}

	var dtsxFile string
	for _, file := range files {
		if !file.IsDir() && len(file.Name()) > 5 && file.Name()[len(file.Name())-5:] == ".dtsx" {
			dtsxFile = "SSIS_EXAMPLES/" + file.Name()
			break
		}
	}

	if dtsxFile == "" {
		t.Skip("No .dtsx files found in SSIS_EXAMPLES directory")
	}

	pkg, err := dtsx.UnmarshalFromFile(dtsxFile)
	if err != nil {
		t.Fatalf("Failed to unmarshal DTSX file: %v", err)
	}

	result := pkg.GetVariables()
	if result == nil {
		t.Fatal("Expected QueryResult, got nil")
	}

	variables, ok := result.Results.([]*schema.VariableType)
	if !ok {
		t.Fatal("Expected []*schema.VariableType, got different type")
	}

	if result.Count != len(variables) {
		t.Fatalf("Count mismatch: expected %d, got %d", len(variables), result.Count)
	}

	t.Logf("Found %d variables in package", result.Count)
}

func TestGetVariableByName(t *testing.T) {
	// Test with a valid DTSX file
	files, err := os.ReadDir("SSIS_EXAMPLES")
	if err != nil || len(files) == 0 {
		t.Skip("No DTSX example files found in SSIS_EXAMPLES directory")
	}

	var dtsxFile string
	for _, file := range files {
		if !file.IsDir() && len(file.Name()) > 5 && file.Name()[len(file.Name())-5:] == ".dtsx" {
			dtsxFile = "SSIS_EXAMPLES/" + file.Name()
			break
		}
	}

	if dtsxFile == "" {
		t.Skip("No .dtsx files found in SSIS_EXAMPLES directory")
	}

	pkg, err := dtsx.UnmarshalFromFile(dtsxFile)
	if err != nil {
		t.Fatalf("Failed to unmarshal DTSX file: %v", err)
	}

	// Test with nil package
	var nilPkg *dtsx.Package
	_, err = nilPkg.GetVariableByName("test")
	if err == nil {
		t.Fatal("Expected error for nil package, got nil")
	}

	// Test with non-existent variable
	_, err = pkg.GetVariableByName("NonExistentVariable")
	if err == nil {
		t.Fatal("Expected error for non-existent variable, got nil")
	}

	// Get all variables and try to find one by name
	variables := pkg.GetVariables()
	if variables.Count > 0 {
		vars := variables.Results.([]*schema.VariableType)
		// Try to find the name of the first variable
		for _, v := range vars {
			for _, prop := range v.Property {
				if prop.NameAttr != nil && *prop.NameAttr == "ObjectName" {
					if prop.PropertyElementBaseType != nil && prop.PropertyElementBaseType.AnySimpleType != nil {
						varName := prop.PropertyElementBaseType.AnySimpleType.Value
						foundVar, err := pkg.GetVariableByName(varName)
						if err != nil {
							t.Fatalf("Failed to find variable %s: %v", varName, err)
						}
						if foundVar != v {
							t.Fatal("Found variable does not match expected variable")
						}
						t.Logf("Successfully found variable: %s", varName)
						return // Test passed
					}
				}
			}
		}
	}

	t.Log("No variables with ObjectName found to test")
}

func TestQueryExecutables(t *testing.T) {
	// Test with a valid DTSX file
	files, err := os.ReadDir("SSIS_EXAMPLES")
	if err != nil || len(files) == 0 {
		t.Skip("No DTSX example files found in SSIS_EXAMPLES directory")
	}

	var dtsxFile string
	for _, file := range files {
		if !file.IsDir() && len(file.Name()) > 5 && file.Name()[len(file.Name())-5:] == ".dtsx" {
			dtsxFile = "SSIS_EXAMPLES/" + file.Name()
			break
		}
	}

	if dtsxFile == "" {
		t.Skip("No .dtsx files found in SSIS_EXAMPLES directory")
	}

	pkg, err := dtsx.UnmarshalFromFile(dtsxFile)
	if err != nil {
		t.Fatalf("Failed to unmarshal DTSX file: %v", err)
	}

	// Test with nil package
	var nilPkg *dtsx.Package
	results := nilPkg.QueryExecutables(func(*schema.AnyNonPackageExecutableType) bool { return true })
	if len(results) != 0 {
		t.Fatal("Expected empty results for nil package")
	}

	// Test with filter that matches all
	allExecutables := pkg.QueryExecutables(func(*schema.AnyNonPackageExecutableType) bool { return true })
	t.Logf("Found %d executables total", len(allExecutables))

	// Test with filter that matches none
	noExecutables := pkg.QueryExecutables(func(*schema.AnyNonPackageExecutableType) bool { return false })
	if len(noExecutables) != 0 {
		t.Fatal("Expected no executables with false filter")
	}

	// Test with specific type filter (if any SQL tasks exist)
	sqlTasks := pkg.QueryExecutables(func(exec *schema.AnyNonPackageExecutableType) bool {
		return exec.ExecutableTypeAttr == "STOCK:SQLTask"
	})
	t.Logf("Found %d SQL tasks", len(sqlTasks))

	// Verify all returned executables match the filter
	for _, exec := range sqlTasks {
		if exec.ExecutableTypeAttr != "STOCK:SQLTask" {
			t.Fatal("QueryExecutables returned executable that doesn't match filter")
		}
	}
}

func TestGetExpressions(t *testing.T) {
	// Test with a valid DTSX file
	files, err := os.ReadDir("SSIS_EXAMPLES")
	if err != nil || len(files) == 0 {
		t.Skip("No DTSX example files found in SSIS_EXAMPLES directory")
	}

	var dtsxFile string
	for _, file := range files {
		if !file.IsDir() && len(file.Name()) > 5 && file.Name()[len(file.Name())-5:] == ".dtsx" {
			dtsxFile = "SSIS_EXAMPLES/" + file.Name()
			break
		}
	}

	if dtsxFile == "" {
		t.Skip("No .dtsx files found in SSIS_EXAMPLES directory")
	}

	pkg, err := dtsx.UnmarshalFromFile(dtsxFile)
	if err != nil {
		t.Fatalf("Failed to unmarshal DTSX file: %v", err)
	}

	result := pkg.GetExpressions()
	if result == nil {
		t.Fatal("Expected QueryResult, got nil")
	}

	expressions, ok := result.Results.([]*dtsx.ExpressionInfo)
	if !ok {
		t.Fatal("Expected []*dtsx.ExpressionInfo, got different type")
	}

	if result.Count != len(expressions) {
		t.Fatalf("Count mismatch: expected %d, got %d", len(expressions), result.Count)
	}

	t.Logf("Found %d expressions in package", result.Count)

	// Verify expression info structure
	for i, expr := range expressions {
		if expr.Expression == "" {
			t.Errorf("Expression %d has empty expression string", i)
		}
		if expr.Location == "" {
			t.Errorf("Expression %d has empty location", i)
		}
		t.Logf("Expression %d: %s at %s (%s)", i+1, expr.Expression, expr.Location, expr.Context)
	}

	// Test with nil package
	var nilPkg *dtsx.Package
	nilResult := nilPkg.GetExpressions()
	if nilResult == nil {
		t.Fatal("Expected QueryResult for nil package, got nil")
	}
	if nilResult.Count != 0 {
		t.Errorf("Expected 0 expressions for nil package, got %d", nilResult.Count)
	}
}

func TestEvaluateExpression(t *testing.T) {
	// Create a package with variables
	pkg := &dtsx.Package{
		ExecutableTypePackage: &schema.ExecutableTypePackage{
			Variables: &schema.VariablesType{
				Variable: []*schema.VariableType{
					{
						NamespaceAttr:  stringPtr("User"),
						ObjectNameAttr: stringPtr("MyVar"),
						VariableValue: &schema.VariableValue{
							Value: "42",
						},
					},
					{
						NamespaceAttr:  stringPtr("User"),
						ObjectNameAttr: stringPtr("StrVar"),
						VariableValue: &schema.VariableValue{
							Value: "hello",
						},
					},
				},
			},
		},
	}

	// Test literal
	result, err := dtsx.EvaluateExpression("123", pkg)
	if err != nil {
		t.Errorf("Failed to evaluate literal: %v", err)
	}
	if result != 123.0 {
		t.Errorf("Expected 123.0, got %v", result)
	}

	// Test variable
	result, err = dtsx.EvaluateExpression("@[User::MyVar]", pkg)
	if err != nil {
		t.Errorf("Failed to evaluate variable: %v", err)
	}
	if result != 42.0 {
		t.Errorf("Expected 42.0, got %v", result)
	}

	// Test string variable
	result, err = dtsx.EvaluateExpression("@[User::StrVar]", pkg)
	if err != nil {
		t.Errorf("Failed to evaluate string variable: %v", err)
	}
	if result != "hello" {
		t.Errorf("Expected 'hello', got %v", result)
	}

	// Test arithmetic
	result, err = dtsx.EvaluateExpression("@[User::MyVar] + 1", pkg)
	if err != nil {
		t.Errorf("Failed to evaluate arithmetic: %v", err)
	}
	if result != 43.0 {
		t.Errorf("Expected 43.0, got %v", result)
	}

	// Test string concatenation
	result, err = dtsx.EvaluateExpression("@[User::StrVar] + \" world\"", pkg)
	if err != nil {
		t.Errorf("Failed to evaluate concatenation: %v", err)
	}
	if result != "hello world" {
		t.Errorf("Expected 'hello world', got %v", result)
	}
}

func stringPtr(s string) *string {
	return &s
}

func TestPackageBuilder(t *testing.T) {
	builder := dtsx.NewPackageBuilder()

	// Add variables
	builder.AddVariable("User", "Var1", "value1")
	builder.AddVariable("System", "Var2", "42")

	// Add connections
	builder.AddConnection("Conn1", "OLEDB", "Server=myServer;Database=myDB")
	builder.AddConnectionExpression("Conn1", "ConnectionString", "@[User::ConnVar]")

	pkg := builder.Build()

	if pkg == nil {
		t.Fatal("Expected package, got nil")
	}

	// Check variables
	vars := pkg.GetVariables()
	if vars.Count != 2 {
		t.Errorf("Expected 2 variables, got %d", vars.Count)
	}

	// Check connections
	conns := pkg.GetConnections()
	if conns.Count != 1 {
		t.Errorf("Expected 1 connection, got %d", conns.Count)
	}

	// Check connection expressions
	connMgr := pkg.ConnectionManagers.ConnectionManager[0]
	if len(connMgr.PropertyExpression) != 1 {
		t.Errorf("Expected 1 connection expression, got %d", len(connMgr.PropertyExpression))
	}
	if connMgr.PropertyExpression[0].NameAttr != "ConnectionString" {
		t.Errorf("Expected expression name 'ConnectionString', got '%s'", connMgr.PropertyExpression[0].NameAttr)
	}
	if connMgr.PropertyExpression[0].AnySimpleType.Value != "@[User::ConnVar]" {
		t.Errorf("Expected expression value '@[User::ConnVar]', got '%s'", connMgr.PropertyExpression[0].AnySimpleType.Value)
	}
}

func TestPackageBuilderWithTypes(t *testing.T) {
	builder := dtsx.NewPackageBuilder()

	// Add variables with different data types
	builder.AddVariableWithType("User", "StringVar", "hello", "String")
	builder.AddVariableWithType("User", "IntVar", "42", "Int32")
	builder.AddVariableWithType("User", "BoolVar", "true", "Boolean")
	builder.AddVariableWithType("User", "DateVar", "2025-01-01", "DateTime")

	pkg := builder.Build()

	if pkg == nil {
		t.Fatal("Expected package, got nil")
	}

	// Check variables
	vars := pkg.GetVariables()
	if vars.Count != 4 {
		t.Errorf("Expected 4 variables, got %d", vars.Count)
	}

	variables := vars.Results.([]*schema.VariableType)

	// Check String variable
	stringVar := variables[0]
	if stringVar.VariableValue.Value != "hello" {
		t.Errorf("Expected string value 'hello', got '%s'", stringVar.VariableValue.Value)
	}
	if stringVar.VariableValue.DataTypeAttr == nil || *stringVar.VariableValue.DataTypeAttr != 8 {
		t.Errorf("Expected data type 8 (DT_WSTR) for string, got %v", stringVar.VariableValue.DataTypeAttr)
	}

	// Check Int32 variable
	intVar := variables[1]
	if intVar.VariableValue.Value != "42" {
		t.Errorf("Expected int value '42', got '%s'", intVar.VariableValue.Value)
	}
	if intVar.VariableValue.DataTypeAttr == nil || *intVar.VariableValue.DataTypeAttr != 3 {
		t.Errorf("Expected data type 3 (DT_I4) for int32, got %v", intVar.VariableValue.DataTypeAttr)
	}

	// Check Boolean variable
	boolVar := variables[2]
	if boolVar.VariableValue.Value != "true" {
		t.Errorf("Expected bool value 'true', got '%s'", boolVar.VariableValue.Value)
	}
	if boolVar.VariableValue.DataTypeAttr == nil || *boolVar.VariableValue.DataTypeAttr != 11 {
		t.Errorf("Expected data type 11 (DT_BOOL) for boolean, got %v", boolVar.VariableValue.DataTypeAttr)
	}

	// Check DateTime variable
	dateVar := variables[3]
	if dateVar.VariableValue.Value != "2025-01-01" {
		t.Errorf("Expected date value '2025-01-01', got '%s'", dateVar.VariableValue.Value)
	}
	if dateVar.VariableValue.DataTypeAttr == nil || *dateVar.VariableValue.DataTypeAttr != 135 {
		t.Errorf("Expected data type 135 (DT_DBTIMESTAMP) for datetime, got %v", dateVar.VariableValue.DataTypeAttr)
	}
}

func TestValidate(t *testing.T) {
	// Create a package with some issues
	pkg := &dtsx.Package{
		ExecutableTypePackage: &schema.ExecutableTypePackage{
			Variables: &schema.VariablesType{
				Variable: []*schema.VariableType{
					{
						NamespaceAttr:  stringPtr("User"),
						ObjectNameAttr: stringPtr("UsedVar"),
						VariableValue: &schema.VariableValue{
							Value: "value",
						},
					},
					{
						NamespaceAttr:  stringPtr("User"),
						ObjectNameAttr: stringPtr("OrphanedVar"),
						VariableValue: &schema.VariableValue{
							Value: "value",
						},
					},
				},
			},
			ConnectionManagers: &schema.ConnectionManagersType{
				ConnectionManager: []*schema.ConnectionManagerType{
					{
						ObjectNameAttr: stringPtr("TestConn"),
						Property: []*schema.Property{
							{
								NameAttr: stringPtr("ConnectionString"),
								PropertyElementBaseType: &schema.PropertyElementBaseType{
									AnySimpleType: &schema.AnySimpleType{
										Value: "test connection",
									},
								},
							},
						},
					},
				},
			},
			PropertyExpression: []*schema.PropertyExpressionElementType{
				{
					NameAttr: "Name",
					AnySimpleType: &schema.AnySimpleType{
						Value: "@[User::UsedVar]",
					},
				},
			},
		},
	}

	errors := pkg.Validate()

	// Should have at least one info about orphaned variable
	foundOrphaned := false
	for _, err := range errors {
		if err.Message == "Variable is not referenced in any expression" && err.Path == "Variables.User::OrphanedVar" {
			foundOrphaned = true
			break
		}
	}
	if !foundOrphaned {
		t.Errorf("Expected to find orphaned variable warning, but didn't. Errors: %+v", errors)
	}

	// Test nil package
	var nilPkg *dtsx.Package
	nilErrors := nilPkg.Validate()
	if len(nilErrors) != 1 || nilErrors[0].Message != "Package is nil" {
		t.Errorf("Expected nil package error, got: %+v", nilErrors)
	}
}

func TestDependencyAnalysis(t *testing.T) {
	// Create a package with dependencies
	pkg := &dtsx.Package{
		ExecutableTypePackage: &schema.ExecutableTypePackage{
			Variables: &schema.VariablesType{
				Variable: []*schema.VariableType{
					{
						NamespaceAttr:  stringPtr("User"),
						ObjectNameAttr: stringPtr("UsedVar"),
						VariableValue: &schema.VariableValue{
							Value: "value",
						},
					},
					{
						NamespaceAttr:  stringPtr("User"),
						ObjectNameAttr: stringPtr("UnusedVar"),
						VariableValue: &schema.VariableValue{
							Value: "value",
						},
					},
				},
			},
			ConnectionManagers: &schema.ConnectionManagersType{
				ConnectionManager: []*schema.ConnectionManagerType{
					{
						ObjectNameAttr: stringPtr("TestConn"),
						Property: []*schema.Property{
							{
								NameAttr: stringPtr("ConnectionString"),
								PropertyElementBaseType: &schema.PropertyElementBaseType{
									AnySimpleType: &schema.AnySimpleType{
										Value: "test connection",
									},
								},
							},
						},
					},
				},
			},
			PropertyExpression: []*schema.PropertyExpressionElementType{
				{
					NameAttr: "TestExpr",
					AnySimpleType: &schema.AnySimpleType{
						Value: "@[User::UsedVar] + 1",
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
									Value: "TestConn",
								},
							},
						},
						{
							NameAttr: stringPtr("SqlStatementSource"),
							PropertyElementBaseType: &schema.PropertyElementBaseType{
								AnySimpleType: &schema.AnySimpleType{
									Value: "SELECT * FROM @[User::UsedVar]",
								},
							},
						},
					},
				},
			},
		},
	}

	// Build dependency graph
	graph := pkg.BuildDependencyGraph()
	if graph == nil {
		t.Fatal("Expected dependency graph, got nil")
	}

	// Check variable dependencies
	impact := graph.GetVariableImpact("User::UsedVar")
	if len(impact) == 0 {
		t.Error("Expected UsedVar to have dependencies")
	}
	found := false
	for _, loc := range impact {
		if strings.Contains(loc, "TestExpr") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected TestExpr in UsedVar impact, got: %v", impact)
	}

	// Check connection dependencies
	connImpact := graph.GetConnectionImpact("TestConn")
	if len(connImpact) == 0 {
		t.Error("Expected TestConn to have dependencies")
	}

	// Check unused variables
	unused := pkg.GetUnusedVariables()
	if len(unused) == 0 {
		t.Error("Expected to find unused variables")
	}
	foundUnused := false
	for _, v := range unused {
		if v == "User::UnusedVar" {
			foundUnused = true
			break
		}
	}
	if !foundUnused {
		t.Errorf("Expected UnusedVar in unused list, got: %v", unused)
	}

	// Test optimization suggestions
	suggestions := pkg.GetOptimizationSuggestions()
	if len(suggestions) == 0 {
		t.Log("No optimization suggestions (expected for simple package)")
	}
}

func TestTemplateSystem(t *testing.T) {
	// Test template registry
	registry := dtsx.NewTemplateRegistry()
	if registry == nil {
		t.Fatal("Expected registry, got nil")
	}

	// Test registering templates
	etlTemplate := dtsx.CreateBasicETLTemplate()
	registry.Register(etlTemplate)

	fileTemplate := dtsx.CreateFileProcessingTemplate()
	registry.Register(fileTemplate)

	// Test listing templates
	templates := registry.List()
	if len(templates) != 2 {
		t.Errorf("Expected 2 templates, got %d", len(templates))
	}

	// Test getting templates
	retrieved := registry.Get("Basic ETL")
	if retrieved == nil {
		t.Error("Expected to retrieve Basic ETL template")
	}
	if retrieved.Name != "Basic ETL" {
		t.Errorf("Expected template name 'Basic ETL', got '%s'", retrieved.Name)
	}

	// Test instantiation with parameters
	params := map[string]interface{}{
		"SourceConnection":      "Server=src;Database=test",
		"DestinationConnection": "Server=dst;Database=test",
		"SourceQuery":           "SELECT * FROM users",
		"DestinationTable":      "users_copy",
		"BatchSize":             "1000",
		"PackageName":           "MyETLPackage",
	}

	pkg, err := etlTemplate.Instantiate(params)
	if err != nil {
		t.Errorf("Failed to instantiate template: %v", err)
	}
	if pkg == nil {
		t.Fatal("Expected package, got nil")
	}

	// Verify parameter substitution
	if len(pkg.Property) > 0 && pkg.Property[0].Value != "MyETLPackage" {
		t.Errorf("Expected package name 'MyETLPackage', got '%s'", pkg.Property[0].Value)
	}

	// Check variables
	if pkg.Variables != nil && len(pkg.Variables.Variable) > 0 {
		if pkg.Variables.Variable[0].VariableValue.Value != "1000" {
			t.Errorf("Expected BatchSize '1000', got '%s'", pkg.Variables.Variable[0].VariableValue.Value)
		}
	}

	// Test default registry
	defaultRegistry := dtsx.GetDefaultTemplateRegistry()
	defaultTemplates := defaultRegistry.List()
	if len(defaultTemplates) < 2 {
		t.Errorf("Expected at least 2 default templates, got %d", len(defaultTemplates))
	}
}

func TestUpdateVariable(t *testing.T) {
	// Create a package with variables
	pkg := &dtsx.Package{
		ExecutableTypePackage: &schema.ExecutableTypePackage{
			Variables: &schema.VariablesType{
				Variable: []*schema.VariableType{
					{
						ObjectNameAttr: stringPtr("TestVar"),
						NamespaceAttr:  stringPtr("User"),
						VariableValue: &schema.VariableValue{
							Value: "OldValue",
						},
					},
				},
			},
		},
	}

	// Test successful update
	err := pkg.UpdateVariable("User", "TestVar", "NewValue")
	if err != nil {
		t.Fatalf("UpdateVariable failed: %v", err)
	}

	// Verify the update
	if pkg.Variables.Variable[0].VariableValue.Value != "NewValue" {
		t.Errorf("Expected variable value 'NewValue', got '%s'", pkg.Variables.Variable[0].VariableValue.Value)
	}

	// Test updating non-existent variable
	err = pkg.UpdateVariable("User", "NonExistent", "SomeValue")
	if err == nil {
		t.Error("Expected error when updating non-existent variable")
	}

	// Test updating with nil package
	var nilPkg *dtsx.Package
	err = nilPkg.UpdateVariable("User", "TestVar", "Value")
	if err == nil {
		t.Error("Expected error when updating variable on nil package")
	}
}

func TestUpdateProperty(t *testing.T) {
	// Create a package with variables, connections, and executables
	pkg := &dtsx.Package{
		ExecutableTypePackage: &schema.ExecutableTypePackage{
			Variables: &schema.VariablesType{
				Variable: []*schema.VariableType{
					{
						ObjectNameAttr: stringPtr("TestVar"),
						NamespaceAttr:  stringPtr("User"),
						VariableValue: &schema.VariableValue{
							Value: "OldValue",
						},
					},
				},
			},
			ConnectionManagers: &schema.ConnectionManagersType{
				ConnectionManager: []*schema.ConnectionManagerType{
					{
						Property: []*schema.Property{
							{
								NameAttr: stringPtr("ObjectName"),
								PropertyElementBaseType: &schema.PropertyElementBaseType{
									AnySimpleType: &schema.AnySimpleType{
										Value: "TestConn",
									},
								},
							},
							{
								NameAttr: stringPtr("ConnectionString"),
								PropertyElementBaseType: &schema.PropertyElementBaseType{
									AnySimpleType: &schema.AnySimpleType{
										Value: "OldConnectionString",
									},
								},
							},
						},
					},
				},
			},
			Executable: []*schema.AnyNonPackageExecutableType{
				{
					Property: []*schema.Property{
						{
							NameAttr: stringPtr("ObjectName"),
							PropertyElementBaseType: &schema.PropertyElementBaseType{
								AnySimpleType: &schema.AnySimpleType{
									Value: "TestExec",
								},
							},
						},
						{
							NameAttr: stringPtr("Description"),
							PropertyElementBaseType: &schema.PropertyElementBaseType{
								AnySimpleType: &schema.AnySimpleType{
									Value: "OldDescription",
								},
							},
						},
					},
				},
			},
		},
	}

	// Test updating package property
	err := pkg.UpdateProperty("package", "", "CreatorName", "NewCreator")
	if err != nil {
		t.Fatalf("UpdateProperty failed for package: %v", err)
	}

	// Verify package property update
	found := false
	for _, prop := range pkg.Property {
		if prop.NameAttr != nil && *prop.NameAttr == "CreatorName" {
			if prop.PropertyElementBaseType != nil && prop.PropertyElementBaseType.AnySimpleType != nil &&
				prop.PropertyElementBaseType.AnySimpleType.Value == "NewCreator" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("Package property 'CreatorName' not updated correctly")
	}

	// Test updating variable property (Value)
	err = pkg.UpdateProperty("variable", "User::TestVar", "Value", "NewVarValue")
	if err != nil {
		t.Fatalf("UpdateProperty failed for variable: %v", err)
	}

	if pkg.Variables.Variable[0].VariableValue.Value != "NewVarValue" {
		t.Errorf("Expected variable value 'NewVarValue', got '%s'", pkg.Variables.Variable[0].VariableValue.Value)
	}

	// Test updating connection property
	err = pkg.UpdateProperty("connection", "TestConn", "ConnectionString", "NewConnectionString")
	if err != nil {
		t.Fatalf("UpdateProperty failed for connection: %v", err)
	}

	// Verify connection property update
	found = false
	for _, prop := range pkg.ConnectionManagers.ConnectionManager[0].Property {
		if prop.NameAttr != nil && *prop.NameAttr == "ConnectionString" {
			if prop.PropertyElementBaseType != nil && prop.PropertyElementBaseType.AnySimpleType != nil &&
				prop.PropertyElementBaseType.AnySimpleType.Value == "NewConnectionString" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("Connection property 'ConnectionString' not updated correctly")
	}

	// Test updating executable property
	err = pkg.UpdateProperty("executable", "TestExec", "Description", "NewDescription")
	if err != nil {
		t.Fatalf("UpdateProperty failed for executable: %v", err)
	}

	// Verify executable property update
	found = false
	for _, prop := range pkg.Executable[0].Property {
		if prop.NameAttr != nil && *prop.NameAttr == "Description" {
			if prop.PropertyElementBaseType != nil && prop.PropertyElementBaseType.AnySimpleType != nil &&
				prop.PropertyElementBaseType.AnySimpleType.Value == "NewDescription" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("Executable property 'Description' not updated correctly")
	}

	// Test invalid target type
	err = pkg.UpdateProperty("invalid", "Test", "Property", "Value")
	if err == nil {
		t.Error("Expected error for invalid target type")
	}

	// Test non-existent elements
	err = pkg.UpdateProperty("variable", "NonExistent::Var", "Value", "Test")
	if err == nil {
		t.Error("Expected error for non-existent variable")
	}

	err = pkg.UpdateProperty("connection", "NonExistent", "Property", "Value")
	if err == nil {
		t.Error("Expected error for non-existent connection")
	}

	err = pkg.UpdateProperty("executable", "NonExistent", "Property", "Value")
	if err == nil {
		t.Error("Expected error for non-existent executable")
	}

	// Test with nil package
	var nilPkg *dtsx.Package
	err = nilPkg.UpdateProperty("package", "", "Property", "Value")
	if err == nil {
		t.Error("Expected error when updating property on nil package")
	}
}
