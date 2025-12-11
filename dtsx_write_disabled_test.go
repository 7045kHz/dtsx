package dtsx

import (
    "bytes"
    "os"
    "testing"
)

func TestInternalMarshalHelpersWork(t *testing.T) {
    pkg := NewPackageBuilder().Build()

    // marshalToWriter should write bytes to the writer
    var buf bytes.Buffer
    if err := marshalToWriter(&buf, pkg); err != nil {
        t.Fatalf("marshalToWriter failed: %v", err)
    }
    if buf.Len() == 0 {
        t.Fatalf("marshalToWriter produced empty output")
    }

    // marshalToFile should create a file
    tmpFile, err := os.CreateTemp("", "dtsx_test_*.dtsx")
    if err != nil {
        t.Fatalf("failed to create temp file: %v", err)
    }
    tmpName := tmpFile.Name()
    tmpFile.Close()
    defer os.Remove(tmpName)

    if err := marshalToFile(tmpName, pkg); err != nil {
        t.Fatalf("marshalToFile failed: %v", err)
    }
    fi, err := os.Stat(tmpName)
    if err != nil {
        t.Fatalf("marshalToFile did not create file: %v", err)
    }
    if fi.Size() == 0 {
        t.Fatalf("marshalToFile produced empty file")
    }
}


func TestInternalMutatingMethodsWork(t *testing.T) {
    // Test updateVariable on a package with a variable
    pkgWithVar := NewPackageBuilder().AddVariableWithType("User", "MyVar", "Old", "String").Build()
    if err := pkgWithVar.updateVariable("User", "MyVar", "new"); err != nil {
        t.Fatalf("updateVariable failed: %v", err)
    }

    // Verify the variable was updated
    if pkgWithVar.Variables == nil || len(pkgWithVar.Variables.Variable) == 0 {
        t.Fatalf("package has no variables after update")
    }
    found := false
    for _, v := range pkgWithVar.Variables.Variable {
        if v.NamespaceAttr != nil && v.ObjectNameAttr != nil && *v.NamespaceAttr == "User" && *v.ObjectNameAttr == "MyVar" {
            if v.VariableValue == nil || v.VariableValue.Value != "new" {
                t.Fatalf("updateVariable did not set value properly, got: %#v", v.VariableValue)
            }
            found = true
            break
        }
    }
    if !found {
        t.Fatalf("variable User::MyVar not found after update")
    }

    // Update connection string: create a connection manager
    pkgWithConn := NewPackageBuilder().AddConnection("MyConn", "OLEDB", "OldConnectionString").Build()
    if err := pkgWithConn.updateConnectionString("MyConn", "newconn"); err != nil {
        t.Fatalf("updateConnectionString failed: %v", err)
    }
    // Verify the connection string was updated
    for _, cm := range pkgWithConn.ConnectionManagers.ConnectionManager {
        var name string
        for _, prop := range cm.Property {
            if prop.NameAttr != nil && *prop.NameAttr == "ObjectName" && prop.PropertyElementBaseType != nil && prop.PropertyElementBaseType.AnySimpleType != nil {
                name = prop.PropertyElementBaseType.AnySimpleType.Value
            }
            if prop.NameAttr != nil && *prop.NameAttr == "ConnectionString" && prop.PropertyElementBaseType != nil && prop.PropertyElementBaseType.AnySimpleType != nil && name == "MyConn" {
                if prop.PropertyElementBaseType.AnySimpleType.Value != "newconn" {
                    t.Fatalf("updateConnectionString did not set value properly, got: %s", prop.PropertyElementBaseType.AnySimpleType.Value)
                }
            }
        }
    }

    // Update expression/internal property tests are similar: create variable
    // and call updateExpression and updateProperty where appropriate
    pkgForExpr := NewPackageBuilder().AddVariableWithType("User", "MyVar", "Old", "String").Build()
    if err := pkgForExpr.updateVariable("User", "MyVar", "expr"); err != nil {
        t.Fatalf("updateVariable (for expression) failed: %v", err)
    }
    if err := pkgForExpr.updateProperty("package", "", "CreatorName", "whoever"); err != nil {
        t.Fatalf("updateProperty failed: %v", err)
    }
}


