package catalog_test

import (
	"errors"
	"reflect"
	"testing"

	"open-niscat/internal/catalog"
)

func partIDs(r catalog.SearchResult) []string {
	ids := make([]string, 0, len(r.Parts))
	for i := range r.Parts {
		ids = append(ids, r.Parts[i].ID)
	}
	return ids
}

func TestSearchPartsByText(t *testing.T) {
	s := openFixture(t)
	r, err := s.Search(ctx, catalog.SearchQuery{Q: "palier", Kind: catalog.SearchParts}, nil, catalog.LangFR)
	if err != nil {
		t.Fatal(err)
	}
	if r.Total != 2 || !reflect.DeepEqual(partIDs(r), []string{"AA3", "AA6"}) || r.Parts[0].Sec != "230A" {
		t.Fatalf("result = %+v", r)
	}
	// Prefix match and English fallback for the AB series.
	r, err = s.Search(ctx, catalog.SearchQuery{Q: "cycl", Kind: catalog.SearchParts}, nil, catalog.LangFR)
	if err != nil || r.Total != 2 || r.Parts[0].Etd != "AB" {
		t.Fatalf("fallback = %+v, %v", r, err)
	}
}

func TestSearchPartsByReference(t *testing.T) {
	s := openFixture(t)
	for _, q := range []string{"23319", "-23319-D97", "23319d9700"} {
		r, err := s.Search(ctx, catalog.SearchQuery{Q: q, Kind: catalog.SearchParts}, nil, catalog.LangEN)
		if err != nil || r.Total == 0 || r.Parts[0].PartKey[:5] != "23319" {
			t.Errorf("%q: %+v, %v", q, r, err)
		}
	}
}

func TestSearchPartsScoped(t *testing.T) {
	s := openFixture(t)
	sc, err := s.ResolveScope(ctx, catalog.ScopeParams{VIN: "VSKBEC220U0990494"}) // 101 and 230A applicable
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Search(ctx, catalog.SearchQuery{Q: "palier", Kind: catalog.SearchParts}, sc, catalog.LangFR)
	if err != nil || !reflect.DeepEqual(partIDs(r), []string{"AA3"}) { // AA6 is in 231, not applicable
		t.Fatalf("scoped = %+v, %v", r, err)
	}
}

func TestSearchPagination(t *testing.T) {
	s := openFixture(t)
	r, err := s.Search(ctx, catalog.SearchQuery{Q: "palier", Kind: catalog.SearchParts, Limit: 1, Offset: 1}, nil, catalog.LangFR)
	if err != nil || r.Total != 2 || !reflect.DeepEqual(partIDs(r), []string{"AA6"}) {
		t.Fatalf("page = %+v, %v", r, err)
	}
}

func TestSearchSections(t *testing.T) {
	s := openFixture(t)
	r, err := s.Search(ctx, catalog.SearchQuery{Q: "alternateur", Kind: catalog.SearchSections}, nil, catalog.LangFR)
	if err != nil || r.Total != 3 || r.Sections[0].Sec != "230" || r.Sections[0].Group != "B" {
		t.Fatalf("sections = %+v, %v", r, err)
	}
	r, err = s.Search(ctx, catalog.SearchQuery{Q: "ALTÉRNATEUR fixation", Kind: catalog.SearchSections}, nil, catalog.LangFR)
	if err != nil || r.Total != 2 {
		t.Fatalf("accents + AND = %+v, %v", r, err)
	}
}

func TestSearchHostileInput(t *testing.T) {
	s := openFixture(t)
	for _, q := range []string{`"`, `*`, `-`, `AND`, `(`, `%`, `_`, `") OR 1=1 --`, `NEAR(a b)`, ``} {
		for _, kind := range []catalog.SearchKind{catalog.SearchParts, catalog.SearchSections} {
			if _, err := s.Search(ctx, catalog.SearchQuery{Q: q, Kind: kind}, nil, catalog.LangEN); err != nil {
				t.Errorf("%s %q: %v", kind, q, err)
			}
		}
	}
	if _, err := catalog.ParseSearchKind("bogus"); !errors.Is(err, catalog.ErrInvalid) {
		t.Errorf("bogus kind: %v", err)
	}
}

func TestLines(t *testing.T) {
	s := openFixture(t)
	r, err := s.Lines(ctx, []string{"AA3", "aa5", "AB1", "AA999", "nope", "AA"}, catalog.LangFR)
	if err != nil {
		t.Fatal(err)
	}
	got := make([]string, 0, len(r.Lines))
	for i := range r.Lines {
		got = append(got, r.Lines[i].ID+"/"+r.Lines[i].Description)
	}
	if !reflect.DeepEqual(got, []string{"AA3/PALIER", "AA5/ROULEMENT", "AB1/CYCLE A"}) {
		t.Fatalf("lines = %v", got)
	}
	if !reflect.DeepEqual(r.Missing, []string{"AA999", "nope", "AA"}) {
		t.Fatalf("missing = %v", r.Missing)
	}
	if _, err := s.Lines(ctx, make([]string, catalog.MaxLines+1), catalog.LangEN); !errors.Is(err, catalog.ErrInvalid) {
		t.Fatalf("too many ids: %v", err)
	}
}
