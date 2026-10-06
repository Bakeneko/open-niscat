package catalog_test

import (
	"errors"
	"testing"

	"open-niscat/internal/catalog"
)

func TestSectionWithVIN(t *testing.T) {
	s := openFixture(t)
	sc, err := s.ResolveScope(ctx, catalog.ScopeParams{VIN: "VSKBEC220U0990494"})
	if err != nil {
		t.Fatal(err)
	}
	sec, err := s.Section(ctx, "aa", "230a", sc, catalog.LangFR)
	if err != nil {
		t.Fatal(err)
	}
	if sec.Plate != "AA230A" || sec.Name != "FIXATION ALTERNATEUR" || sec.Group.Code != "B" || sec.Image != "/files/img/AA/AA230A.png" {
		t.Fatalf("section = %+v", sec)
	}
	if sec.Applicable == nil || !*sec.Applicable || sec.Prev != "" || sec.Next != "" {
		t.Fatalf("applicable/prev/next = %v %q %q", sec.Applicable, sec.Prev, sec.Next)
	}
	if sec.From == nil || *sec.From != "1987-04" || sec.To == nil || *sec.To != "1995-12" {
		t.Errorf("section dates = %v %v", sec.From, sec.To)
	}
	if len(sec.Lines) != 3 || len(sec.Hotspots) != 3 {
		t.Fatalf("lines=%d hotspots=%d", len(sec.Lines), len(sec.Hotspots))
	}
	l1, l2, l3 := sec.Lines[0], sec.Lines[1], sec.Lines[2]
	if l1.ID != "AA1" || l1.Description != "TENDEUR" || l1.From == nil || *l1.From != "1987-04" || l1.To == nil || *l1.To != "1987-04" || l1.InPeriod == nil || *l1.InPeriod {
		t.Errorf("line 1 = %+v", l1)
	}
	if l2.Item != "" || l2.ItemKey != "1" || l2.Level != 2 || l2.Cap != "10" || l2.InPeriod == nil || !*l2.InPeriod {
		t.Errorf("line 2 = %+v", l2)
	}
	if l3.Mark != "#" || l3.Alternative != "-03902204-0" || l3.AlternativeKey != "039022040" || l3.ICA != "2-0" {
		t.Errorf("line 3 = %+v", l3)
	}
	if l3.Latest == nil || l3.Latest.Key != "23319D9799" || l3.Latest.PartNo != "-23319-D9799" {
		t.Errorf("line 3 latest = %+v", l3.Latest)
	}
	if sec.Hotspots[1].Key != "2" || sec.Hotspots[2].Key != "2" {
		t.Errorf("hotspots = %+v", sec.Hotspots)
	}
}

func TestSectionWithoutScopeHasNeighbours(t *testing.T) {
	s := openFixture(t)
	sec, err := s.Section(ctx, "AA", "230A", nil, catalog.LangEN)
	if err != nil {
		t.Fatal(err)
	}
	if sec.Prev != "230" || sec.Next != "231" || sec.Applicable != nil || sec.Lines[0].InPeriod != nil {
		t.Fatalf("prev=%q next=%q applicable=%v", sec.Prev, sec.Next, sec.Applicable)
	}
}

func TestSectionScopedToAnotherSeries(t *testing.T) {
	s := openFixture(t)
	sc, err := s.ResolveScope(ctx, catalog.ScopeParams{VIN: "SJNVC220R00000001"}) // AB vehicle
	if err != nil {
		t.Fatal(err)
	}
	sec, err := s.Section(ctx, "AA", "230A", sc, catalog.LangEN)
	if err != nil {
		t.Fatal(err)
	}
	if sec.Applicable == nil || *sec.Applicable || sec.Lines[0].InPeriod != nil {
		t.Fatalf("applicable=%v inPeriod=%v", sec.Applicable, sec.Lines[0].InPeriod)
	}
}

func TestSectionFromTheOtherPeriodOfTheSameSeries(t *testing.T) {
	s := openFixture(t)
	sc, err := s.ResolveScope(ctx, catalog.ScopeParams{VIN: "VSKLATE0000000001"}) // AA-G02, 03/1995
	if err != nil {
		t.Fatal(err)
	}
	sec, err := s.Section(ctx, "AA", "230A", sc, catalog.LangEN) // 230A only exists in AA-G01
	if err != nil {
		t.Fatal(err)
	}
	if sec.Applicable == nil || *sec.Applicable {
		t.Fatalf("applicable = %v, want false", sec.Applicable)
	}
	for _, l := range sec.Lines {
		if l.InPeriod != nil {
			t.Fatalf("line %s has a period warning in another catalog's section", l.ID)
		}
	}
	if sec.Prev != "230" || sec.Next != "231" {
		t.Fatalf("prev=%q next=%q, want the unfiltered neighbours 230/231", sec.Prev, sec.Next)
	}
}

func TestSectionEnglishOnlySeriesInFrench(t *testing.T) {
	s := openFixture(t)
	sec, err := s.Section(ctx, "AB", "040", nil, catalog.LangFR)
	if err != nil {
		t.Fatal(err)
	}
	if sec.Name != "ENGINE" || len(sec.Lines) != 2 || sec.Lines[0].Description != "CYCLE A" {
		t.Fatalf("section = %+v", sec)
	}
}

func TestSectionNotFound(t *testing.T) {
	s := openFixture(t)
	if _, err := s.Section(ctx, "AA", "999", nil, catalog.LangEN); !errors.Is(err, catalog.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}
