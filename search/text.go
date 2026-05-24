package search

import (
	"math"

	"aveAI/store"
)

// Posting represents a single term occurrence in a document.
type Posting struct {
	EntryID   uint64
	Freq      int    // term frequency in this document
	Positions []int  // positions where term appears
}

// InvertedIndex builds and queries an in-memory inverted index.
type InvertedIndex struct {
	postings map[string][]Posting
	docFreq  map[string]int  // doc frequency per term
	docCount int
	idfCache map[string]float64
}

// NewInvertedIndex creates an empty index.
func NewInvertedIndex() *InvertedIndex {
	return &InvertedIndex{
		postings: make(map[string][]Posting),
		docFreq:  make(map[string]int),
		idfCache: make(map[string]float64),
	}
}

// Build constructs the inverted index from a list of entries.
func (idx *InvertedIndex) Build(entries []store.Entry) {
	idx.postings = make(map[string][]Posting)
	idx.docFreq = make(map[string]int)
	idx.idfCache = make(map[string]float64)
	idx.docCount = len(entries)

	tok := NewTokenizer()

	for _, e := range entries {
		tokens := tok.Tokenize(e.Value)

		// Group tokens by term for this document
		termPositions := make(map[string][]int)
		for _, t := range tokens {
			termPositions[t.Term] = append(termPositions[t.Term], t.Position)
		}

		// Add posting for each term
		for term, positions := range termPositions {
			idx.postings[term] = append(idx.postings[term], Posting{
				EntryID:   e.ID,
				Freq:      len(positions),
				Positions: positions,
			})
			idx.docFreq[term]++
		}
	}

	// Precompute IDF for all terms
	for term, df := range idx.docFreq {
		idx.idfCache[term] = idf(idx.docCount, df)
	}
}

// SearchResult holds a scored search result.
type SearchResult struct {
	EntryID uint64
	Score   float64
}

// Search performs AND search on query terms, returns scored results sorted descending.
func (idx *InvertedIndex) Search(query string) []SearchResult {
	if idx.docCount == 0 {
		return nil
	}

	tokens := NewTokenizer().Tokenize(query)
	if len(tokens) == 0 {
		return nil
	}

	// Get posting lists for each query term
	var postingLists []map[uint64]Posting
	for _, t := range tokens {
		plist, ok := idx.postings[t.Term]
		if !ok {
			// Term not in any document — no results
			return nil
		}
		m := make(map[uint64]Posting)
		for _, p := range plist {
			m[p.EntryID] = p
		}
		postingLists = append(postingLists, m)
	}

	// Intersect: find docs that have ALL query terms
	candidates := postingLists[0]
	for i := 1; i < len(postingLists); i++ {
		intersected := make(map[uint64]Posting)
		for id := range candidates {
			if p, ok := postingLists[i][id]; ok {
				intersected[id] = p
			}
		}
		candidates = intersected
	}

	if len(candidates) == 0 {
		return nil
	}

	// Score each candidate using TF-IDF
	results := make([]SearchResult, 0, len(candidates))
	for id := range candidates {
		var score float64
		for _, t := range tokens {
			plist := idx.postings[t.Term]
			var tf float64
			var posting Posting
			for _, p := range plist {
				if p.EntryID == id {
					posting = p
					break
				}
			}
			// TF: log(1 + freq) — log-scaled term frequency
			tf = math.Log(1 + float64(posting.Freq))
			// IDF from cache
			idf := idx.idfCache[t.Term]
			score += tf * idf
		}
		results = append(results, SearchResult{EntryID: id, Score: score})
	}

	// Sort by score descending
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}

// idf computes inverse document frequency: log(1 + (N - df) / (df + 1))
// This formula ensures IDF is always positive, even when df == N.
// When term appears in all docs (df = N), idf = log(2) ≈ 0.693
func idf(N, df int) float64 {
	return math.Log(1 + float64(N-df+1)/float64(df+1))
}