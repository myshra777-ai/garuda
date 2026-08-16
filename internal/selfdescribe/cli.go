package selfdescribe

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ParseCLICommands scans the codebase for Cobra commands defined in cmd/garuda/.
func ParseCLICommands(root string) CLIInfo {
	var commands []CLICommand
	seen := make(map[string]bool)

	// Only scan cmd/garuda/ directory
	cmdDir := filepath.Join(root, "cmd", "garuda")
	if _, err := os.Stat(cmdDir); os.IsNotExist(err) {
		return CLIInfo{Commands: commands}
	}

	var cmdFiles []string
	err := filepath.Walk(cmdDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			cmdFiles = append(cmdFiles, path)
		}
		return nil
	})
	if err != nil {
		return CLIInfo{Commands: commands}
	}

	// Collect all command variable names that are added to rootCmd
	commandVarNames := make(map[string]bool)

	for _, file := range cmdFiles {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			continue
		}

		// Look for rootCmd.AddCommand(...)
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			// Look for pattern: rootCmd.AddCommand(someCmd)
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == "AddCommand" {
					// The first argument is the command variable
					if len(call.Args) > 0 {
						if ident, ok := call.Args[0].(*ast.Ident); ok {
							commandVarNames[ident.Name] = true
						}
					}
				}
			}
			return true
		})
	}

	// Now find the actual command definitions (var someCmd = &cobra.Command{...})
	for _, file := range cmdFiles {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			continue
		}

		ast.Inspect(f, func(n ast.Node) bool {
			// Look for var declarations
			decl, ok := n.(*ast.GenDecl)
			if !ok || decl.Tok != token.VAR {
				return true
			}
			for _, spec := range decl.Specs {
				valSpec, ok := spec.(*ast.ValueSpec)
				if !ok || len(valSpec.Names) == 0 {
					continue
				}
				varName := valSpec.Names[0].Name
				if !commandVarNames[varName] {
					continue // not added to rootCmd
				}
				// Look for a composite literal of type &cobra.Command{...}
				if len(valSpec.Values) > 0 {
					if unary, ok := valSpec.Values[0].(*ast.UnaryExpr); ok {
						if comp, ok := unary.X.(*ast.CompositeLit); ok {
							// Extract the command's Use field
							cmdName := ""
							for _, elt := range comp.Elts {
								kv, ok := elt.(*ast.KeyValueExpr)
								if !ok {
									continue
								}
								key, ok := kv.Key.(*ast.Ident)
								if !ok || key.Name != "Use" {
									continue
								}
								if val, ok := kv.Value.(*ast.BasicLit); ok && val.Kind == token.STRING {
									cmdName = strings.Trim(val.Value, `"`)
									break
								}
							}
							if cmdName != "" && !seen[cmdName] {
								seen[cmdName] = true
								commands = append(commands, CLICommand{
									Name:        cmdName,
									Description: "Garuda command",
									Flags:       []string{},
								})
							}
						}
					}
				}
			}
			return true
		})
	}

	return CLIInfo{Commands: commands}
}
