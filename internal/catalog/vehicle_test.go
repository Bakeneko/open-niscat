package catalog_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"open-niscat/internal/catalog"
)

var ctx = context.Background()

func TestCatalogs(t *testing.T) {
	s := openFixture(t)
	cats, err := s.Catalogs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(cats) != 3 || cats[0].Cat != "AA-G01" || cats[1].Cat != "AB-G01" || cats[2].Cat != "AA-G02" {
		t.Fatalf("catalogs = %+v", cats)
	}
	if cats[0].Description != "0049 TEST" {
		t.Fatalf("AA = %+v", cats[0])
	}
	if cats[0].From == nil || *cats[0].From != "1987-04" || cats[0].To == nil || *cats[0].To != "1994-11" {
		t.Fatalf("AA dates = %v %v", cats[0].From, cats[0].To)
	}
	if !reflect.DeepEqual(cats[0].Langs, []string{"en", "fr"}) || !reflect.DeepEqual(cats[1].Langs, []string{"en"}) {
		t.Fatalf("langs from data = %v / %v", cats[0].Langs, cats[1].Langs)
	}
	if _, err := s.Catalog(ctx, "ZZ", "G01"); !errors.Is(err, catalog.ErrNotFound) {
		t.Fatalf("unknown catalog: %v", err)
	}
}

func TestIdentifyVINExactWithAttributesInFrench(t *testing.T) {
	s := openFixture(t)
	res, err := s.IdentifyVIN(ctx, " vskbec220u0990494 ", catalog.LangFR)
	if err != nil {
		t.Fatal(err)
	}
	v := res.Vehicle
	if v == nil || res.Candidates != nil {
		t.Fatalf("expected a vehicle, got %+v", res)
	}
	if v.VIN != "VSKBEC220U0990494" || v.Model != "BELC220QSKVX" || v.ProdDate != "1989-05" || v.Catalog.Cat != "AA-G01" {
		t.Fatalf("vehicle = %+v", v)
	}
	want := []catalog.Attribute{
		{Kind: "T", Table: "T01", Name: "MOTEUR", Code: "1", Value: "LD20"},
		{Kind: "T", Table: "T02", Name: "EMPATTEMENT", Code: "1", Value: "COURT"},
		{Kind: "I", Table: "I01", Name: "BODY", Code: "1", Value: "VAN"}, // French missing -> English
	}
	if !reflect.DeepEqual(v.Attributes, want) {
		t.Fatalf("attributes = %+v", v.Attributes)
	}
	if v.VINCount != 2 { // VSKBEC220U0990494 and the malformed 116U0520133
		t.Fatalf("vinCount = %d", v.VINCount)
	}
	if !reflect.DeepEqual(v.Documents, []string{"/files/cinfo/AA/G0101.pdf"}) {
		t.Fatalf("documents = %v", v.Documents)
	}
}

func TestIdentifyVINTail(t *testing.T) {
	s := openFixture(t)
	res, err := s.IdentifyVIN(ctx, "U0520133", catalog.LangEN)
	if err != nil {
		t.Fatal(err)
	}
	want := []catalog.VINMatch{{VIN: "116U0520133", Model: "BELC220QSKVX", Cat: "AA-G01", ProdDate: ptr("1987-07")}}
	if res.Vehicle != nil || !reflect.DeepEqual(res.Candidates, want) {
		t.Fatalf("result = %+v", res)
	}
}

func TestIdentifyVINTailIgnoresDuplicateRows(t *testing.T) {
	s := openFixture(t)
	res, err := s.IdentifyVIN(ctx, "0990494", catalog.LangEN)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Candidates) != 1 {
		t.Fatalf("candidates = %+v, want one (the VIN is stored twice)", res.Candidates)
	}
}

func TestIdentifyVINNotFound(t *testing.T) {
	s := openFixture(t)
	for _, in := range []string{"ZZZZZZZZZ", "0494", "*[?]*"} {
		if _, err := s.IdentifyVIN(ctx, in, catalog.LangEN); !errors.Is(err, catalog.ErrNotFound) {
			t.Errorf("IdentifyVIN(%q) error = %v, want ErrNotFound", in, err)
		}
	}
	if _, err := s.IdentifyVIN(ctx, "   ", catalog.LangEN); !errors.Is(err, catalog.ErrInvalid) {
		t.Errorf("blank VIN must be invalid, got %v", err)
	}
}

func TestEnglishOnlySeriesInFrench(t *testing.T) {
	s := openFixture(t)
	res, err := s.IdentifyVIN(ctx, "SJNVC220R00000001", catalog.LangFR)
	if err != nil {
		t.Fatal(err)
	}
	attrs := res.Vehicle.Attributes
	if len(attrs) != 1 || attrs[0].Name != "ENGINE" || attrs[0].Value != "LD20" {
		t.Fatalf("attributes = %+v", attrs)
	}
}

func TestResolveScope(t *testing.T) {
	s := openFixture(t)
	sc, err := s.ResolveScope(ctx, catalog.ScopeParams{})
	if err != nil || sc != nil {
		t.Fatalf("empty params: %v, %v", sc, err)
	}
	sc, err = s.ResolveScope(ctx, catalog.ScopeParams{Cat: "aa-g01", Model: "BELC220QJKL"})
	if err != nil || sc.Etd != "AA" || sc.Grupo != "G01" || sc.Model != "BELC220QJKL" || sc.ProdDate != 0 || sc.Cat() != "AA-G01" {
		t.Fatalf("model scope = %+v, %v", sc, err)
	}
	sc, err = s.ResolveScope(ctx, catalog.ScopeParams{VIN: "VSKBEC220U0990494"})
	if err != nil || sc.ProdDate != 198905 || sc.Model != "BELC220QSKVX" {
		t.Fatalf("vin scope = %+v, %v", sc, err)
	}
	bad := map[string]catalog.ScopeParams{
		"bad cat":       {Cat: "AAG01"},
		"model w/o cat": {Model: "BELC220QSKVX"},
	}
	for name, p := range bad {
		if _, err := s.ResolveScope(ctx, p); !errors.Is(err, catalog.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
	missing := map[string]catalog.ScopeParams{
		"unknown vin":   {VIN: "NOPE"},
		"unknown cat":   {Cat: "ZZ-G01"},
		"unknown model": {Cat: "AA-G01", Model: "NOPE"},
	}
	for name, p := range missing {
		if _, err := s.ResolveScope(ctx, p); !errors.Is(err, catalog.ErrNotFound) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestModels(t *testing.T) {
	s := openFixture(t)
	models, err := s.Models(ctx, "AA", "G01", catalog.LangEN)
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 2 || models[0].Model != "BELC220QJKL" || models[0].VINCount != 1 || models[1].VINCount != 2 {
		t.Fatalf("models = %+v", models)
	}
	if models[0].Attributes[0].Value != "A15" || models[0].Attributes[1].Value != "LONG" {
		t.Fatalf("attributes = %+v", models[0].Attributes)
	}
}

func ptr(s string) *string { return &s }
