package wkt2

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// tokenKind enumerates the lexical categories produced by the WKT2 lexer.
type tokenKind uint8

const (
	tokEOF tokenKind = iota
	tokKeyword
	tokString
	tokNumber
	tokLBracket // either '[' or '('
	tokRBracket // either ']' or ')'
	tokComma
)

func (k tokenKind) String() string {
	switch k {
	case tokEOF:
		return "EOF"
	case tokKeyword:
		return "keyword"
	case tokString:
		return "string"
	case tokNumber:
		return "number"
	case tokLBracket:
		return "'['"
	case tokRBracket:
		return "']'"
	case tokComma:
		return "','"
	default:
		return fmt.Sprintf("token(%d)", k)
	}
}

// token is a single lexical unit. Offset is the byte offset of the first
// rune of the token in the input string; it is used to make error
// messages diagnosable.
type token struct {
	kind   tokenKind
	value  string // unquoted for tokString; uppercased for tokKeyword
	offset int
}

// lexer turns a WKT2 input string into a stream of tokens. It is a
// hand-rolled scanner — WKT2's lexical surface is small enough that a
// scanner generator would be overkill.
type lexer struct {
	src string
	pos int // byte offset into src
}

func newLexer(src string) *lexer { return &lexer{src: src} }

// SyntaxError carries an offset so callers can point at the exact byte
// in the input that triggered the failure.
type SyntaxError struct {
	Offset int
	Msg    string
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("wkt2: syntax error at offset %d: %s", e.Offset, e.Msg)
}

func errAt(off int, format string, args ...any) error {
	return &SyntaxError{Offset: off, Msg: fmt.Sprintf(format, args...)}
}

// next returns the next token. On end-of-input it returns a tokEOF token
// rather than an error so the parser can decide whether EOF is welcome.
func (l *lexer) next() (token, error) {
	l.skipWhitespace()
	if l.pos >= len(l.src) {
		return token{kind: tokEOF, offset: l.pos}, nil
	}
	off := l.pos
	c := l.src[l.pos]
	switch {
	case c == '[' || c == '(':
		l.pos++
		return token{kind: tokLBracket, value: string(c), offset: off}, nil
	case c == ']' || c == ')':
		l.pos++
		return token{kind: tokRBracket, value: string(c), offset: off}, nil
	case c == ',':
		l.pos++
		return token{kind: tokComma, offset: off}, nil
	case c == '"':
		return l.readString()
	case c == '-' || c == '+' || c == '.' || (c >= '0' && c <= '9'):
		return l.readNumber()
	default:
		if isKeywordStart(rune(c)) {
			return l.readKeyword()
		}
		r, _ := utf8.DecodeRuneInString(l.src[l.pos:])
		return token{}, errAt(off, "unexpected character %q", r)
	}
}

func (l *lexer) skipWhitespace() {
	for l.pos < len(l.src) {
		r, size := utf8.DecodeRuneInString(l.src[l.pos:])
		if !unicode.IsSpace(r) {
			return
		}
		l.pos += size
	}
}

// readString consumes a "double-quoted" string. WKT2 escapes an embedded
// quote by doubling it ("" inside a string), per the spec. Newlines are
// allowed inside strings.
func (l *lexer) readString() (token, error) {
	off := l.pos
	l.pos++ // consume opening "
	var b strings.Builder
	for l.pos < len(l.src) {
		c := l.src[l.pos]
		if c == '"' {
			// Check for "" escape.
			if l.pos+1 < len(l.src) && l.src[l.pos+1] == '"' {
				b.WriteByte('"')
				l.pos += 2
				continue
			}
			l.pos++ // consume closing "
			return token{kind: tokString, value: b.String(), offset: off}, nil
		}
		b.WriteByte(c)
		l.pos++
	}
	return token{}, errAt(off, "unterminated string literal")
}

// readNumber consumes a decimal number, possibly signed, possibly with an
// exponent. We do not parse the value here — the parser converts on
// demand using strconv — we just delimit it.
func (l *lexer) readNumber() (token, error) {
	off := l.pos
	if l.src[l.pos] == '+' || l.src[l.pos] == '-' {
		l.pos++
	}
	digits := 0
	for l.pos < len(l.src) && l.src[l.pos] >= '0' && l.src[l.pos] <= '9' {
		l.pos++
		digits++
	}
	if l.pos < len(l.src) && l.src[l.pos] == '.' {
		l.pos++
		for l.pos < len(l.src) && l.src[l.pos] >= '0' && l.src[l.pos] <= '9' {
			l.pos++
			digits++
		}
	}
	if digits == 0 {
		return token{}, errAt(off, "invalid number")
	}
	if l.pos < len(l.src) && (l.src[l.pos] == 'e' || l.src[l.pos] == 'E') {
		l.pos++
		if l.pos < len(l.src) && (l.src[l.pos] == '+' || l.src[l.pos] == '-') {
			l.pos++
		}
		expDigits := 0
		for l.pos < len(l.src) && l.src[l.pos] >= '0' && l.src[l.pos] <= '9' {
			l.pos++
			expDigits++
		}
		if expDigits == 0 {
			return token{}, errAt(off, "invalid number: missing exponent digits")
		}
	}
	return token{kind: tokNumber, value: l.src[off:l.pos], offset: off}, nil
}

// readKeyword consumes an identifier-like keyword. WKT2 keywords are
// ASCII letters and digits; underscores are uncommon but tolerated for
// extensions. We uppercase as we go so callers can compare directly.
func (l *lexer) readKeyword() (token, error) {
	off := l.pos
	for l.pos < len(l.src) {
		c := l.src[l.pos]
		if isKeywordPart(rune(c)) {
			l.pos++
			continue
		}
		break
	}
	return token{
		kind:   tokKeyword,
		value:  strings.ToUpper(l.src[off:l.pos]),
		offset: off,
	}, nil
}

func isKeywordStart(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r == '_'
}

func isKeywordPart(r rune) bool {
	return isKeywordStart(r) || (r >= '0' && r <= '9')
}
