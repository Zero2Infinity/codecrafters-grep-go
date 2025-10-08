package main

import (
	"fmt"
	"testing"
)

func TestMatchLine(t *testing.T) {
	tests := []struct {
		line      []byte
		pattern   string
		expected  bool
		expectErr bool
	}{
		{[]byte("apple"), "a", true, false},
		{[]byte("apple"), "z", false, false},
		{[]byte("app1e"), "\\d", true, false},
		{[]byte("apple"), "\\d", false, false},
		{[]byte("abc"), "\\w", true, false},
		{[]byte("123"), "\\w", true, false},
		{[]byte("alpha_num3ric"), "\\w", true, false},
		{[]byte("apple"), "[abc]", true, false},
		{[]byte("pple"), "[abc]", false, false},
		{[]byte("cab"), "[^abc]", false, false},
		{[]byte("cat"), "[^abc]", true, false},
		{[]byte("1 apple"), "\\d apple", true, false},
		{[]byte("apple"), "\\d apple", false, false},
		{[]byte("log"), "^log", true, false},
		{[]byte("prolog"), "^log", false, false},
		{[]byte("dog"), "dog$", true, false},
		{[]byte("abc"), "^abc$", true, false},
		{[]byte("cts"), "ca+ts", false, false},
		{[]byte("caats"), "ca+ts", true, false},
		{[]byte("caats"), "ca+ats", true, false},
		{[]byte("dogs"), "dogs?", true, false},
		{[]byte("dogs"), "d?ogs", true, false},
		{[]byte("2 dogs"), "\\d+ dogs", true, false},
		{[]byte("dogs"), "\\d?dogs", true, false},
		{[]byte("dog"), "d.g", true, false},
		{[]byte("dogcat"), "d.+cat", true, false},
		{[]byte("abc"), "(a|b)", true, false},
		{[]byte("wxyz"), "(w|v)xyz", true, false},
		{[]byte("xxx"), "(w|v)xxx", false, false},
		{[]byte("I see 2 dog3"), "^I see \\d+ (cat|dog)s?$", false, false},
		{[]byte("I see 2 dog"), "^I see \\d+ (cat|dog)s?$", true, false},
		{[]byte("I see 2 dogs"), "^I see \\d+ (cat|dog)s?$", true, false},
		{[]byte("sally has 12 apples"), "\\d\\\\d\\\\d apples", false, false},
		{[]byte("[]"), "[banana]", false, false},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("line=%s, pattern=%s", test.line, test.pattern), func(t *testing.T) {

			result, err := matchLine(test.line, test.pattern)

			if (err != nil) != test.expectErr {
				t.Errorf("expected error: %v, got: %v", test.expectErr, err)
			}

			if result != test.expected {
				t.Errorf("expected: %v, got: %v", test.expected, result)
			}
		})
	}
}

/*
func BenchmarkRandInt(b *testing.B) {
	line := []byte("123abc")
	pattern := "\\d"

	_, err := matchLine(line, pattern)
	if err != nil {
		b.Errorf("unexpected error: %v", err)
	}
}
*/
