// Copyright 2014 Christer Sandberg.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.

package css

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// A CSS tokenizer according to http://www.w3.org/TR/css-syntax-3/
type Tokenizer interface {
	NextToken() Token // Returns the next token.
	Position() int    // Returns the position in the input where the next token will be consumed.
}

// Regexp used to preprocess the input.
// See http://www.w3.org/TR/css-syntax-3/#input-preprocessing
var preprocessRegexp = regexp.MustCompile(`\f|\r\n?`)

// NewTokenizer returns a new Tokenizer for the given input.
func NewTokenizer(input string) Tokenizer {
	i := preprocessRegexp.ReplaceAllLiteralString(input, "\n")
	return &tokenizer{
		input:     strings.Replace(i, "\u0000", string(unicode.ReplacementChar), -1),
		pos:       0,
		markedPos: 0,
	}
}

// Returns whether the given rune matches [a-zA-Z].
func IsAlpha(r rune) bool {
	return (r|0x20) >= 'a' && (r|0x20) <= 'z'
}

// Returns whether the given rune matches [0-9].
func IsDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// Returns whether the given rune matches [0-9a-fA-F].
func IsHexDigit(r rune) bool {
	return IsDigit(r) || ((r|0x20) >= 'a' && (r|0x20) <= 'f')
}

// Returns whether the given rune matches [ \t\r\n\f].
func IsSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n' || r == '\f'
}

// Returns whether the given rune is a name-start rune.
// See http://www.w3.org/TR/css-syntax-3/#name-start-code-point
func IsNameStartRune(r rune) bool {
	return r == '_' || r >= 0x80 || IsAlpha(r)
}

// Returns whether the given rune is a name rune.
// See http://www.w3.org/TR/css-syntax-3/#name-code-point
func IsNameRune(r rune) bool {
	return r == '-' || IsNameStartRune(r) || IsDigit(r)
}

// Returns whether the given rune is a non-printable rune.
// See http://www.w3.org/TR/css-syntax-3/#non-printable-code-point
func IsNonPrintable(r rune) bool {
	return (r >= 0x00 && r <= 0x08) || r == 0x0B || (r >= 0x0E && r <= 0x1F) || r == 0x7F
}

// Returns whether the two runes are a valid escape.
// See http://www.w3.org/TR/css-syntax-3/#check-if-two-code-points-are-a-valid-escape
func IsValidEscape(r1, r2 rune) bool {
	return r1 == '\\' && r2 != '\n'
}

// Convert the given hex rune to its numeric value.
func hexValue(r rune) int {
	if r < 'A' {
		return int(r - '0')
	}

	return int(r-'A'+10) & 0xF
}

// A default implementation for Tokenizer.
type tokenizer struct {
	input     string // The input string.
	pos       int    // The current position in the input.
	markedPos int    // The marked position
}

// EOF rune
const eofRune = -1

// isEOF returns whether all runes in the input have been consumed.
func (t *tokenizer) isEOF() bool {
	return t.pos >= len(t.input)
}

// mark marks the current position in the input.
func (t *tokenizer) mark() {
	t.markedPos = t.pos
}

// next consumes and returns the next rune in the input.
func (t *tokenizer) next() rune {
	if t.isEOF() {
		return eofRune
	}

	r, w := utf8.DecodeRuneInString(t.input[t.pos:])
	t.pos += w
	return r
}

// peek returns the next rune in the input without consuming it.
func (t *tokenizer) peek() rune {
	p := t.pos
	r := t.next()
	t.pos = p
	return r
}

// peek2 returns the next two runes in the input without consuming them.
func (t *tokenizer) peek2() (rune, rune) {
	p := t.pos
	r1, r2 := t.next(), t.next()
	t.pos = p
	return r1, r2
}

// peek3 returns the next three runes in the input without consuming them.
func (t *tokenizer) peek3() (rune, rune, rune) {
	p := t.pos
	r1, r2, r3 := t.next(), t.next(), t.next()
	t.pos = p
	return r1, r2, r3
}

// reset resets the position to the previously marked one.
func (t *tokenizer) reset() {
	t.pos = t.markedPos
}

// EOF token.
var eofToken = tt(EOF, -1, "")

// Implementation of NextToken for tokenizer.
func (t *tokenizer) NextToken() Token {
	if t.isEOF() {
		return eofToken
	}

	t.skipComments()
	if t.isEOF() {
		return eofToken
	}

	p := t.pos
	n := t.skipSpace()
	if n > 0 {
		return tt(Whitespace, p, "")
	}

	t.mark()
	r := t.next()
	switch r {
	case '"':
		t.reset()
		return t.consumeStringToken(false)
	case '#':
		if t.isIdentStart() {
			return &HashToken{
				TokenType: Hash,
				Pos:       Pos(p),
				Value:     t.consumeName(),
				ID:        true,
			}
		}

		r1, r2 := t.peek2()
		if IsNameRune(r1) || IsValidEscape(r1, r2) {
			return &HashToken{
				TokenType: Hash,
				Pos:       Pos(p),
				Value:     t.consumeName(),
				ID:        false,
			}
		}

		return tt(Delim, p, "#")
	case '$':
		if t.peek() == '=' {
			t.next()
			return tt(SuffixMatch, p, "$=")
		}

		return tt(Delim, p, "$")
	case '\'':
		t.reset()
		return t.consumeStringToken(true)
	case '(':
		return tt(LeftParen, p, "(")
	case ')':
		return tt(RightParen, p, ")")
	case '*':
		if t.peek() == '=' {
			t.next()
			return tt(SubstringMatch, p, "*=")
		}

		return tt(Delim, p, "*")
	case '+':
		t.reset()
		if t.isNumberStart() {
			return t.consumeNumericToken()
		}

		t.next()
		return tt(Delim, p, "+")
	case ',':
		return tt(Comma, p, ",")
	case '-':
		t.reset()
		if t.isNumberStart() {
			return t.consumeNumericToken()
		}

		if t.isIdentStart() {
			return t.consumeIdentLikeToken()
		}

		if t.consume("-->") {
			return tt(CDC, p, "-->")
		}

		t.next()
		return tt(Delim, p, "-")
	case '.':
		t.reset()
		if t.isNumberStart() {
			return t.consumeNumericToken()
		}

		t.next()
		return tt(Delim, p, ".")
	case ':':
		return tt(Colon, p, ":")
	case ';':
		return tt(Semicolon, p, ";")
	case '<':
		if t.consume("!--") {
			return tt(CDO, p, "<!--")
		}

		return tt(Delim, p, "<")
	case '@':
		if t.isIdentStart() {
			return tt(AtKeyword, p, t.consumeName())
		}

		return tt(Delim, p, "@")
	case '[':
		return tt(LeftSquareBracket, p, "[")
	case ']':
		return tt(RightSquareBracket, p, "]")
	case '\\':
		if IsValidEscape('\\', t.peek()) {
			t.reset()
			return t.consumeIdentLikeToken()
		}

		return tt(Delim, p, "\\")
	case '^':
		if t.peek() == '=' {
			t.next()
			return tt(PrefixMatch, p, "^=")
		}

		return tt(Delim, p, "^")
	case '{':
		return tt(LeftCurlyBracket, p, "{")
	case '}':
		return tt(RightCurlyBracket, p, "}")
	case '|':
		x := t.peek()
		switch x {
		case '=':
			t.next()
			return tt(DashMatch, p, "|=")
		case '|':
			t.next()
			return tt(Column, p, "||")
		}

		return tt(Delim, p, "|")
	case '~':
		if t.peek() == '=' {
			t.next()
			return tt(IncludeMatch, p, "~=")
		}

		return tt(Delim, p, "~")
	}

	if IsDigit(r) {
		t.reset()
		return t.consumeNumericToken()
	}

	if r == 'U' || r == 'u' {
		r1, r2 := t.peek2()
		if r1 == '+' && (r2 == '?' || IsHexDigit(r2)) {
			t.next() // Consume the '+'
			return t.consumeUnicodeRangeToken()
		}

		t.reset()
		return t.consumeIdentLikeToken()
	}

	if IsNameStartRune(r) {
		t.reset()
		return t.consumeIdentLikeToken()
	}

	return tt(Delim, p, string(r))
}

// Implementation of Position for tokenizer.
func (t *tokenizer) Position() int {
	return t.pos
}

// Returns whether the tokenizer could match an identifier at the current position.
// See http://www.w3.org/TR/css-syntax-3/#would-start-an-identifier
func (t *tokenizer) isIdentStart() bool {
	if t.isEOF() {
		return false
	}

	r1, r2, r3 := t.peek3()
	return IsNameStartRune(r1) || IsValidEscape(r1, r2) ||
		(r1 == '-' && (IsNameStartRune(r2) || IsValidEscape(r2, r3)))
}

// Returns whether the tokenizer could match a number at the current position.
// See http://www.w3.org/TR/css-syntax-3/#starts-with-a-number
func (t *tokenizer) isNumberStart() bool {
	if t.isEOF() {
		return false
	}

	r1, r2, r3 := t.peek3()
	if IsDigit(r1) || (r1 == '.' && IsDigit(r2)) {
		return true
	}

	if r1 == '+' || r1 == '-' {
		if IsDigit(r2) {
			return true
		}

		if r2 == '.' && IsDigit(r3) {
			return true
		}
	}

	return false
}

// Skip comments at the current position.
func (t *tokenizer) skipComments() {
	for t.consume("/*") {
		for {
			r := t.next()
			if r == eofRune {
				return
			}

			if r == '*' && t.peek() == '/' {
				t.next() // Eat up the '/'
				return
			}
		}
	}
}

// Skip whitespace at the current position. Returns the number of whitespace runes skipped.
func (t *tokenizer) skipSpace() int {
	n := 0
	for IsSpace(t.peek()) {
		n += 1
		t.next()
	}

	return n
}

// Consume a numeric token.
// It is assumed that the current position of the tokenizer represents a number token.
// See http://www.w3.org/TR/css-syntax-3/#consume-a-numeric-token
func (t *tokenizer) consumeNumericToken() Token {
	token := t.consumeNumber()
	if t.peek() == '%' {
		t.next()
		token.TokenType = Percentage
		return token
	}

	if t.isIdentStart() {
		return &DimensionToken{
			TokenType: Dimension,
			Pos:       token.Pos,
			Value:     token.Value,
			Integer:   token.Integer,
			Unit:      t.consumeName(),
		}
	}

	return token
}

// Consume an ident-like token.
// See http://www.w3.org/TR/css-syntax-3/#consume-an-ident-like-token
func (t *tokenizer) consumeIdentLikeToken() Token {
	p := t.pos
	name := t.consumeName()
	typ := Ident
	if t.peek() == '(' {
		t.next() // Consume the '('
		if strings.EqualFold(name, "url") {
			return t.consumeURLToken()
		} else {
			typ = Function
		}
	}

	return tt(typ, p, name)
}

// Consume a string token.
// See http://www.w3.org/TR/css-syntax-3/#consume-a-string-token
func (t *tokenizer) consumeStringToken(apostrophe bool) *TextToken {
	var s []rune

	p := t.pos
	t.next() // Consume the quote
	for {
		t.mark()
		r := t.next()
		if r == eofRune || (r == '\'' && apostrophe) || (r == '"' && !apostrophe) {
			break
		}

		if r == '\n' {
			t.reset()
			return tt(BadString, p, "")
		}

		if r == '\\' {
			r2 := t.peek()
			if r2 != eofRune {
				if r2 == '\n' {
					t.next() // Consume the newline
				} else {
					s = append(s, t.consumeEscape())
				}
			}
		} else {
			s = append(s, r)
		}
	}

	return tt(String, p, string(s))
}

// Consume a URL token.
// It is assumed that the current position of the tokenizer represents a URL token.
// See http://www.w3.org/TR/css-syntax-3/#consume-a-url-token
func (t *tokenizer) consumeURLToken() *TextToken {
	t.skipSpace()
	p := t.pos
	if t.isEOF() {
		return tt(URL, p, "")
	}

	r := t.peek()
	if r == '\'' || r == '"' {
		token := t.consumeStringToken(r != '"')
		if token.TokenType == BadString {
			p = t.pos
			t.consumeBadURL()
			token.TokenType = BadUrl
			token.Pos = Pos(p)
			return token
		} else {
			t.skipSpace()
			r = t.peek()
			if r == ')' || r == eofRune {
				if r == ')' {
					t.next() // Consume the ')'
				}

				token.TokenType = URL
				token.Pos = Pos(p)
				return token
			}

			p = t.pos
			t.consumeBadURL()
			token.TokenType = BadUrl
			token.Pos = Pos(p)
			return token
		}
	}

	var s []rune
	spaceSeen := false
	for {
		r = t.next()
		if r == ')' || r == eofRune {
			return tt(URL, p, string(s))
		}

		if IsSpace(r) {
			spaceSeen = true
			t.skipSpace()
			continue
		}

		if spaceSeen {
			p = t.pos
			t.consumeBadURL()
			return tt(BadUrl, p, "")
		}

		if r == '\'' || r == '"' || r == '(' || IsNonPrintable(r) {
			p := t.pos
			t.consumeBadURL()
			return tt(BadUrl, p, "")
		}

		if r == '\\' {
			if IsValidEscape(r, t.peek()) {
				s = append(s, t.consumeEscape())
			} else {
				p = t.pos
				t.consumeBadURL()
				return tt(BadUrl, p, "")
			}
		} else {
			s = append(s, r)
		}
	}
}

// Consume a unicode range.
// Is is assumed that the initial u+ has already been consumed
// and that the next input rune has been verified to be a hex digit or a ?
// See http://www.w3.org/TR/css-syntax-3/#consume-a-unicode-range-token
func (t *tokenizer) consumeUnicodeRangeToken() *UnicodeRangeToken {
	p := t.pos
	start, length := 0, 0
	for IsHexDigit(t.peek()) && length < 6 {
		start = (start << 4) + hexValue(t.next())
		length += 1
	}

	q := 0
	if length < 6 {
		for t.peek() == '?' && length < 6 {
			t.next()
			length += 1
			q += 1
		}
	}

	if q != 0 {
		end := start
		for i := 0; i < q; i++ {
			start = (start << 4) + 0
			end = (end << 4) + 15
		}

		return &UnicodeRangeToken{
			TokenType: UnicodeRange,
			Pos:       Pos(p),
			Start:     start,
			End:       end,
		}
	}

	end := 0
	r1, r2 := t.peek2()
	if r1 == '-' && IsHexDigit(r2) {
		t.next() // Consume the '-'
		length := 0
		for IsHexDigit(t.peek()) && length < 6 {
			end = (end << 4) + hexValue(t.next())
			length += 1
		}
	} else {
		end = start
	}

	return &UnicodeRangeToken{
		TokenType: UnicodeRange,
		Pos:       Pos(p),
		Start:     start,
		End:       end,
	}
}

// Consume an escaped rune.
// It is assumed that the U+005C REVERSE SOLIDUS (\) has already been consumed
// and that the next input rune has been verified to not be a newline.
// See http://www.w3.org/TR/css-syntax-3/#consume-an-escaped-code-point
func (t *tokenizer) consumeEscape() rune {
	if t.isEOF() {
		return unicode.ReplacementChar
	}

	if IsHexDigit(t.peek()) {
		uc := 0
		length := 6
		for length > 0 && IsHexDigit(t.peek()) {
			uc = (uc << 4) + hexValue(t.next())
			length -= 1
		}

		if uc == 0 || uc > unicode.MaxRune || (uc >= 0xD800 && uc <= 0xDFFF) {
			uc = unicode.ReplacementChar
		}

		if IsSpace(t.peek()) {
			t.next()
		}

		return rune(uc)
	}

	return t.next()
}

// Consume a name.
// It is assumed that the current position of the tokenizer represents a name.
// See http://www.w3.org/TR/css-syntax-3/#consume-a-name
func (t *tokenizer) consumeName() string {
	var s []rune
	for {
		t.mark()
		r := t.next()
		if IsNameRune(r) {
			s = append(s, r)
		} else if IsValidEscape(r, t.peek()) {
			s = append(s, t.consumeEscape())
		} else {
			t.reset()
			break
		}
	}

	return string(s)
}

// Consume a number.
// It is assumed that the current position of the tokenizer represents a number token.
func (t *tokenizer) consumeNumber() *NumberToken {
	var s []rune
	p := t.pos
	r := t.peek()
	if r == '+' || r == '-' {
		s = append(s, t.next())
	}

	for IsDigit(t.peek()) {
		s = append(s, t.next())
	}

	t.mark()
	i := true
	r1, r2 := t.next(), t.next()
	if r1 == '.' && IsDigit(r2) {
		s = append(s, r1, r2)
		for IsDigit(t.peek()) {
			s = append(s, t.next())
		}

		i = false
	} else {
		t.reset()
	}

	if t.isValidExponent() {
		i = false
		s = append(s, t.next(), t.next())
		for IsDigit(t.peek()) {
			s = append(s, t.next())
		}
	}

	return &NumberToken{
		TokenType: Number,
		Pos:       Pos(p),
		Value:     string(s),
		Integer:   i,
	}
}

// Consume the remnants of a bad URL.
// See http://www.w3.org/TR/css-syntax-3/#consume-the-remnants-of-a-bad-url
func (t *tokenizer) consumeBadURL() {
	for {
		r := t.next()
		if r == ')' || r == eofRune {
			break
		}

		if IsValidEscape(r, t.peek()) {
			t.consumeEscape()
		}
	}
}

// Returns whether the next runes at the current position start the exponential part of a number.
func (t *tokenizer) isValidExponent() bool {
	if t.isEOF() {
		return false
	}

	t.mark()
	defer t.reset()

	r := t.next()
	if r != 'e' && r != 'E' {
		return false
	}

	r = t.next()
	if r == '+' || r == '-' {
		return IsDigit(t.next())
	}

	return IsDigit(r)
}

// Tries to consume s at the current position. Returns true on success.
func (t *tokenizer) consume(s string) bool {
	if !t.isEOF() && strings.HasPrefix(t.input[t.pos:], s) {
		t.pos += len(s)
		return true
	}

	return false
}

// Create a new DefaultToken.
func tt(typ TokenType, pos int, value string) *TextToken {
	return &TextToken{
		TokenType: typ,
		Pos:       Pos(pos),
		Value:     value,
	}
}
