package natives

import "testing"

func TestReadDateForms(t *testing.T) {
	cases := map[string]string{
		"2026-10-19":                     "2026-10-19T00:00:00.000Z",
		"  2026-10-19T08:05  ":           "2026-10-19T08:05:00.000Z",
		"2026-10-19T08:05:06.123456789Z": "2026-10-19T08:05:06.123Z",
		"2026-10-19T08:05:06+0130":       "2026-10-19T06:35:06.000Z",
		"2026-10-19T08:05:06-05:00":      "2026-10-19T13:05:06.000Z",
	}
	for in, want := range cases {
		got, err := nativeToISO(in)
		if err != nil || got != want {
			t.Errorf("toISO(%q) = %v, %v; want %s", in, got, err, want)
		}
	}
	for _, bad := range []interface{}{"2026-10-19T24:00", "2026-10-19T08", "2026-1-19", "2026-10-19Tx", "2026-10-19T08:05+01", true, []interface{}{}} {
		if got, _ := nativeToISO(bad); got != nil {
			t.Errorf("toISO(%v) = %v; want null", bad, got)
		}
	}
}

func TestParseDateIsStrict(t *testing.T) {
	got, _ := nativeParseDate("2026-10-19 08:05:06.007", "YYYY-MM-DD HH:mm:ss.SSS")
	if iso, _ := nativeToISO(got); iso != "2026-10-19T08:05:06.007Z" {
		t.Errorf("parseDate = %v", iso)
	}
	for _, text := range []interface{}{"2026-10-1", "2026-10-190", 20261019, nil} {
		if got, _ := nativeParseDate(text, "YYYY-MM-DD"); got != nil {
			t.Errorf("parseDate(%v) = %v; want null", text, got)
		}
	}
	if _, err := nativeParseDate("2026", 4); err == nil {
		t.Error("a non-string format must raise")
	}
}

func TestMonthsBetweenAcrossYears(t *testing.T) {
	cases := []struct {
		a, b string
		want float64
	}{
		{"2025-12-31T10:00:00Z", "2026-01-31T09:59:59Z", 0},
		{"2025-12-31T10:00:00Z", "2026-01-31T10:00:00Z", 1},
		{"2026-03-01", "2025-03-01", -12},
	}
	for _, c := range cases {
		got, _ := nativeDiffMonths(c.a, c.b)
		if got != c.want {
			t.Errorf("diffMonths(%s, %s) = %v; want %v", c.a, c.b, got, c.want)
		}
	}
	if got, _ := nativeAddMonths("2026-01-15", -13); got != nil {
		if iso, _ := nativeToISO(got); iso != "2024-12-15T00:00:00.000Z" {
			t.Errorf("addMonths(-13) = %v", iso)
		}
	}
}

func TestOutOfRangeDatesAreUnreadable(t *testing.T) {
	if got, _ := nativeToISO(1e30); got != nil {
		t.Errorf("toISO(1e30) = %v; want null", got)
	}
	if got, _ := nativeAddDays("9999-12-31", 1e12); got != nil {
		t.Errorf("addDays far beyond range = %v; want null", got)
	}
	if got, _ := nativeAddYears("2026-01-01", 1e9); got != nil {
		t.Errorf("addYears far beyond range = %v; want null", got)
	}
	if got, _ := nativeYear(8.64e15); got != float64(275760) {
		t.Errorf("year at the range bound = %v; want 275760", got)
	}
}
