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

func TestPeriodContainsAndString(t *testing.T) {
	p := Period{198704, 199206}
	if !p.Contains(198704) || !p.Contains(199206) || p.Contains(198703) || p.Contains(199207) {
		t.Error("bounds must be inclusive")
	}
	if !(Period{0, 198802}).Contains(195001) || !(Period{199411, 0}).Contains(203001) {
		t.Error("open bounds must contain everything on their side")
	}
	cases := map[Period]string{{198704, 199206}: "04/87-06/92", {0, 198802}: "-02/88", {199411, 0}: "11/94-"}
	for p, want := range cases {
		if got := p.String(); got != want {
			t.Errorf("%+v.String() = %q, want %q", p, got, want)
		}
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
	if got := FormatYYYYMM(198905); got != "05/1989" {
		t.Errorf("FormatYYYYMM = %q", got)
	}
}

func TestReverse(t *testing.T) {
	if got := reverse("VSKBEC220U0990494"); got != "4940990U022CEBKSV" {
		t.Errorf("reverse = %q", got)
	}
}
