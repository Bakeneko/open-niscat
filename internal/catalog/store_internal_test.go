package catalog

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
	"testing"

	"open-niscat/internal/catalog/catalogtest"
)

func openInternal(t *testing.T) *Store {
	t.Helper()
	s, err := Open(context.Background(), catalogtest.NewDataDir(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestDataLangFallsBackToEnglish(t *testing.T) {
	s := openInternal(t)
	if got := s.dataLang("AA", LangFR); got != "fr" {
		t.Errorf("AA fr = %q", got)
	}
	if got := s.dataLang("AB", LangFR); got != "en" {
		t.Errorf("AB fr = %q, want en (English-only series)", got)
	}
	if got := s.dataLang("ZZ", LangFR); got != "en" {
		t.Errorf("unknown series = %q", got)
	}
}

func TestLangFilter(t *testing.T) {
	s := openInternal(t)
	cond, args := s.langFilter("p", LangEN)
	if cond != "p.lang = 'en'" || args != nil {
		t.Errorf("en: %q %v", cond, args)
	}
	cond, args = s.langFilter("p", LangFR)
	if !strings.Contains(cond, "p.etd IN (?)") || !reflect.DeepEqual(args, []any{"fr", "AB"}) {
		t.Errorf("fr: %q %v", cond, args)
	}
}

func TestFileURL(t *testing.T) {
	s := openInternal(t)
	if got := s.fileURL("img", "AA", "AA230A.png"); got != "/files/img/AA/AA230A.png" {
		t.Errorf("existing file = %q", got)
	}
	if got := s.fileURL("img", "AA", "nope.png"); got != "" {
		t.Errorf("missing file = %q", got)
	}
}

func TestSmallHelpers(t *testing.T) {
	if got := placeholders(3); got != "?,?,?" {
		t.Errorf("placeholders(3) = %q", got)
	}
	if got := str(sql.NullString{String: " x ", Valid: true}); got != "x" {
		t.Errorf("str = %q", got)
	}
	if got := str(sql.NullString{}); got != "" {
		t.Errorf("str(NULL) = %q", got)
	}
	if got := dsn(`C:\data\data.db`); got != "file:///C:/data/data.db?mode=ro&_pragma=query_only(1)" && !strings.HasPrefix(got, "file:///") {
		t.Errorf("dsn = %q", got)
	}
}
