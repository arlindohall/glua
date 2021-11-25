
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

Declaration := Statement ( ';' )?

Statement := ForStatement
    | Expression

Expression := Term

Term := Factor ( ('+' | '-') Factor )

Factor := Primary ( ('+' | '-') Primary )

Primary := Number | String | Word

Number := [0-9] +

String := '"' StringChar * '"'

# Note: you can backslash escape quotes, more may be added
# Lua uses backslashes, but I haven't bothered to look up all
# the escaped characters
StringChar := ! ( '\' | '"') | '\"'

Word := [a-zA-Z] [a-zA-Z0-9_-] *
```

[1]: https://craftinginterpreters.com/appendix-i.html