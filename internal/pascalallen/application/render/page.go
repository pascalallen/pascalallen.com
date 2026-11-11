package render

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"html/template"
	"io/fs"
	"strings"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/shell"
)

const themeColor = "#0e1116"

type Config struct {
	BaseURL     string
	Env         string
	AssetHashes map[string]string
}

type Page struct {
	baseURL     string
	description string
	title       string
	jsonLD      template.JS
	boot        template.HTML
	prompt      string
	plain       template.HTML
	environment string
	wasmHash    string
}

type Head struct {
	Title       string
	Description string
	Canonical   string
	OGImage     string
	Robots      string
	ThemeColor  string
	JSONLD      template.JS
}

type IndexData struct {
	Head        Head
	Boot        template.HTML
	Prompt      string
	Environment string
	WasmHash    string
}

type PlainData struct {
	Head    Head
	Content template.HTML
}

func NewPage(fsys fs.FS, cfg Config) *Page {
	base := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	p := readProfile(fsys)
	sh := shell.New(fsys)

	hashes := map[string]string{}
	for k, v := range cfg.AssetHashes {
		if v != "" {
			hashes[k] = v
		}
	}
	envJSON, _ := json.Marshal(map[string]any{
		"APP_BASE_URL": cfg.BaseURL,
		"APP_ENV":      cfg.Env,
		"ASSET_HASHES": hashes,
	})

	suffix := p.JobTitle
	if suffix == "" {
		suffix = "pascalallen.com"
	}

	return &Page{
		baseURL:     base,
		description: p.Description,
		title:       p.Name + " - " + suffix,
		jsonLD:      template.JS(jsonLD(p, base)),
		boot:        BootEntry(sh.Boot()),
		prompt:      sh.Prompt(),
		plain:       Plain(fsys),
		environment: base64.StdEncoding.EncodeToString(envJSON),
		wasmHash:    hashes["wasm"],
	}
}

func (p *Page) head(path string, indexable bool) Head {
	h := Head{
		Title:       p.title,
		Description: p.description,
		ThemeColor:  themeColor,
		JSONLD:      p.jsonLD,
	}
	if !indexable {
		h.Robots = "noindex"
	}
	if p.baseURL != "" {
		if indexable {
			h.Canonical = p.baseURL + path
		}
		h.OGImage = p.baseURL + "/static/android-chrome-512x512.png"
	}
	return h
}

func (p *Page) Index(requestPath string) IndexData {
	return IndexData{
		Head:        p.head("/", requestPath == "/"),
		Boot:        p.boot,
		Prompt:      p.prompt,
		Environment: p.environment,
		WasmHash:    p.wasmHash,
	}
}

func (p *Page) PlainView() PlainData {
	return PlainData{Head: p.head("/plain", true), Content: p.plain}
}

func (p *Page) Robots() string {
	out := "User-agent: *\nAllow: /\n"
	if p.baseURL != "" {
		out += "\nSitemap: " + p.baseURL + "/sitemap.xml\n"
	}
	return out
}

type urlEntry struct {
	Loc string `xml:"loc"`
}

type urlSet struct {
	XMLName xml.Name   `xml:"urlset"`
	Xmlns   string     `xml:"xmlns,attr"`
	URLs    []urlEntry `xml:"url"`
}

func (p *Page) Sitemap() ([]byte, error) {
	set := urlSet{Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	if p.baseURL != "" {
		set.URLs = []urlEntry{{Loc: p.baseURL + "/"}, {Loc: p.baseURL + "/plain"}}
	}
	body, err := xml.MarshalIndent(set, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), body...), nil
}
