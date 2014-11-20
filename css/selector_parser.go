// Copyright 2014 Christer Sandberg.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.

package css

import (
	"bytes"
	"fmt"
	"strings"
)

// Parse a SelectorsGroup from Tokenizer t.
func ParseSelector(t Tokenizer) (SelectorsGroup, error) {
	p := &selectorParser{tokenizer: t}
	return p.parseSelectorList()
}

// Parse A SelectorsGroup from the string s.
func ParseSelectorFromString(s string) (SelectorsGroup, error) {
	return ParseSelector(NewTokenizer(s))
}

// Selector parser state.
type selectorParser struct {
	tokenizer Tokenizer // Tokenizer used when parsing.
	saved     Token     // Possibly saved token.
}

// Parse a selector list.
// See http://www.w3.org/TR/selectors/#grouping
func (p *selectorParser) parseSelectorList() (SelectorsGroup, error) {
	s, err := p.parseSelector()
	if err != nil {
		return nil, err
	}

	group := []*Selector{s}
	for {
		tk, _ := p.skipWhitespace()
		typ := tk.Type()
		if typ == EOF {
			break
		}

		if typ != Comma {
			return nil, expected(",", tk)
		}

		s, err = p.parseSelector()
		if err != nil {
			return nil, err
		}

		group = append(group, s)
	}

	return group, nil
}

// Parse a selector.
// See http://www.w3.org/TR/selectors/#selector-syntax
func (p *selectorParser) parseSelector() (*Selector, error) {
	ss, pseudoElement, err := p.parseSimpleSelectors()
	if err != nil {
		return nil, err
	}

	cs := &CompoundSelector{
		SimpleSelectors: ss,
	}

	for pseudoElement == nil {
		tk, skipped := p.skipWhitespace()
		typ := tk.Type()
		if typ == EOF {
			break
		} else if typ == Comma {
			p.saved = tk
			break
		}

		var c Combinator = -1
		if typ == Delim {
			switch tk.String() {
			case ">":
				c = Child
			case "+":
				c = NextSibling
			case "~":
				c = LaterSibling
			}
		}

		if c == -1 {
			if skipped {
				c = Descendant
			} else {
				return nil, expected("one of ' ', '>', '+', '~'", tk)
			}

			p.saved = tk
		}

		ss, pseudoElement, err = p.parseSimpleSelectors()
		if err != nil {
			return nil, err
		} else {
			cs = &CompoundSelector{
				SimpleSelectors: ss,
				Prev: &Prev{
					Combinator:       c,
					CompoundSelector: cs,
				},
			}
		}
	}

	s := &Selector{
		CompoundSelector: cs,
		PseudoElement:    pseudoElement,
	}

	return s, nil
}

// Parse a sequence of simple selectors.
// On successful parsing it returns a sequence of selectors and maybe a pseudo element selector
// indicating if the last (or only) simple selector in the sequence is a pseudo element.
// See http://www.w3.org/TR/selectors/#sequence
func (p *selectorParser) parseSimpleSelectors() ([]SimpleSelector, *PseudoElementSelector, error) {
	pos := p.tokenizer.Position()

	var ss []SimpleSelector
	var pseudoElement *PseudoElementSelector

	empty := true
	tagName, found := p.parseName()
	if found && tagName != "*" {
		ss = append(ss, NewLocalNameSelector(tagName))
		empty = false
	}

	for {
		s, err := p.parseOneSimpleSelector(false)
		if err != nil {
			return nil, nil, err
		}

		if s == nil {
			break
		}

		if x, ok := s.(*PseudoElementSelector); ok {
			empty = false
			pseudoElement = x
			break
		}

		ss = append(ss, s)
		empty = false
	}

	if empty && !found {
		return nil, nil, fmt.Errorf("No simple selectors found at position %d", pos)
	}

	return ss, pseudoElement, nil
}

// Parse the name of an element.
// Returns a name and a boolean indicating if a type selector was found or not.
// See http://www.w3.org/TR/selectors/#type-selectors and http://www.w3.org/TR/selectors/#universal-selector
func (p *selectorParser) parseName() (string, bool) {
	tk, _ := p.skipWhitespace()
	switch tk.Type() {
	case Delim:
		if tk.String() == "*" {
			return "*", true
		}
	case Ident:
		return tk.String(), true
	}

	p.saved = tk
	return "*", false
}

// Parse one simple selector (excluding the type selector).
// Returns (nil, nil) if no SimpleSelector can be parsed, error
// on failure parsing a simple selector.
// See http://www.w3.org/TR/selectors/#simple-selectors
func (p *selectorParser) parseOneSimpleSelector(insideNegation bool) (SimpleSelector, error) {
	tk := p.nextToken()
	switch tk.Type() {
	case Hash:
		h := tk.(*HashToken)
		return NewAttributeSelector(Equals, "id", h.Value), nil
	case Delim:
		if tk.String() == "." {
			tk = p.nextToken()
			if tk.Type() == Ident {
				return NewAttributeSelector(Includes, "class", tk.String()), nil
			} else {
				return nil, expected("class value", tk)
			}
		}

		return nil, expected(".", tk)
	case LeftSquareBracket:
		return p.parseAttribute()
	case Colon:
		tk = p.nextToken()
		switch tk.Type() {
		case Ident:
			v := tk.String()
			switch strings.ToLower(v) {
			case "first-line", "first-letter", "before", "after":
				return NewPseudoElementSelector(v), nil
			default:
				return NewPseudoClassSelector(v), nil
			}
		case Colon:
			tk = p.nextToken()
			if tk.Type() != Ident {
				return nil, expected("pseudo element value", tk)
			}

			return NewPseudoElementSelector(tk.String()), nil
		case Function:
			return p.parseFunctionalPseudoClass(tk.String(), insideNegation)
		}
	}

	p.saved = tk
	return nil, nil
}

// Parse an attribute simple selector.
// See http://www.w3.org/TR/selectors/#attribute-selectors
func (p *selectorParser) parseAttribute() (*AttributeSelector, error) {
	tk, _ := p.skipWhitespace()
	if tk.Type() != Ident {
		return nil, expected("attribute name", tk)
	}

	name := tk.String()
	tk, _ = p.skipWhitespace()
	if tk.Type() == RightSquareBracket {
		return NewAttributeSelector(Exists, name, ""), nil
	}

	var match AttributeMatch
	switch tk.Type() {
	case PrefixMatch:
		match = Begins
	case SuffixMatch:
		match = Ends
	case SubstringMatch:
		match = Contains
	case IncludeMatch:
		match = Includes
	case DashMatch:
		match = Hyphens
	case Delim:
		if tk.String() == "=" {
			match = Equals
		} else {
			return nil, expected("=", tk)
		}
	}

	tk, _ = p.skipWhitespace()
	var value string
	if tk.Type() == Ident || tk.Type() == String {
		value = tk.String()
	} else {
		return nil, expected("attribute value", tk)
	}

	tk, _ = p.skipWhitespace()
	if tk.Type() != RightSquareBracket {
		return nil, expected("]", tk)
	}

	return NewAttributeSelector(match, name, value), nil
}

// Parse a functional pseudo class.
// See http://www.w3.org/TR/selectors/#structural-pseudos
func (p *selectorParser) parseFunctionalPseudoClass(name string, insideNegation bool) (SimpleSelector, error) {
	switch strings.ToLower(name) {
	case "nth-child", "nth-last-child", "nth-of-type", "nth-last-of-type":
		a, b, err := parseNth(p.tokenizer)
		if err != nil {
			return nil, err
		}

		return NewPseudoNthSelector(name, a, b), nil
	case "not":
		if insideNegation {
			return nil, fmt.Errorf("Error at position %d: negations may not be nested", p.tokenizer.Position())
		}

		var s *PseudoNegationSelector
		if tagName, found := p.parseName(); found {
			s = NewPseudoNegationSelector(NewLocalNameSelector(tagName))
		} else {
			ss, err := p.parseOneSimpleSelector(true)
			if err != nil {
				return nil, err
			} else if ss == nil {
				// The failing token was saved when parseOneSimpleSelector didn't find a selector.
				return nil, expected("simple selector", p.nextToken())
			}

			s = NewPseudoNegationSelector(ss)
		}

		tk, _ := p.skipWhitespace()
		if tk.Type() != RightParen {
			return nil, expected(")", tk)
		}

		return s, nil
	default:
		pos := p.tokenizer.Position()

		var b bytes.Buffer
		for {
			tk := p.nextToken()
			typ := tk.Type()
			if typ == EOF {
				return nil, fmt.Errorf("EOF in function expression starting at position %d", pos)
			} else if typ == RightParen {
				break
			} else {
				b.WriteString(tk.String())
			}
		}

		return NewPseudoFunctionSelector(name, b.String()), nil
	}
}

// Returns the next token to parse.
func (p *selectorParser) nextToken() Token {
	if p.saved != nil {
		tk := p.saved
		p.saved = nil
		return tk
	}

	return p.tokenizer.NextToken()
}

// Skips whitespace tokens and returns the next non-whitespace token
// and a boolean indicating if some whitespace was skipped.
func (p *selectorParser) skipWhitespace() (Token, bool) {
	skipped := false
	for {
		tk := p.nextToken()
		if tk.Type() != Whitespace {
			return tk, skipped
		}

		skipped = true
	}
}

// Returns an error of what was expected and what was unexpectedly found.
func expected(what string, tk Token) error {
	return fmt.Errorf("Expected %s at position %d, got %s", what, tk.Position(), tk)
}
