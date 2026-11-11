package render

import (
	"html"
	"html/template"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/shell"
)

var sectionOrder = []string{"about", "skills", "experience", "projects", "writing", "education", "contact"}

func sectionRank(name string) int {
	top, _, _ := strings.Cut(name, "/")
	stem := strings.TrimSuffix(top, path.Ext(top))
	for i, s := range sectionOrder {
		if s == stem {
			return i
		}
	}
	return len(sectionOrder)
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func Plain(fsys fs.FS) template.HTML {
	var files []string
	_ = fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			files = append(files, p)
		}
		return nil
	})
	sort.SliceStable(files, func(i, j int) bool {
		ri, rj := sectionRank(files[i]), sectionRank(files[j])
		if ri != rj {
			return ri < rj
		}
		return files[i] < files[j]
	})

	var b strings.Builder
	for _, name := range files {
		sh := shell.New(fsys)
		result := sh.Exec("cat " + shellQuote(name))
		b.WriteString(`<section class="entry"><h2 class="line s-path">` + html.EscapeString(name) + `</h2>`)
		b.WriteString(string(Blocks(result.Blocks)))
		b.WriteString(`</section>`)
	}
	return template.HTML(b.String())
}
