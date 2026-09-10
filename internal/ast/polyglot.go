package ast

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Language string

const (
	LangGo         Language = "go"
	LangPython     Language = "python"
	LangTypeScript Language = "typescript"
	LangRust       Language = "rust"
)

type CodeEntity struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	Workspace  string
	Path       string
	SymbolName string
	Kind       string
	Language   Language
}

type CodeRelationship struct {
	TenantID uuid.UUID
	SourceID uuid.UUID
	TargetID uuid.UUID
	Kind     string
}

type PolyglotAnalyzer struct {
	pool *pgxpool.Pool
}

func NewPolyglotAnalyzer(pool *pgxpool.Pool) *PolyglotAnalyzer {
	return &PolyglotAnalyzer{pool: pool}
}

func DetectLanguage(path string) Language {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return LangGo
	case ".py":
		return LangPython
	case ".ts", ".tsx", ".js":
		return LangTypeScript
	case ".rs":
		return LangRust
	default:
		return ""
	}
}

func (a *PolyglotAnalyzer) AnalyzeDirectory(ctx context.Context, tenantID uuid.UUID, workspace string, rootDir string) error {
	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, _ = tx.Exec(ctx, `DELETE FROM entities WHERE tenant_id = $1`, tenantID)
	_, _ = tx.Exec(ctx, `DELETE FROM relationships WHERE tenant_id = $1`, tenantID)

	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "vendor" || info.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		lang := DetectLanguage(path)
		if lang == "" {
			return nil
		}

		return a.parseFileSymbols(ctx, tx, tenantID, workspace, path, lang)
	})

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (a *PolyglotAnalyzer) parseFileSymbols(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, workspace string, filePath string, lang Language) error {
	if _, err := os.ReadFile(filePath); err != nil {
		return err
	}

	entityID := uuid.New()
	symbolName := fmt.Sprintf("%s.%s", lang, filepath.Base(filePath))
	
	_, err := tx.Exec(ctx, `
		INSERT INTO entities (id, tenant_id, name, kind, path, language)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT DO NOTHING
	`, entityID, tenantID, symbolName, "module", filePath, string(lang))

	return err
}
