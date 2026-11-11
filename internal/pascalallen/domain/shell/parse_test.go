package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type parseCase struct {
	name  string
	input string
	want  []step
	err   string
}

var parseCases = []parseCase{
	{name: "empty input has no steps", input: "", want: nil},
	{name: "whitespace-only input has no steps", input: " \t  ", want: nil},
	{name: "words split on whitespace", input: "  ls   -la\twriting ", want: []step{{words: []string{"ls", "-la", "writing"}}}},
	{name: "single quotes keep everything literal", input: `echo 'a  "b" \c'`, want: []step{{words: []string{"echo", `a  "b" \c`}}}},
	{name: "double quotes keep spaces", input: `echo "hello   world"`, want: []step{{words: []string{"echo", "hello   world"}}}},
	{name: "double quotes honour backslash before a quote", input: `echo "say \"hi\""`, want: []step{{words: []string{"echo", `say "hi"`}}}},
	{name: "double quotes keep other backslashes", input: `echo "a\nb"`, want: []step{{words: []string{"echo", `a\nb`}}}},
	{name: "a backslash escapes a space outside quotes", input: `cat my\ file`, want: []step{{words: []string{"cat", "my file"}}}},
	{name: "a backslash escapes a special character", input: `echo \| \; \#`, want: []step{{words: []string{"echo", "|", ";", "#"}}}},
	{name: "a trailing backslash is literal", input: `echo a\`, want: []step{{words: []string{"echo", `a\`}}}},
	{name: "adjacent quoted parts join one word", input: `echo a'b'"c"`, want: []step{{words: []string{"echo", "abc"}}}},
	{name: "empty quotes make an empty word", input: `echo ""`, want: []step{{words: []string{"echo", ""}}}},
	{name: "a comment ends the line", input: "ls # list files", want: []step{{words: []string{"ls"}}}},
	{name: "a hash inside a word is literal", input: "echo a#b", want: []step{{words: []string{"echo", "a#b"}}}},
	{name: "a comment-only line has no steps", input: "# nothing", want: nil},
	{name: "semicolons sequence commands", input: "pwd; ls ;whoami", want: []step{{words: []string{"pwd"}}, {words: []string{"ls"}}, {words: []string{"whoami"}}}},
	{name: "and-and joins commands conditionally", input: "cd writing && ls", want: []step{{words: []string{"cd", "writing"}}, {words: []string{"ls"}, andThen: true}}},
	{name: "a trailing semicolon is allowed", input: "ls;", want: []step{{words: []string{"ls"}}}},
	{name: "quoted separators are literal", input: `echo "a;b && c"`, want: []step{{words: []string{"echo", "a;b && c"}}}},
	{name: "an unterminated double quote is an error", input: `echo "abc`, err: "unexpected EOF while looking for matching `\"'"},
	{name: "an unterminated single quote is an error", input: "echo 'abc", err: "unexpected EOF while looking for matching `''"},
	{name: "a leading semicolon is a syntax error", input: "; ls", err: "syntax error near unexpected token `;'"},
	{name: "a double semicolon is a syntax error", input: "ls ;; ls", err: "syntax error near unexpected token `;'"},
	{name: "a leading and-and is a syntax error", input: "&& ls", err: "syntax error near unexpected token `&&'"},
	{name: "a trailing and-and is a syntax error", input: "ls &&", err: "syntax error: unexpected end of line after `&&'"},
	{name: "pipes are unsupported", input: "cat about.md | grep Go", err: "pipes are not supported"},
	{name: "redirection is unsupported", input: "echo hi > file", err: "redirection is not supported"},
	{name: "input redirection is unsupported", input: "cat < file", err: "redirection is not supported"},
	{name: "background jobs are unsupported", input: "ls &", err: "background jobs are not supported"},
	{name: "a single ampersand before a word is a background job", input: "ls &pwd", err: "background jobs are not supported"},
	{name: "variables are unsupported", input: "echo $HOME", err: "variables and substitution are not supported"},
	{name: "variables inside double quotes are unsupported", input: `echo "$HOME"`, err: "variables and substitution are not supported"},
	{name: "command substitution is unsupported", input: "echo `ls`", err: "variables and substitution are not supported"},
	{name: "glob characters are literal", input: "cat *.md [ab]", want: []step{{words: []string{"cat", "*.md", "[ab]"}}}},
	{name: "a question mark in a url is literal", input: "open https://example.com/?q=1", want: []step{{words: []string{"open", "https://example.com/?q=1"}}}},
	{name: "quoted glob characters are literal", input: `grep "*" '?'`, want: []step{{words: []string{"grep", "*", "?"}}}},
	{name: "a dollar sign in single quotes is literal", input: `echo '$HOME'`, want: []step{{words: []string{"echo", "$HOME"}}}},
	{name: "an escaped dollar sign in double quotes is literal", input: `echo "\$HOME"`, want: []step{{words: []string{"echo", "$HOME"}}}},
	{name: "newlines and carriage returns are whitespace", input: "ls\r\nwriting", want: []step{{words: []string{"ls", "writing"}}}},
}

func TestParse(t *testing.T) {
	for _, tc := range parseCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parse(tc.input)
			if tc.err != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
