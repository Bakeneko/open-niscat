package catalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// minVINTail is the shortest input accepted for a search on the end of the VIN.
const minVINTail = 6

// maxCandidates caps the candidates of an end-of-VIN identification.
const maxCandidates = 20

// Info is one catalogue entry (niscat.mdb SERIES).
type Info struct {
	Cat         string   `json:"cat"`
	Etd         string   `json:"etd"`
	Grupo       string   `json:"grupo"`
	Model       string   `json:"model"`
	CModel      string   `json:"cmodel"`
	Drive       string   `json:"drive"`
	From        *string  `json:"from"`
	To          *string  `json:"to"`
	Serie       string   `json:"serie"`
	Description string   `json:"description"`
	Langs       []string `json:"langs"`
}

// Attribute is one vehicle characteristic (csel table: kind T, cinf table: kind I).
type Attribute struct {
	Kind  string `json:"kind"`
	Table string `json:"table"`
	Name  string `json:"name"`
	Code  string `json:"code"`
	Value string `json:"value"`
}

// Vehicle describes a scoped vehicle.
type Vehicle struct {
	VIN        string      `json:"vin,omitempty"`
	Model      string      `json:"model,omitempty"`
	ProdDate   string      `json:"prodDate,omitempty"`
	VINCount   int         `json:"vinCount,omitempty"`
	Catalog    Info        `json:"catalog"`
	Attributes []Attribute `json:"attributes"`
	Documents  []string    `json:"documents"`
}

// VINMatch is a candidate returned by a search on the end of the VIN.
type VINMatch struct {
	VIN      string  `json:"vin"`
	Model    string  `json:"model"`
	Cat      string  `json:"cat"`
	ProdDate *string `json:"prodDate"`
}

// VINResult holds either the identified vehicle or candidates.
type VINResult struct {
	Vehicle    *Vehicle   `json:"vehicle,omitempty"`
	Candidates []VINMatch `json:"candidates,omitempty"`
}

// ModelInfo is a model code of a catalog with its attributes.
type ModelInfo struct {
	Model      string      `json:"model"`
	VINCount   int         `json:"vinCount"`
	Attributes []Attribute `json:"attributes"`
}

const catalogColumns = "etd, grupo, model, cmodel, drive, date_from, date_to, serie, data"

type scanner interface{ Scan(dest ...any) error }

func (s *Store) scanCatalog(row scanner) (Info, error) {
	var f [9]sql.NullString
	if err := row.Scan(&f[0], &f[1], &f[2], &f[3], &f[4], &f[5], &f[6], &f[7], &f[8]); err != nil {
		return Info{}, fmt.Errorf("scan catalog: %w", err)
	}
	info := Info{
		Etd: str(f[0]), Grupo: str(f[1]), Model: str(f[2]), CModel: str(f[3]), Drive: str(f[4]),
		From: yearMonthOf(str(f[5])), To: yearMonthOf(str(f[6])), Serie: str(f[7]), Description: str(f[8]),
	}
	info.Cat = info.Etd + "-" + info.Grupo
	info.Langs = s.availableLangs(info.Etd)
	return info, nil
}

// Catalogs lists all catalogues in NISCAT order.
func (s *Store) Catalogs(ctx context.Context) ([]Info, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+catalogColumns+" FROM catalog ORDER BY orden")
	if err != nil {
		return nil, fmt.Errorf("list catalogs: %w", err)
	}
	defer rows.Close()
	out := []Info{}
	for rows.Next() {
		info, err := s.scanCatalog(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, info)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list catalogs: %w", err)
	}
	return out, nil
}

// Catalog returns one catalogue.
func (s *Store) Catalog(ctx context.Context, etd, grupo string) (Info, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+catalogColumns+" FROM catalog WHERE etd = ? AND grupo = ?", etd, grupo)
	info, err := s.scanCatalog(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Info{}, fmt.Errorf("%w: catalog %s-%s", ErrNotFound, etd, grupo)
	}
	return info, err
}

type attrTable struct{ kind, table, name string }

// attrTables lists attribute tables in modelnis column order: csel (T) tables, then cinf (I) tables.
func (s *Store) attrTables(ctx context.Context, etd, grupo string, lang Lang) ([]attrTable, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT t.kind, t.tabla, COALESCE(l.label, t.label, t.tabla)
		FROM attr_table t
		LEFT JOIN attr_table l ON l.etd = t.etd AND l.grupo = t.grupo AND l.kind = t.kind AND l.tabla = t.tabla AND l.lang = ?
		WHERE t.etd = ? AND t.grupo = ? AND t.lang = 'en'
		ORDER BY CASE t.kind WHEN 'T' THEN 0 ELSE 1 END, t.tabla`, string(lang), etd, grupo)
	if err != nil {
		return nil, fmt.Errorf("attribute tables: %w", err)
	}
	defer rows.Close()
	var out []attrTable
	for rows.Next() {
		var t attrTable
		if err := rows.Scan(&t.kind, &t.table, &t.name); err != nil {
			return nil, fmt.Errorf("attribute tables: %w", err)
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("attribute tables: %w", err)
	}
	return out, nil
}

// attrValueLabels maps "table/cetd" to the value label.
func (s *Store) attrValueLabels(ctx context.Context, etd, grupo string, lang Lang) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT v.tabla, v.cetd, COALESCE(l.label, v.label, v.code, v.cetd)
		FROM attr_value v
		LEFT JOIN attr_value l ON l.etd = v.etd AND l.grupo = v.grupo AND l.tabla = v.tabla AND l.cetd = v.cetd AND l.lang = ?
		WHERE v.etd = ? AND v.grupo = ? AND v.lang = 'en'`, string(lang), etd, grupo)
	if err != nil {
		return nil, fmt.Errorf("attribute values: %w", err)
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var table, cetd, label string
		if err := rows.Scan(&table, &cetd, &label); err != nil {
			return nil, fmt.Errorf("attribute values: %w", err)
		}
		out[table+"/"+cetd] = label
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("attribute values: %w", err)
	}
	return out, nil
}

func buildAttributes(tables []attrTable, labels map[string]string, values []string) []Attribute {
	out := make([]Attribute, 0, len(tables))
	for i, t := range tables {
		if i >= len(values) {
			break
		}
		a := Attribute{Kind: t.kind, Table: t.table, Name: t.name, Code: values[i], Value: labels[t.table+"/"+values[i]]}
		if a.Value == "" {
			a.Value = a.Code
		}
		out = append(out, a)
	}
	return out
}

func (s *Store) attributes(ctx context.Context, etd, grupo string, values []string, lang Lang) ([]Attribute, error) {
	tables, err := s.attrTables(ctx, etd, grupo, lang)
	if err != nil {
		return nil, err
	}
	labels, err := s.attrValueLabels(ctx, etd, grupo, lang)
	if err != nil {
		return nil, err
	}
	return buildAttributes(tables, labels, values), nil
}

// Models lists the model codes of a catalog with their attributes and VIN counts.
func (s *Store) Models(ctx context.Context, etd, grupo string, lang Lang) ([]ModelInfo, error) {
	if _, err := s.Catalog(ctx, etd, grupo); err != nil {
		return nil, err
	}
	tables, err := s.attrTables(ctx, etd, grupo, lang)
	if err != nil {
		return nil, err
	}
	labels, err := s.attrValueLabels(ctx, etd, grupo, lang)
	if err != nil {
		return nil, err
	}
	counts, err := s.vinCounts(ctx, etd)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT codenis, c01, c02, c03, c04, c05, c06, c07, c08, c09, c10
		FROM modelnis WHERE etd = ? AND grupo = ? ORDER BY codenis`, etd, grupo)
	if err != nil {
		return nil, fmt.Errorf("list models: %w", err)
	}
	defer rows.Close()
	out := []ModelInfo{}
	for rows.Next() {
		var model sql.NullString
		var c [10]sql.NullString
		if err := rows.Scan(&model, &c[0], &c[1], &c[2], &c[3], &c[4], &c[5], &c[6], &c[7], &c[8], &c[9]); err != nil {
			return nil, fmt.Errorf("list models: %w", err)
		}
		var values [10]string
		for i := range c {
			values[i] = str(c[i])
		}
		m := str(model)
		out = append(out, ModelInfo{Model: m, VINCount: counts[m], Attributes: buildAttributes(tables, labels, values[:])})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list models: %w", err)
	}
	return out, nil
}

func (s *Store) vinCounts(ctx context.Context, etd string) (map[string]int, error) {
	// DISTINCT: v1t lists some VINs twice.
	rows, err := s.db.QueryContext(ctx, "SELECT codenis, COUNT(DISTINCT vin) FROM vin WHERE etd = ? GROUP BY codenis", etd)
	if err != nil {
		return nil, fmt.Errorf("count VINs: %w", err)
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var model sql.NullString
		var n int
		if err := rows.Scan(&model, &n); err != nil {
			return nil, fmt.Errorf("count VINs: %w", err)
		}
		out[str(model)] = n
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("count VINs: %w", err)
	}
	return out, nil
}

// Vehicle describes the scoped vehicle: catalog, attributes (when the model is known), documents.
func (s *Store) Vehicle(ctx context.Context, sc *Scope, lang Lang) (Vehicle, error) {
	if sc == nil {
		return Vehicle{}, fmt.Errorf("%w: a vehicle scope is required", ErrInvalid)
	}
	info, err := s.Catalog(ctx, sc.Etd, sc.Grupo)
	if err != nil {
		return Vehicle{}, err
	}
	v := Vehicle{VIN: sc.VIN, Model: sc.Model, Catalog: info, Attributes: []Attribute{}}
	if sc.ProdDate > 0 {
		v.ProdDate = *yearMonth(sc.ProdDate)
	}
	if sc.Model != "" {
		if v.Attributes, err = s.attributes(ctx, sc.Etd, sc.Grupo, sc.values[:], lang); err != nil {
			return Vehicle{}, err
		}
		err = s.db.QueryRowContext(ctx, "SELECT COUNT(DISTINCT vin) FROM vin WHERE etd = ? AND codenis = ?", sc.Etd, sc.Model).
			Scan(&v.VINCount)
		if err != nil {
			return Vehicle{}, fmt.Errorf("count VINs: %w", err)
		}
	}
	if v.Documents, err = s.documents(ctx, sc.Etd); err != nil {
		return Vehicle{}, err
	}
	return v, nil
}

func (s *Store) documents(ctx context.Context, etd string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT file FROM cinfo WHERE etd = ? ORDER BY file", etd)
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var file string
		if err := rows.Scan(&file); err != nil {
			return nil, fmt.Errorf("list documents: %w", err)
		}
		if u := s.fileURL("cinfo", etd, file); u != "" {
			out = append(out, u)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}
	return out, nil
}

// IdentifyVIN resolves a VIN exactly, or returns candidates whose VIN ends with the input (>= 6 characters).
func (s *Store) IdentifyVIN(ctx context.Context, input string, lang Lang) (VINResult, error) {
	vin := NormalizeVIN(input)
	if vin == "" {
		return VINResult{}, fmt.Errorf("%w: empty VIN", ErrInvalid)
	}
	sc, err := s.scopeFromVIN(ctx, vin)
	if err == nil {
		v, err := s.Vehicle(ctx, sc, lang)
		if err != nil {
			return VINResult{}, err
		}
		return VINResult{Vehicle: &v}, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return VINResult{}, err
	}
	// The tail goes into a GLOB pattern: drop its metacharacters.
	tail := strings.Map(func(r rune) rune {
		if strings.ContainsRune("*?[]", r) {
			return -1
		}
		return r
	}, vin)
	if len(tail) < minVINTail {
		return VINResult{}, err
	}
	candidates, err := s.vinsEndingWith(ctx, tail, maxCandidates)
	if err != nil {
		return VINResult{}, err
	}
	if len(candidates) == 0 {
		return VINResult{}, fmt.Errorf("%w: VIN %s", ErrNotFound, vin)
	}
	return VINResult{Candidates: candidates}, nil
}

// vinsEndingWith searches a suffix as a prefix GLOB on the reversed VIN, so SQLite can use the vin_rev index.
// v1t has no period: it comes from modelnis (a model code belongs to one period of its series).
func (s *Store) vinsEndingWith(ctx context.Context, tail string, limit int) ([]VINMatch, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT v.vin, v.codenis, v.etd, v.prodata,
		(SELECT m.grupo FROM modelnis m WHERE m.etd = v.etd AND m.codenis = v.codenis LIMIT 1)
		FROM vin v WHERE v.vin_rev GLOB ? ORDER BY v.vin LIMIT ?`, reverse(tail)+"*", limit)
	if err != nil {
		return nil, fmt.Errorf("search VIN tail: %w", err)
	}
	defer rows.Close()
	out := []VINMatch{}
	for rows.Next() {
		var f [5]sql.NullString
		if err := rows.Scan(&f[0], &f[1], &f[2], &f[3], &f[4]); err != nil {
			return nil, fmt.Errorf("search VIN tail: %w", err)
		}
		m := VINMatch{VIN: str(f[0]), Model: str(f[1]), Cat: str(f[2]) + "-" + str(f[4])}
		if d, ok := ParseYYYYMM(str(f[3])); ok {
			m.ProdDate = yearMonth(d)
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("search VIN tail: %w", err)
	}
	return out, nil
}
