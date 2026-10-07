package catalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// Line is a parts list line (one row of <lang>_PAR).
type Line struct {
	ID             string   `json:"id"`
	Pospie         int      `json:"pospie"`
	Mark           string   `json:"mark"`
	Item           string   `json:"item"`
	ItemKey        string   `json:"itemKey"`
	Variant        string   `json:"variant"`
	Callout        string   `json:"callout"`
	Level          int      `json:"level"`
	PartNo         string   `json:"partNo"`
	PartKey        string   `json:"partKey"`
	Description    string   `json:"description"`
	Spec           string   `json:"spec"`
	Qty            string   `json:"qty"`
	Cap            string   `json:"cap"`
	ICA            string   `json:"ica"`
	App            string   `json:"app"`
	From           *string  `json:"from"`
	To             *string  `json:"to"`
	InPeriod       *bool    `json:"inPeriod,omitempty"`
	Alternative    string   `json:"alternative"`
	AlternativeKey string   `json:"alternativeKey"`
	KD             string   `json:"kd"`
	PNC            string   `json:"pnc"`
	Latest         *RefLink `json:"latest,omitempty"`
	period         *Period
}

// LineRef is a line with its series and section, for lists spanning sections.
type LineRef struct {
	Etd string `json:"etd"`
	Sec string `json:"sec"`
	Line
}

// Section is a parts plate: drawing, hotspots, lines and neighbours in its group.
type Section struct {
	Etd        string    `json:"etd"`
	Sec        string    `json:"sec"`
	Plate      string    `json:"plate"`
	Group      Group     `json:"group"`
	Name       string    `json:"name"`
	NameEN     string    `json:"nameEn,omitempty"`
	Notes      string    `json:"notes"`
	From       *string   `json:"from"`
	To         *string   `json:"to"`
	Applicable *bool     `json:"applicable,omitempty"`
	Image      string    `json:"image"`
	Hotspots   []Hotspot `json:"hotspots"`
	Lines      []Line    `json:"lines"`
	Prev       string    `json:"prev,omitempty"`
	Next       string    `json:"next,omitempty"`
}

const lineColumns = "p.etd, p.pospie, p.plate, p.s, p.item, p.item_eff, p.sec, p.ind1, p.ind2, p.ind3, p.ind4, p.ind5, " +
	"p.part_no, p.part_key, p.des, p.esp, p.qty, p.cap, p.ica, p.app, p.dataplic, p.ree, p.ree_key, p.kdf, p.pnc"

func scanLine(rows *sql.Rows) (LineRef, error) {
	var etd, plate, mark, item, itemEff, variant sql.NullString
	var ind [5]sql.NullString
	var f [13]sql.NullString
	var pospie int
	err := rows.Scan(&etd, &pospie, &plate, &mark, &item, &itemEff, &variant, &ind[0], &ind[1], &ind[2], &ind[3], &ind[4],
		&f[0], &f[1], &f[2], &f[3], &f[4], &f[5], &f[6], &f[7], &f[8], &f[9], &f[10], &f[11], &f[12])
	if err != nil {
		return LineRef{}, fmt.Errorf("scan line: %w", err)
	}
	l := Line{
		ID: fmt.Sprintf("%s%d", str(etd), pospie), Pospie: pospie, Mark: str(mark), Item: str(item),
		ItemKey: ItemKey(str(itemEff)), Variant: str(variant), Callout: callout(str(itemEff), str(variant)),
		PartNo: str(f[0]), PartKey: str(f[1]), Description: str(f[2]), Spec: str(f[3]), Qty: str(f[4]), Cap: str(f[5]),
		ICA: str(f[6]), App: str(f[7]), Alternative: str(f[9]), AlternativeKey: str(f[10]),
		KD: str(f[11]), PNC: str(f[12]),
	}
	for i := range ind {
		if str(ind[i]) == "-" {
			l.Level = i + 1
			break
		}
	}
	if p, ok := ParsePeriod(str(f[8])); ok {
		l.From, l.To, l.period = yearMonth(p.From), yearMonth(p.To), &p
	}
	return LineRef{Etd: str(etd), Sec: strings.TrimPrefix(str(plate), str(etd)), Line: l}, nil
}

// callout labels a line "01-02": the drawing number, which NISCAT prints only on an item's first line (item)
// but keeps on every line (item_eff), then the line's number within the item. A dash, not a slash, keeps it
// apart from MM/YY periods.
func callout(itemEff, variant string) string {
	if variant == "" {
		return itemEff
	}
	return itemEff + "-" + variant
}

// Section returns a plate. With a scope on the same catalog (series and period): applicability, period checks
// and neighbours restricted to applicable sections. With a scope on another catalog (another series, or a section
// absent from the scope's period): not applicable, no period checks, unfiltered neighbours.
func (s *Store) Section(ctx context.Context, etd, sec string, sc *Scope, lang Lang) (Section, error) {
	etd, sec = strings.ToUpper(etd), strings.ToUpper(sec)
	var code, name, notes, from, to, en sql.NullString
	err := s.db.QueryRowContext(ctx,
		"SELECT s.codigrup, s.nomsec, s.notas, s.desde, s.hasta, "+englishName+
			" FROM section s WHERE s.etd = ? AND s.lang = ? AND s.numsec = ?",
		etd, s.dataLang(etd, lang), sec).Scan(&code, &name, &notes, &from, &to, &en)
	if errors.Is(err, sql.ErrNoRows) {
		return Section{}, fmt.Errorf("%w: section %s/%s", ErrNotFound, etd, sec)
	}
	if err != nil {
		return Section{}, fmt.Errorf("look up section: %w", err)
	}
	out := Section{
		Etd: etd, Sec: sec, Plate: etd + sec, Name: str(name), NameEN: nameEN(str(name), en), Notes: str(notes), From: yearMonthOf(str(from)), To: yearMonthOf(str(to)),
		Image: s.fileURL("img", etd, etd+sec+".png"),
	}
	if out.Group, err = s.group(ctx, etd, str(code), lang); err != nil {
		return Section{}, err
	}
	if out.Hotspots, err = s.hotspots(ctx, etd, "plate", out.Plate); err != nil {
		return Section{}, err
	}
	if out.Lines, err = s.plateLines(ctx, etd, out.Plate, lang); err != nil {
		return Section{}, err
	}

	sameCatalog := false
	if sc != nil && sc.Etd == etd {
		if sameCatalog, err = s.inGrupo(ctx, etd, sc.Grupo, sec); err != nil {
			return Section{}, err
		}
	}
	var applicable map[string]bool
	if sameCatalog {
		if applicable, err = s.applicableSections(ctx, sc); err != nil {
			return Section{}, err
		}
	}
	switch {
	case sc != nil && !sameCatalog:
		out.Applicable = new(bool)
	case applicable != nil:
		ok := applicable[sec]
		out.Applicable = &ok
	}
	if sameCatalog && sc.ProdDate > 0 {
		for i := range out.Lines {
			if p := out.Lines[i].period; p != nil {
				in := p.Contains(sc.ProdDate)
				out.Lines[i].InPeriod = &in
			}
		}
	}

	grupo := ""
	if sameCatalog {
		grupo = sc.Grupo
	}
	if out.Prev, out.Next, err = s.neighbours(ctx, etd, grupo, out.Group.Code, sec, applicable, lang); err != nil {
		return Section{}, err
	}
	return out, nil
}

// inGrupo reports whether a section belongs to the given period of its series.
func (s *Store) inGrupo(ctx context.Context, etd, grupo, sec string) (bool, error) {
	var one int
	err := s.db.QueryRowContext(ctx, "SELECT 1 FROM infosec WHERE variant = 'F' AND etd = ? AND grupo = ? AND secc = ? LIMIT 1",
		etd, grupo, sec).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("section period: %w", err)
	}
	return true, nil
}

func (s *Store) plateLines(ctx context.Context, etd, plate string, lang Lang) ([]Line, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+lineColumns+" FROM part p WHERE p.etd = ? AND p.plate = ? AND p.lang = ? ORDER BY p.pospie",
		etd, plate, s.dataLang(etd, lang))
	if err != nil {
		return nil, fmt.Errorf("plate lines: %w", err)
	}
	defer rows.Close()
	out := []Line{}
	latest := map[string]*RefLink{}
	for rows.Next() {
		ref, err := scanLine(rows)
		if err != nil {
			return nil, err
		}
		l := ref.Line
		if l.PartKey != "" {
			cached, ok := latest[l.PartKey]
			if !ok {
				if cached, err = s.latest(ctx, l.PartKey); err != nil {
					return nil, err
				}
				latest[l.PartKey] = cached
			}
			l.Latest = cached
		}
		out = append(out, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("plate lines: %w", err)
	}
	return out, nil
}

// neighbours returns the previous and next sections of sec in its group, skipping non-applicable ones.
func (s *Store) neighbours(ctx context.Context, etd, grupo, code, sec string, applicable map[string]bool, lang Lang) (prev, next string, err error) {
	list, err := s.groupSections(ctx, etd, grupo, code, lang)
	if err != nil {
		return "", "", err
	}
	idx := -1
	for i := range list {
		if strings.EqualFold(list[i].Sec, sec) {
			idx = i
			break
		}
	}
	if idx < 0 {
		return "", "", nil
	}
	usable := func(i int) bool { return applicable == nil || applicable[strings.ToUpper(list[i].Sec)] }
	for i := idx - 1; i >= 0; i-- {
		if usable(i) {
			prev = list[i].Sec
			break
		}
	}
	for i := idx + 1; i < len(list); i++ {
		if usable(i) {
			next = list[i].Sec
			break
		}
	}
	return prev, next, nil
}
