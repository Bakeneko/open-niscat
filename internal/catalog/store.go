package catalog

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite" // registers the "sqlite" database/sql driver
)

// SchemaVersion is the data schema this build understands (manifest.json "schema").
const SchemaVersion = 1

var (
	// ErrNotFound reports an unknown VIN, catalog, group, section or part.
	ErrNotFound = errors.New("not found")
	// ErrInvalid reports a malformed argument.
	ErrInvalid = errors.New("invalid argument")
)

// Manifest describes a data directory (data/manifest.json).
type Manifest struct {
	Schema  int    `json:"schema"`
	Version string `json:"version"`
	Edition string `json:"edition"`
	Built   string `json:"built"`
}

// Lang is a supported interface/data language.
type Lang string

// Supported languages.
const (
	LangEN Lang = "en"
	LangFR Lang = "fr"
)

// ParseLang validates a language code.
func ParseLang(s string) (Lang, error) {
	switch l := Lang(s); l {
	case LangEN, LangFR:
		return l, nil
	}
	return "", fmt.Errorf("%w: unsupported language %q", ErrInvalid, s)
}

// Store gives read-only access to a data directory.
type Store struct {
	db       *sql.DB
	dataDir  string
	manifest Manifest
	langs    map[string]map[string]bool // etd -> data languages present
}

// Open checks the data directory (manifest, schema, optimisation pass) and opens data.db read-only.
func Open(ctx context.Context, dataDir string) (*Store, error) {
	abs, err := filepath.Abs(dataDir)
	if err != nil {
		return nil, fmt.Errorf("resolve data directory: %w", err)
	}
	manifest, err := readManifest(abs)
	if err != nil {
		return nil, err
	}
	if manifest.Schema != SchemaVersion {
		return nil, fmt.Errorf("data schema %d is not supported by this build (expected %d)", manifest.Schema, SchemaVersion)
	}
	dbPath := filepath.Join(abs, "data.db")
	if _, err := os.Stat(dbPath); err != nil {
		return nil, fmt.Errorf("open data.db: %w", err)
	}
	db, err := sql.Open("sqlite", dsn(dbPath))
	if err != nil {
		return nil, fmt.Errorf("open data.db: %w", err)
	}
	s := &Store{db: db, dataDir: abs, manifest: manifest}
	if err := s.init(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func readManifest(dir string) (Manifest, error) {
	raw, err := fs.ReadFile(os.DirFS(dir), "manifest.json")
	if err != nil {
		return Manifest{}, fmt.Errorf("read manifest.json in %s: %w", dir, err)
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return Manifest{}, fmt.Errorf("parse manifest.json: %w", err)
	}
	return m, nil
}

// dsn builds a read-only SQLite URI; Windows paths become file:///C:/...
func dsn(dbPath string) string {
	p := filepath.ToSlash(dbPath)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	u := url.URL{Scheme: "file", Path: p, RawQuery: "mode=ro&_pragma=query_only(1)"}
	return u.String()
}

func (s *Store) init(ctx context.Context) error {
	var probe sql.NullString
	err := s.db.QueryRowContext(ctx, "SELECT vin_rev FROM vin LIMIT 1").Scan(&probe)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("data.db lacks the optimisation pass (vin.vin_rev): %w", err)
	}
	rows, err := s.db.QueryContext(ctx, "SELECT DISTINCT etd, lang FROM section")
	if err != nil {
		return fmt.Errorf("load languages: %w", err)
	}
	defer rows.Close()
	s.langs = map[string]map[string]bool{}
	for rows.Next() {
		var etd, lang string
		if err := rows.Scan(&etd, &lang); err != nil {
			return fmt.Errorf("load languages: %w", err)
		}
		if s.langs[etd] == nil {
			s.langs[etd] = map[string]bool{}
		}
		s.langs[etd][lang] = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("load languages: %w", err)
	}
	return nil
}

// Close releases the database.
func (s *Store) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("close data.db: %w", err)
	}
	return nil
}

// Manifest returns the data directory manifest.
func (s *Store) Manifest() Manifest { return s.manifest }

// DataDir returns the absolute data directory.
func (s *Store) DataDir() string { return s.dataDir }

// dataLang returns want if the series has data in that language, else English.
func (s *Store) dataLang(etd string, want Lang) string {
	if s.langs[etd][string(want)] {
		return string(want)
	}
	return string(LangEN)
}

// langFilter returns a condition on alias.lang/alias.etd selecting, per series, want or English when the
// series lacks want, with its arguments.
func (s *Store) langFilter(alias string, want Lang) (cond string, args []any) {
	if want == LangEN {
		return alias + ".lang = 'en'", nil
	}
	missing := make([]string, 0, len(s.langs))
	for etd, langs := range s.langs {
		if !langs[string(want)] {
			missing = append(missing, etd)
		}
	}
	if len(missing) == 0 {
		return alias + ".lang = ?", []any{string(want)}
	}
	sort.Strings(missing)
	args = make([]any, 0, len(missing)+1)
	args = append(args, string(want))
	for _, etd := range missing {
		args = append(args, etd)
	}
	cond = fmt.Sprintf("(%[1]s.lang = ? OR (%[1]s.lang = 'en' AND %[1]s.etd IN (%[2]s)))", alias, placeholders(len(missing)))
	return cond, args
}

// fileURL returns the public URL of a data file, or "" if it does not exist.
func (s *Store) fileURL(kind, etd, name string) string {
	rel := path.Join(kind, etd, name)
	if _, err := fs.Stat(os.DirFS(s.dataDir), rel); err != nil {
		return ""
	}
	return "/files/" + rel
}

func placeholders(n int) string {
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}

func str(v sql.NullString) string {
	return strings.TrimSpace(v.String)
}
