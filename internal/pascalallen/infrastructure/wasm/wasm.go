//go:build js && wasm

package main

import (
	"fmt"
	"strings"
	"syscall/js"
)

// ---------------------------
// Simple in-memory filesystem
// ---------------------------

type node struct {
	name     string
	isDir    bool
	content  string
	children map[string]*node
	parent   *node
}

var (
	fsRoot *node
	cwd    *node
)

func newDir(name string, parent *node) *node {
	return &node{name: name, isDir: true, children: map[string]*node{}, parent: parent}
}

func newFile(name, content string, parent *node) *node {
	return &node{name: name, isDir: false, content: content, parent: parent}
}

func initFS() {
	fsRoot = newDir("/", nil)
	cwd = fsRoot
	// seed with some files
	home := mkdirNode(fsRoot, "home")
	user := mkdirNode(home, "wasm")
	mkdirNode(user, "projects")
	user.children["readme.txt"] = newFile("readme.txt", "Welcome to the in-browser WASM shell.\nType 'help' to see available commands.", user)
}

func mkdirNode(dir *node, name string) *node {
	if dir.children[name] != nil {
		return dir.children[name]
	}
	n := newDir(name, dir)
	dir.children[name] = n
	return n
}

func pathOf(n *node) string {
	if n == fsRoot {
		return "/"
	}
	parts := []string{}
	cur := n
	for cur != nil && cur != fsRoot {
		parts = append([]string{cur.name}, parts...)
		cur = cur.parent
	}
	return "/" + strings.Join(parts, "/")
}

func resolvePath(p string) (*node, *node, string) {
	// returns (targetNode, parentForNew, lastName)
	if p == "" {
		return cwd, cwd.parent, cwd.name
	}
	var start *node
	if strings.HasPrefix(p, "/") {
		start = fsRoot
		p = strings.TrimPrefix(p, "/")
	} else {
		start = cwd
	}
	segments := []string{}
	for _, seg := range strings.Split(p, "/") {
		if seg == "" || seg == "." {
			continue
		}
		if seg == ".." {
			if start.parent != nil {
				start = start.parent
			}
			continue
		}
		segments = append(segments, seg)
	}
	parent := start
	for i, seg := range segments {
		child := parent.children[seg]
		if child == nil {
			// not found
			return nil, parent, strings.Join(append([]string{seg}, segments[i+1:]...), "/")
		}
		if i == len(segments)-1 {
			return child, parent, seg
		}
		parent = child
	}
	return parent, parent.parent, parent.name
}

// ---------------------------
// Shell implementation
// ---------------------------

func shellPrompt() string {
	return fmt.Sprintf("wasm@browser:%s$ ", pathOf(cwd))
}

func shellHandle(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}
	// basic quoting: split by spaces but keep quoted strings
	args := splitArgs(line)
	cmd := args[0]
	args = args[1:]
	switch cmd {
	case "help":
		return "Available commands: help, pwd, ls, cd, mkdir, touch, echo, cat, rm, clear"
	case "pwd":
		return pathOf(cwd)
	case "ls":
		var target *node
		if len(args) == 0 {
			target = cwd
		} else {
			if n, _, _ := resolvePath(args[0]); n != nil && n.isDir {
				target = n
			} else if n != nil && !n.isDir {
				return args[0]
			} else {
				return fmt.Sprintf("ls: cannot access '%s': No such file or directory", args[0])
			}
		}
		names := []string{}
		for name := range target.children {
			names = append(names, name)
		}
		return strings.Join(sortStrings(names), "\t")
	case "cd":
		if len(args) == 0 {
			cwd = fsRoot
			return ""
		}
		if n, _, _ := resolvePath(args[0]); n != nil && n.isDir {
			cwd = n
			return ""
		}
		return fmt.Sprintf("cd: no such file or directory: %s", args[0])
	case "mkdir":
		if len(args) == 0 {
			return "mkdir: missing operand"
		}
		_, parent, name := resolvePath(args[0])
		if parent == nil {
			return fmt.Sprintf("mkdir: cannot create directory ‘%s’", args[0])
		}
		if parent.children[name] != nil {
			return fmt.Sprintf("mkdir: cannot create directory ‘%s’: File exists", args[0])
		}
		mkdirNode(parent, name)
		return ""
	case "touch":
		if len(args) == 0 {
			return "touch: missing file operand"
		}
		for _, p := range args {
			_, parent, name := resolvePath(p)
			if parent == nil {
				continue
			}
			if parent.children[name] == nil {
				parent.children[name] = newFile(name, "", parent)
			}
		}
		return ""
	case "echo":
		return strings.Join(args, " ")
	case "cat":
		if len(args) == 0 {
			return "cat: missing file operand"
		}
		if n, _, _ := resolvePath(args[0]); n != nil && !n.isDir {
			return n.content
		}
		return fmt.Sprintf("cat: %s: No such file", args[0])
	case "rm":
		if len(args) == 0 {
			return "rm: missing operand"
		}
		for _, p := range args {
			if _, parent, name := resolvePath(p); parent != nil {
				if child := parent.children[name]; child != nil {
					if child.isDir && len(child.children) > 0 {
						return fmt.Sprintf("rm: cannot remove '%s': Is a directory", p)
					}
					delete(parent.children, name)
				} else {
					return fmt.Sprintf("rm: cannot remove '%s': No such file or directory", p)
				}
			}
		}
		return ""
	case "clear":
		return "__CLEAR__"
	default:
		return fmt.Sprintf("%s: command not found", cmd)
	}
}

func splitArgs(s string) []string {
	res := []string{}
	cur := strings.Builder{}
	inQuote := false
	quoteChar := byte(0)
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if inQuote {
			if ch == quoteChar {
				inQuote = false
			} else {
				cur.WriteByte(ch)
			}
			continue
		}
		if ch == '\'' || ch == '"' {
			inQuote = true
			quoteChar = ch
			continue
		}
		if ch == ' ' || ch == '\t' {
			if cur.Len() > 0 {
				res = append(res, cur.String())
				cur.Reset()
			}
			continue
		}
		cur.WriteByte(ch)
	}
	if cur.Len() > 0 {
		res = append(res, cur.String())
	}
	return res
}

func sortStrings(a []string) []string {
	// simple insertion sort to avoid importing sort
	for i := 1; i < len(a); i++ {
		j := i
		for j > 0 && a[j-1] > a[j] {
			a[j-1], a[j] = a[j], a[j-1]
			j--
		}
	}
	return a
}

// ---------------------------
// JS bindings
// ---------------------------

func terminalInit(this js.Value, args []js.Value) interface{} {
	initFS()
	return shellPrompt()
}

func terminalHandleInput(this js.Value, args []js.Value) interface{} {
	if len(args) == 0 {
		return ""
	}
	return shellHandle(args[0].String())
}

func terminalGetPrompt(this js.Value, args []js.Value) interface{} {
	return shellPrompt()
}

func terminalReset(this js.Value, args []js.Value) interface{} {
	initFS()
	return ""
}

func main() {
	initFS()
	c := make(chan struct{})
	js.Global().Set("terminal_init", js.FuncOf(terminalInit))
	js.Global().Set("terminal_handle_input", js.FuncOf(terminalHandleInput))
	js.Global().Set("terminal_get_prompt", js.FuncOf(terminalGetPrompt))
	js.Global().Set("terminal_reset", js.FuncOf(terminalReset))
	<-c
}
