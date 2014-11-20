// Copyright 2014 Christer Sandberg.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.

package css

import "fmt"

// TokenType identifies the type of tokens.
type TokenType int

// Type returns itself.
func (t TokenType) Type() TokenType {
	return t
}

const (
	AtKeyword TokenType = iota
	BadString
	BadUrl
	CDC
	CDO
	Colon
	Column
	Comma
	DashMatch
	Delim
	Dimension
	EOF
	Function
	Hash
	Ident
	IncludeMatch
	LeftCurlyBracket
	LeftParen
	LeftSquareBracket
	Number
	Percentage
	PrefixMatch
	RightCurlyBracket
	RightParen
	RightSquareBracket
	Semicolon
	String
	SubstringMatch
	SuffixMatch
	UnicodeRange
	URL
	Whitespace
)

// Pos represents the token position in the input text.
type Pos int

// Position returns ifself.
func (p Pos) Position() Pos {
	return p
}

// Represents a token returned from the tokenizer.
type Token interface {
	Type() TokenType // The type of this token.
	Position() Pos   // The position of this token.
	String() string  // The string value of this token.
}

// Represents the default Token returned from the tokenizer.
type TextToken struct {
	TokenType        // The type of this token.
	Pos              // The position of this token.
	Value     string // The string value of this token.
}

func (t *TextToken) String() string {
	return t.Value
}

// Represents a unicode range token returned from the tokenizer.
type UnicodeRangeToken struct {
	TokenType     // The type of this token.
	Pos           // The position of this token.
	Start     int // The start of this unicode range token.
	End       int // The end of this unicode range token.
}

func (t *UnicodeRangeToken) String() string {
	return fmt.Sprintf("%U-%U", t.Start, t.End)
}

// Represents a hash token returned from the tokenizer.
type HashToken struct {
	TokenType        // The type of this token.
	Pos              // The position of this token.
	Value     string // The string value of this token.
	ID        bool   // If the type flag is ID.
}

func (t *HashToken) String() string {
	return fmt.Sprintf("#%s", t.Value)
}

// Represents a numeric token returned from the tokenizer.
type NumberToken struct {
	TokenType        // The type of this token.
	Pos              // The position of this token.
	Value     string // The string value of this token.
	Integer   bool   // If the type flag is INTEGER.
}

func (t *NumberToken) String() string {
	return t.Value
}

// Represents a dimension token returned from the tokenizer.
type DimensionToken struct {
	TokenType        // The type of this token.
	Pos              // The position of this token.
	Value     string // The string value of this token.
	Integer   bool   // If the type flag is INTEGER.
	Unit      string // The unit of this dimension token.
}

func (t *DimensionToken) String() string {
	return fmt.Sprintf("%s%s", t.Value, t.Unit)
}
