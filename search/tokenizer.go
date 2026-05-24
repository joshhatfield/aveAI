package search

import (
	"strings"
	"unicode"
)

// Token represents a single normalized token with its position.
type Token struct {
	Term     string
	Position int
}

// Tokenizer splits text into normalized tokens.
type Tokenizer struct {
	StopWords map[string]struct{}
}

// NewTokenizer creates a Tokenizer with default English stop words.
func NewTokenizer() *Tokenizer {
	return &Tokenizer{
		StopWords: map[string]struct{}{
			"a": {}, "an": {}, "and": {}, "are": {}, "as": {}, "at": {},
			"be": {}, "by": {}, "for": {}, "from": {}, "has": {}, "he": {},
			"in": {}, "is": {}, "it": {}, "its": {}, "of": {}, "on": {},
			"that": {}, "the": {}, "to": {}, "was": {}, "were": {}, "will": {}, "with": {},
		},
	}
}

// Tokenize splits text into tokens, lowercased and filtered of stop words.
func (t *Tokenizer) Tokenize(text string) []Token {
	var tokens []Token
	position := 0

	var current strings.Builder
	for _, r := range text {
		if unicode.IsSpace(r) || unicode.IsPunct(r) {
			if current.Len() > 0 {
				term := strings.ToLower(current.String())
				if _, ok := t.StopWords[term]; !ok {
					tokens = append(tokens, Token{Term: term, Position: position})
					position++
				}
				current.Reset()
			}
		} else {
			current.WriteRune(r)
		}
	}

	// Flush any remaining token
	if current.Len() > 0 {
		term := strings.ToLower(current.String())
		if _, ok := t.StopWords[term]; !ok {
			tokens = append(tokens, Token{Term: term, Position: position})
		}
	}

	return tokens
}

// TokenizeToTerms is a convenience wrapper that returns only terms (no positions).
func (t *Tokenizer) TokenizeToTerms(text string) []string {
	tokens := t.Tokenize(text)
	terms := make([]string, len(tokens))
	for i, tok := range tokens {
		terms[i] = tok.Term
	}
	return terms
}