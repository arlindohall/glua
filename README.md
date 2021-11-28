
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

## Grammar

I based this grammar on the Lox Grammar[1]

```
Program := Declaration +

# I can't remember semicolon rules in Lua but for now I'm going to
# use them to deliniate the end of an expression
Declaration := Statement ( ';' )?

Statement := AssertStatement
    | Expression

AssertStatement := 'assert' Expression

Expression := LogicOr

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

Exponent := Primary ( '^' Primary )

Primary := Number | String | Identifier | 'nil'

Number := [0-9] +

String := '"' StringChar * '"'

# Note: you can backslash escape quotes, more may be added
# Lua uses backslashes, but I haven't bothered to look up all
# the escaped characters
StringChar := ! ( '\' | '"') | '\"'

Identifier := [a-zA-Z] [a-zA-Z0-9_-] *
```

[1]: https://craftinginterpreters.com/appendix-i.html