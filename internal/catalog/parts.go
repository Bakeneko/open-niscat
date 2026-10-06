package catalog

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
)

// maxChain bounds replacement chain walks.
const maxChain = 50

// RefLink is a reference in a replacement chain.
type RefLink struct {
	Key    string `json:"key"`
	PartNo string `json:"partNo"`
}

// PartInfo describes a reference: where it is used, and every reference it replaces (Previous) or that
// replaces it (Next), nearest first.
type PartInfo struct {
	Key         string    `json:"key"`
	PartNo      string    `json:"partNo"`
	Description string    `json:"description"`
	Occurrences []LineRef `json:"occurrences"`
	Previous    []RefLink `json:"previous"`
	Next        []RefLink `json:"next"`
}

// candidates returns the distinct references next to key, latest period start first (ties: higher key first).
// Forward: rows whose alternative (ree) is key, i.e. its successors. Backward: the alternatives of key's rows,
// i.e. its predecessors.
func (s *Store) candidates(ctx context.Context, key string, forward bool) ([]RefLink, error) {
	query := "SELECT part_key, part_no, dataplic FROM part WHERE ree_key = ? AND part_key IS NOT NULL AND part_key <> ?"
	if !forward {
		query = "SELECT ree_key, ree, dataplic FROM part WHERE part_key = ? AND ree_key IS NOT NULL AND ree_key <> ?"
	}
	rows, err := s.db.QueryContext(ctx, query, key, key)
	if err != nil {
		return nil, fmt.Errorf("replacement chain: %w", err)
	}
	defer rows.Close()
	starts := map[string]int{}
	var out []RefLink
	for rows.Next() {
		var k, no, dataplic sql.NullString
		if err := rows.Scan(&k, &no, &dataplic); err != nil {
			return nil, fmt.Errorf("replacement chain: %w", err)
		}
		start := 0
		if p, ok := ParsePeriod(str(dataplic)); ok {
			start = p.From
		}
		ref := str(k)
		prev, seen := starts[ref]
		if !seen {
			out = append(out, RefLink{Key: ref, PartNo: str(no)})
		}
		if !seen || start > prev {
			starts[ref] = start
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("replacement chain: %w", err)
	}
	sort.SliceStable(out, func(i, j int) bool {
		si, sj := starts[out[i].Key], starts[out[j].Key]
		if si != sj {
			return si > sj
		}
		return out[i].Key > out[j].Key
	})
	return out, nil
}

// chain follows the preferred (first) candidate from key until none is left, a cycle or maxChain.
func (s *Store) chain(ctx context.Context, key string, forward bool) ([]RefLink, error) {
	seen := map[string]bool{key: true}
	out := []RefLink{}
	for cur := key; len(out) < maxChain; {
		next, err := s.candidates(ctx, cur, forward)
		if err != nil {
			return nil, err
		}
		if len(next) == 0 || seen[next[0].Key] {
			break
		}
		seen[next[0].Key] = true
		out = append(out, next[0])
		cur = next[0].Key
	}
	return out, nil
}

// related lists every reference reachable from key in one direction, breadth-first (nearest first), without
// duplicates, bounded by maxChain.
func (s *Store) related(ctx context.Context, key string, forward bool) ([]RefLink, error) {
	seen := map[string]bool{key: true}
	out := []RefLink{}
	for frontier := []string{key}; len(frontier) > 0 && len(out) < maxChain; {
		var level []string
		for _, k := range frontier {
			next, err := s.candidates(ctx, k, forward)
			if err != nil {
				return nil, err
			}
			for _, c := range next {
				if seen[c.Key] || len(out) >= maxChain {
					continue
				}
				seen[c.Key] = true
				out = append(out, c)
				level = append(level, c.Key)
			}
		}
		frontier = level
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
	if info.Previous, err = s.related(ctx, key, false); err != nil {
		return PartInfo{}, err
	}
	if info.Next, err = s.related(ctx, key, true); err != nil {
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
