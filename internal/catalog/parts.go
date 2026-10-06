package catalog

import (
	"context"
	"database/sql"
	"fmt"
)

// maxChain bounds replacement chain walks.
const maxChain = 50

// RefLink is a reference in a replacement chain.
type RefLink struct {
	Key    string `json:"key"`
	PartNo string `json:"partNo"`
}

// PartInfo describes a reference: where it is used and its replacement chain both ways (nearest first).
type PartInfo struct {
	Key         string    `json:"key"`
	PartNo      string    `json:"partNo"`
	Description string    `json:"description"`
	Occurrences []LineRef `json:"occurrences"`
	Previous    []RefLink `json:"previous"`
	Next        []RefLink `json:"next"`
}

// step returns the next reference of key. Forward: rows whose alternative (ree) is key, i.e. its successors.
// Backward: the alternatives of key's rows, i.e. its predecessors. Several candidates: latest period start wins.
func (s *Store) step(ctx context.Context, key string, forward bool) (RefLink, bool, error) {
	query := "SELECT part_key, part_no, dataplic FROM part WHERE ree_key = ? AND part_key IS NOT NULL AND part_key <> ?"
	if !forward {
		query = "SELECT ree_key, ree, dataplic FROM part WHERE part_key = ? AND ree_key IS NOT NULL AND ree_key <> ?"
	}
	rows, err := s.db.QueryContext(ctx, query, key, key)
	if err != nil {
		return RefLink{}, false, fmt.Errorf("replacement chain: %w", err)
	}
	defer rows.Close()
	var best RefLink
	bestStart, found := -1, false
	for rows.Next() {
		var k, no, dataplic sql.NullString
		if err := rows.Scan(&k, &no, &dataplic); err != nil {
			return RefLink{}, false, fmt.Errorf("replacement chain: %w", err)
		}
		start := 0
		if p, ok := ParsePeriod(str(dataplic)); ok {
			start = p.From
		}
		cand := RefLink{Key: str(k), PartNo: str(no)}
		if start > bestStart || (start == bestStart && cand.Key > best.Key) {
			best, bestStart, found = cand, start, true
		}
	}
	if err := rows.Err(); err != nil {
		return RefLink{}, false, fmt.Errorf("replacement chain: %w", err)
	}
	return best, found, nil
}

// chain walks from key in one direction until no further reference, a cycle or maxChain.
func (s *Store) chain(ctx context.Context, key string, forward bool) ([]RefLink, error) {
	seen := map[string]bool{key: true}
	out := []RefLink{}
	for cur := key; len(out) < maxChain; {
		next, ok, err := s.step(ctx, cur, forward)
		if err != nil {
			return nil, err
		}
		if !ok || seen[next.Key] {
			break
		}
		seen[next.Key] = true
		out = append(out, next)
		cur = next.Key
	}
	return out, nil
}

// latest returns the last successor of key, or nil when nothing replaces it.
func (s *Store) latest(ctx context.Context, key string) (*RefLink, error) {
	next, err := s.chain(ctx, key, true)
	if err != nil || len(next) == 0 {
		return nil, err
	}
	last := next[len(next)-1]
	return &last, nil
}

// Part returns every use of a reference and its replacement chain.
func (s *Store) Part(ctx context.Context, ref string, lang Lang) (PartInfo, error) {
	key := NormalizeRef(ref)
	if key == "" {
		return PartInfo{}, fmt.Errorf("%w: empty reference", ErrInvalid)
	}
	cond, args := s.langFilter("p", lang)
	// cond only contains fixed SQL and "?" placeholders (see langFilter).
	rows, err := s.db.QueryContext(ctx, "SELECT "+lineColumns+" FROM part p WHERE p.part_key = ? AND "+cond+ //nolint:gosec // no user input is concatenated
		" ORDER BY p.etd, p.plate, p.pospie", append([]any{key}, args...)...)
	if err != nil {
		return PartInfo{}, fmt.Errorf("part occurrences: %w", err)
	}
	defer rows.Close()
	info := PartInfo{Key: key, PartNo: ref, Occurrences: []LineRef{}}
	for rows.Next() {
		l, err := scanLine(rows)
		if err != nil {
			return PartInfo{}, err
		}
		info.Occurrences = append(info.Occurrences, l)
	}
	if err := rows.Err(); err != nil {
		return PartInfo{}, fmt.Errorf("part occurrences: %w", err)
	}
	if info.Previous, err = s.chain(ctx, key, false); err != nil {
		return PartInfo{}, err
	}
	if info.Next, err = s.chain(ctx, key, true); err != nil {
		return PartInfo{}, err
	}
	if len(info.Occurrences) == 0 && len(info.Previous) == 0 && len(info.Next) == 0 {
		return PartInfo{}, fmt.Errorf("%w: reference %s", ErrNotFound, ref)
	}
	if len(info.Occurrences) > 0 {
		info.PartNo, info.Description = info.Occurrences[0].PartNo, info.Occurrences[0].Description
	}
	return info, nil
}
