package catalog_test

import (
	"errors"
	"testing"

	"open-niscat/internal/catalog"
)

func applicability(sections []catalog.SectionSummary) map[string]any {
	out := map[string]any{}
	for _, s := range sections {
		if s.Applicable == nil {
			out[s.Sec] = nil
		} else {
			out[s.Sec] = *s.Applicable
		}
	}
	return out
}

func TestGroups(t *testing.T) {
	s := openFixture(t)
	groups, err := s.Groups(ctx, "AA", catalog.LangFR)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 2 || groups[1].Code != "B" || groups[1].Label != "SYSTEME ELECTRIQUE MOTEUR" || groups[1].Image != "/files/gindex/AA/B.png" {
		t.Fatalf("groups = %+v", groups)
	}
	ab, err := s.Groups(ctx, "AB", catalog.LangFR)
	if err != nil || len(ab) != 1 || ab[0].Label != "ENGINE" || ab[0].Image != "" {
		t.Fatalf("AB groups = %+v, %v", ab, err)
	}
	if _, err := s.Groups(ctx, "ZZ", catalog.LangEN); !errors.Is(err, catalog.ErrNotFound) {
		t.Fatalf("unknown etd: %v", err)
	}
}

func TestGroupDetailWithVIN(t *testing.T) {
	s := openFixture(t)
	sc, err := s.ResolveScope(ctx, catalog.ScopeParams{VIN: "VSKBEC220U0990494"})
	if err != nil {
		t.Fatal(err)
	}
	d, err := s.GroupDetail(ctx, "AA", "G01", "B", sc, catalog.LangFR)
	if err != nil {
		t.Fatal(err)
	}
	// 230: excluded by date (198704-198812), 230A: applicable, 231: other engine.
	got := applicability(d.Sections)
	if len(got) != 3 || got["230"] != false || got["230A"] != true || got["231"] != false {
		t.Fatalf("applicability = %v", got)
	}
	if d.Sections[1].Name != "FIXATION ALTERNATEUR" || d.Sections[1].Notes != "LD20-II" {
		t.Fatalf("section = %+v", d.Sections[1])
	}
	if len(d.Hotspots) != 2 || d.Hotspots[0].Caption != "230" || d.Group.Image != "/files/gindex/AA/B.png" {
		t.Fatalf("detail = %+v", d)
	}
}

func TestGroupDetailScopes(t *testing.T) {
	s := openFixture(t)
	// Model scope without date: the 198704-198812 window of 230 no longer excludes it.
	sc, err := s.ResolveScope(ctx, catalog.ScopeParams{Cat: "AA-G01", Model: "BELC220QSKVX"})
	if err != nil {
		t.Fatal(err)
	}
	d, err := s.GroupDetail(ctx, "AA", "G01", "B", sc, catalog.LangEN)
	if err != nil {
		t.Fatal(err)
	}
	if got := applicability(d.Sections); got["230"] != true || got["230A"] != true || got["231"] != false {
		t.Fatalf("model scope = %v", got)
	}
	// Other engine (c01 = 2), 1993.
	sc, err = s.ResolveScope(ctx, catalog.ScopeParams{VIN: "VSKBEC220U0111111"})
	if err != nil {
		t.Fatal(err)
	}
	d, err = s.GroupDetail(ctx, "AA", "G01", "B", sc, catalog.LangEN)
	if err != nil {
		t.Fatal(err)
	}
	if got := applicability(d.Sections); got["230"] != false || got["230A"] != false || got["231"] != true {
		t.Fatalf("other engine = %v", got)
	}
	// No scope: no applicability information at all.
	d, err = s.GroupDetail(ctx, "AA", "G01", "B", nil, catalog.LangEN)
	if err != nil {
		t.Fatal(err)
	}
	if got := applicability(d.Sections); got["230"] != nil || len(got) != 3 {
		t.Fatalf("no scope = %v", got)
	}
	if _, err := s.GroupDetail(ctx, "AA", "G01", "Z", nil, catalog.LangEN); !errors.Is(err, catalog.ErrNotFound) {
		t.Fatalf("unknown group: %v", err)
	}
}
