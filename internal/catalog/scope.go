package catalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// ScopeParams are the vehicle-scope query parameters: vin, or cat [+ model].
type ScopeParams struct {
	VIN   string
	Cat   string
	Model string
}

// IsZero reports whether no scope was requested.
func (p ScopeParams) IsZero() bool { return p.VIN == "" && p.Cat == "" && p.Model == "" }

// Scope is a resolved vehicle context.
type Scope struct {
	VIN      string
	Etd      string
	Grupo    string
	Model    string
	ProdDate int // YYYYMM, 0 when unknown
	values   [10]string
}

// Cat returns the catalog id, e.g. "AA-G01".
func (sc *Scope) Cat() string { return sc.Etd + "-" + sc.Grupo }

// ParseCat splits a catalog id "AA-G01" into etd and grupo.
func ParseCat(cat string) (etd, grupo string, err error) {
	etd, grupo, ok := strings.Cut(strings.ToUpper(strings.TrimSpace(cat)), "-")
	if !ok || len(etd) != 2 || grupo == "" {
		return "", "", fmt.Errorf("%w: catalog %q (expected e.g. AA-G01)", ErrInvalid, cat)
	}
	return etd, grupo, nil
}

// ResolveScope validates scope parameters against the data. It returns nil, nil when no scope is requested.
func (s *Store) ResolveScope(ctx context.Context, p ScopeParams) (*Scope, error) {
	switch {
	case p.IsZero():
		return nil, nil
	case p.VIN != "":
		return s.scopeFromVIN(ctx, NormalizeVIN(p.VIN))
	case p.Cat == "":
		return nil, fmt.Errorf("%w: a model scope needs cat", ErrInvalid)
	}
	etd, grupo, err := ParseCat(p.Cat)
	if err != nil {
		return nil, err
	}
	if _, err := s.Catalog(ctx, etd, grupo); err != nil {
		return nil, err
	}
	sc := &Scope{Etd: etd, Grupo: grupo}
	if p.Model == "" {
		return sc, nil
	}
	_, values, err := s.modelRow(ctx, etd, grupo, p.Model)
	if err != nil {
		return nil, err
	}
	sc.Model, sc.values = p.Model, values
	return sc, nil
}

func (s *Store) scopeFromVIN(ctx context.Context, vin string) (*Scope, error) {
	var model, etd, prod sql.NullString
	err := s.db.QueryRowContext(ctx, "SELECT codenis, etd, prodata FROM vin WHERE vin = ? LIMIT 1", vin).
		Scan(&model, &etd, &prod)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: VIN %s", ErrNotFound, vin)
	}
	if err != nil {
		return nil, fmt.Errorf("look up VIN: %w", err)
	}
	grupo, values, err := s.modelRow(ctx, str(etd), "", str(model))
	if err != nil {
		return nil, err
	}
	date, _ := ParseYYYYMM(str(prod))
	return &Scope{VIN: vin, Etd: str(etd), Grupo: grupo, Model: str(model), ProdDate: date, values: values}, nil
}

// modelRow returns the grupo and c01..c10 of a model code; grupo "" matches any period (a model code belongs
// to a single period, so a VIN, which has none, resolves to it).
func (s *Store) modelRow(ctx context.Context, etd, grupo, model string) (modelGrupo string, values [10]string, err error) {
	var g sql.NullString
	var c [10]sql.NullString
	err = s.db.QueryRowContext(ctx, `SELECT grupo, c01, c02, c03, c04, c05, c06, c07, c08, c09, c10
		FROM modelnis WHERE etd = ? AND codenis = ? AND (? = '' OR grupo = ?) LIMIT 1`, etd, model, grupo, grupo).
		Scan(&g, &c[0], &c[1], &c[2], &c[3], &c[4], &c[5], &c[6], &c[7], &c[8], &c[9])
	if errors.Is(err, sql.ErrNoRows) {
		return "", values, fmt.Errorf("%w: model %s in %s", ErrNotFound, model, etd)
	}
	if err != nil {
		return "", values, fmt.Errorf("look up model: %w", err)
	}
	for i := range c {
		values[i] = str(c[i])
	}
	return str(g), values, nil
}
