package search

import (
	"math"
	"testing"

	"aveAI/store"
)

func TestInvertedIndexBuildEmpty(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Build([]store.Entry{})
	if idx.docCount != 0 {
		t.Errorf("expected 0 docs, got %d", idx.docCount)
	}
}

func TestInvertedIndexBuildSingle(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Build([]store.Entry{
		{ID: 1, Value: "always wrap errors %w"},
	})

	if idx.docCount != 1 {
		t.Errorf("expected docCount=1, got %d", idx.docCount)
	}

	// Should have postings for: "always", "wrap", "errors", "w" (no stop words in this phrase)
	// "with" would be a stop word but it's not in this phrase
	expectedTerms := []string{"always", "wrap", "errors"}
	for _, term := range expectedTerms {
		if _, ok := idx.postings[term]; !ok {
			t.Errorf("expected posting for term '%s'", term)
		}
	}
}

func TestInvertedIndexBuildMultiple(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Build([]store.Entry{
		{ID: 1, Value: "wrap errors with %w"},
		{ID: 2, Value: "never use panic"},
		{ID: 3, Value: "errors should be wrapped"},
	})

	if idx.docCount != 3 {
		t.Errorf("expected docCount=3, got %d", idx.docCount)
	}

	// "errors" appears in docs 1 and 3
	plist, ok := idx.postings["errors"]
	if !ok {
		t.Fatalf("expected postings for 'errors'")
	}
	if len(plist) != 2 {
		t.Errorf("expected 2 postings for 'errors', got %d", len(plist))
	}

	// "wrap" appears in doc 1 (exact match), doc 3 has "wrapped" — different term
	plist, ok = idx.postings["wrap"]
	if !ok {
		t.Fatalf("expected postings for 'wrap'")
	}
	if len(plist) != 1 {
		t.Errorf("expected 1 posting for 'wrap' (doc1), got %d", len(plist))
	}
}

func TestInvertedIndexSearchSingleTerm(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Build([]store.Entry{
		{ID: 1, Value: "wrap errors with %w"},
		{ID: 2, Value: "never use panic"},
		{ID: 3, Value: "errors should be wrapped"},
	})

	results := idx.Search("errors")
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestInvertedIndexSearchMultipleTerms(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Build([]store.Entry{
		{ID: 1, Value: "wrap errors with %w"},
		{ID: 2, Value: "never use panic"},
		{ID: 3, Value: "errors should be wrapped"},
	})

	// AND search — "wrapped" ≠ "wrap", so only doc 1 has both "errors" AND "wrap"
	results := idx.Search("errors wrap")
	if len(results) != 1 {
		t.Errorf("expected 1 result (doc1), got %d", len(results))
	}
	if results[0].EntryID != 1 {
		t.Errorf("expected entry 1, got %d", results[0].EntryID)
	}
}

func TestInvertedIndexSearchNoMatch(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Build([]store.Entry{
		{ID: 1, Value: "wrap errors with %w"},
	})

	results := idx.Search("nonexistent term")
	if results != nil {
		t.Errorf("expected nil for no match, got %v", results)
	}
}

func TestInvertedIndexSearchEmpty(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Build([]store.Entry{
		{ID: 1, Value: "wrap errors with %w"},
	})

	results := idx.Search("")
	if results != nil {
		t.Errorf("expected nil for empty query, got %v", results)
	}
}

func TestInvertedIndexSearchScores(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Build([]store.Entry{
		{ID: 1, Value: "apple"},                              // doc with unique word
		{ID: 2, Value: "apple apple apple apple apple"},     // apple 5 times
		{ID: 3, Value: "apple banana"},                       // apple + another word
	})

	results := idx.Search("apple")

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Doc 2 should score highest (TF=5, same df as others)
	if results[0].EntryID != 2 {
		t.Errorf("expected doc 2 first (highest TF), got %d", results[0].EntryID)
	}

	// Doc 1 and 3 have lower TF — order should reflect TF
	// (tie-break by ID order if same TF)
}

func TestInvertedIndexIDFCalculation(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Build([]store.Entry{
		{ID: 1, Value: "term appears in doc1 only"},
		{ID: 2, Value: "other content"},
		{ID: 3, Value: "more content"},
	})

	// N=3, df(term in 1 doc) = 1
	// idf = log(1 + (3-1+1)/(1+1)) = log(1 + 3/2) = log(2.5) ≈ 0.916
	idf := idx.idfCache["term"]
	expectedIDF := 0.916290732 // math.Log(2.5)
	if math.Abs(idf-expectedIDF) > 0.0001 {
		t.Errorf("expected IDF=%f, got %f", expectedIDF, idf)
	}
}

func TestInvertedIndexOrderByScore(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Build([]store.Entry{
		{ID: 1, Value: "word here"},                                           // "word" appears once
		{ID: 2, Value: "word word word word word here"},                      // "word" appears 5 times — highest
		{ID: 3, Value: "word word here"},                                     // "word" appears 2 times
	})

	results := idx.Search("word")

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Doc 2 should be first (TF=5)
	if results[0].EntryID != 2 {
		t.Errorf("expected doc 2 first, got %d (scores: %v)", results[0].EntryID, results)
	}
	// Doc 3 should be second (TF=2)
	if results[1].EntryID != 3 {
		t.Errorf("expected doc 3 second, got %d", results[1].EntryID)
	}
	// Doc 1 should be third (TF=1)
	if results[2].EntryID != 1 {
		t.Errorf("expected doc 1 third, got %d", results[2].EntryID)
	}

	// Verify all scores are positive
	for _, r := range results {
		if r.Score <= 0 {
			t.Errorf("expected positive score, got %f for doc %d", r.Score, r.EntryID)
		}
	}
}

func TestInvertedIndexSortResults(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Build([]store.Entry{
		{ID: 1, Value: "apple banana"},
		{ID: 2, Value: "apple apple banana"},
		{ID: 3, Value: "apple apple apple banana"},
	})

	results := idx.Search("apple banana")

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Doc 3 has both terms most frequently → highest score
	// Doc 2 second, doc 1 lowest
	if results[0].EntryID != 3 {
		t.Errorf("expected doc 3 first (highest TF), got %d", results[0].EntryID)
	}

	// Verify descending order
	for i := 0; i < len(results)-1; i++ {
		if results[i+1].Score > results[i].Score {
			t.Errorf("results not sorted: results[%d]=%f > results[%d]=%f",
				i+1, results[i+1].Score, i, results[i].Score)
		}
	}
}

func TestInvertedIndexNoStopWordsInPostings(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Build([]store.Entry{
		{ID: 1, Value: "the quick brown fox"},
	})

	// "the" is a stop word — should not appear in index
	if _, ok := idx.postings["the"]; ok {
		t.Errorf("stop word 'the' should not be in postings")
	}

	// "quick", "brown", "fox" should be there
	for _, term := range []string{"quick", "brown", "fox"} {
		if _, ok := idx.postings[term]; !ok {
			t.Errorf("expected '%s' in postings", term)
		}
	}
}