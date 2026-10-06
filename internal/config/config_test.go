package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"open-niscat/internal/config"
)

func writeConfig(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, config.FileName)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDefaultsWithoutFile(t *testing.T) {
	base := t.TempDir()
	cfg, err := config.Load(nil, base)
	if err != nil {
		t.Fatal(err)
	}
	want := config.Config{Data: filepath.Join(base, "data"), Addr: "127.0.0.1:8080", OpenBrowser: true, DefaultLang: "en"}
	if cfg != want {
		t.Fatalf("got %+v, want %+v", cfg, want)
	}
}

func TestFileNextToBinaryOverridesDefaults(t *testing.T) {
	base := t.TempDir()
	writeConfig(t, base, "data = \"catalog\"\naddr = \"0.0.0.0:9000\"\nopen_browser = false\ndefault_lang = \"fr\"\n")
	cfg, err := config.Load(nil, base)
	if err != nil {
		t.Fatal(err)
	}
	want := config.Config{Data: filepath.Join(base, "catalog"), Addr: "0.0.0.0:9000", OpenBrowser: false, DefaultLang: "fr"}
	if cfg != want {
		t.Fatalf("got %+v, want %+v", cfg, want)
	}
}

func TestFlagsOverrideFile(t *testing.T) {
	base := t.TempDir()
	writeConfig(t, base, "addr = \"0.0.0.0:9000\"\nopen_browser = false\n")
	dataDir := t.TempDir()
	cfg, err := config.Load([]string{"--addr", "127.0.0.1:7000", "--open-browser=true", "--data", dataDir}, base)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != "127.0.0.1:7000" || !cfg.OpenBrowser || cfg.Data != dataDir {
		t.Fatalf("flags not applied: %+v", cfg)
	}
}

func TestExplicitConfigFileAndRelativeData(t *testing.T) {
	base, other := t.TempDir(), t.TempDir()
	path := writeConfig(t, other, "data = \"d\"\n")
	cfg, err := config.Load([]string{"--config", path}, base)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Data != filepath.Join(other, "d") {
		t.Fatalf("data should be relative to the config file, got %s", cfg.Data)
	}
}

func TestErrors(t *testing.T) {
	base := t.TempDir()
	cases := map[string]struct {
		file string
		args []string
		want string
	}{
		"unknown key":          {file: "colour = \"red\"\n", want: "unknown key"},
		"invalid lang in file": {file: "default_lang = \"de\"\n", want: "default_lang"},
		"invalid lang flag":    {args: []string{"--default-lang", "es"}, want: "default_lang"},
		"empty addr":           {file: "addr = \"\"\n", want: "addr"},
		"missing explicit":     {args: []string{"--config", filepath.Join(base, "nope.toml")}, want: "nope.toml"},
		"bad toml":             {file: "addr = \n", want: "open-niscat.toml"},
		"positional argument":  {args: []string{"./data", "--addr", "0.0.0.0:9000"}, want: "unexpected argument \"./data\""},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			if tc.file != "" {
				writeConfig(t, dir, tc.file)
			}
			_, err := config.Load(tc.args, dir)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want error containing %q", err, tc.want)
			}
		})
	}
}
