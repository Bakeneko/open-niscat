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

func TestPartNotFound(t *testing.T) {
	s := openFixture(t)
	if _, err := s.Part(ctx, "00000-00000", catalog.LangEN); !errors.Is(err, catalog.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
	if _, err := s.Part(ctx, "---", catalog.LangEN); !errors.Is(err, catalog.ErrInvalid) {
		t.Fatalf("err = %v", err)
	}
}
