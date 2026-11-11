package content_test

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/content"
)

var publications = []string{
	"https://medium.com/@pascalallen/how-to-run-a-go-web-server-with-apple-container-0fd1e7182b80",
	"https://medium.com/@pascalallen/event-sourcing-in-go-a-full-cqrs-implementation-with-eventstoredb-and-rabbitmq-1ac73782ef3a",
	"https://medium.com/@pascalallen/i-turned-a-35-computer-into-a-live-security-camera-with-go-f1e061dacc33",
	"https://pascalallen.medium.com/structuring-go-projects-with-clean-architecture-2c46d7e58ac3",
	"https://pascalallen.medium.com/how-to-build-an-mcp-server-in-go-50ed8a4c9ed4",
	"https://pascalallen.medium.com/go-ing-beyond-the-ordinary-discovering-the-unique-superpowers-of-the-go-language-372bfbf54445",
	"https://pascalallen.medium.com/mastering-the-fundamentals-of-algorithms-a-beginners-guide-364adaf31556",
	"https://pascalallen.medium.com/automate-your-deployments-with-cron-7174ecb9f52f",
	"https://pascalallen.medium.com/how-to-build-a-grpc-server-in-go-943f337c4e05",
	"https://pascalallen.medium.com/developing-a-framework-for-any-project-9cf7dac82ffe",
	"https://pascalallen.medium.com/streaming-server-sent-events-with-go-8cc1f615d561",
	"https://pascalallen.medium.com/dispatching-events-with-react-and-typescript-89f80f07635f",
	"https://pascalallen.medium.com/jwt-authentication-with-go-242215a9b4f8",
	"https://pascalallen.medium.com/how-to-compile-a-webassembly-module-from-go-a9ed5f831582",
	"https://pascalallen.medium.com/how-to-deploy-to-kubernetes-76c42e5ea28c",
	"https://pascalallen.medium.com/how-to-build-a-containerized-web-app-in-go-73f42619a193",
	"https://pascalallen.medium.com/releasing-packages-to-github-and-the-npm-registry-8ff6c3789bc8",
	"https://pascalallen.medium.com/scrum-simplified-880113ed0db",
	"https://www.bizjournals.com/sanantonio/news/2016/11/23/divergent-career-paths-how-tech-talent-is-leaking.html",
}

var resumeSections = []string{
	"# Summary",
	"# Technical Skills",
	"# Professional Experience",
	"# Consulting Projects",
	"# Open Source & Writing",
	"# Education & Certifications",
}

var experience = map[string]string{
	"experience/01-sectigo.md":              "Sectigo",
	"experience/02-intuit-mailchimp.md":     "Intuit Mailchimp",
	"experience/03-monumental.md":           "Monumental",
	"experience/04-onlinemeded.md":          "OnlineMedEd",
	"experience/05-far-west-capital.md":     "Far West Capital",
	"experience/06-crimson-drive-design.md": "Crimson Drive Design",
}

var projects = map[string]string{
	"projects/pgmcp.md":                  "https://github.com/pascalallen/pgmcp",
	"projects/hmac.md":                   "https://pkg.go.dev/github.com/pascalallen/hmac/v2",
	"projects/pubsub.md":                 "https://pkg.go.dev/github.com/pascalallen/pubsub/v2",
	"projects/pgqueue.md":                "https://pkg.go.dev/github.com/pascalallen/pgqueue",
	"projects/ollama-cloud-go.md":        "https://github.com/pascalallen/ollama-cloud-go",
	"projects/mcp-server-go.md":          "https://github.com/pascalallen/mcp-server-go",
	"projects/es-go.md":                  "https://github.com/pascalallen/es-go",
	"projects/rpi-security-cam.md":       "https://github.com/pascalallen/rpi-security-cam",
	"projects/adk-agent.md":              "https://github.com/pascalallen/adk_agent",
	"projects/vllm-server.md":            "https://github.com/pascalallen/vllm-server",
	"projects/react-form-components.md":  "https://www.npmjs.com/package/@pascalallen/react-form-components",
	"projects/react-event-dispatcher.md": "https://www.npmjs.com/package/@pascalallen/react-event-dispatcher",
	"projects/pascalallen.com.md":        "https://github.com/pascalallen/pascalallen.com",
}

var contact = []string{
	"pascal.allen88@gmail.com",
	"https://pascalallen.com",
	"https://github.com/pascalallen",
	"https://pascalallen.medium.com",
	"https://www.linkedin.com/in/pascal-allen-942749112",
}

var certifications = []string{
	"Codeup — Full-Stack Web Development",
	"Harvard Online — CS50",
	"Google — Google Cloud Sales Credential",
	"Intuit — Secure Coding for Web Applications",
	"Anthropic — Claude Code 101",
	"Anthropic — Claude Code in Action",
	"Docker — Docker Foundations Professional Certificate",
}

var siteStack = []string{"Go", "WebAssembly", "TypeScript", "Docker", "Raspberry Pi"}

func read(t *testing.T, name string) string {
	t.Helper()
	b, err := fs.ReadFile(content.FS(), name)
	require.NoError(t, err, name)
	return string(b)
}

func all(t *testing.T) map[string]string {
	t.Helper()
	files := map[string]string{}
	err := fs.WalkDir(content.FS(), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			files[p] = read(t, p)
		}
		return nil
	})
	require.NoError(t, err)
	return files
}

func TestContent(t *testing.T) {
	t.Run("the home directory holds the documented top-level files", func(t *testing.T) {
		for _, name := range []string{"about.md", "contact.md", "skills.md", "resume.md", "education.md", "README.md"} {
			read(t, name)
		}
	})

	t.Run("every publication is a writing entry with its url", func(t *testing.T) {
		entries, err := fs.ReadDir(content.FS(), "writing")
		require.NoError(t, err)
		require.Len(t, entries, len(publications))
		for i, url := range publications {
			name := "writing/" + entries[i].Name()
			assert.Contains(t, read(t, name), url, name)
		}
	})

	t.Run("every resume section is in resume.md", func(t *testing.T) {
		resume := read(t, "resume.md")
		for _, section := range resumeSections {
			assert.Contains(t, resume, "\n"+section+"\n", section)
		}
	})

	t.Run("every employer has an experience entry newest first", func(t *testing.T) {
		entries, err := fs.ReadDir(content.FS(), "experience")
		require.NoError(t, err)
		require.Len(t, entries, len(experience))
		for name, company := range experience {
			assert.True(t, strings.HasPrefix(read(t, name), "# "+company), name)
			assert.Contains(t, read(t, "resume.md"), company, name)
		}
	})

	t.Run("every open-source entry has a project file with its url", func(t *testing.T) {
		entries, err := fs.ReadDir(content.FS(), "projects")
		require.NoError(t, err)
		require.Len(t, entries, len(projects))
		for name, url := range projects {
			assert.Contains(t, read(t, name), url, name)
		}
	})

	t.Run("contact lists email, site, github, medium and linkedin", func(t *testing.T) {
		text := read(t, "contact.md")
		for _, want := range contact {
			assert.Contains(t, text, want)
		}
	})

	t.Run("every certification is in education.md and resume.md", func(t *testing.T) {
		for _, cert := range certifications {
			assert.Contains(t, read(t, "education.md"), cert)
			assert.Contains(t, read(t, "resume.md"), cert)
		}
	})

	t.Run("the readme names the stack this site runs on", func(t *testing.T) {
		text := read(t, "README.md")
		for _, tech := range siteStack {
			assert.Contains(t, text, tech)
		}
	})

	t.Run("a prose line is never a mid-sentence continuation of the line before it", func(t *testing.T) {
		for name, text := range all(t) {
			lines := strings.Split(text, "\n")
			for i, line := range lines {
				if i == 0 || strings.TrimSpace(line) == "" || startsBlock(line) {
					continue
				}
				prev := lines[i-1]
				assert.True(t, mayPrecedeProse(prev) || isLabel(prev, line), name+": "+line)
			}
		}
	})

	t.Run("no file publishes a phone number", func(t *testing.T) {
		for name, text := range all(t) {
			assert.False(t, hasPhoneNumber(text), name)
		}
	})

	t.Run("every file is valid utf-8 ending in a newline", func(t *testing.T) {
		for name, text := range all(t) {
			assert.True(t, utf8.ValidString(text), name)
			assert.True(t, strings.HasSuffix(text, "\n"), name)
		}
	})
}

func hasPhoneNumber(text string) bool {
	digits := 0
	for _, r := range text {
		switch {
		case r >= '0' && r <= '9':
			digits++
			if digits >= 10 {
				return true
			}
		case r == ' ' || r == '-' || r == '.' || r == '(' || r == ')' || r == '+':
		default:
			digits = 0
		}
	}
	return false
}

var (
	keyValue = regexp.MustCompile(`\S\s{2,}\S`)
	meta     = regexp.MustCompile(`· \d{4}`)
)

func startsBlock(line string) bool {
	return strings.HasPrefix(line, "#") ||
		strings.HasPrefix(line, "- ") ||
		strings.HasPrefix(line, "Stack:") ||
		strings.HasPrefix(strings.TrimSpace(line), "https://") ||
		keyValue.MatchString(line) ||
		meta.MatchString(line)
}

func mayPrecedeProse(line string) bool {
	return strings.TrimSpace(line) == "" ||
		strings.HasPrefix(line, "#") ||
		strings.HasPrefix(strings.TrimSpace(line), "https://") ||
		keyValue.MatchString(line) ||
		meta.MatchString(line)
}

func isLabel(prev, line string) bool {
	return strings.HasPrefix(line, "  ") && !strings.HasPrefix(prev, " ") && !startsBlock(prev)
}
