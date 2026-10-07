package catalog_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"open-niscat/internal/catalog"
	"open-niscat/internal/catalog/catalogtest"
)

func openFixture(t *testing.T) *catalog.Store {
	t.Helper()
	s, err := catalog.Open(context.Background(), catalogtest.NewDataDir(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestOpen(t *testing.T) {
	s := openFixture(t)
	m := s.Manifest()
	if m.Version != "2015.01-9" || m.Source.Edition != "2015-01" || m.Source.Name != "NISCAT" || m.Build.Revision != 9 || m.Build.Date != "2026-10-07" {
		t.Fatalf("manifest = %+v", m)
	}
	if !filepath.IsAbs(s.DataDir()) {
		t.Fatalf("DataDir must be absolute: %s", s.DataDir())
	}
}

func TestOpenRejectsBadDataDirs(t *testing.T) {
	cases := map[string]struct {
		prepare func(dir string)
		want    string
	}{
		"no manifest":  {func(dir string) { _ = os.Remove(filepath.Join(dir, "manifest.json")) }, "manifest.json"},
		"other schema": {func(dir string) { writeFile(t, filepath.Join(dir, "manifest.json"), `{"schema":2}`) }, "schema 2"},
		"old format": {func(dir string) {
			writeFile(t, filepath.Join(dir, "manifest.json"), `{"schema":1,"version":"x","edition":"Ed. TEST","built":"2026-10-06"}`)
		}, "rebuild"},
		"no database": {func(dir string) { _ = os.Remove(filepath.Join(dir, "data.db")) }, "data.db"},
		"no vin_rev": {func(dir string) {
			_ = os.Remove(filepath.Join(dir, "data.db"))
			writeFile(t, filepath.Join(dir, "data.db"), "")
		}, "vin_rev"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			dir := catalogtest.NewDataDir(t)
			tc.prepare(dir)
			s, err := catalog.Open(context.Background(), dir)
			if err == nil {
				_ = s.Close()
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error %q does not mention %q", err, tc.want)
			}
		})
	}
}

func TestParseLang(t *testing.T) {
	for in, want := range map[string]catalog.Lang{"en": catalog.LangEN, "fr": catalog.LangFR, "es": catalog.LangES, "de": catalog.LangDE} {
		if l, err := catalog.ParseLang(in); err != nil || l != want {
			t.Errorf("ParseLang(%s) = %q, %v", in, l, err)
		}
	}
	if _, err := catalog.ParseLang("it"); err == nil {
		t.Fatal("it must be rejected: the data has no Italian")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
