package catalog

import (
	"context"
	"database/sql"
	"reflect"
	"runtime"
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
}

func TestDSN(t *testing.T) {
	const query = "?mode=ro&_pragma=query_only(1)"
	cases := map[string]string{
		"/srv/data/data.db":  "file:///srv/data/data.db" + query,
		"/srv/my data/#1.db": "file:///srv/my%20data/%231.db" + query,
		"/srv/50%/data.db":   "file:///srv/50%25/data.db" + query,
	}
	if runtime.GOOS == "windows" {
		cases[`C:\data\data.db`] = "file:///C:/data/data.db" + query
	}
	for in, want := range cases {
		if got := dsn(in); got != want {
			t.Errorf("dsn(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestStoreIsReadOnly(t *testing.T) {
	s := openInternal(t)
	ctx := context.Background()
	if _, err := s.db.ExecContext(ctx, "INSERT INTO cinfo VALUES ('x', 'y')"); err == nil {
		t.Fatal("INSERT succeeded on a store that must be read-only")
	}
	if _, err := s.db.ExecContext(ctx, "PRAGMA query_only = 0"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.ExecContext(ctx, "INSERT INTO cinfo VALUES ('x', 'y')"); err == nil {
		t.Fatal("INSERT succeeded after disabling query_only: the file is not opened read-only")
	}
}

func TestAvailableLangs(t *testing.T) {
	s := openInternal(t)
	if got := s.availableLangs("AA"); !reflect.DeepEqual(got, []string{"en", "fr"}) {
		t.Errorf("AA = %v", got)
	}
	if got := s.availableLangs("AB"); !reflect.DeepEqual(got, []string{"en"}) {
		t.Errorf("AB = %v", got)
	}
}
