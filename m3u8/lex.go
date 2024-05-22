package m3u8

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

type item struct {
	typ itemType
	val string
}

func (it item) String() string {
	if it.typ == itemNewline {
		return "newline"
	}
	return fmt.Sprintf("%s: %s", it.typ, it.val)
}

type itemType int

const (
	itemError itemType = iota
	itemTag
	itemAttrName
	itemEquals
	itemNumber
	itemString
	itemComma
	itemURL
	itemNewline
	itemEOF
)

func (t itemType) String() string {
	switch t {
	case itemError:
		return "error"
	case itemTag:
		return "tag"
	case itemAttrName:
		return "attribute name"
	case itemEquals:
		return "equals"
	case itemNumber:
		return "number"
	case itemString:
		return "string"
	case itemComma:
		return "comma"
	case itemURL:
		return "url"
	case itemNewline:
		return "newline"
	case itemEOF:
		return "EOF"
	}
	return "unknown item type"
}

const tagStart = "#EXT"

// A lexer... TODO
// The design is described in "Lexical Scanning in Go" by Rob Pike:
// https://www.youtube.com/watch?v=HxaD_trXwRE
type lexer struct {
	sc    *bufio.Scanner
	input string
	start int
	pos   int
	width int
	items chan item
}

type stateFn func(*lexer) stateFn

func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		return -1
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// ignore skips the current rune.
func (l *lexer) ignore() { l.start = l.pos }

// backup steps the lexer back one rune.
func (l *lexer) backup() { l.pos -= l.width }

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) errorf(format string, a ...any) stateFn {
	err := fmt.Sprintf(format, a...)
	l.items <- item{itemError, err}
	return nil
}

func (l *lexer) run() {
	for state := lexStart; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	// fmt.Println(item{t, l.input[l.start:l.pos]})
	l.start = l.pos
}

func newLexer(r io.Reader) *lexer {
	return &lexer{
		sc:    bufio.NewScanner(r),
		items: make(chan item),
	}
}

func lexStart(l *lexer) stateFn {
	for l.sc.Scan() {
		if l.sc.Text() == "" {
			continue // ignore blank lines
		}
		l.input = l.sc.Text() + "\n"
		l.pos = 0
		l.start = 0
		if strings.HasPrefix(l.input, tagStart) {
			return lexTag(l)
		} else if strings.HasPrefix(l.input, "#") {
			continue // ignore comments
		}
		// not a tag, so must be a URL.
		// emit the URL, then the newline we appended ourselves.
		l.pos = len(l.sc.Text())
		l.emit(itemURL)
		l.emit(itemNewline)
	}
	if err := l.sc.Err(); err != nil {
		panic(err)
	}
	return nil
}

func lexTag(l *lexer) stateFn {
	r := l.next()
	if r != '#' {
		return l.errorf("missing starting #")
	}
	return lexTagName(l)
}

func lexTagName(l *lexer) stateFn {
	for {
		r := l.peek()
		if isTagNameChar(r) {
			l.next()
			continue
		}
		switch r {
		case '\n':
			l.emit(itemTag)
			l.next()
			l.emit(itemNewline)
			return lexStart(l)
		case ':':
			l.emit(itemTag)
			l.next()
			l.ignore()
			return lexAttrs(l)
		}
		return l.errorf("illegal tag character %q", r)
	}
}

func isTagNameChar(r rune) bool {
	if r >= 'A' && r <= 'Z' {
		return true
	}
	switch r {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	case '-':
		return true
	}
	return false
}

func lexAttrs(l *lexer) stateFn {
	for {
		switch r := l.peek(); {
		case isTagNameChar(r):
			l.next()
			continue
		case r == '\n':
			if len(l.input[l.start:l.pos]) != 0 {
				l.emit(itemAttrName)
			}
			l.next()
			l.emit(itemNewline)
			return lexStart(l)
		case r == '=':
			l.emit(itemAttrName)
			l.next()
			l.emit(itemEquals)
			return lexAttrValue(l)
		case r == ',':
			l.next()
			l.emit(itemComma)
			return lexAttrs(l)
		case r == '.':
			return lexAttrValue(l)
		case r == '@':
			return lexAttrValue(l)
		default:
			return l.errorf("illegal character %q in attribute name", r)
		}
	}
}

func lexAttrValue(l *lexer) stateFn {
	r := l.next()
	switch r {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.':
		return lexNumber(l)
	case '"':
		return lexQString(l)
	}
	if isTagNameChar(r) {
		return lexRawString(l)
	}
	return l.errorf("unquoted string starting with illegal character %q", r)
}

func lexNumber(l *lexer) stateFn {
	for {
		switch r := l.peek(); r {
		case 'x', '@':
			// are we lexing a resolution? e.g. 640x480
			// or a byte range? e.g. 69@3000
			return lexRawString(l)
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.':
			l.next()
			continue
		default:
			l.emit(itemNumber)
			return lexAttrs(l)
		}
	}
}

func lexQString(l *lexer) stateFn {
	for {
		r := l.next()
		if r == '"' {
			l.emit(itemString)
			return lexAttrs(l)
		} else if r == '\n' {
			return l.errorf("unterminated quoted string")
		}
	}
}

func lexRawString(l *lexer) stateFn {
	for {
		if l.peek() == ',' || l.peek() == '\n' {
			break
		}
		l.next()
	}
	l.emit(itemString)
	return lexAttrs(l)
}
