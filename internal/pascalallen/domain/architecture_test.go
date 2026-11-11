package domain_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var purePackages = []string{"shell", "content"}

var forbiddenImports = map[string]string{
	"os":            "no file system or process access",
	"net":           "no network access",
	"net/http":      "no network access",
	"syscall":       "no system calls",
	"syscall/js":    "the browser bridge lives in infrastructure",
	"unsafe":        "no unsafe memory access",
	"reflect":       "keeps a TinyGo build possible",
	"encoding/json": "keeps a TinyGo build possible",
	"regexp":        "keeps a TinyGo build possible",
	"text/template": "keeps a TinyGo build possible",
	"html/template": "keeps a TinyGo build possible",
	"math/rand":     "output must be deterministic",
	"math/rand/v2":  "output must be deterministic",
	"crypto/rand":   "output must be deterministic",
}

type violation struct {
	file string
	what string
}

func sources(t *testing.T, dir string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_test.go") {
			files = append(files, p)
		}
		return nil
	})
	require.NoError(t, err)
	return files
}

func inspect(t *testing.T, path string) []violation {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly|parser.SkipObjectResolution)
	require.NoError(t, err)
	full, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
	require.NoError(t, err)
	var found []violation
	timeName := ""
	for _, imp := range file.Imports {
		p, err := strconv.Unquote(imp.Path.Value)
		require.NoError(t, err)
		first := strings.SplitN(p, "/", 2)[0]
		if strings.Contains(first, ".") {
			found = append(found, violation{path, "imports non-standard package " + p})
		}
		if reason, bad := forbiddenImports[p]; bad || strings.HasPrefix(p, "net/") || strings.HasPrefix(p, "os/") {
			found = append(found, violation{path, "imports " + p + ": " + reason})
		}
		if p == "time" {
			timeName = "time"
			if imp.Name != nil {
				timeName = imp.Name.Name
			}
		}
	}
	ast.Inspect(full, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.GoStmt:
			found = append(found, violation{path, "starts a goroutine at " + fset.Position(n.Pos()).String()})
		case *ast.SelectorExpr:
			if id, ok := n.X.(*ast.Ident); ok && timeName != "" && id.Name == timeName && (n.Sel.Name == "Now" || n.Sel.Name == "Since" || n.Sel.Name == "Until") {
				found = append(found, violation{path, "reads the wall clock with time." + n.Sel.Name + " at " + fset.Position(n.Pos()).String()})
			}
		}
		return true
	})
	return found
}

func TestPurePackages(t *testing.T) {
	for _, pkg := range purePackages {
		t.Run(pkg+" imports only allowed standard library packages and never reads the clock", func(t *testing.T) {
			files := sources(t, pkg)
			require.NotEmpty(t, files)
			for _, f := range files {
				for _, v := range inspect(t, f) {
					assert.Fail(t, v.file+": "+v.what)
				}
			}
		})
	}

	t.Run("the checker catches every forbidden construct", func(t *testing.T) {
		dir := t.TempDir()
		src := "package bad\n\nimport (\n\t\"os\"\n\tclock \"time\"\n\t\"github.com/x/y\"\n\t\"net/url\"\n)\n\nfunc f() {\n\tgo func() {}()\n\t_ = clock.Now()\n\t_ = os.Args\n}\n"
		path := filepath.Join(dir, "bad.go")
		require.NoError(t, os.WriteFile(path, []byte(src), 0o644))
		var whats []string
		for _, v := range inspect(t, path) {
			whats = append(whats, v.what)
		}
		joined := strings.Join(whats, "\n")
		assert.Contains(t, joined, "imports os:")
		assert.Contains(t, joined, "imports non-standard package github.com/x/y")
		assert.Contains(t, joined, "imports net/url")
		assert.Contains(t, joined, "starts a goroutine")
		assert.Contains(t, joined, "reads the wall clock with time.Now")
	})
}
