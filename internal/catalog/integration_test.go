package catalog_test

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"open-niscat/internal/catalog"
)

// These tests pin what was checked in the real NISCAT application. They need the licensed data in
// data/data.db and skip when it is absent, so the repository never contains Nissan data.

func openReal(t *testing.T) *catalog.Store {
	t.Helper()
	dir := filepath.Join("..", "..", "data")
	if _, err := os.Stat(filepath.Join(dir, "data.db")); err != nil {
		t.Skip("real data not available (data/data.db)")
	}
	s, err := catalog.Open(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func values(attrs []catalog.Attribute) []string {
	out := make([]string, len(attrs))
	for i, a := range attrs {
		out[i] = a.Value
	}
	return out
}

func TestRealVanetteIdentification(t *testing.T) {
	s := openReal(t)
	res, err := s.IdentifyVIN(ctx, "VSKBEC220U0990494", catalog.LangEN)
	if err != nil {
		t.Fatal(err)
	}
	v := res.Vehicle
	if v.Catalog.Cat != "AA-G01" || v.Model != "BELC220QSKVX" || v.ProdDate != "1989-05" || v.VINCount != 8047 {
		t.Fatalf("vehicle = %+v", v)
	}
	want := []string{"LD20", "SHORT", "5 DOOR", "STANDARD", "5 SPEED", "VAN", "VANETTE C220", "VAN", "SPAIN"}
	if got := values(v.Attributes); !reflect.DeepEqual(got, want) {
		t.Fatalf("attributes = %v", got)
	}
}

func TestRealPatrolAndCabstarIdentification(t *testing.T) {
	s := openReal(t)
	cases := []struct {
		vin, cat, model, date string
		first                 []string
	}{
		{"VSKAVU260U0618518", "AC-G02", "AVPULQF260TPAA---ASPA", "2000-11", []string{"HIGH ROOF VAN", "TD27T", "LONG WHEELBASE"}},
		{"VWASBFTL01A142282", "AL-G01", "SBC3LQFTL0CQG8Q6-3ITA", "2001-05", []string{"FIX CAB", "BD30TI", "3500KG", "EUROPE", "GENERAL", "LONG (3400MM)"}},
	}
	for _, tc := range cases {
		res, err := s.IdentifyVIN(ctx, tc.vin, catalog.LangEN)
		if err != nil {
			t.Fatal(err)
		}
		v := res.Vehicle
		if v.Catalog.Cat != tc.cat || v.Model != tc.model || v.ProdDate != tc.date {
			t.Errorf("%s = %+v", tc.vin, v)
		}
		if got := values(v.Attributes)[:len(tc.first)]; !reflect.DeepEqual(got, tc.first) {
			t.Errorf("%s attributes = %v", tc.vin, got)
		}
	}
}

func TestRealVanetteGroupB(t *testing.T) {
	s := openReal(t)
	sc, err := s.ResolveScope(ctx, catalog.ScopeParams{VIN: "VSKBEC220U0990494"})
	if err != nil {
		t.Fatal(err)
	}
	d, err := s.GroupDetail(ctx, "AA", "G01", "B", sc, catalog.LangFR)
	if err != nil {
		t.Fatal(err)
	}
	var shown, hidden []string
	for _, sec := range d.Sections {
		if *sec.Applicable {
			shown = append(shown, sec.Sec)
		} else {
			hidden = append(hidden, sec.Sec)
		}
	}
	sort.Strings(shown)
	sort.Strings(hidden)
	if !reflect.DeepEqual(shown, []string{"230A", "230B", "231", "231A", "233C", "233D"}) {
		t.Errorf("shown = %v", shown)
	}
	if !reflect.DeepEqual(hidden, []string{"221", "230", "231B", "233", "233A", "233B"}) {
		t.Errorf("hidden = %v", hidden)
	}
}

func TestRealStarterBearingLine(t *testing.T) {
	s := openReal(t)
	sec, err := s.Section(ctx, "AA", "233C", nil, catalog.LangFR)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range sec.Lines {
		if l.Item == "04" && l.Variant == "01" {
			if l.PartNo != "-23319-D9700" || l.Alternative != "-03902204-0" || l.ICA != "2-0" || l.Qty != "2" || l.Mark != "#" {
				t.Fatalf("line = %+v", l)
			}
			return
		}
	}
	t.Fatal("item 04/01 not found in AA233C")
}

func TestRealCatalogIDsAreUnique(t *testing.T) {
	s := openReal(t)
	cats, err := s.Catalogs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, c := range cats {
		if seen[c.Cat] {
			t.Errorf("duplicate catalog id %s", c.Cat)
		}
		seen[c.Cat] = true
	}
	if len(cats) != 22 {
		t.Errorf("catalogs = %d, want 22", len(cats))
	}
}
