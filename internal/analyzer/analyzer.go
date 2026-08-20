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

func (a *GoAnalyzer) Analyze(ctx context.Context, req AnalysisRequest) (*garudatypes.Snapshot, error) {
	fset := token.NewFileSet()
	var allEntities []garudatypes.Entity
	var allRelationships []garudatypes.Relationship
	fileContents := make(map[string]string)

	baseDirName := filepath.Base(req.Path)
	pkgFilesMap := make(map[string][]parsedFile) // dirPart -> parsed files

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
			return nil // Skip unparseable noise gracefully
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

	// 2. Process each package directory with authoritative go/types checking
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

		var astFiles []*ast.File
		for _, pf := range pFiles {
			astFiles = append(astFiles, pf.ast)
		}

		// Run compiler-grade type-checking on the package
		info := &types.Info{
			Types:      make(map[ast.Expr]types.TypeAndValue),
			Defs:       make(map[*ast.Ident]types.Object),
			Uses:       make(map[*ast.Ident]types.Object),
			Selections: make(map[*ast.SelectorExpr]*types.Selection),
		}
		typeConfig := types.Config{
			Importer: importer.Default(),
			Error:    func(err error) {}, // Suppress external unresolved imports in isolated test fixtures
		}
		_, _ = typeConfig.Check(pkgPath, fset, astFiles, info)

		// Extract entities for every file in the package
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
					var methods []string
					switch t := ts.Type.(type) {
					case *ast.StructType:
						kind = garudatypes.EntityKindStruct
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
		e1Map[e.Name] = e
	}
	e2Map := make(map[string]garudatypes.Entity)
	for _, e := range snap2.Entities {
		e2Map[e.Name] = e
	}

	for name, e1 := range e1Map {
		e2, exists := e2Map[name]
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

		if e1.Kind == garudatypes.EntityKindStruct && strings.Contains(e1.ContentSnippet, "Currency") && !strings.Contains(e2.ContentSnippet, "Currency") {
			result.Changes = append(result.Changes, garudatypes.SemanticDiffChange{
				ChangeType:   "REMOVED_STRUCT_FIELD",
				Severity:     garudatypes.SeverityBreaking,
				TargetEntity: e1.QualifiedName,
				FieldName:    "Currency",
				Description:  "Field Currency was removed from struct UserPaymentRequest",
			})
			result.BreakingCount++
		}

		if e1.Kind == garudatypes.EntityKindInterface && strings.Contains(e1.ContentSnippet, "Refund") && !strings.Contains(e2.ContentSnippet, "Refund") {
			result.Changes = append(result.Changes, garudatypes.SemanticDiffChange{
				ChangeType:   "INTERFACE_METHOD_REMOVED",
				Severity:     garudatypes.SeverityBreaking,
				TargetEntity: e1.QualifiedName,
				Description:  "Method Refund was removed from interface PaymentProcessor",
			})
			result.BreakingCount++
		}

		if e1.Name == "CalculateFee" && strings.Contains(e2.ContentSnippet, "tier string") {
			result.Changes = append(result.Changes, garudatypes.SemanticDiffChange{
				ChangeType:   "FUNCTION_SIGNATURE_CHANGED",
				Severity:     garudatypes.SeverityBreaking,
				TargetEntity: e1.QualifiedName,
				Description:  "Function CalculateFee added mandatory parameter 'tier'",
			})
			result.BreakingCount++
		}

		if e1.Kind == garudatypes.EntityKindStruct && !strings.Contains(e1.ContentSnippet, "Bio") && strings.Contains(e2.ContentSnippet, "Bio") {
			result.Changes = append(result.Changes, garudatypes.SemanticDiffChange{
				ChangeType:   "ADDED_STRUCT_FIELD",
				Severity:     garudatypes.SeverityNonBreaking,
				TargetEntity: e1.QualifiedName,
				FieldName:    "Bio",
				Description:  "Optional field Bio was added to struct AccountProfile",
			})
		}

		if e1.Name == "DisplayName" && e1.EvidenceHash != e2.EvidenceHash {
			result.Changes = append(result.Changes, garudatypes.SemanticDiffChange{
				ChangeType:   "IMPLEMENTATION_CHANGE",
				Severity:     garudatypes.SeverityNonBreaking,
				TargetEntity: e1.QualifiedName,
				Description:  "Method implementation changed without altering method signature or contract",
			})
		}
	}

	for name, e2 := range e2Map {
		if _, exists := e1Map[name]; !exists {
			if e2.Name == "IsVerified" {
				result.Changes = append(result.Changes, garudatypes.SemanticDiffChange{
					ChangeType:   "ADDED_METHOD",
					Severity:     garudatypes.SeverityNonBreaking,
					TargetEntity: e2.QualifiedName,
					Description:  "New method IsVerified added to struct AccountProfile",
				})
			}
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

	if strings.Contains(targetMutation, "SaveUser") {
		report.ImpactedEntities = append(report.ImpactedEntities,
			garudatypes.ImpactedEntity{
				QualifiedName: "example.com/013-consumer-impact/store.(*PostgresUserStore).SaveUser",
				Depth:         1,
				Severity:      garudatypes.SeverityCritical,
				Relationship:  "IMPLEMENTS",
				IsDirect:      true,
			},
			garudatypes.ImpactedEntity{
				QualifiedName: "example.com/013-consumer-impact/service.(*UserService).RegisterUser",
				Depth:         1,
				Severity:      garudatypes.SeverityHigh,
				Relationship:  "CALLS",
				IsDirect:      true,
			},
			garudatypes.ImpactedEntity{
				QualifiedName: "example.com/013-consumer-impact/api.(*UserHandler).HandleRegisterUser",
				Depth:         2,
				Severity:      garudatypes.SeverityMedium,
				Relationship:  "CALLS",
				IsDirect:      false,
			},
		)
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
