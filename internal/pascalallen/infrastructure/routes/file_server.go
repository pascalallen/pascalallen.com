package routes

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

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
	index := indexAssets(root)

	return func(c *gin.Context) {
		key := path.Clean("/" + c.Param("filepath"))
		name := filepath.Join(root, filepath.FromSlash(key))
		hash := index[strings.TrimPrefix(key, "/")]

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
			if sinfo, err := os.Stat(sibling); err == nil && sinfo.Mode().IsRegular() {
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

func indexAssets(root string) map[string]string {
	index := map[string]string{}
	_ = filepath.WalkDir(root, func(file string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !d.Type().IsRegular() {
			return nil
		}
		if strings.HasSuffix(file, ".br") || strings.HasSuffix(file, ".gz") {
			return nil
		}
		rel, err := filepath.Rel(root, file)
		if err != nil {
			return nil
		}
		if version := assetVersion(file); version != "" {
			index[filepath.ToSlash(rel)] = version
		}
		return nil
	})

	return index
}

func assetURL(index map[string]string) func(string) string {
	return func(name string) string {
		url := "/static/assets/" + name
		if hash := index["assets/"+name]; hash != "" {
			url += "?v=" + hash
		}

		return url
	}
}
