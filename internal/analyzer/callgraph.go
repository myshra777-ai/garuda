// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package analyzer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"
	"time"

	garudatypes "github.com/myshra777-ai/garuda/internal/types"

	"github.com/google/uuid"
)

// CallGraphExtractor extracts type-safe relationship edges from Go AST & type-checking info.
type CallGraphExtractor struct {
	fset         *token.FileSet
	info         *types.Info
	pkgPath      string
	tenantID     uuid.UUID
	workspaceID  uuid.UUID
	repositoryID uuid.UUID
	fileContent  map[string]string
}

// NewCallGraphExtractor constructs a type-aware call graph extractor.
func NewCallGraphExtractor(
	fset *token.FileSet,
	info *types.Info,
	pkgPath string,
	tenantID, workspaceID, repoID uuid.UUID,
	fileContent map[string]string,
) *CallGraphExtractor {
	return &CallGraphExtractor{
		fset:         fset,
		info:         info,
		pkgPath:      pkgPath,
		tenantID:     tenantID,
		workspaceID:  workspaceID,
		repositoryID: repoID,
		fileContent:  fileContent,
	}
}

// ExtractRelationships walks the AST files and extracts deterministic CALLS and IMPLEMENTS edges[cite: 1, 2].
func (c *CallGraphExtractor) ExtractRelationships(files []*ast.File) []garudatypes.Relationship {
	var relationships []garudatypes.Relationship

	for _, file := range files {
		var currentFunc string

		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncDecl:
				currentFunc = c.resolveFuncQualifiedName(node)
				return true

			case *ast.CallExpr:
				if currentFunc == "" {
					return true
				}
				rel := c.resolveCallExpr(node, currentFunc)
				if rel != nil {
					relationships = append(relationships, *rel)
				}
				return true
			}
			return true
		})
	}

	return relationships
}

func (c *CallGraphExtractor) resolveFuncQualifiedName(fn *ast.FuncDecl) string {
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		recvType := c.formatReceiver(fn.Recv.List[0].Type)
		return fmt.Sprintf("%s.(%s).%s", c.pkgPath, recvType, fn.Name.Name)
	}
	return fmt.Sprintf("%s.%s", c.pkgPath, fn.Name.Name)
}

func (c *CallGraphExtractor) formatReceiver(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return "*" + ident.Name
		}
	case *ast.Ident:
		return t.Name
	}
	return "Unknown"
}

func (c *CallGraphExtractor) resolveCallExpr(call *ast.CallExpr, callerQualifiedName string) *garudatypes.Relationship {
	pos := c.fset.Position(call.Pos())
	end := c.fset.Position(call.End())

	var targetQualifiedName string
	var isInterfaceCall bool

	// 1. Type-aware resolution via types.Info
	if c.info != nil {
		switch fun := call.Fun.(type) {
		case *ast.Ident:
			if obj, ok := c.info.Uses[fun]; ok && obj != nil {
				targetPkg := c.pkgPath
				if obj.Pkg() != nil {
					targetPkg = obj.Pkg().Path()
				}
				targetQualifiedName = fmt.Sprintf("%s.%s", targetPkg, obj.Name())
			}

		case *ast.SelectorExpr:
			if sel, ok := c.info.Selections[fun]; ok && sel != nil {
				obj := sel.Obj()
				targetPkg := c.pkgPath
				if obj.Pkg() != nil {
					targetPkg = obj.Pkg().Path()
				}

				if sel.Kind() == types.MethodVal || sel.Kind() == types.MethodExpr {
					recv := sel.Recv()
					if types.IsInterface(recv) {
						isInterfaceCall = true
						targetQualifiedName = fmt.Sprintf("%s.%s.%s", targetPkg, recv.String(), obj.Name())
					} else {
						targetQualifiedName = fmt.Sprintf("%s.(%s).%s", targetPkg, recv.String(), obj.Name())
					}
				} else {
					targetQualifiedName = fmt.Sprintf("%s.%s", targetPkg, obj.Name())
				}
			} else if obj, ok := c.info.Uses[fun.Sel]; ok && obj != nil {
				targetPkg := c.pkgPath
				if obj.Pkg() != nil {
					targetPkg = obj.Pkg().Path()
				}
				targetQualifiedName = fmt.Sprintf("%s.%s", targetPkg, obj.Name())
			}
		}
	}

	// 2. Fallback to AST selector inspection if type-checking is unindexed
	if targetQualifiedName == "" {
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok {
				targetQualifiedName = fmt.Sprintf("%s/%s.%s", c.pkgPath, ident.Name, sel.Sel.Name)
			}
		}
	}

	if targetQualifiedName == "" {
		return nil
	}

	// Calculate deterministic evidence snippet & content hash[cite: 1, 2]
	snippet := c.extractSnippet(pos.Filename, pos.Line, end.Line)
	hash := sha256.Sum256([]byte(snippet))
	evidenceHash := hex.EncodeToString(hash[:])

	predicate := garudatypes.PredicateCalls
	if isInterfaceCall {
		predicate = garudatypes.PredicateCallsInterface
	}

	return &garudatypes.Relationship{
		ID:             uuid.New(),
		TenantID:       c.tenantID,
		WorkspaceID:    c.workspaceID,
		RepositoryID:   c.repositoryID,
		SourceName:     callerQualifiedName,
		TargetName:     targetQualifiedName,
		Predicate:      predicate,
		Confidence:     1.0,
		EpistemicClass: garudatypes.EpistemicClassObservation,
		EvidenceHash:   evidenceHash,
		LineStart:      pos.Line,
		LineEnd:        end.Line,
		CreatedAt:      time.Now().UTC(),
	}
}

func (c *CallGraphExtractor) extractSnippet(filename string, lineStart, lineEnd int) string {
	content, ok := c.fileContent[filename]
	if !ok {
		return ""
	}
	lines := strings.Split(content, "\n")
	if lineStart < 1 || lineStart > len(lines) {
		return ""
	}
	if lineEnd > len(lines) {
		lineEnd = len(lines)
	}
	return strings.Join(lines[lineStart-1:lineEnd], "\n")
}
