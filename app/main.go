package main

import (
	"bytes"
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

// TODO: this big if-else block is wrong, we need to combining patterns!!
func matchLine(line []byte, pattern string) (bool, error) {
	if utf8.RuneCountInString(pattern) == 0 {
		return false, fmt.Errorf("unsupported pattern: %q", pattern)
	}

	var ok bool
	ll, pl := len(line), len(pattern)

	if pattern == "\\d" {
		ok = bytes.ContainsAny(line, "1234567890")
	} else if pattern == "\\w" {
		ok = bytes.ContainsAny(line, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890_")
	} else if strings.HasPrefix(pattern, "[^") && strings.HasSuffix(pattern, "]") {
		ok = !containsAll(line, pattern[2:pl-1])
	} else if strings.HasPrefix(pattern, "[") && strings.HasSuffix(pattern, "]") {
		ok = bytes.ContainsAny(line, pattern[1:pl-1])
	} else if strings.HasPrefix(pattern, "^") && strings.HasSuffix(pattern, "$") {
		ok = strings.Compare(string(line), pattern[1:pl-1]) == 0
	} else if strings.HasPrefix(pattern, "^") {
		ok = containsAll(line[0:pl-1], pattern[1:pl-1])
	} else if strings.HasSuffix(pattern, "$") {
		ok = containsAll(line[ll-(pl-1):], pattern[0:pl-2])
	} else if strings.Contains(pattern, "+") {
		ok = containsOneOrMore(line, []byte(pattern))
	} else if strings.Contains(pattern, "?") {
		ok = containsZeroOrOne(line, []byte(pattern))
	} else if strings.Contains(pattern, ".") {
		ok = containsAnyChar(line, []byte(pattern))
	} else if strings.ContainsAny(pattern, "(|)") {
		ok = alternation(string(line), pattern)
	} else {
		// ok = bytes.ContainsAny(line, pattern)
		ok = combineCharClasses(line, []byte(pattern))
	}

	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

	return ok, nil
}

// convert []byte to []rune to support UTF-8
func combineCharClasses(line, pat []byte) bool {
	i, j := 0, 0
	for ; i < len(line) && j < len(pat); i++ {
		// fmt.Printf("%d == %d\n", i, j)
		if j+1 == len(pat) && pat[j] == '\\' {
			// fmt.Printf("\\$ == %c\n", line[i])
			if line[i] == pat[j] {
				j++
			}
		} else if pat[j] == '\\' && pat[j+1] == 'd' {
			if line[i] >= '0' && line[i] <= '9' {
				// fmt.Printf("%c is digit\n", line[i])
				j += 2
			}
		} else if pat[j] == '\\' && pat[j+1] == 'w' {
			if (line[i] >= 'a' && line[i] <= 'z') || (line[i] >= 'A' && line[i] <= 'Z') ||
				(line[i] >= '0' && line[i] <= '9') || line[i] == '_' {
				// fmt.Printf("%c is char\n", line[i])
				j += 2
			}
		} else if line[i] == pat[j] {
			// fmt.Printf("%c == %c\n", line[i], pat[j])
			j++
		} else if j != 0 && line[i] != pat[j] {
			// fmt.Printf("%c != %c -- reset\n", line[i], pat[j])
			j++
			j = 0
		}
	}

	return j == len(pat) // return based on if we found pattern
}

// special function to match each
func containsAll(b []byte, pat string) bool {
	patArr := []byte(pat)
	for i := 0; i < len(pat); i++ {
		if !bytes.Contains(b, patArr[i:i+1]) {
			return false
		}
	}
	return true
}

func containsOneOrMore(line, pat []byte) bool {
	i, j := 0, 0
	for i < len(line) {
		if backtrack(line, pat, i, j) {
			return true
		}
		i++
	}
	return false
}

func containsZeroOrOne(line, pat []byte) bool {
	i, j := 0, 0
	for i < len(line) {
		if backtrack(line, pat, i, j) {
			return true
		}
		i++
	}
	return false
}

func backtrack(line, pat []byte, i, j int) bool {
	if j == len(pat) {
		return true
	}

	if j+1 < len(pat) {
		switch pat[j+1] {
		case '+':
			if pat[j] == '.' {
				for i < len(line) {
					if backtrack(line, pat, i+1, j+2) {
						return true
					}
					i++
				}
			} else if line[i] == pat[j] {
				for i < len(line) && line[i] == pat[j] {
					if backtrack(line, pat, i+1, j+2) {
						return true
					}
					i++
				}
			}
		case '?':
			if i < len(line) && line[i] != pat[j] {
				return backtrack(line, pat, i, j+2)
			} else {
				return backtrack(line, pat, i+1, j+2)
			}
		}
	}

	if i < len(line) && line[i] == pat[j] {
		return backtrack(line, pat, i+1, j+1)
	}

	return false
}

func containsAnyChar(line, pat []byte) bool {
	i, j := 0, 0
	for i < len(line) && j < len(pat) {
		if line[i] == pat[j] || pat[j] == '.' {
			j++
		} else {
			j = 0
		}
		i++
	}
	return j == len(pat)
}

func alternation(line, pat string) bool {
	openIdx := strings.Index(pat, "(")
	closeIdx := strings.Index(pat, ")")
	tokens := strings.Split(pat[openIdx+1:closeIdx], "|")

	i, j := 0, 0
	for i < len(line) && j < len(pat) {
		if pat[j] == '(' {
			count := 0
			for _, t := range tokens {
				if strings.Compare(t, line[i:i+len(t)]) == 0 {
					count++
					i += len(t)
					j = closeIdx + 1
					break
				}
			}
			if count == 0 {
				return false
			}
		} else if line[i] == pat[j] {
			i++
			j++
		}
	}

	return j == len(pat)
}
