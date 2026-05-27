package main

import "strings"

const bioText = "Senior software engineer in Austin, TX — 10+ years across the full stack, with a deep focus on Go, distributed systems, and domain-driven design. I build in the open, write about what I learn, and occasionally over-engineer things for fun."

var filesystem = map[string]string{
	"bio.txt": bioText,
}

func executeCommandLogic(cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return ""
	}

	switch parts[0] {
	case "ls":
		return "bio.txt"
	case "cat":
		if len(parts) < 2 {
			return "cat: missing file operand"
		}
		content, ok := filesystem[parts[1]]
		if !ok {
			return "cat: " + parts[1] + ": No such file or directory"
		}
		return content
	case "whoami":
		return "pascal"
	case "help":
		return "available commands: cat, clear, help, ls, whoami"
	case "clear":
		return "__CLEAR__"
	default:
		return "bash: " + parts[0] + ": command not found"
	}
}
