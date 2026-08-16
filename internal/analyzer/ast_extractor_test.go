package analyzer

import (
	"os"
	"testing"
)

func TestExtractRelationships(t *testing.T) {
	// Create a temporary Go file with sample code
	content := `
package test

import "fmt"

type User struct {
    Name string
}

func (u *User) Greet() {
    fmt.Println("Hello")
}

func main() {
    u := User{}
    u.Greet()
}
`
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test.go"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	// Also create a go.mod file (required for go/packages)
	if err := os.WriteFile(tmpDir+"/go.mod", []byte("module test\n\ngo 1.21"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Extract(tmpDir)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Check entities
	expectedEntityNames := map[string]bool{
		"test.User":       true,
		"test.main":       true,
		"test.Greet":      true,
		"test":            true, // package
		tmpFile:           true, // file
		"test.fmt":        true, // external import
		"test.User.Greet": true, // method (if we handle)
	}
	for _, e := range result.Entities {
		if _, ok := expectedEntityNames[e.ID]; !ok {
			t.Logf("Unexpected entity: %s", e.ID)
		}
	}

	// Check relationships
	found := false
	for _, rel := range result.Relationships {
		if rel.Type == RelCalls && rel.From == "test.main" && rel.To == "test.Greet" {
			found = true
		}
	}
	if !found {
		t.Error("Expected CALLS relationship from main to Greet not found")
	}
}
