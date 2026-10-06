package catalog

import "testing"

func TestNormalizeRef(t *testing.T) {
	cases := map[string]string{
		"-23300-D9701":  "23300D9701",
		"23300 d9701":   "23300D9701",
		"D-4100-17C90":  "D410017C90",
		" -03902204-0 ": "039022040",
		"":              "",
	}
	for in, want := range cases {
		if got := NormalizeRef(in); got != want {
			t.Errorf("NormalizeRef(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeVIN(t *testing.T) {
	cases := map[string]string{
		"vskbec220u0990494":   "VSKBEC220U0990494",
		" VSKBEC220U0990494 ": "VSKBEC220U0990494",
		"1    16  U0520133":   "116U0520133",
	}
	for in, want := range cases {
		if got := NormalizeVIN(in); got != want {
			t.Errorf("NormalizeVIN(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestItemKey(t *testing.T) {
	cases := []struct{ in, want string }{
		{"01", "1"}, {"1", "1"}, {"10", "10"}, {" 04 ", "4"}, {"", ""}, {"00", "0"},
	}
	for _, tc := range cases {
		if got := ItemKey(tc.in); got != tc.want {
			t.Errorf("ItemKey(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestParsePeriod(t *testing.T) {
	cases := []struct {
		in   string
		want Period
		ok   bool
	}{
		{"0487-0692", Period{198704, 199206}, true},
		{"1194-0115", Period{199411, 201501}, true},
		{"-0288", Period{0, 198802}, true},
		{"1194-", Period{199411, 0}, true},
		{"", Period{}, false},
		{"-", Period{}, false},
		{"1387-0101", Period{}, false},
		{"abcd-0101", Period{}, false},
		{"0487", Period{}, false},
	}
	for _, tc := range cases {
		got, ok := ParsePeriod(tc.in)
		if got != tc.want || ok != tc.ok {
			t.Errorf("ParsePeriod(%q) = %+v, %v; want %+v, %v", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestPeriodContains(t *testing.T) {
	p := Period{198704, 199206}
	if !p.Contains(198704) || !p.Contains(199206) || p.Contains(198703) || p.Contains(199207) {
		t.Error("bounds must be inclusive")
	}
	if !(Period{0, 198802}).Contains(195001) || !(Period{199411, 0}).Contains(203001) {
		t.Error("open bounds must contain everything on their side")
	}
}

func TestYYYYMM(t *testing.T) {
	if v, ok := ParseYYYYMM("198905"); !ok || v != 198905 {
		t.Errorf("ParseYYYYMM(198905) = %d, %v", v, ok)
	}
	for _, bad := range []string{"", "1989", "198913", "19890a"} {
		if _, ok := ParseYYYYMM(bad); ok {
			t.Errorf("ParseYYYYMM(%q) should fail", bad)
		}
	}
}

func TestReverse(t *testing.T) {
	if got := reverse("VSKBEC220U0990494"); got != "4940990U022CEBKSV" {
		t.Errorf("reverse = %q", got)
	}
}

func TestYearMonthAndParseMonthYear(t *testing.T) {
	if got := yearMonth(198905); got == nil || *got != "1989-05" {
		t.Errorf("yearMonth(198905) = %v", got)
	}
	if yearMonth(0) != nil {
		t.Error("yearMonth(0) must be nil")
	}
	cases := []struct {
		in   string
		want int
		ok   bool
	}{
		{"04/87", 198704, true},
		{"04-87", 198704, true},
		{" 03/93", 199303, true},
		{"12/15", 201512, true},
		{"", 0, false},
		{"13/87", 0, false},
		{"0487", 0, false},
		{"04/1987", 0, false},
	}
	for _, tc := range cases {
		got, ok := parseMonthYear(tc.in)
		if got != tc.want || ok != tc.ok {
			t.Errorf("parseMonthYear(%q) = %d, %v", tc.in, got, ok)
		}
	}
}
