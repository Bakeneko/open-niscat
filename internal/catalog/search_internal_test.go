package catalog

import (
	"context"
	"testing"
)

func TestScopedSearchFiltersBeforeTheCap(t *testing.T) {
	s := openInternal(t)
	ctx := context.Background()
	old := maxMatches
	maxMatches = 1
	t.Cleanup(func() { maxMatches = old })

	// Engine c01=2: only 231 is applicable in group B, so "palier" must find AA6 (231), not lose it
	// behind AA3 (230A) which comes first and is not applicable.
	sc, err := s.ResolveScope(ctx, ScopeParams{VIN: "VSKBEC220U0111111"})
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Search(ctx, SearchQuery{Q: "palier", Kind: SearchParts}, sc, LangFR)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Parts) != 1 || r.Parts[0].ID != "AA6" || r.Truncated {
		t.Fatalf("scoped = %+v", r)
	}

	r, err = s.Search(ctx, SearchQuery{Q: "palier", Kind: SearchParts}, nil, LangFR)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Truncated || r.Total != 1 {
		t.Fatalf("unscoped capped search must report truncation: %+v", r)
	}

	sr, err := s.Search(ctx, SearchQuery{Q: "alternateur", Kind: SearchSections}, sc, LangFR)
	if err != nil {
		t.Fatal(err)
	}
	if len(sr.Sections) != 1 || sr.Sections[0].Sec != "231" {
		t.Fatalf("scoped sections = %+v", sr)
	}
}

func TestReferenceLikeInputAlsoSearchesText(t *testing.T) {
	s := openInternal(t)
	// 23358 is the PNC of the AA5 bearing (part 03902204-0): it looks like a reference but is not one.
	r, err := s.Search(context.Background(), SearchQuery{Q: "23358", Kind: SearchParts}, nil, LangEN)
	if err != nil {
		t.Fatal(err)
	}
	if r.Total != 1 || r.Parts[0].ID != "AA5" {
		t.Fatalf("PNC search = %+v", r)
	}
}
