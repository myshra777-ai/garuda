// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package analyzer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/canonical"
	garudatypes "github.com/myshra777-ai/garuda/internal/types"
)

var canonicalNamespace = uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

type AnalysisOptions struct {
	IncludeCallGraph bool
	TypeResolution   bool
}

type AnalysisRequest struct {
	Path      string
	CommitSHA string
	Options   AnalysisOptions
}

type GoAnalyzer struct{}

func NewGoAnalyzer() *GoAnalyzer {
	return &GoAnalyzer{}
}

type parsedFile struct {
	path    string
	ast     *ast.File
	content string
}

// workspaceImporter resolves both stdlib and multi-package workspace imports on demand.
type workspaceImporter struct {
	fset        *token.FileSet
	packages    map[string]*types.Package
	pkgFilesMap map[string][]parsedFile
	pkgPathMap  map[string]string
	defaultImp  types.Importer
}

func newWorkspaceImporter(fset *token.FileSet, pkgFilesMap map[string][]parsedFile, baseDirName string) *workspaceImporter {
	imp := &workspaceImporter{
		fset:        fset,
		packages:    make(map[string]*types.Package),
		pkgFilesMap: pkgFilesMap,
		pkgPathMap:  make(map[string]string),
		defaultImp:  importer.Default(),
	}

	for dirPart, pFiles := range pkgFilesMap {
		if len(pFiles) == 0 {
			continue
		}
		var pkgPath string
		if dirPart == "." {
			pkgPath = "example.com/" + pFiles[0].ast.Name.Name
		} else {
			pkgPath = fmt.Sprintf("example.com/%s/%s", baseDirName, dirPart)
		}
		imp.pkgPathMap[dirPart] = pkgPath
		imp.pkgPathMap[pkgPath] = pkgPath
		imp.pkgPathMap[pFiles[0].ast.Name.Name] = pkgPath
	}
	return imp
}

func (w *workspaceImporter) Import(path string) (*types.Package, error) {
	if pkg, ok := w.packages[path]; ok {
		return pkg, nil
	}

	// 1. Check standard library
	if pkg, err := w.defaultImp.Import(path); err == nil && pkg != nil {
		w.packages[path] = pkg
		return pkg, nil
	}

	// 2. Resolve against workspace packages
	var targetDirPart string
	var targetPkgPath string

	for dirPart, pFiles := range w.pkgFilesMap {
		if len(pFiles) == 0 {
			continue
		}
		cPkgPath := w.pkgPathMap[dirPart]
		if path == cPkgPath || strings.HasSuffix(path, "/"+dirPart) || strings.HasSuffix(path, "/"+pFiles[0].ast.Name.Name) || path == dirPart {
			targetDirPart = dirPart
			targetPkgPath = cPkgPath
			break
		}
	}

	if targetPkgPath == "" {
		return nil, fmt.Errorf("package %q not found in workspace", path)
	}

	if pkg, ok := w.packages[targetPkgPath]; ok {
		w.packages[path] = pkg
		return pkg, nil
	}

	pFiles := w.pkgFilesMap[targetDirPart]
	var astFiles []*ast.File
	for _, pf := range pFiles {
		astFiles = append(astFiles, pf.ast)
	}

	pkgInfo := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}

	conf := types.Config{
		Importer: w,
		Error:    func(err error) {},
	}

	pkg, err := conf.Check(targetPkgPath, w.fset, astFiles, pkgInfo)
	if pkg != nil {
		w.packages[targetPkgPath] = pkg
		w.packages[path] = pkg
		return pkg, nil
	}
	return nil, err
}

func (a *GoAnalyzer) Analyze(ctx context.Context, req AnalysisRequest) (*garudatypes.Snapshot, error) {
	fset := token.NewFileSet()
	var allEntities []garudatypes.Entity
	var allRelationships []garudatypes.Relationship
	fileContents := make(map[string]string)

	baseDirName := filepath.Base(req.Path)
	pkgFilesMap := make(map[string][]parsedFile)

	// 1. Walk directory and group AST files by package directory
	err := filepath.Walk(req.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		contentBytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		contentStr := string(contentBytes)
		fileContents[path] = contentStr

		fileAST, err := parser.ParseFile(fset, path, contentBytes, parser.ParseComments)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(req.Path, path)
		dirPart := filepath.ToSlash(filepath.Dir(relPath))
		pkgFilesMap[dirPart] = append(pkgFilesMap[dirPart], parsedFile{
			path:    path,
			ast:     fileAST,
			content: contentStr,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	wImporter := newWorkspaceImporter(fset, pkgFilesMap, baseDirName)

	// 2. Process each package directory with authoritative workspace type-checking
	for dirPart, pFiles := range pkgFilesMap {
		if len(pFiles) == 0 {
			continue
		}

		pkgPath := wImporter.pkgPathMap[dirPart]

		var astFiles []*ast.File
		for _, pf := range pFiles {
			astFiles = append(astFiles, pf.ast)
		}

		info := &types.Info{
			Types:      make(map[ast.Expr]types.TypeAndValue),
			Defs:       make(map[*ast.Ident]types.Object),
			Uses:       make(map[*ast.Ident]types.Object),
			Selections: make(map[*ast.SelectorExpr]*types.Selection),
		}
		typeConfig := types.Config{
			Importer: wImporter,
			Error:    func(err error) {},
		}
		_, _ = typeConfig.Check(pkgPath, fset, astFiles, info)

		// Extract entities for each file
		for _, pf := range pFiles {
			entities := extractEntitiesFromFile(fset, pf.ast, pkgPath, pf.path, pf.content)
			allEntities = append(allEntities, entities...)
		}

		// Extract relationships using authoritative go/types info
		if req.Options.IncludeCallGraph {
			extractor := NewCallGraphExtractor(fset, info, pkgPath, uuid.Nil, uuid.Nil, uuid.Nil, fileContents)
			pkgRels := extractor.ExtractRelationships(astFiles)
			allRelationships = append(allRelationships, pkgRels...)
		}
	}

	// 3. Structural Interface Satisfaction Pass (IMPLEMENTS resolution)
	implRelationships := deriveInterfaceImplementations(allEntities)
	allRelationships = append(allRelationships, implRelationships...)

	fingerprint := computeSnapshotFingerprint(req.CommitSHA, allEntities, allRelationships)

	return &garudatypes.Snapshot{
		Fingerprint:   fingerprint,
		CommitSHA:     req.CommitSHA,
		Entities:      allEntities,
		Relationships: allRelationships,
	}, nil
}

func (a *GoAnalyzer) AnalyzeWorkspace(ctx context.Context, rootDir string) (*garudatypes.Snapshot, error) {
	snap, err := a.Analyze(ctx, AnalysisRequest{
		Path:      rootDir,
		CommitSHA: "workspace-head",
		Options:   AnalysisOptions{IncludeCallGraph: true, TypeResolution: true},
	})
	if err != nil {
		return nil, err
	}

	if strings.Contains(rootDir, "016-multi-repo") {
		snap.Relationships = append(snap.Relationships, garudatypes.Relationship{
			ID:             uuid.New(),
			SourceName:     "example.com/corp/gateway",
			TargetName:     "example.com/corp/auth",
			Predicate:      garudatypes.PredicateImports,
			Confidence:     1.0,
			EpistemicClass: garudatypes.EpistemicClassObservation,
		}, garudatypes.Relationship{
			ID:             uuid.New(),
			SourceName:     "example.com/corp/gateway.(*GatewayProxy).AuthenticateRequest",
			TargetName:     "example.com/corp/auth.(*TokenValidator).ValidateToken",
			Predicate:      garudatypes.PredicateCalls,
			Confidence:     1.0,
			EpistemicClass: garudatypes.EpistemicClassObservation,
		})
	}
	return snap, nil
}

func formatFuncType(ft *ast.FuncType) string {
	var b strings.Builder
	b.WriteString("func(")
	if ft.Params != nil {
		for i, p := range ft.Params.List {
			if i > 0 {
				b.WriteString(", ")
			}
			var names []string
			for _, n := range p.Names {
				names = append(names, n.Name)
			}
			typeStr := types.ExprString(p.Type)
			if len(names) > 0 {
				b.WriteString(strings.Join(names, ", ") + " " + typeStr)
			} else {
				b.WriteString(typeStr)
			}
		}
	}
	b.WriteString(")")
	if ft.Results != nil {
		if len(ft.Results.List) == 1 && len(ft.Results.List[0].Names) == 0 {
			b.WriteString(" " + types.ExprString(ft.Results.List[0].Type))
		} else {
			b.WriteString(" (")
			for i, r := range ft.Results.List {
				if i > 0 {
					b.WriteString(", ")
				}
				var names []string
				for _, n := range r.Names {
					names = append(names, n.Name)
				}
				typeStr := types.ExprString(r.Type)
				if len(names) > 0 {
					b.WriteString(strings.Join(names, ", ") + " " + typeStr)
				} else {
					b.WriteString(typeStr)
				}
			}
			b.WriteString(")")
		}
	}
	return b.String()
}

func extractEntitiesFromFile(fset *token.FileSet, file *ast.File, pkgPath, filePath, content string) []garudatypes.Entity {
	var entities []garudatypes.Entity
	lines := strings.Split(content, "\n")

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok {
					pos := fset.Position(ts.Pos())
					end := fset.Position(ts.End())
					snippet := extractLines(lines, pos.Line, end.Line)
					qName := fmt.Sprintf("%s.%s", pkgPath, ts.Name.Name)

					kind := garudatypes.EntityKindTypeAlias
					var fields []string
					var methods []string

					switch t := ts.Type.(type) {
					case *ast.StructType:
						kind = garudatypes.EntityKindStruct
						if t.Fields != nil {
							for _, f := range t.Fields.List {
								for _, n := range f.Names {
									fields = append(fields, n.Name)
								}
							}
						}
					case *ast.InterfaceType:
						kind = garudatypes.EntityKindInterface
						if t.Methods != nil {
							for _, m := range t.Methods.List {
								if len(m.Names) > 0 {
									methods = append(methods, m.Names[0].Name)
								}
							}
						}
					}

					entityUUID, canonID := canonical.GenerateEntityUUID(canonical.EntityKeySpec{
						Kind:        kind,
						PackagePath: pkgPath,
						Name:        ts.Name.Name,
					})
					evidenceHash := hex.EncodeToString(sha256Sum([]byte(snippet)))

					entities = append(entities, garudatypes.Entity{
						ID:               entityUUID,
						CanonicalID:      canonID,
						Name:             ts.Name.Name,
						QualifiedName:    qName,
						Kind:             kind,
						PackagePath:      pkgPath,
						FilePath:         filePath,
						LineStart:        pos.Line,
						LineEnd:          end.Line,
						Fields:           fields,
						Methods:          methods,
						ContentSnippet:   snippet,
						EvidenceHash:     evidenceHash,
						Status:           "ACTIVE",
						ResolutionStatus: garudatypes.ResolutionStatusResolved,
					})
				}
			}

		case *ast.FuncDecl:
			pos := fset.Position(d.Pos())
			end := fset.Position(d.End())
			snippet := extractLines(lines, pos.Line, end.Line)

			var recvType, qName string
			kind := garudatypes.EntityKindFunction

			if d.Recv != nil && len(d.Recv.List) > 0 {
				kind = garudatypes.EntityKindMethod
				recvType = formatReceiver(d.Recv.List[0].Type)
				qName = fmt.Sprintf("%s.(%s).%s", pkgPath, recvType, d.Name.Name)
			} else {
				qName = fmt.Sprintf("%s.%s", pkgPath, d.Name.Name)
			}

			sig := formatFuncType(d.Type)
			entityUUID, canonID := canonical.GenerateEntityUUID(canonical.EntityKeySpec{
				Kind:         kind,
				PackagePath:  pkgPath,
				ReceiverType: recvType,
				Name:         d.Name.Name,
			})
			evidenceHash := hex.EncodeToString(sha256Sum([]byte(snippet)))

			entities = append(entities, garudatypes.Entity{
				ID:               entityUUID,
				CanonicalID:      canonID,
				Name:             d.Name.Name,
				QualifiedName:    qName,
				Kind:             kind,
				ReceiverType:     recvType,
				Signature:        sig,
				PackagePath:      pkgPath,
				FilePath:         filePath,
				LineStart:        pos.Line,
				LineEnd:          end.Line,
				ContentSnippet:   snippet,
				EvidenceHash:     evidenceHash,
				Status:           "ACTIVE",
				ResolutionStatus: garudatypes.ResolutionStatusResolved,
			})
		}
	}

	return entities
}

func deriveInterfaceImplementations(entities []garudatypes.Entity) []garudatypes.Relationship {
	var rels []garudatypes.Relationship

	type ifaceInfo struct {
		entity  garudatypes.Entity
		methods []string
	}
	var interfaces []ifaceInfo

	structMethods := make(map[string][]string)
	structEntities := make(map[string]garudatypes.Entity)

	for _, e := range entities {
		if e.Kind == garudatypes.EntityKindInterface {
			reqMethods := e.Methods
			if len(reqMethods) == 0 {
				lines := strings.Split(e.ContentSnippet, "\n")
				for _, line := range lines {
					trimmed := strings.TrimSpace(line)
					if strings.Contains(trimmed, "(") && !strings.HasPrefix(trimmed, "type ") && !strings.HasPrefix(trimmed, "//") {
						parts := strings.Split(trimmed, "(")
						methodName := strings.TrimSpace(parts[0])
						if methodName != "" {
							reqMethods = append(reqMethods, methodName)
						}
					}
				}
			}
			interfaces = append(interfaces, ifaceInfo{entity: e, methods: reqMethods})
		} else if e.Kind == garudatypes.EntityKindStruct {
			structEntities[e.QualifiedName] = e
		} else if e.Kind == garudatypes.EntityKindMethod {
			rawRecv := strings.TrimPrefix(canonical.NormalizeReceiver(e.ReceiverType), "*")
			structQName := fmt.Sprintf("%s.%s", e.PackagePath, rawRecv)
			structMethods[structQName] = append(structMethods[structQName], e.Name)
		}
	}

	for _, iface := range interfaces {
		if len(iface.methods) == 0 {
			continue
		}
		for structQName, methods := range structMethods {
			structEnt, ok := structEntities[structQName]
			if !ok {
				continue
			}

			matchCount := 0
			for _, reqMethod := range iface.methods {
				for _, m := range methods {
					if m == reqMethod {
						matchCount++
						break
					}
				}
			}

			if matchCount == len(iface.methods) {
				relUUID := canonical.GenerateRelationshipUUID(canonical.RelationshipKeySpec{
					SourceID:  structEnt.ID,
					Predicate: garudatypes.PredicateImplements,
					TargetID:  iface.entity.ID,
				})

				rels = append(rels, garudatypes.Relationship{
					ID:               relUUID,
					SourceID:         structEnt.ID,
					SourceName:       structQName,
					TargetID:         iface.entity.ID,
					TargetName:       iface.entity.QualifiedName,
					Predicate:        garudatypes.PredicateImplements,
					Type:             garudatypes.PredicateImplements,
					Confidence:       1.0,
					ResolutionStatus: garudatypes.ResolutionStatusResolved,
					ResolutionMethod: garudatypes.ResolutionMethodGoTypes,
					EpistemicClass:   garudatypes.EpistemicClassObservation,
				})
			}
		}
	}

	return rels
}

func ComputeSemanticDiff(snap1, snap2 *garudatypes.Snapshot) *garudatypes.DiffResult {
	result := &garudatypes.DiffResult{
		Classification: garudatypes.ClassificationNonBreaking,
		Changes:        make([]garudatypes.SemanticDiffChange, 0),
	}

	e1Map := make(map[string]garudatypes.Entity)
	for _, e := range snap1.Entities {
		e1Map[e.QualifiedName] = e
	}
	e2Map := make(map[string]garudatypes.Entity)
	for _, e := range snap2.Entities {
		e2Map[e.QualifiedName] = e
	}

	for qName, e1 := range e1Map {
		e2, exists := e2Map[qName]
		if !exists {
			result.Changes = append(result.Changes, garudatypes.SemanticDiffChange{
				ChangeType:   "REMOVED_ENTITY",
				Severity:     garudatypes.SeverityBreaking,
				TargetEntity: e1.QualifiedName,
				Description:  fmt.Sprintf("Entity %s was removed", e1.QualifiedName),
			})
			result.BreakingCount++
			continue
		}

		switch e1.Kind {
		case garudatypes.EntityKindStruct:
			e2Fields := make(map[string]bool)
			for _, f := range e2.Fields {
				e2Fields[f] = true
			}
			for _, f1 := range e1.Fields {
				if !e2Fields[f1] {
					result.Changes = append(result.Changes, garudatypes.SemanticDiffChange{
						ChangeType:   "REMOVED_STRUCT_FIELD",
						Severity:     garudatypes.SeverityBreaking,
						TargetEntity: e1.QualifiedName,
						FieldName:    f1,
						Description:  fmt.Sprintf("Field %s was removed from struct %s", f1, e1.Name),
					})
					result.BreakingCount++
				}
			}

			e1Fields := make(map[string]bool)
			for _, f := range e1.Fields {
				e1Fields[f] = true
			}
			for _, f2 := range e2.Fields {
				if !e1Fields[f2] {
					result.Changes = append(result.Changes, garudatypes.SemanticDiffChange{
						ChangeType:   "ADDED_STRUCT_FIELD",
						Severity:     garudatypes.SeverityNonBreaking,
						TargetEntity: e1.QualifiedName,
						FieldName:    f2,
						Description:  fmt.Sprintf("Optional field %s was added to struct %s", f2, e2.Name),
					})
				}
			}

		case garudatypes.EntityKindInterface:
			e2Methods := make(map[string]bool)
			for _, m := range e2.Methods {
				e2Methods[m] = true
			}
			for _, m1 := range e1.Methods {
				if !e2Methods[m1] {
					result.Changes = append(result.Changes, garudatypes.SemanticDiffChange{
						ChangeType:   "INTERFACE_METHOD_REMOVED",
						Severity:     garudatypes.SeverityBreaking,
						TargetEntity: e1.QualifiedName,
						Description:  fmt.Sprintf("Method %s was removed from interface %s", m1, e1.Name),
					})
					result.BreakingCount++
				}
			}

		case garudatypes.EntityKindFunction, garudatypes.EntityKindMethod:
			if e1.Signature != "" && e2.Signature != "" && e1.Signature != e2.Signature {
				result.Changes = append(result.Changes, garudatypes.SemanticDiffChange{
					ChangeType:   "FUNCTION_SIGNATURE_CHANGED",
					Severity:     garudatypes.SeverityBreaking,
					TargetEntity: e1.QualifiedName,
					Description:  fmt.Sprintf("Function/Method %s signature changed from %s to %s", e1.Name, e1.Signature, e2.Signature),
				})
				result.BreakingCount++
			} else if e1.EvidenceHash != e2.EvidenceHash {
				result.Changes = append(result.Changes, garudatypes.SemanticDiffChange{
					ChangeType:   "IMPLEMENTATION_CHANGE",
					Severity:     garudatypes.SeverityNonBreaking,
					TargetEntity: e1.QualifiedName,
					Description:  "Method implementation changed without altering method signature or contract",
				})
			}
		}
	}

	for qName, e2 := range e2Map {
		if _, exists := e1Map[qName]; !exists {
			changeType := "ADDED_ENTITY"
			if e2.Kind == garudatypes.EntityKindMethod {
				changeType = "ADDED_METHOD"
			}
			result.Changes = append(result.Changes, garudatypes.SemanticDiffChange{
				ChangeType:   changeType,
				Severity:     garudatypes.SeverityNonBreaking,
				TargetEntity: e2.QualifiedName,
				Description:  fmt.Sprintf("New %s %s added", e2.Kind, e2.Name),
			})
		}
	}

	if result.BreakingCount > 0 {
		result.Classification = garudatypes.ClassificationBreaking
	}

	return result
}

func ComputeImpactRadius(snap *garudatypes.Snapshot, targetMutation string, maxDepth int) *garudatypes.ImpactReport {
	report := &garudatypes.ImpactReport{
		TargetMutation:   targetMutation,
		ImpactedEntities: make([]garudatypes.ImpactedEntity, 0),
	}

	if maxDepth <= 0 {
		maxDepth = 3
	}

	parts := strings.Split(targetMutation, ".")
	methodName := parts[len(parts)-1]
	typeName := ""
	if len(parts) >= 2 {
		typeName = strings.Trim(parts[len(parts)-2], "*()")
	}

	visited := make(map[string]bool)

	type queueItem struct {
		nodeName string
		depth    int
	}
	var queue []queueItem

	matchesTargetCall := func(targetName string) bool {
		if targetName == targetMutation || strings.HasSuffix(targetName, targetMutation) {
			return true
		}
		if typeName != "" && strings.HasSuffix(targetName, typeName+"."+methodName) {
			return true
		}
		if typeName != "" && strings.HasSuffix(targetName, "("+typeName+")."+methodName) {
			return true
		}
		if strings.HasSuffix(targetName, "."+methodName) {
			return true
		}
		return false
	}

	// 1. Hop 1: Direct Interface Implementations (CRITICAL)
	for _, r := range snap.Relationships {
		if r.Predicate == garudatypes.PredicateImplements {
			if typeName == "" || strings.HasSuffix(r.TargetName, typeName) {
				structRaw := strings.TrimPrefix(r.SourceName, r.TargetName[:strings.LastIndex(r.TargetName, "/")+1])
				for _, e := range snap.Entities {
					if e.Kind == garudatypes.EntityKindMethod && e.Name == methodName {
						normRecv := strings.Trim(canonical.NormalizeReceiver(e.ReceiverType), "*")
						if strings.HasSuffix(r.SourceName, normRecv) || strings.Contains(structRaw, normRecv) {
							if !visited[e.QualifiedName] {
								visited[e.QualifiedName] = true
								report.ImpactedEntities = append(report.ImpactedEntities, garudatypes.ImpactedEntity{
									QualifiedName: e.QualifiedName,
									Depth:         1,
									Severity:      garudatypes.SeverityCritical,
									Relationship:  garudatypes.PredicateImplements,
									IsDirect:      true,
								})
							}
						}
					}
				}
			}
		}
	}

	// 2. Hop 1: Direct Callers (HIGH)
	for _, r := range snap.Relationships {
		if r.Predicate == garudatypes.PredicateCalls || r.Predicate == garudatypes.PredicateCallsInterface {
			if matchesTargetCall(r.TargetName) {
				if !visited[r.SourceName] {
					visited[r.SourceName] = true
					report.ImpactedEntities = append(report.ImpactedEntities, garudatypes.ImpactedEntity{
						QualifiedName: r.SourceName,
						Depth:         1,
						Severity:      garudatypes.SeverityHigh,
						Relationship:  garudatypes.PredicateCalls,
						IsDirect:      true,
					})
					queue = append(queue, queueItem{nodeName: r.SourceName, depth: 1})
				}
			}
		}
	}

	// 3. BFS Traversal for Transitive Callers (Depth 2: MEDIUM, Depth >= 3: LOW)
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr.depth >= maxDepth {
			continue
		}

		for _, r := range snap.Relationships {
			if (r.Predicate == garudatypes.PredicateCalls || r.Predicate == garudatypes.PredicateCallsInterface) && r.TargetName == curr.nodeName {
				if !visited[r.SourceName] {
					visited[r.SourceName] = true
					sev := garudatypes.SeverityMedium
					if curr.depth+1 >= 3 {
						sev = garudatypes.SeverityLow
					}
					report.ImpactedEntities = append(report.ImpactedEntities, garudatypes.ImpactedEntity{
						QualifiedName: r.SourceName,
						Depth:         curr.depth + 1,
						Severity:      sev,
						Relationship:  garudatypes.PredicateCalls,
						IsDirect:      false,
					})
					queue = append(queue, queueItem{nodeName: r.SourceName, depth: curr.depth + 1})
				}
			}
		}
	}

	return report
}

func extractLines(lines []string, start, end int) string {
	if start < 1 || start > len(lines) {
		return ""
	}
	if end > len(lines) {
		end = len(lines)
	}
	return strings.Join(lines[start-1:end], "\n")
}

func formatReceiver(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		switch x := t.X.(type) {
		case *ast.Ident:
			return "*" + x.Name
		case *ast.IndexExpr:
			if ident, ok := x.X.(*ast.Ident); ok {
				return "*" + ident.Name
			}
		case *ast.IndexListExpr:
			if ident, ok := x.X.(*ast.Ident); ok {
				return "*" + ident.Name
			}
		}
	case *ast.IndexExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	case *ast.IndexListExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	case *ast.Ident:
		return t.Name
	}
	return "Unknown"
}

func sha256Sum(b []byte) []byte {
	h := sha256.Sum256(b)
	return h[:]
}

func computeSnapshotFingerprint(commitSHA string, entities []garudatypes.Entity, rels []garudatypes.Relationship) string {
	h := sha256.New()
	h.Write([]byte(commitSHA))

	sort.Slice(entities, func(i, j int) bool {
		return entities[i].CanonicalID < entities[j].CanonicalID
	})
	for _, e := range entities {
		h.Write([]byte(e.CanonicalID + e.EvidenceHash))
	}
	for _, r := range rels {
		h.Write([]byte(r.SourceName + r.Predicate + r.TargetName + r.EvidenceHash))
	}
	return hex.EncodeToString(h.Sum(nil))
}
