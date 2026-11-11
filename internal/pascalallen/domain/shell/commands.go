package shell

import "strings"

type argKind int

const (
	argNone argKind = iota
	argPath
	argDir
	argCommand
	argTheme
)

type command struct {
	name     string
	synopsis string
	summary  string
	help     string
	args     argKind
	maxArgs  int
	hidden   bool
	run      func(s *Shell, o *output, args []string)
}

func registry() []command {
	return []command{
		{name: "help", synopsis: "help [command]", summary: "list every command", args: argCommand, maxArgs: 1, run: cmdHelp,
			help: "Lists every command with a one-line summary.\nWith a command name, shows its manual page, like man."},
		{name: "man", synopsis: "man <command>", summary: "show the manual page for a command", args: argCommand, maxArgs: 1, run: cmdMan,
			help: "Shows the name, synopsis and description of a command."},
		{name: "ls", synopsis: "ls [-la] [path]", summary: "list directory contents", args: argPath, maxArgs: -1, run: cmdLs,
			help: "Lists the files in a directory, the current one by default.\n-l  long format: permissions, owner and size\n-a  include . and .."},
		{name: "cd", synopsis: "cd [path|-]", summary: "change the working directory", args: argDir, maxArgs: 1, run: cmdCd,
			help: "Changes the working directory. With no path, goes home (~).\n- returns to the previous directory. You cannot climb above home."},
		{name: "pwd", synopsis: "pwd", summary: "print the working directory", maxArgs: 0, run: cmdPwd,
			help: "Prints the absolute path of the working directory."},
		{name: "cat", synopsis: "cat <file>...", summary: "print files", args: argPath, maxArgs: -1, run: cmdCat,
			help: "Prints one or more files. Links in the text are clickable."},
		{name: "tree", synopsis: "tree [path]", summary: "show the directory tree", args: argDir, maxArgs: 1, run: cmdTree,
			help: "Draws every file and directory below a path, the current one by default."},
		{name: "head", synopsis: "head [-n lines] <file>...", summary: "print the first lines of files", args: argPath, maxArgs: -1, run: headTail("head", false),
			help: "Prints the first 10 lines of each file.\n-n N   print N lines instead (also -N)\n-n -N  print all but the last N lines"},
		{name: "tail", synopsis: "tail [-n lines] <file>...", summary: "print the last lines of files", args: argPath, maxArgs: -1, run: headTail("tail", true),
			help: "Prints the last 10 lines of each file.\n-n N   print N lines instead (also -N)\n-n +N  print from line N to the end"},
		{name: "grep", synopsis: "grep [-i] <text> [path]", summary: "search files for text", args: argPath, maxArgs: -1, run: cmdGrep,
			help: "Prints every line containing the text, searching all files below the path.\nThe text is matched literally, not as a pattern.\n-i  ignore case"},
		{name: "whoami", synopsis: "whoami", summary: "print your user name", maxArgs: 0, run: cmdWhoami,
			help: "Prints the name you are logged in as."},
		{name: "echo", synopsis: "echo [text]...", summary: "print text", maxArgs: -1, run: cmdEcho,
			help: "Prints its arguments separated by spaces."},
		{name: "date", synopsis: "date", summary: "print the date and time", maxArgs: 0, run: cmdDate,
			help: "Prints the current date and time."},
		{name: "history", synopsis: "history [n]", summary: "list recent commands", maxArgs: 1, run: cmdHistory,
			help: "Lists the commands you have run, oldest first, or only the last n.\nThe Up and Down keys walk the same list."},
		{name: "clear", synopsis: "clear", summary: "clear the screen", maxArgs: 0, run: cmdClear,
			help: "Clears the screen. Ctrl-L does the same."},
		{name: "open", synopsis: "open <url|file>", summary: "open a link in a new tab", args: argPath, maxArgs: 1, run: cmdOpen,
			help: "Opens an http, https or mailto link in a new tab.\nGiven a file, opens the first web link in it."},
		{name: "about", synopsis: "about", summary: "who Pascal is", maxArgs: 0, run: showFile("about", "about.md"),
			help: "Prints ~/about.md."},
		{name: "resume", synopsis: "resume", summary: "the full resume", maxArgs: 0, run: showFile("resume", "resume.md"),
			help: "Prints ~/resume.md."},
		{name: "skills", synopsis: "skills", summary: "technical skills", maxArgs: 0, run: showFile("skills", "skills.md"),
			help: "Prints ~/skills.md."},
		{name: "experience", synopsis: "experience", summary: "work history, newest first", maxArgs: 0, run: listing("experience", "Experience", true),
			help: "Lists ~/experience. Tap an entry, or cat it, to read it."},
		{name: "projects", synopsis: "projects", summary: "open-source projects", maxArgs: 0, run: listing("projects", "Projects", false),
			help: "Lists ~/projects. Tap an entry, or cat it, to read it."},
		{name: "writing", synopsis: "writing", summary: "articles and press, newest first", maxArgs: 0, run: listing("writing", "Writing", false),
			help: "Lists ~/writing. Tap an entry, or cat it, to read it."},
		{name: "contact", synopsis: "contact", summary: "email and social links", maxArgs: 0, run: showFile("contact", "contact.md"),
			help: "Prints ~/contact.md."},
		{name: "education", synopsis: "education", summary: "education and certifications", maxArgs: 0, run: showFile("education", "education.md"),
			help: "Prints ~/education.md."},
		{name: "repos", synopsis: "repos", summary: "live list of GitHub repositories", maxArgs: 0, run: fetchFeed(FeedRepos, "fetching repositories from GitHub..."),
			help: "Fetches the most recently updated public repositories from GitHub."},
		{name: "packages", synopsis: "packages", summary: "live list of npm packages", maxArgs: 0, run: fetchFeed(FeedPackages, "fetching packages from npm..."),
			help: "Fetches the published npm packages."},
		{name: "camera", synopsis: "camera", summary: "live camera stream from a Raspberry Pi", maxArgs: 0, run: cmdCamera,
			help: "Shows the live MJPEG stream from the Raspberry Pi that serves this site."},
		{name: "theme", synopsis: "theme [name]", summary: "list or set the color theme", args: argTheme, maxArgs: 1, run: cmdTheme,
			help: "With no name, lists the themes and marks the current one.\nWith a name, switches to it: " + strings.Join(themes, ", ") + "."},
		{name: "exit", synopsis: "exit", summary: "leave the shell", maxArgs: 0, run: cmdExit,
			help: "Ends the session, as far as a web page can."},
		{name: "sudo", synopsis: "sudo <command>", summary: "run a command as root", args: argCommand, maxArgs: -1, hidden: true, run: cmdSudo,
			help: "Runs a command as root, if you are in the sudoers file."},
		{name: "vim", synopsis: "vim [file]", summary: "edit a file", args: argPath, maxArgs: 1, hidden: true, run: cmdVim,
			help: "Opens a file in the editor."},
		{name: "rm", synopsis: "rm [-rf] <file>...", summary: "remove files", args: argPath, maxArgs: -1, hidden: true, run: cmdRm,
			help: "Removes files. The file system is read-only."},
		{name: "mkdir", synopsis: "mkdir <dir>...", summary: "make directories", args: argPath, maxArgs: -1, hidden: true, run: readOnly("mkdir", "cannot create directory"),
			help: "Creates directories. The file system is read-only."},
		{name: "touch", synopsis: "touch <file>...", summary: "create files", args: argPath, maxArgs: -1, hidden: true, run: readOnly("touch", "cannot touch"),
			help: "Creates files. The file system is read-only."},
		{name: "mv", synopsis: "mv <source> <dest>", summary: "move files", args: argPath, maxArgs: -1, hidden: true, run: cmdMv,
			help: "Moves files. The file system is read-only."},
	}
}
