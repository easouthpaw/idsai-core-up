package architecture_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestNoJSONTagsOutsideHTTP(t *testing.T) {
	repoRoot := findRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "internal", "services"),
		filepath.Join(repoRoot, "internal", "domain"),
	}

	var violations []string
	for _, root := range targets {
		files, err := listGoFiles(root)
		if err != nil {
			t.Fatalf("list go files in %s: %v", root, err)
		}
		for _, file := range files {
			rel, _ := filepath.Rel(repoRoot, file)
			if err := inspectStructTags(file, func(tag string, pos token.Position) {
				if strings.Contains(tag, `json:"`) {
					violations = append(violations, fmt.Sprintf("%s:%d contains json struct tag outside HTTP transport", filepath.ToSlash(rel), pos.Line))
				}
			}); err != nil {
				t.Fatalf("inspect struct tags in %s: %v", rel, err)
			}
		}
	}

	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("transport boundary violations:\n- %s", strings.Join(violations, "\n- "))
	}
}

func TestHTTPDTOsAvoidAny(t *testing.T) {
	repoRoot := findRepoRoot(t)
	root := filepath.Join(repoRoot, "internal", "http", "dto")
	files, err := listGoFiles(root)
	if err != nil {
		t.Fatalf("list go files: %v", err)
	}

	var violations []string
	fset := token.NewFileSet()
	for _, file := range files {
		parsed, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		rel, _ := filepath.Rel(repoRoot, file)
		ast.Inspect(parsed, func(n ast.Node) bool {
			spec, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			structType, ok := spec.Type.(*ast.StructType)
			if !ok {
				return true
			}
			for _, field := range structType.Fields.List {
				if fieldUsesAny(field.Type) {
					pos := fset.Position(field.Pos())
					violations = append(violations, fmt.Sprintf("%s:%d type %s uses any in transport DTO", filepath.ToSlash(rel), pos.Line, spec.Name.Name))
				}
			}
			return true
		})
	}

	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("http dto type violations:\n- %s", strings.Join(violations, "\n- "))
	}
}

func TestHandlersDoNotDeclareTransportDTOs(t *testing.T) {
	repoRoot := findRepoRoot(t)
	root := filepath.Join(repoRoot, "internal", "http", "handlers")
	files, err := listGoFiles(root)
	if err != nil {
		t.Fatalf("list go files: %v", err)
	}

	var violations []string
	fset := token.NewFileSet()
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		rel, _ := filepath.Rel(repoRoot, file)
		ast.Inspect(parsed, func(n ast.Node) bool {
			spec, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			if !isTransportDTOName(spec.Name.Name) {
				return true
			}
			if _, ok := spec.Type.(*ast.StructType); !ok {
				return true
			}
			pos := fset.Position(spec.Pos())
			violations = append(violations, fmt.Sprintf("%s:%d declares transport DTO %s in handlers package", filepath.ToSlash(rel), pos.Line, spec.Name.Name))
			return true
		})
	}

	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("handler transport dto violations:\n- %s", strings.Join(violations, "\n- "))
	}
}

func inspectStructTags(file string, visit func(tag string, pos token.Position)) error {
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, file, nil, 0)
	if err != nil {
		return err
	}
	ast.Inspect(parsed, func(n ast.Node) bool {
		field, ok := n.(*ast.Field)
		if !ok || field.Tag == nil {
			return true
		}
		visit(field.Tag.Value, fset.Position(field.Tag.Pos()))
		return true
	})
	return nil
}

func fieldUsesAny(expr ast.Expr) bool {
	switch node := expr.(type) {
	case *ast.Ident:
		return node.Name == "any"
	case *ast.ArrayType:
		return fieldUsesAny(node.Elt)
	case *ast.StarExpr:
		return fieldUsesAny(node.X)
	case *ast.MapType:
		return fieldUsesAny(node.Key) || fieldUsesAny(node.Value)
	case *ast.SelectorExpr:
		return false
	case *ast.InterfaceType:
		return len(node.Methods.List) == 0
	default:
		return false
	}
}

func isTransportDTOName(name string) bool {
	suffixes := []string{"Request", "Response", "ListResponse", "EnvelopeResponse"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}
