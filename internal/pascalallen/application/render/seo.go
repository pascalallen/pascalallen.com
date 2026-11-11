package render

import (
	"encoding/json"
	"io/fs"
	"strings"
)

type profile struct {
	Name        string
	JobTitle    string
	Description string
	SameAs      []string
	KnowsAbout  []string
}

func readProfile(fsys fs.FS) profile {
	p := profile{Name: "Pascal Allen"}
	if data, err := fs.ReadFile(fsys, "about.md"); err == nil {
		var prose []string
		tagline := false
		paragraphDone := false
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			switch {
			case strings.HasPrefix(line, "# "):
				p.Name = strings.TrimSpace(strings.TrimPrefix(line, "# "))
			case line == "":
				if len(prose) > 0 {
					paragraphDone = true
				}
			case paragraphDone:
			case !tagline:
				tagline = true
				title, _, _ := strings.Cut(line, " · ")
				p.JobTitle = strings.TrimSpace(title)
			default:
				prose = append(prose, line)
			}
		}
		p.Description = truncate(strings.Join(prose, " "), 160)
	}
	if p.Description == "" {
		p.Description = p.Name + ", software engineer."
	}
	if data, err := fs.ReadFile(fsys, "contact.md"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			fields := strings.Fields(line)
			if len(fields) != 2 {
				continue
			}
			switch fields[0] {
			case "GitHub", "LinkedIn", "Medium":
				if _, ok := safeHref(fields[1]); ok && strings.HasPrefix(fields[1], "http") {
					p.SameAs = append(p.SameAs, fields[1])
				}
			}
		}
	}
	if data, err := fs.ReadFile(fsys, "skills.md"); err == nil {
		p.KnowsAbout = parseSkills(string(data))
	}
	return p
}

func parseSkills(text string) []string {
	var items, group []string
	flush := func() {
		for _, part := range strings.Split(strings.Join(group, " "), ",") {
			if part = strings.TrimSpace(part); part != "" {
				items = append(items, part)
			}
		}
		group = nil
	}
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "  ") && strings.TrimSpace(line) != "" {
			group = append(group, strings.TrimSpace(line))
			continue
		}
		flush()
	}
	flush()
	seen := map[string]bool{}
	var out []string
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			out = append(out, item)
		}
	}
	return out
}

func truncate(s string, limit int) string {
	r := []rune(s)
	if len(r) <= limit {
		return s
	}
	cut := string(r[:limit-1])
	if i := strings.LastIndex(cut, " "); i > 0 {
		cut = cut[:i]
	}
	return strings.TrimRight(cut, " ,.;") + "…"
}

func jsonLD(p profile, baseURL string) string {
	doc := map[string]any{
		"@context": "https://schema.org",
		"@type":    "Person",
		"name":     p.Name,
	}
	if p.JobTitle != "" {
		doc["jobTitle"] = p.JobTitle
	}
	if baseURL != "" {
		doc["url"] = baseURL
	}
	if len(p.SameAs) > 0 {
		doc["sameAs"] = p.SameAs
	}
	if len(p.KnowsAbout) > 0 {
		doc["knowsAbout"] = p.KnowsAbout
	}
	out, err := json.Marshal(doc)
	if err != nil {
		return "{}"
	}
	return string(out)
}
