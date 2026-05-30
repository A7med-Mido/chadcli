// internal/search/search.go

// Package search provides fuzzy filtering over a list of explorer entries.
// It wraps github.com/sahilm/fuzzy to give ranked, typo-tolerant results.
package search



import (
	"github.com/sahilm/fuzzy"
	"github.com/A7med-Mido/chadcli/internal/explorer"
)

// Result pairs a matched entry with its match score and highlighted indices.
type Result struct {
	Entry      explorer.Entry
	MatchIndex int    // index in the original slice
	Score      int
	Positions  []int // character positions that matched, for highlighting
}

// Filter returns entries that fuzzy-match the query, ranked by score.
// If the query is empty, all entries are returned in their original order.
func Filter(query string, entries []explorer.Entry) []Result {
	if query == "" {
		results := make([]Result, len(entries))
		for i, e := range entries {
			results[i] = Result{Entry: e, MatchIndex: i, Score: 0}
		}
		return results
	}

	// Build a string source for the fuzzy library.
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name
	}

	matches := fuzzy.Find(query, names)

	results := make([]Result, 0, len(matches))
	for _, m := range matches {
		results = append(results, Result{
			Entry:      entries[m.Index],
			MatchIndex: m.Index,
			Score:      m.Score,
			Positions:  m.MatchedIndexes,
		})
	}
	return results
}