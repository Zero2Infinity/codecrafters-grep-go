# Project Overview

This project is a Go implementation of the `grep` command-line utility, as part of the CodeCrafters "Build Your Own grep" challenge. The goal is to build a tool that can search for patterns in text using regular expressions.

The main application logic is in `app/main.go` and associated tests is in `app/main_test.go`. 

The project uses Go version 1.24.0, as specified in the `go.mod` file.

# Building and Running

The project can be built and run using the provided shell script:

```sh
./your_program.sh
```
The project can be built and test using the provided go command:
```sh
go test
```

This script compiles the Go code located in `app/main.go` and then executes the resulting binary.

# Development Conventions

The code implements a basic version of `grep` that supports some regular expression features:
* Match a literal character
    - matches a single literal character.
* Match digits
    - `\d` matches any digit.
    - Example: "\d" should match "3", but not "c".
* Match word characters
    - `\w` matches any alphanumeric character (a-z, A-Z, 0-9) and underscore _.
    - Example: "\w" should match "foo101", but not "$!?".
* Positive Character Groups
    - Positive character groups match any character that is present within a pair of square brackets.
    - Example: "[abc]" should match "apple", but not "dog".
* Negative Character Groups
    - Negative character groups match any character that is not present within a pair of square brackets.
    - Example: "[^abc]" should match "cat", since "t" is not in the set "a", "b", or "c".
    - Example "[^abc]" should not match "cab", since all characters are in the set.
* Combining Character Classes
    - Examples:
        - "\d apple" should match "1 apple", but not "1 orange".
        - "\d\d\d apple" should match "100 apples", but not "1 apple".
        - "\d \w\w\ws" should match "3 dogs" and "4 cats" but not "1 dog" (because the "s" is not present at the end).
* Start of string anchor
    - In this stage, we'll add support for ^, the Start of String or Line anchor. `^` doesn't match a character, it matches the start of a line.
    - Example: "^log" should match "log", but not "slog".
* End of string anchor
    - In this stage, we'll add support for $, the End of String or Line anchor. `$` doesn't match a character, it matches the end of a line.
    - Example: "dog$" should match "dog", but not "dogs".
* Match one or more times
    - In this stage, we'll add support for +, the one or more quantifier.
    - Example: "a+" should match "apple" and "SaaS", but not "dog".
* Match zero or one times
    - In this stage, we'll add support for ?, the zero or one quantifier (also known as the "optional" quantifier).
    - Example: "dogs?" should match "dogs" and "dog", but not "cat".
* Wildcard 
    - In this stage, we'll add support for ., which matches any character.
    - Example: "d.g" should match "dog", but not "cog".
* Alternation
    - In this stage, we'll add support for the | keyword, which allows combining multiple patterns in an either/or fashion.
    - Example: "(cat|dog)" should match "dog" and "cat", but not "apple".

The `main` function parses command-line arguments and reads from standard input. 
The `matchLine` function is the core of the implementation, responsible for matching the input against the provided pattern.
The code includes a `tokenizer` to break down the regex pattern and a `matchPatterns` function to recursively match the tokens against the input line.


