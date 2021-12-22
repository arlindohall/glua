
# Glua

A toy programming language based on Lua and built with Go.

## Building and running programs

To lint the project I use:

```
golangci-lint run
```

And to run the code I use

```
go run . <filename>
```

## Missing Features List

- Weak tables
- `next` builtin function

## Grammar

_This is really out of date but I don't want to bother fixing it right now_.

I based this grammar on the Lox Grammar[1]

```
Program := Declaration +

# I can't remember semicolon rules in Lua but for now I'm going to
# use them to deliniate the end of an expression
Declaration := GlobalDeclaration
    | LocalDeclaration
    | FunctionDeclaration
    | Statement ( ';' )?

Statement := AssertStatement
    | WhileStatement
    | Expression

AssertStatement := 'assert' Expression

Expression := Assignment

Assignment := ( Call '.' )? Identifier ( '=' Assignment )
    | LogicOr

# Logic and is higher precedence than or; or is the lowest
# precedence operator
#
# x or y and z will evaluate as x or (y and z)
LogicOr := LogicAnd ( 'or' LogicAnd ) *

LogicAnd := Comparison ( 'and' Comparison ) *

# Skip concatenation operator for now
Comparison := Term ( ('<' | '>' | '<=' | '>=' | '~=' | '==' ) Term ) *

Term := Factor ( ('+' | '-') Factor ) *

Factor := Unary ( ('*' | '/') Unary ) *

# The way I'd represent is...
# Unary := ('-' | '!') * Exponent
# but that would require an explicit stack to implement as written.
#
# Instead use a recursive definition
Unary := ('-' | '!') Unary | Exponent

Exponent := Call ( '^' Call )

# Name is call because this is where the precedence for function
# calls will go, highest precedence except for literals and
# identifiers
Call := Primary ( '.' Identifier ) *

Primary := Number | String | Identifier | 'nil' | Table

Number := [0-9] +

String := '"' StringChar * '"'

# Note: you can backslash escape quotes, more may be added
# Lua uses backslashes, but I haven't bothered to look up all
# the escaped characters
StringChar := ! ( '\' | '"') | '\"'

Identifier := [a-zA-Z] [a-zA-Z0-9_-] *

Table := '{' Pair * '}'

Pair := StringPair | LiteralPair | Value

StringPair := Identifier '=' Value

LiteralPair := LiteralKey '=' Value

Value := Expression

LiteralKey := '[' Expression ']'
```

[1]: https://craftinginterpreters.com/appendix-i.html