package catalog

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

// Group is a main group of the general index (INDIGRAL).
type Group struct {
	Code  string `json:"code"`
	Label string `json:"label"`
	Image string `json:"image,omitempty"`
}

// Hotspot is a clickable zone of a drawing, in source pixels. Key is the caption without leading zeros.
type Hotspot struct {
	Caption string `json:"caption"`
	Key     string `json:"key"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
	W       int    `json:"w"`
	H       int    `json:"h"`
}

// SectionSummary is a section in a list. Applicable is nil when the scope cannot filter.
type SectionSummary struct {
	Sec        string  `json:"sec"`
	Name       string  `json:"name"`
	Notes      string  `json:"notes"`
	From       *string `json:"from"`
	To         *string `json:"to"`
	Applicable *bool   `json:"applicable,omitempty"`
}

// GroupDetail is a group index: drawing hotspots (captions are section numbers) and sections.
type GroupDetail struct {
	Group    Group            `json:"group"`
	Hotspots []Hotspot        `json:"hotspots"`
	Sections []SectionSummary `json:"sections"`
}

type infosecRow struct {
	secc     string
	from, to int
	cols     [10]string
}

// applies implements the NISCAT rule (INFOSECF): dates when known, '0' and '-' are wildcards.
func applies(r *infosecRow, vals []string, prod int) bool {
	if prod > 0 && (prod < r.from || prod > r.to) {
		return false
	}
	for i, v := range r.cols {
		if v == "" || v == "0" || v == "-" {
			continue
		}
		if i >= len(vals) || v != vals[i] {
			return false
		}
	}
	return true
}

// applicableSections returns the upper-cased section numbers applicable to sc, or nil when sc has no model.
func (s *Store) applicableSections(ctx context.Context, sc *Scope) (map[string]bool, error) {
	if sc == nil || sc.Model == "" {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT secc, desdat, finsdat, c01, c02, c03, c04, c05, c06, c07, c08, c09, c10
		FROM infosec WHERE etd = ? AND grupo = ? AND variant = 'F'`, sc.Etd, sc.Grupo)
	if err != nil {
		return nil, fmt.Errorf("section applicability: %w", err)
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var secc, from, to sql.NullString
		var c [10]sql.NullString
		if err := rows.Scan(&secc, &from, &to, &c[0], &c[1], &c[2], &c[3], &c[4], &c[5], &c[6], &c[7], &c[8], &c[9]); err != nil {
			return nil, fmt.Errorf("section applicability: %w", err)
		}
		r := infosecRow{secc: strings.ToUpper(str(secc)), from: atoiOr(str(from), 0), to: atoiOr(str(to), 999999)}
		for i := range c {
			r.cols[i] = str(c[i])
		}
		if applies(&r, sc.values[:], sc.ProdDate) {
			out[r.secc] = true
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("section applicability: %w", err)
	}
	return out, nil
}

func atoiOr(s string, fallback int) int {
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return fallback
}

// Groups lists the main groups of a series with their index drawing.
func (s *Store) Groups(ctx context.Context, etd string, lang Lang) ([]Group, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT g.cetd, COALESCE(l.label, g.label, g.cetd)
		FROM main_group g
		LEFT JOIN main_group l ON l.etd = g.etd AND l.cetd = g.cetd AND l.lang = ?
		WHERE g.etd = ? AND g.lang = 'en' ORDER BY g.rowid`, string(lang), etd)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	defer rows.Close()
	var out []Group
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.Code, &g.Label); err != nil {
			return nil, fmt.Errorf("list groups: %w", err)
		}
		g.Image = s.fileURL("gindex", etd, g.Code+".png")
		out = append(out, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%w: series %s", ErrNotFound, etd)
	}
	return out, nil
}

func (s *Store) group(ctx context.Context, etd, code string, lang Lang) (Group, error) {
	groups, err := s.Groups(ctx, etd, lang)
	if err != nil {
		return Group{}, err
	}
	for _, g := range groups {
		if strings.EqualFold(g.Code, code) {
			return g, nil
		}
	}
	return Group{}, fmt.Errorf("%w: group %s/%s", ErrNotFound, etd, code)
}

func (s *Store) hotspots(ctx context.Context, etd, kind, image string) ([]Hotspot, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT caption, x, y, w, h FROM hotspot WHERE etd = ? AND kind = ? AND image = ? ORDER BY rowid", etd, kind, image)
	if err != nil {
		return nil, fmt.Errorf("hotspots: %w", err)
	}
	defer rows.Close()
	out := []Hotspot{}
	for rows.Next() {
		var h Hotspot
		var caption sql.NullString
		if err := rows.Scan(&caption, &h.X, &h.Y, &h.W, &h.H); err != nil {
			return nil, fmt.Errorf("hotspots: %w", err)
		}
		h.Caption = str(caption)
		h.Key = ItemKey(h.Caption)
		out = append(out, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("hotspots: %w", err)
	}
	return out, nil
}

// groupSections lists the sections of a group in NISCAT order; with grupo, only sections of that period.
func (s *Store) groupSections(ctx context.Context, etd, grupo, code string, lang Lang) ([]SectionSummary, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT s.numsec, s.nomsec, s.notas, s.desde, s.hasta FROM section s
		WHERE s.etd = ? AND s.lang = ? AND s.codigrup = ?
		  AND (? = '' OR EXISTS (SELECT 1 FROM infosec i
		        WHERE i.variant = 'F' AND i.etd = s.etd AND i.grupo = ? AND i.secc = s.numsec))
		ORDER BY s.rowid`, etd, s.dataLang(etd, lang), code, grupo, grupo)
	if err != nil {
		return nil, fmt.Errorf("group sections: %w", err)
	}
	defer rows.Close()
	out := []SectionSummary{}
	for rows.Next() {
		var f [5]sql.NullString
		if err := rows.Scan(&f[0], &f[1], &f[2], &f[3], &f[4]); err != nil {
			return nil, fmt.Errorf("group sections: %w", err)
		}
		out = append(out, SectionSummary{Sec: str(f[0]), Name: str(f[1]), Notes: str(f[2]), From: yearMonthOf(str(f[3])), To: yearMonthOf(str(f[4]))})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("group sections: %w", err)
	}
	return out, nil
}

// GroupDetail returns a group index. grupo may be "" (no catalog period known). Applicability is set when
// sc has a model; sections of another series than sc are marked not applicable.
func (s *Store) GroupDetail(ctx context.Context, etd, grupo, code string, sc *Scope, lang Lang) (GroupDetail, error) {
	g, err := s.group(ctx, etd, code, lang)
	if err != nil {
		return GroupDetail{}, err
	}
	hotspots, err := s.hotspots(ctx, etd, "group", g.Code)
	if err != nil {
		return GroupDetail{}, err
	}
	sections, err := s.groupSections(ctx, etd, grupo, g.Code, lang)
	if err != nil {
		return GroupDetail{}, err
	}
	if err := s.markApplicable(ctx, etd, sc, sections); err != nil {
		return GroupDetail{}, err
	}
	return GroupDetail{Group: g, Hotspots: hotspots, Sections: sections}, nil
}

func (s *Store) markApplicable(ctx context.Context, etd string, sc *Scope, sections []SectionSummary) error {
	if sc == nil {
		return nil
	}
	if sc.Etd != etd {
		for i := range sections {
			sections[i].Applicable = new(bool)
		}
		return nil
	}
	m, err := s.applicableSections(ctx, sc)
	if err != nil || m == nil {
		return err
	}
	for i := range sections {
		ok := m[strings.ToUpper(sections[i].Sec)]
		sections[i].Applicable = &ok
	}
	return nil
}
