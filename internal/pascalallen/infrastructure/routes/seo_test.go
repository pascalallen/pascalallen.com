package routes

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"html"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/content"
)

var (
	scriptTag  = regexp.MustCompile(`(?s)<script([^>]*)>(.*?)</script>`)
	styleTag   = regexp.MustCompile(`(?i)<style|\sstyle=`)
	inlineJSON = regexp.MustCompile(`(?s)<script type="application/ld\+json">(.*?)</script>`)
)

func seoRouter(t *testing.T) Router {
	t.Setenv("APP_BASE_URL", "https://example.test")
	t.Setenv("APP_ENV", "test")
	return newTestRouter(t)
}

func TestIndexPage(t *testing.T) {
	r := seoRouter(t)
	body := get(r, "/").Body.String()

	t.Run("the boot screen is rendered into the output container with the marker", func(t *testing.T) {
		assert.Contains(t, body, `id="output" role="log"`)
		assert.Contains(t, body, `<div class="entry entry-boot" data-prerendered="">`)
		assert.Contains(t, body, "Full-stack software engineer with 10+ years of experience")
		assert.Contains(t, body, `<label class="prompt" id="prompt" for="input">visitor@pascalallen.com:~$</label>`)
		assert.Contains(t, body, `<input class="input" id="input"`)
	})

	t.Run("every boot command and every template chip is a deep link", func(t *testing.T) {
		for _, name := range []string{"help", "about", "resume", "experience", "projects", "writing", "contact", "repos", "camera"} {
			assert.Contains(t, body, `href="/#`+name+`" data-run="`+name+`"`, name)
		}
		for _, name := range []string{"help", "about", "projects", "writing", "repos", "contact", "theme", "clear"} {
			assert.Contains(t, body, `<a class="chip" href="/#`+name+`" data-run="`+name+`">`, name)
		}
	})

	t.Run("a visible link leads to the plain text version", func(t *testing.T) {
		assert.Contains(t, body, `<a class="chip" href="/plain" rel="alternate">plain text version</a>`)
	})

	t.Run("no script carries inline code and no style is inline", func(t *testing.T) {
		for _, m := range scriptTag.FindAllStringSubmatch(body, -1) {
			if strings.Contains(m[1], "application/ld+json") {
				continue
			}
			assert.Contains(t, m[1], "src=", m[0])
			assert.Empty(t, strings.TrimSpace(m[2]), m[0])
		}
		assert.False(t, styleTag.MatchString(body))
	})

	t.Run("the head carries title, description, canonical, social tags and theme color", func(t *testing.T) {
		assert.Contains(t, body, "<title>Pascal Allen - Senior Software Engineer</title>")
		assert.Contains(t, body, `<meta name="description" content="Full-stack software engineer`)
		assert.Contains(t, body, `<link rel="canonical" href="https://example.test/">`)
		assert.Contains(t, body, `<meta property="og:title"`)
		assert.Contains(t, body, `<meta property="og:image" content="https://example.test/static/og-image.png">`)
		assert.Contains(t, body, `<meta property="og:image:width" content="1200">`)
		assert.Contains(t, body, `<meta property="og:image:height" content="630">`)
		assert.Contains(t, body, `<meta name="twitter:card" content="summary_large_image">`)
		assert.Contains(t, body, `<meta name="twitter:image" content="https://example.test/static/og-image.png">`)
		assert.Contains(t, body, `<link rel="icon" type="image/svg+xml" href="/static/favicon.svg">`)
		assert.Contains(t, body, `<link rel="icon" href="/static/favicon.ico" sizes="48x48">`)
		assert.Contains(t, body, `<meta name="theme-color" content="#0e1116">`)
		assert.Contains(t, body, `<link rel="manifest" href="/static/site.webmanifest">`)
		assert.NotContains(t, body, "Content-Security-Policy")
		assert.NotContains(t, body, "noindex")
	})

	t.Run("the json-ld block parses as a person", func(t *testing.T) {
		m := inlineJSON.FindStringSubmatch(body)
		require.Len(t, m, 2)
		var doc map[string]any
		require.NoError(t, json.Unmarshal([]byte(html.UnescapeString(m[1])), &doc))

		assert.Equal(t, "Person", doc["@type"])
		assert.Equal(t, "Pascal Allen", doc["name"])
		assert.Equal(t, "https://example.test", doc["url"])
		assert.Len(t, doc["sameAs"], 3)
		assert.NotEmpty(t, doc["knowsAbout"])
	})
}

func TestIndexPageWithoutBaseURL(t *testing.T) {
	t.Setenv("APP_BASE_URL", "")
	t.Setenv("APP_ENV", "test")
	body := get(newTestRouter(t), "/").Body.String()

	t.Run("without an image the social card falls back to a summary", func(t *testing.T) {
		assert.Contains(t, body, `<meta name="twitter:card" content="summary">`)
		assert.NotContains(t, body, "summary_large_image")
		assert.NotContains(t, body, "og:image")
	})
}

func TestPlainRoute(t *testing.T) {
	r := seoRouter(t)
	rec := get(r, "/plain")
	body := rec.Body.String()

	t.Run("it renders every publication with its link", func(t *testing.T) {
		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")
		assert.GreaterOrEqual(t, len(publicationURLs(t)), 19)
		for _, url := range publicationURLs(t) {
			assert.Contains(t, body, `href="`+html.EscapeString(url)+`"`, url)
		}
	})

	t.Run("it is indexable with its own canonical and loads no script", func(t *testing.T) {
		assert.Contains(t, body, `<link rel="canonical" href="https://example.test/plain">`)
		assert.NotContains(t, body, "noindex")
		assert.NotContains(t, body, "app.js")
		assert.NotContains(t, body, `data-prerendered`)
	})
}

func publicationURLs(t *testing.T) []string {
	t.Helper()
	fsys := content.FS()
	files, err := fs.Glob(fsys, "writing/*.md")
	require.NoError(t, err)
	require.NotEmpty(t, files)
	var urls []string
	for _, name := range files {
		data, err := fs.ReadFile(fsys, name)
		require.NoError(t, err)
		var found int
		for _, f := range strings.Fields(string(data)) {
			if strings.HasPrefix(f, "http") {
				urls = append(urls, f)
				found++
			}
		}
		require.Equal(t, 1, found, name)
	}
	return urls
}

func assertSecurityHeaders(t *testing.T, h http.Header, label string) {
	t.Helper()

	assert.Equal(t, "nosniff", h.Get("X-Content-Type-Options"), label)
	assert.Equal(t, "strict-origin-when-cross-origin", h.Get("Referrer-Policy"), label)
	assert.Equal(t, "camera=(), microphone=(), geolocation=()", h.Get("Permissions-Policy"), label)
	csp := h.Get("Content-Security-Policy")
	for _, directive := range []string{
		"default-src 'self'",
		"script-src 'self' 'wasm-unsafe-eval'",
		"img-src 'self' data:",
		"connect-src 'self'",
		"style-src 'self'",
		"base-uri 'none'",
		"frame-ancestors 'none'",
		"form-action 'self'",
	} {
		assert.Contains(t, csp, directive, label)
	}
	assert.NotContains(t, csp, "unsafe-inline", label)
}

func TestSecurityHeaders(t *testing.T) {
	r := seoRouter(t)

	for _, path := range []string{"/", "/plain", "/plain/", "/robots.txt", "/sitemap.xml", "/static/favicon.ico", "/api/v1/camera/status", "/api/v1/nope", "/no/such/path"} {
		t.Run("the headers are present on "+path, func(t *testing.T) {
			assertSecurityHeaders(t, get(r, path).Header(), path)
		})
	}

	t.Run("the headers are present on a 405", func(t *testing.T) {
		assertSecurityHeaders(t, do(r, http.MethodPost, "/plain", "", "").Header(), "POST /plain")
	})

	t.Run("the headers are present on the camera stream", func(t *testing.T) {
		frames := make(chan []byte, 1)
		stream := assembledRouter(t, stubStream{frames: frames, available: true, released: make(chan struct{})})
		ctx, cancel := context.WithCancel(t.Context())
		rec := httptest.NewRecorder()
		done := make(chan struct{})
		go func() {
			stream.engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/camera/stream", nil).WithContext(ctx))
			close(done)
		}()
		frames <- []byte("--frame\r\nchunk")
		assert.Eventually(t, func() bool { return len(frames) == 0 }, 2*time.Second, 5*time.Millisecond)
		cancel()
		<-done

		assertSecurityHeaders(t, rec.Header(), "camera stream")
		assert.Equal(t, "multipart/x-mixed-replace; boundary=frame", rec.Header().Get("Content-Type"))
		assert.Contains(t, rec.Body.String(), "chunk")
	})
}

func TestMethods(t *testing.T) {
	r := seoRouter(t)

	for path, contentType := range map[string]string{
		"/":             "text/html",
		"/plain":        "text/html",
		"/robots.txt":   "text/plain",
		"/sitemap.xml":  "application/xml",
		"/no/such/path": "text/html",
	} {
		t.Run("HEAD "+path+" answers with the headers and no body", func(t *testing.T) {
			rec := do(r, http.MethodHead, path, "", "")

			require.Equal(t, http.StatusOK, rec.Code)
			assert.Empty(t, rec.Body.String())
			assert.Contains(t, rec.Header().Get("Content-Type"), contentType)
			assertSecurityHeaders(t, rec.Header(), path)
		})
	}

	for _, tc := range []struct{ method, path string }{
		{http.MethodPost, "/"},
		{http.MethodPost, "/plain"},
		{http.MethodPut, "/robots.txt"},
		{http.MethodDelete, "/sitemap.xml"},
		{http.MethodPost, "/no/such/path"},
	} {
		t.Run(tc.method+" "+tc.path+" is a plain 405 with an Allow header", func(t *testing.T) {
			rec := do(r, tc.method, tc.path, "", "")

			require.Equal(t, http.StatusMethodNotAllowed, rec.Code)
			assert.Equal(t, "GET, HEAD", rec.Header().Get("Allow"))
			assert.NotContains(t, rec.Body.String(), "<html")
			assert.Contains(t, rec.Header().Get("Content-Type"), "text/plain")
		})
	}

	t.Run("a wrong method under the api is a JSend 405", func(t *testing.T) {
		rec := do(r, http.MethodDelete, "/api/v1/repos", "", "")

		require.Equal(t, http.StatusMethodNotAllowed, rec.Code)
		assert.JSONEq(t, `{"status":"error","message":"method not allowed"}`, rec.Body.String())
	})

	t.Run("an unknown api path is a JSON 404 and not the page", func(t *testing.T) {
		for _, path := range []string{"/api/v1/nope", "/api/zzz", "/api"} {
			rec := get(r, path)

			require.Equal(t, http.StatusNotFound, rec.Code, path)
			assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
			assert.JSONEq(t, `{"status":"error","message":"not found"}`, rec.Body.String(), path)
		}
	})

	t.Run("a trailing slash is not redirected", func(t *testing.T) {
		rec := get(r, "/plain/")

		assert.NotEqual(t, http.StatusMovedPermanently, rec.Code)
		assert.Empty(t, rec.Header().Get("Location"))
	})
}

func TestRobotsAndSitemap(t *testing.T) {
	r := seoRouter(t)

	t.Run("robots allows all and names the sitemap", func(t *testing.T) {
		rec := get(r, "/robots.txt")

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "text/plain")
		assert.Equal(t, "User-agent: *\nAllow: /\n\nSitemap: https://example.test/sitemap.xml\n", rec.Body.String())
	})

	t.Run("the sitemap is well formed and lists the canonical urls", func(t *testing.T) {
		rec := get(r, "/sitemap.xml")

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "xml")
		var set struct {
			XMLName xml.Name `xml:"urlset"`
			Locs    []string `xml:"url>loc"`
		}
		require.NoError(t, xml.Unmarshal(rec.Body.Bytes(), &set))
		assert.Equal(t, "http://www.sitemaps.org/schemas/sitemap/0.9", set.XMLName.Space)
		assert.Equal(t, []string{"https://example.test/", "https://example.test/plain"}, set.Locs)
	})
}

func TestMethodMatrix(t *testing.T) {
	r := seoRouter(t)

	for _, tc := range []struct {
		method, path string
		status       int
		allow        string
		contentType  string
	}{
		{http.MethodPut, "/api/v1/visits", http.StatusMethodNotAllowed, "POST", "application/json"},
		{http.MethodGet, "/api/v1/visits", http.StatusMethodNotAllowed, "POST", "application/json"},
		{http.MethodOptions, "/api/v1/repos", http.StatusMethodNotAllowed, "GET, HEAD", "application/json"},
		{http.MethodDelete, "/api/v1/packages", http.StatusMethodNotAllowed, "GET, HEAD", "application/json"},
		{http.MethodPost, "/api/v1/camera/status", http.StatusMethodNotAllowed, "GET, HEAD", "application/json"},
		{http.MethodHead, "/api/v1/camera/stream", http.StatusMethodNotAllowed, "GET", "application/json"},
		{http.MethodHead, "/api/v1/camera/status", http.StatusOK, "", "application/json"},
		{http.MethodPost, "/plain", http.StatusMethodNotAllowed, "GET, HEAD", "text/plain"},
		{http.MethodPost, "/no/such/path", http.StatusMethodNotAllowed, "GET, HEAD", "text/plain"},
		{http.MethodPost, "/api/v1/nope", http.StatusNotFound, "", "application/json"},
		{http.MethodDelete, "/api/zzz", http.StatusNotFound, "", "application/json"},
		{http.MethodGet, "/api/v1/nope", http.StatusNotFound, "", "application/json"},
	} {
		t.Run(tc.method+" "+tc.path+" answers "+http.StatusText(tc.status), func(t *testing.T) {
			rec := do(r, tc.method, tc.path, "", "")

			require.Equal(t, tc.status, rec.Code)
			assert.Equal(t, tc.allow, rec.Header().Get("Allow"))
			assert.Contains(t, rec.Header().Get("Content-Type"), tc.contentType)
			assertSecurityHeaders(t, rec.Header(), tc.method+" "+tc.path)
		})
	}

	t.Run("HEAD on the feed routes is routed like GET", func(t *testing.T) {
		for _, path := range []string{"/api/v1/repos", "/api/v1/packages"} {
			rec := do(r, http.MethodHead, path, "", "")

			assert.NotEqual(t, http.StatusNotFound, rec.Code, path)
			assert.NotEqual(t, http.StatusMethodNotAllowed, rec.Code, path)
		}
	})
}

func TestConcurrentRequests(t *testing.T) {
	r := seoRouter(t)
	want := get(r, "/").Body.String()
	plain := get(r, "/plain").Body.String()

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.Equal(t, want, get(r, "/").Body.String())
			assert.Equal(t, plain, get(r, "/plain").Body.String())
			assert.NotEmpty(t, get(r, "/sitemap.xml").Body.String())
		}()
	}
	wg.Wait()
}
