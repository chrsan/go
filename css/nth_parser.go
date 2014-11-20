// Copyright 2014 Christer Sandberg.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.

package css

import (
	"fmt"
	"strconv"
	"strings"
)

// Parse an An+B notation at the current position in Tokenizer t.
// Returns the value for A and B on successful parse.
func parseNth(t Tokenizer) (int, int, error) {
	var a, b int
	var ok bool

	err := fmt.Errorf("Invalid nth arguments at position %v", t.Position())

	tk := skipWhitespace(t)
	switch tk.Type() {
	case Number:
		n := tk.(*NumberToken)
		if !n.Integer || !closingParen(t) {
			return 0, 0, err
		}

		if b, ok = parseInt(n.Value); ok {
			return 0, b, nil
		} else {
			return 0, 0, err
		}
	case Dimension:
		d := tk.(*DimensionToken)
		if !d.Integer {
			return 0, 0, err
		}

		a, ok = parseInt(d.Value)
		if !ok {
			return 0, 0, err
		}

		unit := strings.ToLower(d.Unit)
		if unit == "n" {
			b, ok = parseB(t)
		} else if unit == "n-" {
			b, ok = parseSignlessB(t, -1)
		} else {
			b, ok = parseNDashDigits(unit)
			ok = ok && closingParen(t)
		}

		if !ok {
			return 0, 0, err
		}

		return a, b, nil
	case Ident:
		ident := strings.ToLower(tk.String())
		switch ident {
		case "even":
			a, b = 2, 0
			ok = closingParen(t)
		case "odd":
			a, b = 2, 1
			ok = closingParen(t)
		case "n":
			a = 1
			b, ok = parseB(t)
		case "-n":
			a = -1
			b, ok = parseB(t)
		case "n-":
			a = 1
			b, ok = parseSignlessB(t, -1)
		case "-n-":
			a = -1
			b, ok = parseSignlessB(t, -1)
		default:
			if strings.HasPrefix(ident, "-") {
				a = -1
				b, ok = parseNDashDigits(ident[1:])
				ok = ok && closingParen(t)
			} else {
				a = 1
				b, ok = parseNDashDigits(ident)
				ok = ok && closingParen(t)
			}
		}

		if !ok {
			return 0, 0, err
		}

		return a, b, nil
	case Delim:
		if tk.String() != "+" {
			return 0, 0, err
		}

		tk = t.NextToken()
		if tk.Type() != Ident {
			return 0, 0, err
		}

		ident := strings.ToLower(tk.String())
		switch ident {
		case "n":
			a = 1
			b, ok = parseB(t)
		case "n-":
			a = 1
			b, ok = parseSignlessB(t, -1)
		default:
			a = 1
			b, ok = parseNDashDigits(ident)
			ok = ok && closingParen(t)
		}

		if !ok {
			return 0, 0, err
		}

		return a, b, nil
	default:
		return 0, 0, err
	}
}

func parseB(t Tokenizer) (int, bool) {
	tk := skipWhitespace(t)
	switch tk.Type() {
	case RightParen:
		return 0, true
	case Delim:
		switch tk.String() {
		case "+":
			return parseSignlessB(t, 1)
		case "-":
			return parseSignlessB(t, -1)
		default:
			return 0, false
		}
	case Number:
		n := tk.(*NumberToken)
		if !n.Integer || !hasSign(n.Value) || !closingParen(t) {
			return 0, false
		}

		return parseInt(n.Value)
	default:
		return 0, false
	}
}

func parseSignlessB(t Tokenizer, sign int) (int, bool) {
	tk := skipWhitespace(t)
	n, ok := tk.(*NumberToken)
	if !ok || !n.Integer || hasSign(n.Value) || !closingParen(t) {
		return 0, false
	}

	if b, ok := parseInt(n.Value); ok {
		return sign * b, true
	}

	return 0, false
}

func parseNDashDigits(s string) (int, bool) {
	if len(s) >= 3 && strings.HasPrefix(s, "n-") {
		return parseInt(s[1:])
	}

	return 0, false
}

func hasSign(s string) bool {
	c := s[0:1]
	return c == "+" || c == "-"
}

func parseInt(s string) (int, bool) {
	if n, err := strconv.ParseInt(s, 10, 0); err == nil {
		return int(n), true
	} else {
		return 0, false
	}
}

func closingParen(t Tokenizer) bool {
	return skipWhitespace(t).Type() == RightParen
}

func skipWhitespace(t Tokenizer) Token {
	for {
		tk := t.NextToken()
		if tk.Type() != Whitespace {
			return tk
		}
	}
}
