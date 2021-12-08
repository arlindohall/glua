package scanner

import (
	"fmt"
	"os"
)

func (tt TokenType) String() string {
	switch tt {
	case TokenError:
		return "TokenError"
	case TokenAssert:
		return "TokenAssert"
	case TokenGlobal:
		return "TokenGlobal"
	case TokenLocal:
		return "TokenLocal"
	case TokenIdentifier:
		return "TokenIdentifier"
	case TokenNumber:
		return "TokenNumber"
	case TokenString:
		return "TokenString"
	case TokenEof:
		return "TokenEof"
	case TokenPlus:
		return "TokenPlus"
	case TokenMinus:
		return "TokenMinus"
	case TokenBang:
		return "TokenBang"
	case TokenSemicolon:
		return "TokenSemicolon"
	case TokenComma:
		return "TokenComma"
	case TokenDot:
		return "TokenDot"
	case TokenSlash:
		return "TokenSlash"
	case TokenStar:
		return "TokenStar"
	case TokenTrue:
		return "TokenTrue"
	case TokenFalse:
		return "TokenFalse"
	case TokenNil:
		return "TokenNil"
	case TokenEqualEqual:
		return "TokenEqualEqual"
	case TokenEqual:
		return "TokenEqual"
	case TokenLess:
		return "TokenLess"
	case TokenWhile:
		return "TokenWhile"
	case TokenDo:
		return "TokenDo"
	case TokenEnd:
		return "TokenEnd"
	case TokenFunction:
		return "TokenFunction"
	case TokenReturn:
		return "TokenReturn"
	case TokenAnd:
		return "TokenAnd"
	case TokenOr:
		return "TokenOr"
	case TokenLeftBrace:
		return "TokenLeftBrace"
	case TokenRightBrace:
		return "TokenRightBrace"
	case TokenLeftBracket:
		return "TokenLeftBracket"
	case TokenRightBracket:
		return "TokenRightBracket"
	case TokenLeftParen:
		return "TokenLeftParen"
	case TokenRightParen:
		return "TokenRightParen"
	default:
		return fmt.Sprint("UnrecognizedToken/", int(tt))
	}
}

func (t Token) String() string {
	return fmt.Sprintf("%v/\"%v\"", t.Type, t.Text)
}

// todo: if using goroutines to publish tokens, how to debug?
func DebugTokens(tokens []Token) {
	for _, token := range tokens {
		if token.Type == TokenSemicolon {
			fmt.Fprintln(os.Stderr, ";")
			continue
		}
		fmt.Fprint(os.Stderr, token, " ")
	}

	fmt.Fprintln(os.Stderr)
}
