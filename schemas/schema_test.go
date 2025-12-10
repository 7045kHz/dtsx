package schema

import (
	"encoding/xml"
	"testing"
)

// Helper function for string pointers in tests
func stringPtr(s string) *string {
	return &s
}

// Helper function for int pointers in tests
func intPtr(i int) *int {
	return &i
}

func TestSchemaStructs(t *testing.T) {
	t.Run("ExecutableTypePackage", func(t *testing.T) {
		// Test creation and field access
		pkg := &ExecutableTypePackage{
			ExecutableTypeAttr: stringPtr("STOCK:SSISPackage"),
			Property: []*Property{
				{
					NameAttr: stringPtr("Name"),
					PropertyElementBaseType: &PropertyElementBaseType{
						AnySimpleType: &AnySimpleType{
							Value: "TestPackage",
						},
					},
				},
			},
			Variables: &VariablesType{
				Variable: []*VariableType{
					{
						NamespaceAttr:  stringPtr("User"),
						ObjectNameAttr: stringPtr("TestVar"),
						VariableValue: &VariableValue{
							Value: "test value",
						},
					},
				},
			},
		}

		if pkg.ExecutableTypeAttr == nil || *pkg.ExecutableTypeAttr != "STOCK:SSISPackage" {
			t.Error("ExecutableTypeAttr not set correctly")
		}

		if len(pkg.Property) != 1 || pkg.Property[0].NameAttr == nil || *pkg.Property[0].NameAttr != "Name" {
			t.Error("Property not set correctly")
		}

		if pkg.Variables == nil || len(pkg.Variables.Variable) != 1 {
			t.Error("Variables not set correctly")
		}
	})

	t.Run("AnyNonPackageExecutableType", func(t *testing.T) {
		exec := &AnyNonPackageExecutableType{
			RefIdAttr:          stringPtr("Package.Task1"),
			ExecutableTypeAttr: "ExecuteSQLTask",
			ObjectNameAttr:     stringPtr("Task1"),
			Property: []*Property{
				{
					NameAttr: stringPtr("Connection"),
					PropertyElementBaseType: &PropertyElementBaseType{
						AnySimpleType: &AnySimpleType{
							Value: "TestConn",
						},
					},
				},
			},
		}

		if exec.RefIdAttr == nil || *exec.RefIdAttr != "Package.Task1" {
			t.Error("RefIdAttr not set correctly")
		}

		if exec.ExecutableTypeAttr != "ExecuteSQLTask" {
			t.Error("ExecutableTypeAttr not set correctly")
		}

		if len(exec.Property) != 1 {
			t.Error("Property not set correctly")
		}
	})

	t.Run("VariableType", func(t *testing.T) {
		variable := &VariableType{
			NamespaceAttr:  stringPtr("User"),
			ObjectNameAttr: stringPtr("MyVar"),
			VariableValue: &VariableValue{
				Value: "42",
			},
			Property: []*Property{
				{
					NameAttr: stringPtr("EvaluateAsExpression"),
					PropertyElementBaseType: &PropertyElementBaseType{
						AnySimpleType: &AnySimpleType{
							Value: "false",
						},
					},
				},
			},
		}

		if variable.NamespaceAttr == nil || *variable.NamespaceAttr != "User" {
			t.Error("NamespaceAttr not set correctly")
		}

		if variable.VariableValue == nil || variable.VariableValue.Value != "42" {
			t.Error("VariableValue not set correctly")
		}
	})

	t.Run("ConnectionManagerType", func(t *testing.T) {
		conn := &ConnectionManagerType{
			RefIdAttr:        stringPtr("Package.Connections[TestConn]"),
			ObjectNameAttr:   stringPtr("TestConn"),
			CreationNameAttr: stringPtr("OLEDB"),
			Property: []*Property{
				{
					NameAttr: stringPtr("ConnectionString"),
					PropertyElementBaseType: &PropertyElementBaseType{
						AnySimpleType: &AnySimpleType{
							Value: "Server=test;Database=test",
						},
					},
				},
			},
		}

		if conn.ObjectNameAttr == nil || *conn.ObjectNameAttr != "TestConn" {
			t.Error("ObjectNameAttr not set correctly")
		}

		if conn.CreationNameAttr == nil || *conn.CreationNameAttr != "OLEDB" {
			t.Error("CreationNameAttr not set correctly")
		}
	})

	t.Run("PrecedenceConstraintType", func(t *testing.T) {
		pc := &PrecedenceConstraintType{
			Property: []*Property{
				{
					NameAttr: stringPtr("Value"),
					PropertyElementBaseType: &PropertyElementBaseType{
						AnySimpleType: &AnySimpleType{
							Value: "Success",
						},
					},
				},
			},
			Executable: []*PrecedenceConstraintExecutableReferenceType{
				{
					IDREFAttr:  stringPtr("Package.Task2"),
					IsFromAttr: intPtr(0),
				},
			},
		}

		if len(pc.Executable) != 1 {
			t.Error("Executable not set correctly")
		}

		if pc.Executable[0].IDREFAttr == nil || *pc.Executable[0].IDREFAttr != "Package.Task2" {
			t.Error("Executable IDREFAttr not set correctly")
		}
	})

	t.Run("PropertyExpressionElementType", func(t *testing.T) {
		expr := &PropertyExpressionElementType{
			NameAttr: "SqlStatementSource",
			AnySimpleType: &AnySimpleType{
				Value: "@[User::Var1] + @[User::Var2]",
			},
		}

		if expr.NameAttr != "SqlStatementSource" {
			t.Error("NameAttr not set correctly")
		}

		if expr.AnySimpleType == nil || expr.AnySimpleType.Value != "@[User::Var1] + @[User::Var2]" {
			t.Error("AnySimpleType not set correctly")
		}
	})

	t.Run("XMLMarshaling", func(t *testing.T) {
		// Test XML marshaling/unmarshaling of a simple package
		original := &ExecutableTypePackage{
			ExecutableTypeAttr: stringPtr("STOCK:SSISPackage"),
			Property: []*Property{
				{
					NameAttr: stringPtr("Name"),
					PropertyElementBaseType: &PropertyElementBaseType{
						AnySimpleType: &AnySimpleType{
							Value: "TestPackage",
						},
					},
				},
			},
		}

		// Marshal to XML
		data, err := xml.MarshalIndent(original, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal XML: %v", err)
		}

		// Unmarshal back
		var unmarshaled ExecutableTypePackage
		err = xml.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Fatalf("Failed to unmarshal XML: %v", err)
		}

		// Verify the data matches
		if unmarshaled.ExecutableTypeAttr == nil || *unmarshaled.ExecutableTypeAttr != "STOCK:SSISPackage" {
			t.Error("Unmarshaled ExecutableTypeAttr does not match")
		}

		if len(unmarshaled.Property) != 1 || unmarshaled.Property[0].NameAttr == nil || *unmarshaled.Property[0].NameAttr != "Name" {
			t.Error("Unmarshaled Property does not match")
		}
	})

	t.Run("TypeAliases", func(t *testing.T) {
		// Test the Executable type alias
		var exec Executable
		pkg := &ExecutableTypePackage{
			ExecutableTypeAttr: stringPtr("STOCK:SSISPackage"),
		}
		exec = pkg

		if exec == nil {
			t.Error("Executable type alias assignment failed")
		}

		// Test dereferencing
		if exec.ExecutableTypeAttr == nil || *exec.ExecutableTypeAttr != "STOCK:SSISPackage" {
			t.Error("Type alias dereference not working")
		}
	})

	t.Run("TaskSpecificSchemas", func(t *testing.T) {
		// Test SQL Task Data type alias
		var sqlTask SqlTaskData
		sqlData := &SqlTaskDataType{
			ParameterBinding: []*SqlTaskParameterBindingType{
				{
					ParameterNameAttr:   stringPtr("@Param1"),
					DtsVariableNameAttr: stringPtr("User::Var1"),
				},
			},
		}
		sqlTask = sqlData

		if sqlTask == nil {
			t.Error("SqlTaskData type alias assignment failed")
		}

		if len(sqlTask.ParameterBinding) != 1 {
			t.Error("SqlTaskData ParameterBinding not set correctly")
		}

		// Test Bulk Insert Task Data
		bulkData := &BulkInsertTaskDataType{
			SourceConnectionNameAttr:      stringPtr("TestConn"),
			DestinationTableNameAttr:      stringPtr("dbo.TestTable"),
			DestinationConnectionNameAttr: stringPtr("DestConn"),
		}

		if bulkData.SourceConnectionNameAttr == nil || *bulkData.SourceConnectionNameAttr != "TestConn" {
			t.Error("BulkInsertTaskDataType SourceConnectionNameAttr not set correctly")
		}

		if bulkData.DestinationTableNameAttr == nil || *bulkData.DestinationTableNameAttr != "dbo.TestTable" {
			t.Error("BulkInsertTaskDataType DestinationTableNameAttr not set correctly")
		}

		// Test Send Mail Task Data
		mailData := &SendMailTaskType{
			SMTPServerAttr: stringPtr("smtp.example.com"),
			ToAttr:         stringPtr("test@example.com"),
			SubjectAttr:    stringPtr("Test Subject"),
		}

		if mailData.SMTPServerAttr == nil || *mailData.SMTPServerAttr != "smtp.example.com" {
			t.Error("SendMailTaskType SMTPServerAttr not set correctly")
		}

		if mailData.SubjectAttr == nil || *mailData.SubjectAttr != "Test Subject" {
			t.Error("SendMailTaskType SubjectAttr not set correctly")
		}
	})
}
