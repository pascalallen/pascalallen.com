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

var allowedImports = map[string]bool{
	"embed":        true,
	"errors":       true,
	"io/fs":        true,
	"sort":         true,
	"strconv":      true,
	"strings":      true,
	"time":         true,
	"unicode":      true,
	"unicode/utf8": true,
}

var forbiddenTime = map[string]bool{
	"Now":       true,
	"Since":     true,
	"Until":     true,
	"After":     true,
	"AfterFunc": true,
	"Sleep":     true,
	"Tick":      true,
	"NewTimer":  true,
	"NewTicker": true,
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

func inspect(t *testing.T, name, src string) []string {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, name, src, parser.SkipObjectResolution)
	require.NoError(t, err)
	var found []string
	timeNames := map[string]bool{}
	for _, imp := range file.Imports {
		p, err := strconv.Unquote(imp.Path.Value)
		require.NoError(t, err)
		if !allowedImports[p] {
			found = append(found, "imports "+p+", which is not on the allow-list")
		}
		local := p[strings.LastIndex(p, "/")+1:]
		if imp.Name != nil {
			switch imp.Name.Name {
			case ".":
				found = append(found, "dot-imports "+p)
			case "_":
				found = append(found, "blank-imports "+p)
			}
			local = imp.Name.Name
		}
		if p == "time" {
			timeNames[local] = true
		}
	}
	ast.Inspect(file, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.GoStmt:
			found = append(found, "starts a goroutine at "+fset.Position(n.Pos()).String())
		case *ast.SelectorExpr:
			if id, ok := n.X.(*ast.Ident); ok && timeNames[id.Name] && forbiddenTime[n.Sel.Name] {
				found = append(found, "uses time."+n.Sel.Name+" at "+fset.Position(n.Pos()).String())
			}
		}
		return true
	})
	return found
}

func TestPurePackages(t *testing.T) {
	for _, pkg := range purePackages {
		t.Run(pkg+" imports only allow-listed standard library packages and never reads the clock", func(t *testing.T) {
			files := sources(t, pkg)
			require.NotEmpty(t, files)
			for _, f := range files {
				src, err := os.ReadFile(f)
				require.NoError(t, err)
				for _, v := range inspect(t, f, string(src)) {
					assert.Fail(t, f+": "+v)
				}
			}
		})
	}
}

func TestPurityChecker(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{"an os import is reported", `import "os"; var _ = os.Args`, "imports os,"},
		{"a log import is reported", `import "log"; var _ = log.Println`, "imports log,"},
		{"a net import is reported", `import "net"; var _ net.IP`, "imports net,"},
		{"a net/url import is reported", `import "net/url"; var _ url.URL`, "imports net/url,"},
		{"an io/ioutil import is reported", `import "io/ioutil"; var _ = ioutil.Discard`, "imports io/ioutil,"},
		{"a syscall import is reported", `import "syscall"; var _ = syscall.Getpid`, "imports syscall,"},
		{"a syscall/js import is reported", `import "syscall/js"; var _ js.Value`, "imports syscall/js,"},
		{"an unsafe import is reported", `import "unsafe"; var _ unsafe.Pointer`, "imports unsafe,"},
		{"a runtime import is reported", `import "runtime"; var _ = runtime.GOOS`, "imports runtime,"},
		{"a reflect import is reported", `import "reflect"; var _ reflect.Type`, "imports reflect,"},
		{"an encoding/json import is reported", `import "encoding/json"; var _ = json.Marshal`, "imports encoding/json,"},
		{"a math/rand import is reported", `import "math/rand"; var _ = rand.Int`, "imports math/rand,"},
		{"a module import is reported", `import "github.com/x/y"; var _ = y.Z`, "imports github.com/x/y,"},
		{"a dot import is reported", `import . "time"; var _ = Unix`, "dot-imports time"},
		{"a blank import is reported", `import _ "embed"`, "blank-imports embed"},
		{"time.Now is reported", `import "time"; var _ = time.Now`, "uses time.Now"},
		{"an aliased time.Now is reported", `import clock "time"; var _ = clock.Now()`, "uses time.Now"},
		{"time.Since is reported", `import "time"; var _ = time.Since`, "uses time.Since"},
		{"time.Until is reported", `import "time"; var _ = time.Until`, "uses time.Until"},
		{"time.After is reported", `import "time"; var _ = time.After`, "uses time.After"},
		{"time.AfterFunc is reported", `import "time"; var _ = time.AfterFunc`, "uses time.AfterFunc"},
		{"time.Sleep is reported", `import "time"; var _ = time.Sleep`, "uses time.Sleep"},
		{"time.Tick is reported", `import "time"; var _ = time.Tick`, "uses time.Tick"},
		{"time.NewTimer is reported", `import "time"; var _ = time.NewTimer`, "uses time.NewTimer"},
		{"time.NewTicker is reported", `import "time"; var _ = time.NewTicker`, "uses time.NewTicker"},
		{"a goroutine is reported", `func f() { go f() }`, "starts a goroutine"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			found := inspect(t, "bad.go", "package bad\n"+strings.ReplaceAll(tc.src, "; ", "\n")+"\n")
			assert.Contains(t, strings.Join(found, "\n"), tc.want)
		})
	}

	t.Run("allowed imports and time values pass", func(t *testing.T) {
		src := "package ok\nimport (\n\t\"strings\"\n\t\"time\"\n)\nvar _ = strings.Fields\nvar _ = time.Unix(0, 0).Format(time.UnixDate)\nfunc now() time.Time { return time.Time{} }\n"
		assert.Empty(t, inspect(t, "ok.go", src))
	})
}
