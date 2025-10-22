package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

// Usage: echo <input_text> | your_program.sh -E <pattern>
func main() {
	if len(os.Args) < 3 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
		os.Exit(2) // 1 means no lines were selected, >1 means error
	}

	pattern := os.Args[2]

	line, err := io.ReadAll(os.Stdin) // assume we're only dealing with a single line
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
		os.Exit(2)
	}

	ok, err := matchLine(line, pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	if !ok {
		os.Exit(1)
	}

	// default exit code is 0 which means success
}

func matchLine(line []byte, pattern string) (bool, error) {
	if utf8.RuneCountInString(pattern) == 0 {
		return false, fmt.Errorf("unsupported pattern: %q", pattern)
	}

	rLine := []rune(string(line))
	tokens, err := tokenizer(pattern)
	if err != nil {
		return false, err
	}

	if len(tokens) > 0 && tokens[0] == "^" {
		// Anchored search: must match from the start of the line.
		return matchPatterns(rLine, tokens[1:]), nil
	}

	// Unanchored search: check for a match at every position in the line.
	for i := range rLine {
		if matchPatterns(rLine[i:], tokens) {
			return true, nil
		}
	}

	return false, nil
}

func tokenizer(pattern string) ([]string, error) {
	var tokens []string

	pat := []rune(pattern)
	startIdx := 0
	currIdx := startIdx
	endIdx := len(pat) - 1

	if len(pat) > 0 && pat[endIdx] == '$' {
		endIdx -= 1
	}

	if len(pat) > 0 && pat[startIdx] == '^' {
		tokens = append(tokens, "^")
		currIdx = startIdx + 1
	}

	for currIdx <= endIdx {
		char := pat[currIdx]
		if char == '\\' && currIdx+1 <= endIdx && pat[currIdx+1] != '\\' {
			switch pat[currIdx+1] {
			case 'd', 'w':
				token := string(pat[currIdx : currIdx+2])
				if currIdx+2 <= endIdx && (pat[currIdx+2] == '+' || pat[currIdx+2] == '?') {
					tokens = append(tokens, token+string(pat[currIdx+2]))
					currIdx += 3
				} else {
					tokens = append(tokens, token)
					currIdx += 2
				}
			default:
				tokens = append(tokens, string(pat[currIdx+1]))
				currIdx += 2
			}
		} else if isWordCharacters(char) {
			token := string(char)
			if currIdx+1 <= endIdx && (pat[currIdx+1] == '+' || pat[currIdx+1] == '?') {
				tokens = append(tokens, token+string(pat[currIdx+1]))
				currIdx += 2
			} else {
				tokens = append(tokens, token)
				currIdx += 1
			}
		} else if char == '.' {
			token := "."
			if currIdx+1 <= endIdx && (pat[currIdx+1] == '+' || pat[currIdx+1] == '?') {
				tokens = append(tokens, token+string(pat[currIdx+1]))
				currIdx += 2
			} else {
				tokens = append(tokens, token)
				currIdx += 1
			}
		} else if char == '[' {
			token, nextIdx, err := parseGroup(pat, currIdx, endIdx, '[', ']')
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
			currIdx = nextIdx
		} else if char == '(' {
			token, nextIdx, err := parseGroup(pat, currIdx, endIdx, '(', ')')
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
			currIdx = nextIdx
		} else {
			tokens = append(tokens, string(char))
			currIdx += 1
		}
	}

	if len(pat) > 0 && pat[len(pat)-1] == '$' {
		tokens = append(tokens, "$")
	}

	return tokens, nil
}

// matchPatterns checks if the pattern (tokens) matches from the beginning of the line.
func matchPatterns(line []rune, tokens []string) bool {
	if len(tokens) == 0 {
		return true // All tokens matched
	}

	token := tokens[0]
	remainingTokens := tokens[1:]

	if token == "$" {
		return len(line) == 0
	}

	if strings.HasSuffix(token, "?") {
		baseToken := token[:len(token)-1]
		if matchPatterns(line, remainingTokens) {
			return true // Zero-match case
		}
		if len(line) > 0 && matchChar(line[0], baseToken) {
			return matchPatterns(line[1:], remainingTokens) // One-match case
		}
		return false
	}

	if strings.HasSuffix(token, "+") {
		baseToken := token[:len(token)-1]
		if len(line) == 0 || !matchChar(line[0], baseToken) {
			return false // Must match at least once
		}
		// After one match, it can either be followed by the rest of the pattern...
		if matchPatterns(line[1:], remainingTokens) {
			return true
		}
		// ...or by more instances of the same token.
		return matchPatterns(line[1:], tokens)
	}

	if strings.HasPrefix(token, "[") {
		if len(line) == 0 {
			return false
		}
		char := line[0]
		content := token[1 : len(token)-1]
		matched := false
		if strings.HasPrefix(content, "^") {
			if !strings.ContainsRune(content[1:], char) {
				matched = true
			}
		} else {
			if strings.ContainsRune(content, char) {
				matched = true
			}
		}
		if matched {
			return matchPatterns(line[1:], remainingTokens)
		}
		return false
	}

	if strings.HasPrefix(token, "(") {
		options := strings.Split(token[1:len(token)-1], "|")
		for _, opt := range options {
			if strings.HasPrefix(string(line), opt) {
				if matchPatterns(line[len(opt):], remainingTokens) {
					return true
				}
			}
		}
		return false
	}

	// Simple character match
	if len(line) > 0 && matchChar(line[0], token) {
		return matchPatterns(line[1:], remainingTokens)
	}

	return false
}

func matchChar(r rune, token string) bool {
	switch token {
	case ".":
		return true
	case "\\d":
		return isDigit(r)
	case "\\w":
		return isWordCharacters(r)
	default:
		return string(r) == token
	}
}

func isDigit(r rune) bool {
	return (r >= '0' && r <= '9')
}

func isWordCharacters(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') || (r == '_')
}

func parseGroup(pat []rune, startIdx int, endIdx int, openDelim rune, closeDelim rune) (string, int, error) {
	var t []rune
	currIdx := startIdx
	for currIdx <= endIdx {
		char := pat[currIdx]
		t = append(t, char)
		if char == closeDelim {
			return string(t), currIdx + 1, nil
		}
		currIdx++
	}
	return "", startIdx, fmt.Errorf("unmatched '%c' in pattern", openDelim)
}
