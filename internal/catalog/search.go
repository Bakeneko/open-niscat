package catalog

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// Search limits.
const (
	defaultLimit = 50
	maxLimit     = 200
	// MaxLines is the maximum number of line ids accepted by Lines.
	MaxLines = 500
)

// maxMatches caps the rows a search reads; a variable so tests can lower it.
var maxMatches = 2000

// SearchKind selects what Search looks for.
type SearchKind string

// Search kinds.
const (
	SearchParts    SearchKind = "parts"
	SearchSections SearchKind = "sections"
)

// ParseSearchKind validates a search kind; "" means parts.
func ParseSearchKind(s string) (SearchKind, error) {
	switch k := SearchKind(s); k {
	case SearchParts, SearchSections:
		return k, nil
	case "":
		return SearchParts, nil
	}
	return "", fmt.Errorf("%w: search type %q", ErrInvalid, s)
}

// SearchQuery is a free-text search.
type SearchQuery struct {
	Q      string
	Kind   SearchKind
	Limit  int
	Offset int
}

// SectionHit is a section search result.
type SectionHit struct {
	Etd   string `json:"etd"`
	Sec   string `json:"sec"`
	Group string `json:"group"`
	Name  string `json:"name"`
	Notes string `json:"notes"`
}

// SearchResult is one page of results; Total counts all matches (capped at maxMatches).
type SearchResult struct {
	Total     int          `json:"total"`
	Truncated bool         `json:"truncated"`
	Parts     []LineRef    `json:"parts"`
	Sections  []SectionHit `json:"sections"`
}

// LinesResult resolves cart line ids; unknown or malformed ids are listed in Missing.
type LinesResult struct {
	Lines   []LineRef `json:"lines"`
	Missing []string  `json:"missing"`
}

var accents = strings.NewReplacer(
	"à", "a", "â", "a", "ä", "a", "é", "e", "è", "e", "ê", "e", "ë", "e", "î", "i", "ï", "i",
	"ô", "o", "ö", "o", "ù", "u", "û", "u", "ü", "u", "ç", "c",
	"À", "A", "Â", "A", "Ä", "A", "É", "E", "È", "E", "Ê", "E", "Ë", "E", "Î", "I", "Ï", "I",
	"Ô", "O", "Ö", "O", "Ù", "U", "Û", "U", "Ü", "U", "Ç", "C",
)

// words splits a query into accent-free letter/digit tokens.
func words(q string) []string {
	return strings.FieldsFunc(accents.Replace(q), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
}

// looksLikeRef reports whether q should be searched as a part reference prefix.
func looksLikeRef(q string) bool {
	q = strings.TrimSpace(q)
	key := NormalizeRef(q)
	return !strings.ContainsAny(q, " \t") && len(key) >= 5 && strings.ContainsAny(key, "0123456789")
}

// Search runs a free-text search; with a scope on a model, only applicable sections are searched.
// At most maxMatches rows are read; Truncated reports that more exist.
func (s *Store) Search(ctx context.Context, q SearchQuery, sc *Scope, lang Lang) (SearchResult, error) {
	limit, offset := q.Limit, max(q.Offset, 0)
	if limit <= 0 {
		limit = defaultLimit
	}
	limit = min(limit, maxLimit)
	applicable, err := s.applicableSections(ctx, sc)
	if err != nil {
		return SearchResult{}, err
	}
	res := SearchResult{Parts: []LineRef{}, Sections: []SectionHit{}}
	switch q.Kind {
	case SearchSections:
		hits, err := s.searchSections(ctx, q.Q, sc, applicable, lang)
		if err != nil {
			return SearchResult{}, err
		}
		hits, res.Truncated = capped(hits)
		res.Total = len(hits)
		res.Sections = page(hits, offset, limit)
	case SearchParts:
		hits, err := s.searchParts(ctx, q.Q, sc, applicable, lang)
		if err != nil {
			return SearchResult{}, err
		}
		hits, res.Truncated = capped(hits)
		res.Total = len(hits)
		res.Parts = page(hits, offset, limit)
	}
	return res, nil
}

// capped trims a result read with LIMIT maxMatches+1 and reports whether it was cut.
func capped[T any](items []T) ([]T, bool) {
	if len(items) > maxMatches {
		return items[:maxMatches], true
	}
	return items, false
}

func page[T any](items []T, offset, limit int) []T {
	if offset >= len(items) {
		return []T{}
	}
	return items[offset:min(offset+limit, len(items))]
}

// scopeFilter restricts a query to the scope's series and, when known, to its applicable sections, before any
// LIMIT applies. column is the plate or section-number column; prefixEtd builds plates ("AA" + "230A").
func scopeFilter(sc *Scope, applicable map[string]bool, etdColumn, column string, prefixEtd bool) (conds []string, args []any) {
	if sc == nil {
		return nil, nil
	}
	conds, args = []string{etdColumn + " = ?"}, []any{sc.Etd}
	if applicable == nil {
		return conds, args
	}
	secs := make([]string, 0, len(applicable))
	for sec := range applicable {
		secs = append(secs, sec)
	}
	sort.Strings(secs)
	if len(secs) == 0 {
		return append(conds, "0"), args // nothing applies
	}
	for _, sec := range secs {
		if prefixEtd {
			sec = sc.Etd + sec
		}
		args = append(args, sec)
	}
	return append(conds, column+" IN ("+placeholders(len(secs))+")"), args
}

func ftsExpression(q string) string {
	tokens := words(q)
	expr := make([]string, len(tokens))
	for i, t := range tokens {
		expr[i] = `"` + t + `"*`
	}
	return strings.Join(expr, " ")
}

func (s *Store) searchParts(ctx context.Context, q string, sc *Scope, applicable map[string]bool, lang Lang) ([]LineRef, error) {
	cond, args := s.langFilter("p", lang)
	where := []string{cond}
	expr := ftsExpression(q)
	const fts = "p.rowid IN (SELECT rowid FROM part_fts WHERE part_fts MATCH ?)"
	switch {
	case looksLikeRef(q) && expr != "":
		where = append(where, "(p.part_key GLOB ? OR "+fts+")")
		args = append(args, NormalizeRef(q)+"*", expr)
	case expr != "":
		where = append(where, fts)
		args = append(args, expr)
	default:
		return []LineRef{}, nil
	}
	scopeConds, scopeArgs := scopeFilter(sc, applicable, "p.etd", "p.plate", true)
	where = append(where, scopeConds...)
	args = append(args, scopeArgs...)
	query := "SELECT " + lineColumns + " FROM part p WHERE " + strings.Join(where, " AND ") + //nolint:gosec // fixed fragments and placeholders only
		" ORDER BY p.etd, p.plate, p.pospie LIMIT " + strconv.Itoa(maxMatches+1)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search parts: %w", err)
	}
	defer rows.Close()
	out := []LineRef{}
	for rows.Next() {
		l, err := scanLine(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("search parts: %w", err)
	}
	return out, nil
}

func (s *Store) searchSections(ctx context.Context, q string, sc *Scope, applicable map[string]bool, lang Lang) ([]SectionHit, error) {
	tokens := words(q)
	if len(tokens) == 0 {
		return []SectionHit{}, nil
	}
	cond, args := s.langFilter("s", lang)
	where := []string{cond}
	escape := strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`)
	for _, t := range tokens {
		like := "%" + escape.Replace(t) + "%"
		where = append(where, `(s.nomsec LIKE ? ESCAPE '\' OR s.notas LIKE ? ESCAPE '\' OR s.numsec LIKE ? ESCAPE '\')`)
		args = append(args, like, like, like)
	}
	scopeConds, scopeArgs := scopeFilter(sc, applicable, "s.etd", "s.numsec", false)
	where = append(where, scopeConds...)
	args = append(args, scopeArgs...)
	query := "SELECT s.etd, s.numsec, s.codigrup, s.nomsec, s.notas FROM section s WHERE " + strings.Join(where, " AND ") + //nolint:gosec // fixed fragments and placeholders only
		" ORDER BY s.etd, s.rowid LIMIT " + strconv.Itoa(maxMatches+1)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search sections: %w", err)
	}
	defer rows.Close()
	out := []SectionHit{}
	for rows.Next() {
		var f [5]sql.NullString
		if err := rows.Scan(&f[0], &f[1], &f[2], &f[3], &f[4]); err != nil {
			return nil, fmt.Errorf("search sections: %w", err)
		}
		out = append(out, SectionHit{Etd: str(f[0]), Sec: str(f[1]), Group: str(f[2]), Name: str(f[3]), Notes: str(f[4])})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("search sections: %w", err)
	}
	return out, nil
}

// parseLineID splits "AA2605" into "AA" and 2605.
func parseLineID(id string) (etd string, pospie int, ok bool) {
	id = strings.ToUpper(strings.TrimSpace(id))
	if len(id) < 3 || id[0] < 'A' || id[0] > 'Z' || id[1] < 'A' || id[1] > 'Z' {
		return "", 0, false
	}
	n, err := strconv.Atoi(id[2:])
	if err != nil || n <= 0 {
		return "", 0, false
	}
	return id[:2], n, true
}

// Lines resolves cart line ids, keeping the requested order.
func (s *Store) Lines(ctx context.Context, ids []string, lang Lang) (LinesResult, error) {
	if len(ids) > MaxLines {
		return LinesResult{}, fmt.Errorf("%w: at most %d lines", ErrInvalid, MaxLines)
	}
	res := LinesResult{Lines: []LineRef{}, Missing: []string{}}
	for _, id := range ids {
		etd, pospie, ok := parseLineID(id)
		if !ok {
			res.Missing = append(res.Missing, id)
			continue
		}
		l, found, err := s.line(ctx, etd, pospie, lang)
		if err != nil {
			return LinesResult{}, err
		}
		if !found {
			res.Missing = append(res.Missing, id)
			continue
		}
		res.Lines = append(res.Lines, l)
	}
	return res, nil
}

func (s *Store) line(ctx context.Context, etd string, pospie int, lang Lang) (LineRef, bool, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+lineColumns+" FROM part p WHERE p.etd = ? AND p.pospie = ? AND p.lang = ? LIMIT 1",
		etd, pospie, s.dataLang(etd, lang))
	if err != nil {
		return LineRef{}, false, fmt.Errorf("look up line: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return LineRef{}, false, fmt.Errorf("look up line: %w", err)
		}
		return LineRef{}, false, nil
	}
	l, err := scanLine(rows)
	if err != nil {
		return LineRef{}, false, err
	}
	return l, true, nil
}
