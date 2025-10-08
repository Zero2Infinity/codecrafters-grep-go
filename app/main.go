package main

import (
	"fmt"
	"io"
	"os"
	"slices"
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

	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

	rLine := []rune(string(line))
	tokens, err := tokenizer(pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	fmt.Println(tokens)
	if tokens[0] == "^" { 
		return matchPatterns(rLine, tokens, 0, 1), nil
	} else {
		return matchPatterns(rLine, tokens, 0, 0), nil
	}
}

// TODO: lacks pattern validation
func tokenizer(pattern string) ([]string, error) {
	var tokens []string

	pat := []rune(pattern)
	startIdx := 0
	currIdx := startIdx
	endIdx := len(pat) - 1

	if pat[endIdx] == '$' {
		endIdx -= 1
	}

	if pat[startIdx] == '^' {
		tokens = append(tokens, "^")
		currIdx = startIdx + 1
	}

	for currIdx <= endIdx {
		if pat[currIdx] == '\\' && pat[currIdx+1] != '\\' {
			switch pat[currIdx+1] {
			case 'd':
				if currIdx+2 <= endIdx && pat[currIdx+2] == '+' {
					tokens = append(tokens, "\\d+")
					currIdx += 3
				} else if currIdx+2 <= endIdx && pat[currIdx+2] == '?' {
					tokens = append(tokens, "\\d?")
					currIdx += 3
				} else {
					tokens = append(tokens, "\\d")
					currIdx += 2
				}
			case 'w':
				if currIdx+2 <= endIdx && pat[currIdx+2] == '+' {
					tokens = append(tokens, "\\w+")
					currIdx += 3
				} else if currIdx+2 <= endIdx && pat[currIdx+2] == '?' {
					tokens = append(tokens, "\\w?")
					currIdx += 3
				} else {
					tokens = append(tokens, "\\w")
					currIdx += 2
				}
			}
		} else if (pat[currIdx] >= 'a' && pat[currIdx] <= 'z') ||
			(pat[currIdx] >= 'A' && pat[currIdx] <= 'Z') ||
			(pat[currIdx] >= '0' && pat[currIdx] <= '9') {
			switch {
			case currIdx+1 <= endIdx && pat[currIdx+1] == '+':
				tokens = append(tokens, string([]rune{pat[currIdx], '+'}))
				currIdx += 2
			case currIdx+1 <= endIdx && pat[currIdx+1] == '?':
				tokens = append(tokens, string([]rune{pat[currIdx], '?'}))
				currIdx += 2
			default: 
				tokens = append(tokens, string(pat[currIdx]))
				currIdx += 1
			}
		} else if pat[currIdx] == '.' {
			switch {
			case currIdx+1 <= endIdx && pat[currIdx+1] == '+':
				tokens = append(tokens, ".+")
				currIdx += 2
			case currIdx+1 <= endIdx && pat[currIdx+1] == '?':
				tokens = append(tokens, ".?")
				currIdx += 2
			}
			tokens = append(tokens, ".")
			currIdx += 1
		} else if pat[currIdx] == '[' {
			// TODO: duplicate codes b/w [] and ()
			var t []rune
			for currIdx <= endIdx {
				if pat[currIdx] == ']' {
					t = append(t, ']')
					currIdx += 1
					tokens = append(tokens, string(t))
					break
				}
				t = append(t, pat[currIdx])
				currIdx += 1
			}
		} else if pat[currIdx] == '(' {
			// TODO: duplicate codes b/w [] and ()
			var t []rune
			for currIdx <= endIdx {
				if pat[currIdx] == ')' {
					t = append(t, ')')
					currIdx += 1
					tokens = append(tokens, string(t))
					break
				}
				t = append(t, pat[currIdx])
				currIdx += 1
			}
		} else {
			tokens = append(tokens, string(pat[currIdx]))
			currIdx += 1
		}
	}

	if pat[len(pat)-1] == '$' {
		tokens = append(tokens, "$")
		currIdx += 1
	}

	if currIdx < endIdx {
		return nil, fmt.Errorf("unable to parse pattern: %q", pattern)
	}

	return tokens, nil
}

func matchPatterns(line []rune, tokens []string, i, j int) bool {
	var ok bool

    if i < len(line) {
		fmt.Print(string(line[i])); 
	} else {
		fmt.Print("[empty]")
	}
	fmt.Print(", "); 

	if i < len(line) && j < len(tokens) {
		token := tokens[j]
		fmt.Println(token)
		switch {
		case token == "$":
			return false
		case token == ".":
			ok = matchPatterns(line, tokens, i+1, j+1)
		case token == "\\d":
			if isDigit(line[i]) {
				ok = matchPatterns(line, tokens, i+1, j+1)
			} else {
				ok = matchPatterns(line, tokens, i+1, j)
			}
		case token == "\\w":
			if isWordCharacters(line[i]) {
				ok = matchPatterns(line, tokens, i+1, j+1)
			} else {
				ok = matchPatterns(line, tokens, i+1, j)
			}
		case strings.HasPrefix(token, "[^"):
			p := []rune(token)
			if !slices.Contains(p[2:], line[i]) {
				ok = matchPatterns(line, tokens, i+1, j+1)
			} else {
				ok = matchPatterns(line, tokens, i+1, j)
			}
		case strings.HasPrefix(token, "["):
			q := []rune(token)
			if slices.Contains(q[1:], line[i]) {
				ok = matchPatterns(line, tokens, i+1, j+1)
			} else {
				return false
			}
		case strings.ContainsRune(token, '+'):
			// e.g. a+ or .+ => one or more
			for i < len(line) && 
				(strings.ContainsRune(token, line[i]) || 
				strings.ContainsRune(token, '.') || 
				(strings.Contains(token, "\\d") && isDigit(line[i]))) {
				if matchPatterns(line, tokens, i+1, j+1) {
					return true
				}
				i++
			}
		case strings.ContainsRune(token, '?'):
			// e.g. a? or .? => zero or one
			if strings.Contains(token, "\\d") && isDigit(line[i]) {
				ok = matchPatterns(line, tokens, i+1, j+1)
			} else if strings.Contains(token, "\\d") && !isDigit(line[i]) {
				ok = matchPatterns(line, tokens, i, j+1)
			} else if !strings.ContainsRune(token, line[i]) {
				return matchPatterns(line, tokens, i, j+1)
			} else {
				return matchPatterns(line, tokens, i+1, j+1)
			}
		case strings.HasPrefix(token, "("):
			words := strings.Split(token[1:len(token)-1], "|")
			for _, word := range words {
				if string(line[i:i+len(word)]) == word {
					ok = matchPatterns(line, tokens, i+len(word), j+1)
				}
			}
		case utf8.RuneCountInString(token) == 1:
			if j == 1 && tokens[j-1] == "^" {
				if !strings.ContainsRune(token, line[i]) {
					return false 
				} else {
					ok = matchPatterns(line, tokens, i+1, j+1)
				}
			} else if token == string(line[i]) {
				ok = matchPatterns(line, tokens, i+1, j+1)
			} else {
				ok = matchPatterns(line, tokens, i+1, j)
			}
		}
	}

	if j == len(tokens) {
		// matched all tokens
		ok = true
	} else if i == len(line) && tokens[j] == "$" {
		// e.g. 'a' == 'a$'
		ok = true
	} else if i == len(line) && strings.ContainsRune(tokens[j], '?') {
		// e.g. 'a' == 'as?$'
		ok = true
	}

	return ok
}


func isDigit(r rune) bool {
	return (r >= '0' && r <= '9')
}

func isWordCharacters(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') || (r == '_')
}

