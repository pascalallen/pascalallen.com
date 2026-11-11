package routes

import (
	"crypto/sha256"
	"encoding/hex"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	staticRoot       = "./web/static"
	wasmModule       = "./web/static/assets/shell.wasm"
	cacheImmutable   = "public, max-age=31536000, immutable"
	cacheRevalidate  = "no-cache"
	wasmContentType  = "application/wasm"
	versionHexLength = 12
)

var encodings = []struct {
	token  string
	suffix string
}{
	{"br", ".br"},
	{"gzip", ".gz"},
}

func (r Router) Fileserver() {
	handler := serveStatic(staticRoot)
	r.engine.GET("/static/*filepath", handler)
	r.engine.HEAD("/static/*filepath", handler)
}

func serveStatic(root string) gin.HandlerFunc {
	index := newAssetIndex(root)

	return func(c *gin.Context) {
		key := path.Clean("/" + c.Param("filepath"))
		name := filepath.Join(root, filepath.FromSlash(key))
		hash := index.hash(strings.TrimPrefix(key, "/"))

		info, err := os.Stat(name)
		if err != nil || !info.Mode().IsRegular() {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		contentType := mime.TypeByExtension(filepath.Ext(name))
		if filepath.Ext(name) == ".wasm" {
			contentType = wasmContentType
		}

		served := name
		encoding := ""
		hasSibling := false
		accepted := acceptedEncodings(c.GetHeader("Accept-Encoding"))
		for _, e := range encodings {
			sibling := name + e.suffix
			if sinfo, err := os.Stat(sibling); err == nil && sinfo.Mode().IsRegular() && !sinfo.ModTime().Before(info.ModTime()) {
				hasSibling = true
				if encoding == "" && accepted[e.token] {
					served = sibling
					encoding = e.token
				}
			}
		}

		file, err := os.Open(served)
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		defer file.Close()
		sinfo, err := file.Stat()
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		header := c.Writer.Header()
		if contentType != "" {
			header.Set("Content-Type", contentType)
		}
		if hasSibling {
			header.Set("Vary", "Accept-Encoding")
		}
		if encoding != "" {
			header.Set("Content-Encoding", encoding)
			header.Set("Content-Length", strconv.FormatInt(sinfo.Size(), 10))
		}
		if hash != "" && c.Query("v") == hash {
			header.Set("Cache-Control", cacheImmutable)
		} else {
			header.Set("Cache-Control", cacheRevalidate)
		}
		if hash != "" {
			tag := hash
			if encoding != "" {
				tag += "-" + encoding
			}
			header.Set("ETag", `"`+tag+`"`)
		}

		http.ServeContent(c.Writer, c.Request, filepath.Base(name), sinfo.ModTime(), file)
	}
}

func acceptedEncodings(header string) map[string]bool {
	accepted := map[string]bool{}
	for _, part := range strings.Split(header, ",") {
		fields := strings.Split(part, ";")
		token := strings.ToLower(strings.TrimSpace(fields[0]))
		if token == "" {
			continue
		}
		weight := 1.0
		for _, param := range fields[1:] {
			key, value, found := strings.Cut(strings.TrimSpace(param), "=")
			if found && strings.EqualFold(strings.TrimSpace(key), "q") {
				if q, err := strconv.ParseFloat(strings.TrimSpace(value), 64); err == nil {
					weight = q
				}
			}
		}
		accepted[token] = weight > 0
	}

	return accepted
}

func assetVersion(file string) string {
	data, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)

	return hex.EncodeToString(sum[:])[:versionHexLength]
}

type assetEntry struct {
	size  int64
	mtime time.Time
	hash  string
}

type assetIndex struct {
	mu      sync.Mutex
	root    string
	entries map[string]assetEntry
}

func newAssetIndex(root string) *assetIndex {
	return &assetIndex{root: root, entries: map[string]assetEntry{}}
}

func (i *assetIndex) hash(rel string) string {
	rel = strings.TrimPrefix(path.Clean("/"+rel), "/")
	if rel == "" || strings.HasSuffix(rel, ".br") || strings.HasSuffix(rel, ".gz") {
		return ""
	}
	file := filepath.Join(i.root, filepath.FromSlash(rel))

	i.mu.Lock()
	defer i.mu.Unlock()

	info, err := os.Stat(file)
	if err != nil || !info.Mode().IsRegular() {
		delete(i.entries, rel)
		return ""
	}
	if e, ok := i.entries[rel]; ok && e.size == info.Size() && e.mtime.Equal(info.ModTime()) {
		return e.hash
	}
	hash := assetVersion(file)
	if hash == "" {
		delete(i.entries, rel)
		return ""
	}
	i.entries[rel] = assetEntry{size: info.Size(), mtime: info.ModTime(), hash: hash}

	return hash
}

func assetURL(index *assetIndex) func(string) string {
	return func(name string) string {
		url := "/static/assets/" + name
		if hash := index.hash("assets/" + name); hash != "" {
			url += "?v=" + hash
		}

		return url
	}
}
