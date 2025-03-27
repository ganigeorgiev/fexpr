package fexpr

import (
	"bytes"
	"fmt"
	"strings"
	"unicode/utf8"
)

// eof represents a marker rune for the end of the reader.
const eof = rune(0)

// JoinOp represents a join type operator.
type JoinOp string

// supported join type operators
const (
	JoinAnd JoinOp = "&&"
	JoinOr  JoinOp = "||"
)

// SignOp represents an expression sign operator.
type SignOp string

// supported expression sign operators
const (
	SignEq    SignOp = "="
	SignNeq   SignOp = "!="
	SignLike  SignOp = "~"
	SignNlike SignOp = "!~"
	SignLt    SignOp = "<"
	SignLte   SignOp = "<="
	SignGt    SignOp = ">"
	SignGte   SignOp = ">="

	// array/any operators
	SignAnyEq    SignOp = "?="
	SignAnyNeq   SignOp = "?!="
	SignAnyLike  SignOp = "?~"
	SignAnyNlike SignOp = "?!~"
	SignAnyLt    SignOp = "?<"
	SignAnyLte   SignOp = "?<="
	SignAnyGt    SignOp = "?>"
	SignAnyGte   SignOp = "?>="
)

// TokenType represents a Token type.
type TokenType string

// token type constants
const (
	TokenUnexpected TokenType = "unexpected"
	TokenEOF        TokenType = "eof"
	TokenWS         TokenType = "whitespace"
	TokenJoin       TokenType = "join"
	TokenSign       TokenType = "sign"
	TokenIdentifier TokenType = "identifier" // variable, column name, placeholder, etc.
	TokenFunction   TokenType = "function"   // function
	TokenNumber     TokenType = "number"
	TokenText       TokenType = "text"  // ' or " quoted string
	TokenGroup      TokenType = "group" // groupped/nested tokens
	TokenComment    TokenType = "comment"
)

// Token represents a single scanned literal (one or more combined runes).
type Token struct {
	Meta    interface{}
	Type    TokenType
	Literal string
}

// NewScanner creates and returns a new scanner instance loaded with the specified data.
func NewScanner(data []byte) *Scanner {
	return &Scanner{
		data:         data,
		maxFuncDepth: 3,
	}
}

// Scanner represents a filter and lexical scanner.
type Scanner struct {
	data         []byte
	pos          int
	maxFuncDepth int
}

// Scan reads and returns the next available token value from the scanner's buffer.
func (s *Scanner) Scan() (Token, error) {
	ch := s.read()

	if ch == eof {
		return Token{Type: TokenEOF, Literal: ""}, nil
	}

	if isWhitespaceRune(ch) {
		s.unread()
		return s.scanWhitespace()
	}

	if isGroupStartRune(ch) {
		s.unread()
		return s.scanGroup()
	}

	if isIdentifierStartRune(ch) {
		s.unread()
		return s.scanIdentifier(s.maxFuncDepth)
	}

	if isNumberStartRune(ch) {
		s.unread()
		return s.scanNumber()
	}

	if isTextStartRune(ch) {
		s.unread()
		return s.scanText(false)
	}

	if isSignStartRune(ch) {
		s.unread()
		return s.scanSign()
	}

	if isJoinStartRune(ch) {
		s.unread()
		return s.scanJoin()
	}

	if isCommentStartRune(ch) {
		s.unread()
		return s.scanComment()
	}

	return Token{Type: TokenUnexpected, Literal: string(ch)}, fmt.Errorf("unexpected character %q", ch)
}

// scanWhitespace consumes all contiguous whitespace runes.
func (s *Scanner) scanWhitespace() (Token, error) {
	var buf bytes.Buffer

	// Reads every subsequent whitespace character into the buffer.
	// Non-whitespace runes and EOF will cause the loop to exit.
	for {
		ch := s.read()

		if ch == eof {
			break
		}

		if !isWhitespaceRune(ch) {
			s.unread()
			break
		}

		// write the whitespace rune
		buf.WriteRune(ch)
	}

	return Token{Type: TokenWS, Literal: buf.String()}, nil
}

// scanNumber consumes all contiguous digit runes
// (complex numbers and scientific notations are not supported).
func (s *Scanner) scanNumber() (Token, error) {
	var buf bytes.Buffer

	var hadDot bool

	// Read every subsequent digit rune into the buffer.
	// Non-digit runes and EOF will cause the loop to exit.
	for {
		ch := s.read()

		if ch == eof {
			break
		}

		// not a digit rune
		if !isDigitRune(ch) &&
			// minus sign but not at the beginning
			(ch != '-' || buf.Len() != 0) &&
			// dot but there was already another dot
			(ch != '.' || hadDot) {
			s.unread()
			break
		}

		// write the rune
		buf.WriteRune(ch)

		if ch == '.' {
			hadDot = true
		}
	}

	total := buf.Len()
	literal := buf.String()

	var err error
	// only "-" or starts with "." or ends with "."
	if (total == 1 && literal[0] == '-') || literal[0] == '.' || literal[total-1] == '.' {
		err = fmt.Errorf("invalid number %q", literal)
	}

	return Token{Type: TokenNumber, Literal: buf.String()}, err
}

// scanText consumes all contiguous quoted text runes.
func (s *Scanner) scanText(preserveQuotes bool) (Token, error) {
	var buf bytes.Buffer

	// read the first rune to determine the quotes type
	firstCh := s.read()
	buf.WriteRune(firstCh)
	var prevCh rune
	var hasMatchingQuotes bool

	// Read every subsequent text rune into the buffer.
	// EOF and matching unescaped ending quote will cause the loop to exit.
	for {
		ch := s.read()

		if ch == eof {
			break
		}

		// write the text rune
		buf.WriteRune(ch)

		// unescaped matching quote, aka. the end
		if ch == firstCh && prevCh != '\\' {
			hasMatchingQuotes = true
			break
		}

		prevCh = ch
	}

	literal := buf.String()

	var err error
	if !hasMatchingQuotes {
		err = fmt.Errorf("invalid quoted text %q", literal)
	} else if !preserveQuotes {
		// unquote
		literal = literal[1 : len(literal)-1]
		// remove escaped quotes prefix (aka. \)
		firstChStr := string(firstCh)
		literal = strings.ReplaceAll(literal, `\`+firstChStr, firstChStr)
	}

	return Token{Type: TokenText, Literal: literal}, err
}

// scanComment consumes all contiguous single line comment runes until
// a new character (\n) or EOF is reached.
func (s *Scanner) scanComment() (Token, error) {
	var buf bytes.Buffer

	// Read the first 2 characters without writting them to the buffer.
	if !isCommentStartRune(s.read()) || !isCommentStartRune(s.read()) {
		return Token{Type: TokenComment}, ErrInvalidComment
	}

	// Read every subsequent comment text rune into the buffer.
	// \n and EOF will cause the loop to exit.
	for i := 0; ; i++ {
		ch := s.read()

		if ch == eof || ch == '\n' {
			break
		}

		buf.WriteRune(ch)
	}

	return Token{Type: TokenComment, Literal: strings.TrimSpace(buf.String())}, nil
}

// scanIdentifier consumes all contiguous ident runes.
func (s *Scanner) scanIdentifier(funcDepth int) (Token, error) {
	var buf bytes.Buffer

	// read the first rune in case it is a special start identifier character
	buf.WriteRune(s.read())

	// Read every subsequent identifier rune into the buffer.
	// Non-ident runes and EOF will cause the loop to exit.
	for {
		ch := s.read()

		if ch == eof {
			break
		}

		// func
		if ch == '(' {
			funcName := buf.String()
			if funcDepth <= 0 {
				return Token{Type: TokenFunction, Literal: funcName}, fmt.Errorf("max nested function arguments reached (max: %d)", s.maxFuncDepth)
			}
			if !isValidIdentifier(funcName) {
				return Token{Type: TokenFunction, Literal: funcName}, fmt.Errorf("invalid function name %q", funcName)
			}
			s.unread()
			return s.scanFunctionArgs(funcName, funcDepth)
		}

		// not an identifier character
		if !isLetterRune(ch) && !isDigitRune(ch) && !isIdentifierCombineRune(ch) && ch != '_' {
			s.unread()
			break
		}

		// write the identifier rune
		buf.WriteRune(ch)
	}

	literal := buf.String()

	var err error
	if !isValidIdentifier(literal) {
		err = fmt.Errorf("invalid identifier %q", literal)
	}

	return Token{Type: TokenIdentifier, Literal: literal}, err
}

// scanSign consumes all contiguous sign operator runes.
func (s *Scanner) scanSign() (Token, error) {
	var buf bytes.Buffer

	// Read every subsequent sign rune into the buffer.
	// Non-sign runes and EOF will cause the loop to exit.
	for {
		ch := s.read()

		if ch == eof {
			break
		}

		if !isSignStartRune(ch) {
			s.unread()
			break
		}

		// write the sign rune
		buf.WriteRune(ch)
	}

	literal := buf.String()

	var err error
	if !isSignOperator(literal) {
		err = fmt.Errorf("invalid sign operator %q", literal)
	}

	return Token{Type: TokenSign, Literal: literal}, err
}

// scanJoin consumes all contiguous join operator runes.
func (s *Scanner) scanJoin() (Token, error) {
	var buf bytes.Buffer

	// Read every subsequent join operator rune into the buffer.
	// Non-join runes and EOF will cause the loop to exit.
	for {
		ch := s.read()

		if ch == eof {
			break
		}

		if !isJoinStartRune(ch) {
			s.unread()
			break
		}

		// write the join operator rune
		buf.WriteRune(ch)
	}

	literal := buf.String()

	var err error
	if !isJoinOperator(literal) {
		err = fmt.Errorf("invalid join operator %q", literal)
	}

	return Token{Type: TokenJoin, Literal: literal}, err
}

// scanGroup consumes all runes within a group/parenthesis.
func (s *Scanner) scanGroup() (Token, error) {
	var buf bytes.Buffer

	// read the first group bracket without writing it to the buffer
	firstChar := s.read()
	openGroups := 1

	// Read every subsequent text rune into the buffer.
	// EOF and matching unescaped ending quote will cause the loop to exit.
	for {
		ch := s.read()

		if ch == eof {
			break
		}

		if isGroupStartRune(ch) {
			// nested group
			openGroups++
			buf.WriteRune(ch)
		} else if isTextStartRune(ch) {
			s.unread()
			t, err := s.scanText(true) // with quotes to preserve the exact text start/end runes
			if err != nil {
				// write the errored literal as it is
				buf.WriteString(t.Literal)
				return Token{Type: TokenGroup, Literal: buf.String()}, err
			}

			buf.WriteString(t.Literal)
		} else if ch == ')' {
			openGroups--

			if openGroups <= 0 {
				// main group end
				break
			} else {
				buf.WriteRune(ch)
			}
		} else {
			buf.WriteRune(ch)
		}
	}

	literal := buf.String()

	var err error
	if !isGroupStartRune(firstChar) || openGroups > 0 {
		err = fmt.Errorf("invalid formatted group - missing %d closing bracket(s)", openGroups)
	}

	return Token{Type: TokenGroup, Literal: literal}, err
}

// scanFunctionArgs consumes all contiguous function call runes to
// extract its arguments and returns a function token with the found
// Token arguments loaded in Token.Meta.
func (s *Scanner) scanFunctionArgs(funcName string, funcDepth int) (Token, error) {
	var args []Token

	var expectComma, isComma, isClosed bool

	ch := s.read()
	if ch != '(' {
		return Token{Type: TokenFunction, Literal: funcName}, fmt.Errorf("invalid or incomplete function call %q", funcName)
	}

	// Read every subsequent rune until ')' or EOF has been reached.
	for {
		ch := s.read()

		if ch == eof {
			break
		}

		if ch == ')' {
			isClosed = true
			break
		}

		// skip whitespaces
		if isWhitespaceRune(ch) {
			_, err := s.scanWhitespace()
			if err != nil {
				return Token{Type: TokenFunction, Literal: funcName, Meta: args}, fmt.Errorf("failed to scan whitespaces in function %q: %w", funcName, err)
			}
			continue
		}

		// skip comments
		if isCommentStartRune(ch) {
			s.unread()
			_, err := s.scanComment()
			if err != nil {
				return Token{Type: TokenFunction, Literal: funcName, Meta: args}, fmt.Errorf("failed to scan comment in function %q: %w", funcName, err)
			}
			continue
		}

		isComma = ch == ','

		if expectComma && !isComma {
			return Token{Type: TokenFunction, Literal: funcName, Meta: args}, fmt.Errorf("expected comma after the last argument in function %q", funcName)
		}

		if !expectComma && isComma {
			return Token{Type: TokenFunction, Literal: funcName, Meta: args}, fmt.Errorf("unexpected comma in function %q", funcName)
		}

		expectComma = false // reset

		if isComma {
			continue
		}

		if isIdentifierStartRune(ch) {
			s.unread()
			t, err := s.scanIdentifier(funcDepth - 1)
			if err != nil {
				return Token{Type: TokenFunction, Literal: funcName, Meta: args}, fmt.Errorf("invalid identifier argument %q in function %q: %w", t.Literal, funcName, err)
			}
			args = append(args, t)
			expectComma = true
		} else if isNumberStartRune(ch) {
			s.unread()
			t, err := s.scanNumber()
			if err != nil {
				return Token{Type: TokenFunction, Literal: funcName, Meta: args}, fmt.Errorf("invalid number argument %q in function %q: %w", t.Literal, funcName, err)
			}
			args = append(args, t)
			expectComma = true
		} else if isTextStartRune(ch) {
			s.unread()
			t, err := s.scanText(false)
			if err != nil {
				return Token{Type: TokenFunction, Literal: funcName, Meta: args}, fmt.Errorf("invalid text argument %q in function %q: %w", t.Literal, funcName, err)
			}
			args = append(args, t)
			expectComma = true
		} else {
			return Token{Type: TokenFunction, Literal: funcName, Meta: args}, fmt.Errorf("unsupported argument character %q in function %q", ch, funcName)
		}
	}

	if !isClosed {
		return Token{Type: TokenFunction, Literal: funcName, Meta: args}, fmt.Errorf("invalid or incomplete function %q (expected ')')", funcName)
	}

	return Token{Type: TokenFunction, Literal: funcName, Meta: args}, nil
}

// unread unreads the last character and revert the position 1 step back.
func (s *Scanner) unread() {
	if s.pos > 0 {
		s.pos = s.pos - 1
	}
}

// read reads the next rune and moves the position forward.
func (s *Scanner) read() rune {
	if s.pos >= len(s.data) {
		return eof
	}

	ch, n := utf8.DecodeRune(s.data[s.pos:])
	s.pos += n

	return ch
}

// Lexical helpers:
// -------------------------------------------------------------------

// isWhitespaceRune checks if a rune is a space, tab, or newline.
func isWhitespaceRune(ch rune) bool { return ch == ' ' || ch == '\t' || ch == '\n' }

// isLetterRune checks if a rune is a letter.
func isLetterRune(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// isDigitRune checks if a rune is a digit.
func isDigitRune(ch rune) bool {
	return (ch >= '0' && ch <= '9')
}

// isTextStartRune checks if a rune is a valid quoted text first character
// (aka. single or double quote).
func isTextStartRune(ch rune) bool {
	return ch == '\'' || ch == '"'
}

// isNumberStartRune checks if a rune is a valid number start character (aka. digit).
func isNumberStartRune(ch rune) bool {
	return ch == '-' || isDigitRune(ch)
}

// isSignStartRune checks if a rune is a valid sign operator start character.
func isSignStartRune(ch rune) bool {
	return ch == '=' ||
		ch == '?' ||
		ch == '!' ||
		ch == '>' ||
		ch == '<' ||
		ch == '~'
}

// isJoinStartRune checks if a rune is a valid join type start character.
func isJoinStartRune(ch rune) bool {
	return ch == '&' || ch == '|'
}

// isGroupStartRune checks if a rune is a valid group/parenthesis start character.
func isGroupStartRune(ch rune) bool {
	return ch == '('
}

// isCommentStartRune checks if a rune is a valid comment start character.
func isCommentStartRune(ch rune) bool {
	return ch == '/'
}

// isIdentifierStartRune checks if a rune is valid identifier's first character.
func isIdentifierStartRune(ch rune) bool {
	return isLetterRune(ch) || isIdentifierSpecialStartRune(ch)
}

// isIdentifierSpecialStartRune checks if a rune is valid identifier's first special character.
func isIdentifierSpecialStartRune(ch rune) bool {
	return ch == '@' || ch == '_' || ch == '#'
}

// isIdentifierCombineRune checks if a rune is valid identifier's combine character.
func isIdentifierCombineRune(ch rune) bool {
	return ch == '.' || ch == ':'
}

// isSignOperator checks if a literal is a valid sign operator.
func isSignOperator(literal string) bool {
	switch SignOp(literal) {
	case
		SignEq,
		SignNeq,
		SignLt,
		SignLte,
		SignGt,
		SignGte,
		SignLike,
		SignNlike,
		SignAnyEq,
		SignAnyNeq,
		SignAnyLike,
		SignAnyNlike,
		SignAnyLt,
		SignAnyLte,
		SignAnyGt,
		SignAnyGte:
		return true
	}

	return false
}

// isJoinOperator checks if a literal is a valid join type operator.
func isJoinOperator(literal string) bool {
	switch JoinOp(literal) {
	case
		JoinAnd,
		JoinOr:
		return true
	}

	return false
}

// isValidIdentifier validates the literal against common identifier requirements.
func isValidIdentifier(literal string) bool {
	length := len(literal)

	return (
	// doesn't end with combine rune
	!isIdentifierCombineRune(rune(literal[length-1])) &&
		// is not just a special start rune
		(length != 1 || !isIdentifierSpecialStartRune(rune(literal[0]))))
}
