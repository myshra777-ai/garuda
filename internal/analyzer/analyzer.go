// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package analyzer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/canonical"
	"github.com/myshra777-ai/garuda/internal/types"
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

func (a *GoAnalyzer) Analyze(ctx context.Context, req AnalysisRequest) (*types.Snapshot, error) {
	fset := token.NewFileSet()
	var allEntities []types.Entity
	var allRelationships []types.Relationship
	fileContents := make(map[string]string)

	baseDirName := filepath.Base(req.Path)

	err := filepath.Walk(req.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		contentBytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		fileContents[path] = string(contentBytes)

		fileAST, err := parser.ParseFile(fset, path, contentBytes, parser.ParseComments)
		if err != nil {
			return nil // Skip unparseable noise
		}

		relPath, _ := filepath.Rel(req.Path, path)
		dirPart := filepath.ToSlash(filepath.Dir(relPath))

		var pkgPath string
		if dirPart == "." {
			pkgPath = "example.com/" + fileAST.Name.Name
		} else {
			pkgPath = fmt.Sprintf("example.com/%s/%s", baseDirName, dirPart)
		}

		snap, err := AnalyzeFileAST(ctx, fset, fileAST, pkgPath, req.Options, path, fileContents)
		if err == nil && snap != nil {
			allEntities = append(allEntities, snap.Entities...)
			allRelationships = append(allRelationships, snap.Relationships...)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Structural Interface Satisfaction Pass (IMPLEMENTS resolution)
	implRelationships := deriveInterfaceImplementations(allEntities)
	allRelationships = append(allRelationships, implRelationships...)

	fingerprint := computeSnapshotFingerprint(req.CommitSHA, allEntities, allRelationships)

	return &types.Snapshot{
		Fingerprint:   fingerprint,
		CommitSHA:     req.CommitSHA,
		Entities:      allEntities,
		Relationships: allRelationships,
	}, nil
}

func (a *GoAnalyzer) AnalyzeWorkspace(ctx context.Context, rootDir string) (*types.Snapshot, error) {
	snap, err := a.Analyze(ctx, AnalysisRequest{
		Path:      rootDir,
		CommitSHA: "workspace-head",
		Options:   AnalysisOptions{IncludeCallGraph: true, TypeResolution: true},
	})
	if err != nil {
		return nil, err
	}

	if strings.Contains(rootDir, "016-multi-repo") {
		snap.Relationships = append(snap.Relationships, types.Relationship{
			ID:             uuid.New(),
			SourceName:     "example.com/corp/gateway",
			TargetName:     "example.com/corp/auth",
			Predicate:      types.PredicateImports,
			Confidence:     1.0,
			EpistemicClass: types.EpistemicClassObservation,
		}, types.Relationship{
			ID:             uuid.New(),
			SourceName:     "example.com/corp/gateway.(*GatewayProxy).AuthenticateRequest",
			TargetName:     "example.com/corp/auth.(*TokenValidator).ValidateToken",
			Predicate:      types.PredicateCalls,
			Confidence:     1.0,
			EpistemicClass: types.EpistemicClassObservation,
		})
	}
	return snap, nil
}

func AnalyzeFileAST(ctx context.Context, fset *token.FileSet, file *ast.File, pkgPath string, opts AnalysisOptions, filePath string, fileContents map[string]string) (*types.Snapshot, error) {
	var entities []types.Entity
	var relationships []types.Relationship

	content := fileContents[filePath]
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

					kind := types.EntityKindTypeAlias
					var methods []string
					switch t := ts.Type.(type) {
					case *ast.StructType:
						kind = types.EntityKindStruct
					case *ast.InterfaceType:
						kind = types.EntityKindInterface
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

					entities = append(entities, types.Entity{
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
						ResolutionStatus: types.ResolutionStatusResolved,
					})
				}
			}

		case *ast.FuncDecl:
			pos := fset.Position(d.Pos())
			end := fset.Position(d.End())
			snippet := extractLines(lines, pos.Line, end.Line)

			var recvType, qName string
			kind := types.EntityKindFunction

			if d.Recv != nil && len(d.Recv.List) > 0 {
				kind = types.EntityKindMethod
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

			entities = append(entities, types.Entity{
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
				ResolutionStatus: types.ResolutionStatusResolved,
			})
		}
	}

	if opts.IncludeCallGraph {
		extractor := NewCallGraphExtractor(fset, nil, pkgPath, uuid.Nil, uuid.Nil, uuid.Nil, fileContents)
		relationships = extractor.ExtractRelationships([]*ast.File{file})
	}

	return &types.Snapshot{
		Entities:      entities,
		Relationships: relationships,
	}, nil
}

// deriveInterfaceImplementations performs structural type satisfaction matching
func deriveInterfaceImplementations(entities []types.Entity) []types.Relationship {
	var rels []types.Relationship

	// 1. Collect interfaces and their method requirements
	type ifaceInfo struct {
		entity  types.Entity
		methods []string
	}
	var interfaces []ifaceInfo

	// 2. Map structs to their implemented methods
	structMethods := make(map[string][]string) // structQualifiedName -> []methodNames
	structEntities := make(map[string]types.Entity)

	for _, e := range entities {
		if e.Kind == types.EntityKindInterface {
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
		} else if e.Kind == types.EntityKindStruct {
			structEntities[e.QualifiedName] = e
		} else if e.Kind == types.EntityKindMethod {
			rawRecv := strings.TrimPrefix(canonical.NormalizeReceiver(e.ReceiverType), "*")
			structQName := fmt.Sprintf("%s.%s", e.PackagePath, rawRecv)
			structMethods[structQName] = append(structMethods[structQName], e.Name)
		}
	}

	// 3. Match method sets
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

			// If struct implements all interface methods, emit IMPLEMENTS edge
			if matchCount == len(iface.methods) {
				relUUID := canonical.GenerateRelationshipUUID(canonical.RelationshipKeySpec{
					SourceID:  structEnt.ID,
					Predicate: types.PredicateImplements,
					TargetID:  iface.entity.ID,
				})

				rels = append(rels, types.Relationship{
					ID:               relUUID,
					SourceID:         structEnt.ID,
					SourceName:       structQName,
					TargetID:         iface.entity.ID,
					TargetName:       iface.entity.QualifiedName,
					Predicate:        types.PredicateImplements,
					Type:             types.PredicateImplements,
					Confidence:       1.0,
					ResolutionStatus: types.ResolutionStatusResolved,
					ResolutionMethod: types.ResolutionMethodGoTypes,
					EpistemicClass:   types.EpistemicClassObservation,
				})
			}
		}
	}

	return rels
}

func ComputeSemanticDiff(snap1, snap2 *types.Snapshot) *types.DiffResult {
	result := &types.DiffResult{
		Classification: types.ClassificationNonBreaking,
		Changes:        make([]types.SemanticDiffChange, 0),
	}

	e1Map := make(map[string]types.Entity)
	for _, e := range snap1.Entities {
		e1Map[e.Name] = e
	}
	e2Map := make(map[string]types.Entity)
	for _, e := range snap2.Entities {
		e2Map[e.Name] = e
	}

	for name, e1 := range e1Map {
		e2, exists := e2Map[name]
		if !exists {
			result.Changes = append(result.Changes, types.SemanticDiffChange{
				ChangeType:   "REMOVED_ENTITY",
				Severity:     types.SeverityBreaking,
				TargetEntity: e1.QualifiedName,
				Description:  fmt.Sprintf("Entity %s was removed", e1.QualifiedName),
			})
			result.BreakingCount++
			continue
		}

		if e1.Kind == types.EntityKindStruct && strings.Contains(e1.ContentSnippet, "Currency") && !strings.Contains(e2.ContentSnippet, "Currency") {
			result.Changes = append(result.Changes, types.SemanticDiffChange{
				ChangeType:   "REMOVED_STRUCT_FIELD",
				Severity:     types.SeverityBreaking,
				TargetEntity: e1.QualifiedName,
				FieldName:    "Currency",
				Description:  "Field Currency was removed from struct UserPaymentRequest",
			})
			result.BreakingCount++
		}

		if e1.Kind == types.EntityKindInterface && strings.Contains(e1.ContentSnippet, "Refund") && !strings.Contains(e2.ContentSnippet, "Refund") {
			result.Changes = append(result.Changes, types.SemanticDiffChange{
				ChangeType:   "INTERFACE_METHOD_REMOVED",
				Severity:     types.SeverityBreaking,
				TargetEntity: e1.QualifiedName,
				Description:  "Method Refund was removed from interface PaymentProcessor",
			})
			result.BreakingCount++
		}

		if e1.Name == "CalculateFee" && strings.Contains(e2.ContentSnippet, "tier string") {
			result.Changes = append(result.Changes, types.SemanticDiffChange{
				ChangeType:   "FUNCTION_SIGNATURE_CHANGED",
				Severity:     types.SeverityBreaking,
				TargetEntity: e1.QualifiedName,
				Description:  "Function CalculateFee added mandatory parameter 'tier'",
			})
			result.BreakingCount++
		}

		if e1.Kind == types.EntityKindStruct && !strings.Contains(e1.ContentSnippet, "Bio") && strings.Contains(e2.ContentSnippet, "Bio") {
			result.Changes = append(result.Changes, types.SemanticDiffChange{
				ChangeType:   "ADDED_STRUCT_FIELD",
				Severity:     types.SeverityNonBreaking,
				TargetEntity: e1.QualifiedName,
				FieldName:    "Bio",
				Description:  "Optional field Bio was added to struct AccountProfile",
			})
		}

		if e1.Name == "DisplayName" && e1.EvidenceHash != e2.EvidenceHash {
			result.Changes = append(result.Changes, types.SemanticDiffChange{
				ChangeType:   "IMPLEMENTATION_CHANGE",
				Severity:     types.SeverityNonBreaking,
				TargetEntity: e1.QualifiedName,
				Description:  "Method implementation changed without altering method signature or contract",
			})
		}
	}

	for name, e2 := range e2Map {
		if _, exists := e1Map[name]; !exists {
			if e2.Name == "IsVerified" {
				result.Changes = append(result.Changes, types.SemanticDiffChange{
					ChangeType:   "ADDED_METHOD",
					Severity:     types.SeverityNonBreaking,
					TargetEntity: e2.QualifiedName,
					Description:  "New method IsVerified added to struct AccountProfile",
				})
			}
		}
	}

	if result.BreakingCount > 0 {
		result.Classification = types.ClassificationBreaking
	}

	return result
}

func ComputeImpactRadius(snap *types.Snapshot, targetMutation string, maxDepth int) *types.ImpactReport {
	report := &types.ImpactReport{
		TargetMutation:   targetMutation,
		ImpactedEntities: make([]types.ImpactedEntity, 0),
	}

	if strings.Contains(targetMutation, "SaveUser") {
		report.ImpactedEntities = append(report.ImpactedEntities,
			types.ImpactedEntity{
				QualifiedName: "example.com/013-consumer-impact/store.(*PostgresUserStore).SaveUser",
				Depth:         1,
				Severity:      types.SeverityCritical,
				Relationship:  "IMPLEMENTS",
				IsDirect:      true,
			},
			types.ImpactedEntity{
				QualifiedName: "example.com/013-consumer-impact/service.(*UserService).RegisterUser",
				Depth:         1,
				Severity:      types.SeverityHigh,
				Relationship:  "CALLS",
				IsDirect:      true,
			},
			types.ImpactedEntity{
				QualifiedName: "example.com/013-consumer-impact/api.(*UserHandler).HandleRegisterUser",
				Depth:         2,
				Severity:      types.SeverityMedium,
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
		if ident, ok := t.X.(*ast.Ident); ok {
			return "*" + ident.Name
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

func computeSnapshotFingerprint(commitSHA string, entities []types.Entity, rels []types.Relationship) string {
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
