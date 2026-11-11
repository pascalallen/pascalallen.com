package shell

import (
	"errors"
	"strings"
)

type step struct {
	words   []string
	andThen bool
}

var (
	errPipes        = errors.New("pipes are not supported in this shell")
	errRedirect     = errors.New("redirection is not supported in this shell")
	errBackground   = errors.New("background jobs are not supported in this shell")
	errSubstitution = errors.New("variables and substitution are not supported in this shell")
	errTrailingAnd  = errors.New("syntax error: unexpected end of line after `&&'")
)

func errUnterminated(quote rune) error {
	return errors.New("unexpected EOF while looking for matching `" + string(quote) + "'")
}

func errUnexpected(token string) error {
	return errors.New("syntax error near unexpected token `" + token + "'")
}

type parser struct {
	steps      []step
	words      []string
	word       strings.Builder
	inWord     bool
	pendingAnd bool
}

func (p *parser) flushWord() {
	if !p.inWord {
		return
	}
	p.words = append(p.words, p.word.String())
	p.word.Reset()
	p.inWord = false
}

func (p *parser) endStep(sep string) error {
	p.flushWord()
	if len(p.words) == 0 {
		return errUnexpected(sep)
	}
	p.steps = append(p.steps, step{words: p.words, andThen: p.pendingAnd})
	p.words = nil
	p.pendingAnd = sep == "&&"
	return nil
}

func (p *parser) singleQuoted(rs []rune, i int) (int, error) {
	for j := i + 1; j < len(rs); j++ {
		if rs[j] == '\'' {
			p.word.WriteString(string(rs[i+1 : j]))
			return j, nil
		}
	}
	return 0, errUnterminated('\'')
}

func (p *parser) doubleQuoted(rs []rune, i int) (int, error) {
	for j := i + 1; j < len(rs); j++ {
		c := rs[j]
		switch {
		case c == '"':
			return j, nil
		case c == '\\' && j+1 < len(rs) && strings.ContainsRune("\"\\$`", rs[j+1]):
			p.word.WriteRune(rs[j+1])
			j++
		case c == '$' || c == '`':
			return 0, errSubstitution
		default:
			p.word.WriteRune(c)
		}
	}
	return 0, errUnterminated('"')
}

func parse(line string) ([]step, error) {
	p := &parser{}
	rs := []rune(line)
	var err error
	for i := 0; i < len(rs); i++ {
		r := rs[i]
		switch r {
		case ' ', '\t', '\n', '\r':
			p.flushWord()
		case '\'':
			p.inWord = true
			i, err = p.singleQuoted(rs, i)
		case '"':
			p.inWord = true
			i, err = p.doubleQuoted(rs, i)
		case '\\':
			p.inWord = true
			if i+1 < len(rs) {
				i++
				p.word.WriteRune(rs[i])
			} else {
				p.word.WriteRune('\\')
			}
		case '#':
			if !p.inWord {
				i = len(rs)
				continue
			}
			p.word.WriteRune(r)
		case ';':
			err = p.endStep(";")
		case '&':
			if i+1 < len(rs) && rs[i+1] == '&' {
				i++
				err = p.endStep("&&")
			} else {
				err = errBackground
			}
		case '|':
			err = errPipes
		case '<', '>':
			err = errRedirect
		case '$', '`':
			err = errSubstitution
		default:
			p.inWord = true
			p.word.WriteRune(r)
		}
		if err != nil {
			return nil, err
		}
	}
	p.flushWord()
	if len(p.words) > 0 {
		p.steps = append(p.steps, step{words: p.words, andThen: p.pendingAnd})
	} else if p.pendingAnd {
		return nil, errTrailingAnd
	}
	return p.steps, nil
}
