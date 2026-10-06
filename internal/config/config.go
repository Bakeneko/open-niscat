// Package config loads open-niscat settings from defaults, an optional TOML file and command-line flags.
package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

// FileName is the configuration file looked up next to the binary.
const FileName = "open-niscat.toml"

// Config holds the effective settings.
type Config struct {
	Data        string // absolute data directory
	Addr        string // listen address
	OpenBrowser bool   // open the default browser on startup
	DefaultLang string // language offered on "/" ("en" or "fr")
}

// Defaults returns the built-in settings (Data is relative to the binary directory).
func Defaults() Config {
	return Config{Data: "data", Addr: "127.0.0.1:8080", OpenBrowser: true, DefaultLang: "en"}
}

type fileConfig struct {
	Data        *string `toml:"data"`
	Addr        *string `toml:"addr"`
	OpenBrowser *bool   `toml:"open_browser"`
	DefaultLang *string `toml:"default_lang"`
}

// Load returns the configuration: defaults, then the TOML file (--config, or FileName in baseDir if present),
// then flags. Relative paths resolve against baseDir (defaults), the file's directory (file) or the working
// directory (flags).
func Load(args []string, baseDir string) (Config, error) {
	flags := flag.NewFlagSet("open-niscat", flag.ContinueOnError)
	configPath := flags.String("config", "", "path to the TOML configuration file")
	data := flags.String("data", "", "data directory")
	addr := flags.String("addr", "", "listen address, e.g. 127.0.0.1:8080")
	openBrowser := flags.Bool("open-browser", true, "open the default browser on startup")
	defaultLang := flags.String("default-lang", "", "language offered on / (en or fr)")
	if err := flags.Parse(args); err != nil {
		return Config{}, fmt.Errorf("parse flags: %w", err)
	}
	if flags.NArg() > 0 {
		return Config{}, fmt.Errorf("unexpected argument %q (use --data to choose the data directory)", flags.Arg(0))
	}
	set := map[string]bool{}
	flags.Visit(func(f *flag.Flag) { set[f.Name] = true })

	cfg := Defaults()
	cfg.Data = filepath.Join(baseDir, cfg.Data)

	path := filepath.Join(baseDir, FileName)
	if set["config"] {
		path = *configPath
	}
	if err := applyFile(&cfg, path, set["config"]); err != nil {
		return Config{}, err
	}

	if set["data"] {
		cfg.Data = *data
	}
	if set["addr"] {
		cfg.Addr = *addr
	}
	if set["open-browser"] {
		cfg.OpenBrowser = *openBrowser
	}
	if set["default-lang"] {
		cfg.DefaultLang = *defaultLang
	}

	abs, err := filepath.Abs(cfg.Data)
	if err != nil {
		return Config{}, fmt.Errorf("resolve data directory: %w", err)
	}
	cfg.Data = abs
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func applyFile(cfg *Config, path string, required bool) error {
	if _, err := os.Stat(path); err != nil {
		if !required && errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("config file %s: %w", path, err)
	}
	var fc fileConfig
	meta, err := toml.DecodeFile(path, &fc)
	if err != nil {
		return fmt.Errorf("config file %s: %w", path, err)
	}
	if undecoded := meta.Undecoded(); len(undecoded) > 0 {
		keys := make([]string, 0, len(undecoded))
		for _, k := range undecoded {
			keys = append(keys, k.String())
		}
		sort.Strings(keys)
		return fmt.Errorf("config file %s: unknown key(s): %s", path, strings.Join(keys, ", "))
	}
	if fc.Data != nil {
		cfg.Data = *fc.Data
		if !filepath.IsAbs(cfg.Data) {
			cfg.Data = filepath.Join(filepath.Dir(path), cfg.Data)
		}
	}
	if fc.Addr != nil {
		cfg.Addr = *fc.Addr
	}
	if fc.OpenBrowser != nil {
		cfg.OpenBrowser = *fc.OpenBrowser
	}
	if fc.DefaultLang != nil {
		cfg.DefaultLang = *fc.DefaultLang
	}
	return nil
}

func (c *Config) validate() error {
	if strings.TrimSpace(c.Addr) == "" {
		return errors.New("addr must not be empty")
	}
	if c.DefaultLang != "en" && c.DefaultLang != "fr" {
		return fmt.Errorf("default_lang must be \"en\" or \"fr\", got %q", c.DefaultLang)
	}
	return nil
}
