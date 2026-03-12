package architecture_test

import (
	"bufio"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
	"testing"
)

type depRule struct {
	importerPrefix string
	forbidden      []string
}

// Temporary legacy allowances. Keep this list short and remove entries during refactoring.
var legacyAllowed = map[string][]string{}

func TestDependencyRules(t *testing.T) {
	repoRoot := findRepoRoot(t)
	modulePath := readModulePath(t, repoRoot)

	rules := []depRule{
		{
			importerPrefix: modulePath + "/internal/domain",
			forbidden: []string{
				modulePath + "/internal/http",
				modulePath + "/internal/repos",
				modulePath + "/internal/infra",
				modulePath + "/internal/app",
			},
		},
		{
			importerPrefix: modulePath + "/internal/services",
			forbidden: []string{
				modulePath + "/internal/http",
				modulePath + "/internal/app",
				modulePath + "/internal/repos/postgres",
				"github.com/gin-gonic/gin",
				"github.com/jackc/pgx/v5",
			},
		},
		{
			importerPrefix: modulePath + "/internal/http",
			forbidden: []string{
				modulePath + "/internal/repos/postgres",
			},
		},
		{
			importerPrefix: modulePath + "/internal/http/handlers",
			forbidden: []string{
				"github.com/jackc/pgx/v5",
				"github.com/jackc/pgx/v5/pgconn",
			},
		},
		{
			importerPrefix: modulePath + "/internal/repos",
			forbidden: []string{
				modulePath + "/internal/http",
			},
		},
	}

	internalRoot := filepath.Join(repoRoot, "internal")
	files, err := listGoFiles(internalRoot)
	if err != nil {
		t.Fatalf("list go files: %v", err)
	}

	var violations []string
	for _, file := range files {
		importerPath := importPathForFile(modulePath, repoRoot, file)
		imports, err := fileImports(file)
		if err != nil {
			t.Fatalf("parse imports in %s: %v", file, err)
		}
		for _, imp := range imports {
			for _, rule := range rules {
				if !hasPrefixPath(importerPath, rule.importerPrefix) {
					continue
				}
				if !matchesAnyPrefix(imp, rule.forbidden) {
					continue
				}
				if isLegacyAllowed(modulePath, importerPath, imp) {
					continue
				}
				rel, _ := filepath.Rel(repoRoot, file)
				violations = append(violations, fmt.Sprintf("%s imports forbidden package %q", filepath.ToSlash(rel), imp))
			}
		}
	}

	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("architecture dependency violations:\n- %s", strings.Join(violations, "\n- "))
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate current file")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("cannot find go.mod from %s: %v", root, err)
	}
	return root
}

func readModulePath(t *testing.T, repoRoot string) string {
	t.Helper()
	f, err := os.Open(filepath.Join(repoRoot, "go.mod"))
	if err != nil {
		t.Fatalf("open go.mod: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan go.mod: %v", err)
	}
	t.Fatal("module directive not found in go.mod")
	return ""
}

func listGoFiles(root string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		out = append(out, path)
		return nil
	})
	return out, err
}

func importPathForFile(modulePath, repoRoot, file string) string {
	dir := filepath.Dir(file)
	rel, err := filepath.Rel(repoRoot, dir)
	if err != nil {
		return modulePath
	}
	rel = filepath.ToSlash(rel)
	if rel == "." {
		return modulePath
	}
	return modulePath + "/" + rel
}

func fileImports(file string) ([]string, error) {
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, file, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(parsed.Imports))
	for _, imp := range parsed.Imports {
		out = append(out, strings.Trim(imp.Path.Value, `"`))
	}
	return out, nil
}

func hasPrefixPath(path, prefix string) bool {
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}

func matchesAnyPrefix(target string, prefixes []string) bool {
	for _, p := range prefixes {
		if hasPrefixPath(target, p) {
			return true
		}
	}
	return false
}

func isLegacyAllowed(modulePath, importerPath, importedPath string) bool {
	for relPrefix, allowedImports := range legacyAllowed {
		fullPrefix := modulePath + "/" + relPrefix
		if !hasPrefixPath(importerPath, fullPrefix) {
			continue
		}
		return slices.ContainsFunc(allowedImports, func(p string) bool {
			return hasPrefixPath(importedPath, p)
		})
	}
	return false
}
