package catalog

import "testing"

func row(from, to int, cols ...string) *infosecRow {
	r := &infosecRow{secc: "X", from: from, to: to}
	copy(r.cols[:], cols)
	for i := len(cols); i < len(r.cols); i++ {
		r.cols[i] = "-"
	}
	return r
}

func TestApplies(t *testing.T) {
	// Vehicle: c01=1, c02=2, c03=3; c04..c10 empty (NULL in modelnis).
	vals := []string{"1", "2", "3", "", "", "", "", "", "", ""}
	cases := []struct {
		name string
		r    *infosecRow
		prod int
		want bool
	}{
		{"all wildcards", row(198701, 199912, "0", "0", "0"), 198905, true},
		{"dash wildcard on columns the model leaves empty", row(198701, 199912, "1", "2", "3", "-", "-"), 198905, true},
		{"exact match", row(198701, 199912, "1", "2", "3"), 198905, true},
		{"mismatch on c01", row(198701, 199912, "2", "0", "0"), 198905, false},
		{"mismatch on c03", row(198701, 199912, "0", "0", "4"), 198905, false},
		{"value required on a column the model leaves empty", row(198701, 199912, "0", "0", "0", "1"), 198905, false},
		{"production date equals start", row(198905, 199912, "0"), 198905, true},
		{"production date equals end", row(198701, 198905, "0"), 198905, true},
		{"produced before start", row(198906, 199912, "0"), 198905, false},
		{"produced after end", row(198701, 198904, "0"), 198905, false},
		{"unknown production date ignores dates", row(198906, 199912, "0"), 0, true},
		{"open end", row(198701, 999999, "0"), 201501, true},
	}
	for _, tc := range cases {
		if got := applies(tc.r, vals, tc.prod); got != tc.want {
			t.Errorf("%s: applies = %v, want %v", tc.name, got, tc.want)
		}
	}
}
