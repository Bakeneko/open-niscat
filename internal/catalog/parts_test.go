package catalog_test

import (
	"errors"
	"reflect"
	"testing"

	"open-niscat/internal/catalog"
)

func TestPartChainBothWays(t *testing.T) {
	s := openFixture(t)
	p, err := s.Part(ctx, "23319-D9700", catalog.LangFR)
	if err != nil {
		t.Fatal(err)
	}
	if p.Key != "23319D9700" || p.PartNo != "-23319-D9700" || p.Description != "PALIER" {
		t.Fatalf("part = %+v", p)
	}
	if len(p.Occurrences) != 1 || p.Occurrences[0].Etd != "AA" || p.Occurrences[0].Sec != "230A" || p.Occurrences[0].ID != "AA3" {
		t.Fatalf("occurrences = %+v", p.Occurrences)
	}
	if !reflect.DeepEqual(p.Previous, []catalog.RefLink{{Key: "039022040", PartNo: "-03902204-0"}}) {
		t.Fatalf("previous = %+v", p.Previous)
	}
	if !reflect.DeepEqual(p.Next, []catalog.RefLink{{Key: "23319D9799", PartNo: "-23319-D9799"}}) {
		t.Fatalf("next = %+v", p.Next)
	}
}

func TestPartChainStopsOnCycles(t *testing.T) {
	s := openFixture(t)
	p, err := s.Part(ctx, "9000000001", catalog.LangEN)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Next) != 1 || p.Next[0].Key != "9000000002" || len(p.Previous) != 1 || p.Previous[0].Key != "9000000002" {
		t.Fatalf("next=%+v previous=%+v", p.Next, p.Previous)
	}
}

func keys(links []catalog.RefLink) []string {
	out := make([]string, 0, len(links))
	for _, l := range links {
		out = append(out, l.Key)
	}
	return out
}

func TestPartListsEveryRelatedReference(t *testing.T) {
	s := openFixture(t)
	p, err := s.Part(ctx, "90000-00003", catalog.LangEN)
	if err != nil {
		t.Fatal(err)
	}
	// Both alternatives it replaces, and both successors, latest period start first.
	if got := keys(p.Previous); !reflect.DeepEqual(got, []string{"9000000005", "9000000004"}) && !reflect.DeepEqual(got, []string{"9000000004", "9000000005"}) {
		t.Errorf("previous = %v, want both 9000000004 and 9000000005", got)
	}
	if got := keys(p.Next); !reflect.DeepEqual(got, []string{"9000000007", "9000000006"}) {
		t.Errorf("next = %v, want [9000000007 9000000006]", got)
	}
}

func TestPartNotFound(t *testing.T) {
	s := openFixture(t)
	if _, err := s.Part(ctx, "00000-00000", catalog.LangEN); !errors.Is(err, catalog.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
	if _, err := s.Part(ctx, "---", catalog.LangEN); !errors.Is(err, catalog.ErrInvalid) {
		t.Fatalf("err = %v", err)
	}
}
