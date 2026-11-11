package pascalallen

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const modulePath = "github.com/pascalallen/pascalallen.com"

func importsUnder(t *testing.T, dir string) map[string][]string {
	t.Helper()

	found := map[string][]string{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, spec := range file.Imports {
			value, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			found[path] = append(found[path], value)
		}

		return nil
	})
	require.NoError(t, err)

	return found
}

func isStdlib(importPath string) bool {
	first, _, _ := strings.Cut(importPath, "/")

	return !strings.Contains(first, ".")
}

func TestArchitecture(t *testing.T) {
	t.Run("domain packages import only the standard library", func(t *testing.T) {
		files := importsUnder(t, "domain")
		require.NotEmpty(t, files)

		for path, imports := range files {
			for _, imp := range imports {
				assert.Truef(t, isStdlib(imp), "%s imports %s", path, imp)
			}
		}
	})

	t.Run("application packages never import infrastructure", func(t *testing.T) {
		files := importsUnder(t, "application")
		require.NotEmpty(t, files)

		for path, imports := range files {
			for _, imp := range imports {
				assert.Falsef(t, strings.HasPrefix(imp, modulePath+"/internal/pascalallen/infrastructure"), "%s imports %s", path, imp)
			}
		}
	})
}
