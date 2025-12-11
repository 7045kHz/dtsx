//go:build ignore

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

// This example edits the DTSX XML file in place (writes a copy) without
// re-marshaling, preserving the original DTSX structure and attributes.
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run examples/modify_sql_command_inplace.go <dtsx-file>")
		os.Exit(1)
	}

	src := os.Args[1]
	dest := "ConfigFile2_inplace.dtsx"

	data, err := ioutil.ReadFile(src)
	if err != nil {
		log.Fatalf("read: %v", err)
	}
	s := string(data)

	// Locate the data flow package executable block by scanning for the string 'production_product_color_to_csv'
	pos := strings.Index(s, "production_product_color_to_csv")
	if pos == -1 {
		log.Fatalf("data flow exec identifier not found")
	}

	// find the start of the <Executable ...> tag that contains this refId
	startExecIdx := strings.LastIndex(s[:pos], "<DTS:Executable")
	if startExecIdx == -1 {
		startExecIdx = strings.LastIndex(s[:pos], "<Executable")
	}
	if startExecIdx == -1 {
		log.Fatalf("executable start not found")
	}

	// find the end of the component </component>
	// find the end of the Executable block so we restrict our search to that executable
	execEndIdx := strings.Index(s[pos:], "</DTS:Executable>")
	if execEndIdx == -1 {
		execEndIdx = strings.Index(s[pos:], "</Executable>")
	}
	if execEndIdx == -1 {
		log.Fatalf("executable end not found")
	}
	execEndIdx = pos + execEndIdx + len("</DTS:Executable>")

	// Search for component 'get_colors_all' within this executable block
	compStartRel := strings.Index(s[startExecIdx:execEndIdx], "name=\"get_colors_all\"")
	if compStartRel == -1 {
		log.Fatalf("component name get_colors_all not found in exec")
	}
	compStart := startExecIdx + strings.LastIndex(s[startExecIdx:startExecIdx+compStartRel], "<component")
	if compStart < 0 {
		log.Fatalf("component start not found")
	}
	startIdx := compStart

	// find the end of the component </component>
	endRel := strings.Index(s[startIdx:execEndIdx], "</component>")
	if endRel == -1 {
		log.Fatalf("component end not found")
	}
	endIdx := startIdx + endRel + len("</component>")

	comp := s[startIdx:endIdx]

	// Find SqlCommand property in this component
	// The property may include attributes across multiple lines
	re := regexp.MustCompile(`(?s)<property[^>]*name="SqlCommand"[^>]*>(.*?)</property>`) // DOTALL
	match := re.FindStringSubmatch(comp)
	if len(match) < 2 {
		log.Fatalf("SqlCommand property not found in component")
	}

	oldVal := match[1]
	// Build new value; encode single quotes as &#39;
	newSQL := "select * from production.product where color=&#39;Blue&#39;"

	newComp := re.ReplaceAllString(comp, strings.Replace(match[0], oldVal, newSQL, 1))

	// Replace component content in package
	newS := s[:startIdx] + newComp + s[endIdx:]

	// Save to destination
	if err := ioutil.WriteFile(dest, []byte(newS), 0644); err != nil {
		log.Fatalf("write: %v", err)
	}

	fmt.Println("Updated SqlCommand and saved to", dest)
}
