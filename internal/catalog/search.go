package catalog

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Search limits.
const (
	defaultLimit = 50
	maxLimit     = 200
	maxMatches   = 2000
	// MaxLines is the maximum number of line ids accepted by Lines.
	MaxLines = 500
)

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
	Total    int          `json:"total"`
	Parts    []LineRef    `json:"parts"`
	Sections []SectionHit `json:"sections"`
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

// Search runs a free-text search; with a scope on a model, only applicable sections are returned.
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
		res.Total = len(hits)
		res.Sections = page(hits, offset, limit)
	case SearchParts:
		hits, err := s.searchParts(ctx, q.Q, sc, applicable, lang)
		if err != nil {
			return SearchResult{}, err
		}
		res.Total = len(hits)
		res.Parts = page(hits, offset, limit)
	}
	return res, nil
}

func page[T any](items []T, offset, limit int) []T {
	if offset >= len(items) {
		return []T{}
	}
	return items[offset:min(offset+limit, len(items))]
}

func (s *Store) searchParts(ctx context.Context, q string, sc *Scope, applicable map[string]bool, lang Lang) ([]LineRef, error) {
	cond, args := s.langFilter("p", lang)
	where := []string{cond}
	if looksLikeRef(q) {
		where = append(where, "p.part_key GLOB ?")
		args = append(args, NormalizeRef(q)+"*")
	} else {
		tokens := words(q)
		if len(tokens) == 0 {
			return []LineRef{}, nil
		}
		expr := make([]string, len(tokens))
		for i, t := range tokens {
			expr[i] = `"` + t + `"*`
		}
		where = append(where, "p.rowid IN (SELECT rowid FROM part_fts WHERE part_fts MATCH ?)")
		args = append(args, strings.Join(expr, " "))
	}
	if sc != nil {
		where = append(where, "p.etd = ?")
		args = append(args, sc.Etd)
	}
	// Only fixed SQL fragments and "?" placeholders are concatenated.
	query := "SELECT " + lineColumns + " FROM part p WHERE " + strings.Join(where, " AND ") + //nolint:gosec // fixed fragments and placeholders only
		" ORDER BY p.etd, p.plate, p.pospie LIMIT " + strconv.Itoa(maxMatches)
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
		if applicable != nil && !applicable[strings.ToUpper(l.Sec)] {
			continue
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
	if sc != nil {
		where = append(where, "s.etd = ?")
		args = append(args, sc.Etd)
	}
	// Only fixed SQL fragments and "?" placeholders are concatenated.
	query := "SELECT s.etd, s.numsec, s.codigrup, s.nomsec, s.notas FROM section s WHERE " + strings.Join(where, " AND ") + //nolint:gosec // fixed fragments and placeholders only
		" ORDER BY s.etd, s.rowid LIMIT " + strconv.Itoa(maxMatches)
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
		h := SectionHit{Etd: str(f[0]), Sec: str(f[1]), Group: str(f[2]), Name: str(f[3]), Notes: str(f[4])}
		if applicable != nil && !applicable[strings.ToUpper(h.Sec)] {
			continue
		}
		out = append(out, h)
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
