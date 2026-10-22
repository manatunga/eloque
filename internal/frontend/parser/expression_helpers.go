package parser

import (
	"strings"

	"github.com/manatunga/eloque/internal/frontend/lexer"
)

// Expression binding powers define the precedence of Eloque operators.
// Higher values indicate tighter binding. These constants follow the
// precedence rules established in the language specification and are
// used by the Pratt parser to determine expression grouping.
const (
	bindingPowerAdditive       = 50
	bindingPowerMultiplicative = 60
	bindingPowerUnary          = 70
	bindingPowerExponentiation = 80
)

// binaryBindingPower returns the binding power of a supported
// binary arithmetic operator.
//
// The boolean result indicates whether the token kind represents
// an operator supported by the current expression parser.
//
// Operators with higher binding powers bind more tightly.
// Addition, subtraction, multiplication, division, floor division,
// and modulo are left-associative. Exponentiation is right-associative.
func binaryBindingPower(kind lexer.TokenKind) (int, bool) {
	switch kind {
	case lexer.TokenPlus, lexer.TokenMinus:
		return bindingPowerAdditive, true
	case lexer.TokenStar, lexer.TokenSlash, lexer.TokenFloorDiv, lexer.TokenPercent:
		return bindingPowerMultiplicative, true
	case lexer.TokenPower:
		return bindingPowerExponentiation, true
	default:
		return 0, false
	}
}

// decodeStringLiteral converts a validated Eloque string token's
// lexeme into its decoded string value.
//
// The input includes the original surrounding quotation marks.
// Escape sequences are converted to their corresponding characters.
// The lexer is responsible for rejecting malformed string literals
// before this function is called.
func decodeStringLiteral(lexeme string) string {
	innerString := lexeme[1 : len(lexeme)-1]
	var result strings.Builder
	result.Grow(len(innerString))

	for i := 0; i < len(innerString); i++ {
		char := innerString[i]
		if char == '\\' && i+1 < len(innerString) {
			nextChar := innerString[i+1]
			switch nextChar {
			case 'n':
				result.WriteByte('\n')
			case 't':
				result.WriteByte('\t')
			case 'r':
				result.WriteByte('\r')
			case '\\':
				result.WriteByte('\\')
			case '\'':
				result.WriteByte('\'')
			case '"':
				result.WriteByte('"')
			default:
				result.WriteByte('\\')
				result.WriteByte(nextChar)
			}
			i++
		} else {
			result.WriteByte(char)
		}
	}
	return result.String()
}
