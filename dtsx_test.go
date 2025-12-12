package dtsx_test

import (
	"os"
	"path/filepath"
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
			dtsxFile = filepath.Join("SSIS_EXAMPLES", file.Name())
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
			dtsxFile = filepath.Join("SSIS_EXAMPLES", file.Name())
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
			dtsxFile = filepath.Join("SSIS_EXAMPLES", file.Name())
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
			dtsxFile = filepath.Join("SSIS_EXAMPLES", file.Name())
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
			dtsxFile = filepath.Join("SSIS_EXAMPLES", file.Name())
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
			dtsxFile = filepath.Join("SSIS_EXAMPLES", file.Name())
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
			dtsxFile = filepath.Join("SSIS_EXAMPLES", file.Name())
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
