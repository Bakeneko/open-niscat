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
	if got := s.Manifest().Edition; got != "Ed. TEST" {
		t.Fatalf("edition = %q", got)
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
		"no database":  {func(dir string) { _ = os.Remove(filepath.Join(dir, "data.db")) }, "data.db"},
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
	if l, err := catalog.ParseLang("fr"); err != nil || l != catalog.LangFR {
		t.Fatalf("ParseLang(fr) = %q, %v", l, err)
	}
	if _, err := catalog.ParseLang("de"); err == nil {
		t.Fatal("de must be rejected")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
