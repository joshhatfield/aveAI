package search

import (
	"reflect"
	"testing"
)

func TestTokenizeBasic(t *testing.T) {
	tok := NewTokenizer()
	tokens := tok.Tokenize("hello world this is a test")

	// "is" and "a" are stop words and filtered out
	// Expected: "hello", "world", "this", "test"
	if len(tokens) != 4 {
		t.Errorf("expected 4 tokens, got %d", len(tokens))
	}

	want := []Token{
		{Term: "hello", Position: 0},
		{Term: "world", Position: 1},
		{Term: "this", Position: 2},
		{Term: "test", Position: 3},
	}
	for i, tok := range tokens {
		if tok.Term != want[i].Term {
			t.Errorf("token[%d]: got '%s', want '%s'", i, tok.Term, want[i].Term)
		}
		if tok.Position != want[i].Position {
			t.Errorf("token[%d] position: got %d, want %d", i, tok.Position, want[i].Position)
		}
	}
}

func TestTokenizeStopWordsFiltered(t *testing.T) {
	tok := NewTokenizer()
	tokens := tok.Tokenize("the quick brown fox jumps over the lazy dog")

	terms := make([]string, len(tokens))
	for i, t := range tokens {
		terms[i] = t.Term
	}

	// Known stop words: "the", "a", "over" (not in current list but checking)
	// Current stop words: a, an, and, are, as, at, be, by, for, from, has, he, in, is, it, its, of, on, that, the, to, was, were, will, with
	for _, sw := range []string{"the", "a"} {
		for _, term := range terms {
			if term == sw {
				t.Errorf("stop word '%s' should be filtered", sw)
			}
		}
	}

	// "over" is NOT in our stop word list, so it should remain
	found := false
	for _, term := range terms {
		if term == "over" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("'over' should NOT be filtered (not in stop word list)")
	}
}

func TestTokenizePunctuationSplit(t *testing.T) {
	tok := NewTokenizer()
	tokens := tok.Tokenize("error: can't wrap nil")

	terms := make([]string, len(tokens))
	for i, t := range tokens {
		terms[i] = t.Term
	}

	// "can't" → apostrophe splits → "can" + "t"
	// "can" is a stop word, so only "t" remains
	found := false
	for _, term := range terms {
		if term == "t" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 't' in tokens (from can't), got %v", terms)
	}
}

func TestTokenizeEmpty(t *testing.T) {
	tok := NewTokenizer()
	tokens := tok.Tokenize("")
	if len(tokens) != 0 {
		t.Errorf("expected empty tokens, got %d", len(tokens))
	}
}

func TestTokenizeToTerms(t *testing.T) {
	tok := NewTokenizer()
	terms := tok.TokenizeToTerms("one two three")
	if len(terms) != 3 {
		t.Errorf("expected 3 terms, got %d", len(terms))
	}
}

func TestTokenizeLowercase(t *testing.T) {
	tok := NewTokenizer()
	tokens := tok.Tokenize("ERROR WRAPPING")
	if tokens[0].Term != "error" || tokens[1].Term != "wrapping" {
		t.Errorf("expected lowercase, got %v", tokens)
	}
}

func TestTokenizeNumbers(t *testing.T) {
	tok := NewTokenizer()
	tokens := tok.Tokenize("go1.22 is released")
	// "go1.22" — period splits, so we get "go1" and "22"
	if len(tokens) < 2 {
		t.Errorf("expected tokens from 'go1.22', got %v", tokens)
	}
}

func TestTokenizePositions(t *testing.T) {
	tok := NewTokenizer()
	tokens := tok.Tokenize("quick brown fox")

	if tokens[0].Position != 0 {
		t.Errorf("expected position 0, got %d", tokens[0].Position)
	}
	if tokens[1].Position != 1 {
		t.Errorf("expected position 1, got %d", tokens[1].Position)
	}
	if tokens[2].Position != 2 {
		t.Errorf("expected position 2, got %d", tokens[2].Position)
	}
}

func TestTokenizePositionsSkippingStopWords(t *testing.T) {
	tok := NewTokenizer()
	tokens := tok.Tokenize("the quick brown")

	// "the" is filtered, so positions should be 0, 1 for quick, brown
	if len(tokens) != 2 {
		t.Errorf("expected 2 tokens, got %d", len(tokens))
	}
	if tokens[0].Term != "quick" || tokens[0].Position != 0 {
		t.Errorf("quick should be at position 0, got %v", tokens[0])
	}
	if tokens[1].Term != "brown" || tokens[1].Position != 1 {
		t.Errorf("brown should be at position 1, got %v", tokens[1])
	}
}

func TestTokenizeMultipleSpaces(t *testing.T) {
	tok := NewTokenizer()
	tokens := tok.Tokenize("hello    world")
	if len(tokens) != 2 {
		t.Errorf("expected 2 tokens, got %d", len(tokens))
	}
}

var _ = reflect.DeepEqual // use reflect in tests