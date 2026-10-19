package natives

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// ============ Date/Time functions ============
//
// A date argument is either milliseconds since the epoch or ISO-8601 text
// (YYYY-MM-DD, optionally followed by a time and a zone). Every calendar
// computation is done in UTC so that a script gives the same answer on every
// host and in every implementation. A date that cannot be read yields null.

const msPerDay = 86_400_000

// maxMillis bounds a date to ±100 000 000 days around 1970, the range every
// implementation can represent (it is JavaScript's): beyond it a date is unreadable.
const maxMillis = 8.64e15

// readDate reads a date argument; ok is false when the value is not a date.
func readDate(v interface{}) (time.Time, bool) {
	switch d := v.(type) {
	case float64:
		if math.IsNaN(d) || math.Abs(d) > maxMillis {
			return time.Time{}, false
		}
		return time.UnixMilli(int64(d)).UTC(), true
	case int:
		return readDate(float64(d))
	case int64:
		return readDate(float64(d))
	case string:
		t, ok := parseISO(strings.TrimSpace(d))
		return t, ok && inRange(t)
	}
	return time.Time{}, false
}

func inRange(t time.Time) bool {
	ms := t.UnixMilli()
	return float64(ms) >= -maxMillis && float64(ms) <= maxMillis && t.Year() > -300000 && t.Year() < 300000
}

// dateValue is a computed date in milliseconds, or nil when it leaves the representable range.
func dateValue(t time.Time) interface{} {
	if !inRange(t) {
		return nil
	}
	return millis(t)
}

// parseISO reads YYYY-MM-DD[(T| )HH:mm[:ss[.fraction]][Z|±HH:mm|±HHmm]].
func parseISO(s string) (time.Time, bool) {
	p := &scanner{s: s}
	year, ok1 := p.digits(4)
	ok2 := p.literal('-')
	month, ok3 := p.digits(2)
	ok4 := p.literal('-')
	day, ok5 := p.digits(2)
	if !(ok1 && ok2 && ok3 && ok4 && ok5) {
		return time.Time{}, false
	}
	hour, minute, second, nanos := 0, 0, 0, 0
	offset := 0
	if !p.done() {
		if !p.literal('T') && !p.literal(' ') {
			return time.Time{}, false
		}
		var ok bool
		if hour, ok = p.digits(2); !ok || !p.literal(':') {
			return time.Time{}, false
		}
		if minute, ok = p.digits(2); !ok {
			return time.Time{}, false
		}
		if p.literal(':') {
			if second, ok = p.digits(2); !ok {
				return time.Time{}, false
			}
			if p.literal('.') {
				start := p.pos
				for !p.done() && isDigit(p.s[p.pos]) && p.pos-start < 9 {
					p.pos++
				}
				frac := p.s[start:p.pos]
				if frac == "" {
					return time.Time{}, false
				}
				frac += strings.Repeat("0", 9-len(frac))
				nanos, _ = strconv.Atoi(frac)
			}
		}
		if !p.done() {
			if p.literal('Z') {
				// UTC
			} else if c := p.s[p.pos]; c == '+' || c == '-' {
				p.pos++
				oh, ok := p.digits(2)
				if !ok {
					return time.Time{}, false
				}
				p.literal(':')
				om, ok := p.digits(2)
				if !ok || oh > 23 || om > 59 {
					return time.Time{}, false
				}
				offset = oh*3600 + om*60
				if c == '-' {
					offset = -offset
				}
			} else {
				return time.Time{}, false
			}
		}
	}
	if !p.done() {
		return time.Time{}, false
	}
	t, ok := buildDate(year, month, day, hour, minute, second, nanos)
	if !ok {
		return time.Time{}, false
	}
	return t.Add(-time.Duration(offset) * time.Second), true
}

// buildDate rejects out-of-range fields instead of letting time.Date normalise them.
func buildDate(year, month, day, hour, minute, second, nanos int) (time.Time, bool) {
	if month < 1 || month > 12 || day < 1 || day > daysIn(year, time.Month(month)) ||
		hour > 23 || minute > 59 || second > 59 {
		return time.Time{}, false
	}
	t := time.Date(year, time.Month(month), day, hour, minute, second, nanos, time.UTC)
	return t.Truncate(time.Millisecond), true
}

func daysIn(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

type scanner struct {
	s   string
	pos int
}

func (p *scanner) done() bool { return p.pos >= len(p.s) }

func (p *scanner) literal(c byte) bool {
	if !p.done() && p.s[p.pos] == c {
		p.pos++
		return true
	}
	return false
}

func (p *scanner) digits(n int) (int, bool) {
	if p.pos+n > len(p.s) {
		return 0, false
	}
	chunk := p.s[p.pos : p.pos+n]
	for i := 0; i < n; i++ {
		if !isDigit(chunk[i]) {
			return 0, false
		}
	}
	p.pos += n
	v, _ := strconv.Atoi(chunk)
	return v, true
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

func millis(t time.Time) float64 { return float64(t.UnixMilli()) }

// count reads a whole-number argument, truncated toward zero.
func count(name string, v interface{}) (int, error) {
	n, ok := toFloat(v)
	if !ok {
		return 0, fmt.Errorf("%s requires a number as second argument", name)
	}
	return int(n), nil
}

// dateOrNow reads an optional date argument: absent means now.
func dateOrNow(args []interface{}) (time.Time, bool) {
	if len(args) == 0 {
		return time.Now().UTC(), true
	}
	return readDate(args[0])
}

// ---- Current date and time ----

func nativeNow(args ...interface{}) (interface{}, error) {
	return float64(time.Now().UnixMilli()), nil
}

func nativeDate(args ...interface{}) (interface{}, error) {
	return time.Now().UTC().Format("2006-01-02"), nil
}

func nativeTime(args ...interface{}) (interface{}, error) {
	return time.Now().UTC().Format("15:04:05"), nil
}

func nativeDatetime(args ...interface{}) (interface{}, error) {
	return isoString(time.Now().UTC()), nil
}

func nativeTimestamp(args ...interface{}) (interface{}, error) {
	t, ok := dateOrNow(args)
	if !ok {
		return nil, nil
	}
	return millis(t), nil
}

// ---- Reading and writing dates ----

func isoString(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}

func nativeToISO(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("toISO requires 1 argument (date)")
	}
	t, ok := readDate(args[0])
	if !ok {
		return nil, nil
	}
	return isoString(t), nil
}

func nativeIsDate(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("isDate requires 1 argument")
	}
	_, ok := readDate(args[0])
	return ok, nil
}

// dateTokens are tried longest first at each position of a format.
var dateTokens = []string{"YYYY", "SSS", "YY", "MM", "DD", "HH", "mm", "ss"}

func tokenAt(format string, i int) string {
	for _, tok := range dateTokens {
		if strings.HasPrefix(format[i:], tok) {
			return tok
		}
	}
	return ""
}

func nativeFormatDate(args ...interface{}) (interface{}, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("formatDate requires 1 or 2 arguments (date, [format])")
	}
	format := "YYYY-MM-DD"
	if len(args) == 2 {
		f, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("formatDate requires a string as second argument")
		}
		format = f
	}
	t, ok := readDate(args[0])
	if !ok {
		return nil, nil
	}
	var b strings.Builder
	for i := 0; i < len(format); {
		tok := tokenAt(format, i)
		switch tok {
		case "YYYY":
			fmt.Fprintf(&b, "%04d", t.Year())
		case "YY":
			fmt.Fprintf(&b, "%02d", t.Year()%100)
		case "MM":
			fmt.Fprintf(&b, "%02d", int(t.Month()))
		case "DD":
			fmt.Fprintf(&b, "%02d", t.Day())
		case "HH":
			fmt.Fprintf(&b, "%02d", t.Hour())
		case "mm":
			fmt.Fprintf(&b, "%02d", t.Minute())
		case "ss":
			fmt.Fprintf(&b, "%02d", t.Second())
		case "SSS":
			fmt.Fprintf(&b, "%03d", t.Nanosecond()/1_000_000)
		default:
			b.WriteByte(format[i])
			i++
			continue
		}
		i += len(tok)
	}
	return b.String(), nil
}

func nativeParseDate(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("parseDate requires 2 arguments (text, format)")
	}
	format, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("parseDate requires a string format as second argument")
	}
	text, ok := args[0].(string)
	if !ok {
		return nil, nil
	}
	year, month, day, hour, minute, second, ms := 1970, 1, 1, 0, 0, 0, 0
	p := &scanner{s: text}
	for i := 0; i < len(format); {
		tok := tokenAt(format, i)
		var field *int
		width := 0
		switch tok {
		case "YYYY":
			field, width = &year, 4
		case "MM":
			field, width = &month, 2
		case "DD":
			field, width = &day, 2
		case "HH":
			field, width = &hour, 2
		case "mm":
			field, width = &minute, 2
		case "ss":
			field, width = &second, 2
		case "SSS":
			field, width = &ms, 3
		}
		if field == nil {
			if !p.literal(format[i]) {
				return nil, nil
			}
			i++
			continue
		}
		v, ok := p.digits(width)
		if !ok {
			return nil, nil
		}
		*field = v
		i += len(tok)
	}
	if !p.done() {
		return nil, nil
	}
	t, ok := buildDate(year, month, day, hour, minute, second, ms*1_000_000)
	if !ok {
		return nil, nil
	}
	return dateValue(t), nil
}

// ---- Calendar fields ----

func calendarField(name string, field func(time.Time) int) NativeFunc {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) > 1 {
			return nil, fmt.Errorf("%s requires 0 or 1 argument ([date])", name)
		}
		t, ok := dateOrNow(args)
		if !ok {
			return nil, nil
		}
		return float64(field(t)), nil
	}
}

var (
	nativeYear      = calendarField("year", func(t time.Time) int { return t.Year() })
	nativeMonth     = calendarField("month", func(t time.Time) int { return int(t.Month()) })
	nativeDay       = calendarField("day", func(t time.Time) int { return t.Day() })
	nativeHour      = calendarField("hour", func(t time.Time) int { return t.Hour() })
	nativeMinute    = calendarField("minute", func(t time.Time) int { return t.Minute() })
	nativeSecond    = calendarField("second", func(t time.Time) int { return t.Second() })
	nativeDayOfWeek = calendarField("dayOfWeek", func(t time.Time) int { return int(t.Weekday()) })
)

func nativeDaysInMonth(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("daysInMonth requires 1 argument (date)")
	}
	t, ok := readDate(args[0])
	if !ok {
		return nil, nil
	}
	return float64(daysIn(t.Year(), t.Month())), nil
}

func nativeIsWeekend(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("isWeekend requires 1 argument (date)")
	}
	t, ok := readDate(args[0])
	if !ok {
		return nil, nil
	}
	return t.Weekday() == time.Saturday || t.Weekday() == time.Sunday, nil
}

// ---- Arithmetic ----

func shift(name string, apply func(time.Time, int) time.Time) NativeFunc {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("%s requires 2 arguments (date, n)", name)
		}
		n, err := count(name, args[1])
		if err != nil {
			return nil, err
		}
		t, ok := readDate(args[0])
		if !ok {
			return nil, nil
		}
		return dateValue(apply(t, n)), nil
	}
}

// addMonthsClamped keeps the day when it exists in the target month, else takes its last day.
func addMonthsClamped(t time.Time, n int) time.Time {
	total := int(t.Month()) - 1 + n
	year := t.Year() + floorDiv(total, 12)
	month := time.Month(total - floorDiv(total, 12)*12 + 1)
	day := t.Day()
	if last := daysIn(year, month); day > last {
		day = last
	}
	return time.Date(year, month, day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
}

func floorDiv(a, b int) int {
	q := a / b
	if (a%b != 0) && ((a < 0) != (b < 0)) {
		q--
	}
	return q
}

var (
	nativeAddDays    = shift("addDays", func(t time.Time, n int) time.Time { return t.AddDate(0, 0, n) })
	nativeAddHours   = shift("addHours", func(t time.Time, n int) time.Time { return t.Add(time.Duration(n) * time.Hour) })
	nativeAddMinutes = shift("addMinutes", func(t time.Time, n int) time.Time { return t.Add(time.Duration(n) * time.Minute) })
	nativeAddSeconds = shift("addSeconds", func(t time.Time, n int) time.Time { return t.Add(time.Duration(n) * time.Second) })
	nativeAddMonths  = shift("addMonths", addMonthsClamped)
	nativeAddYears   = shift("addYears", func(t time.Time, n int) time.Time { return addMonthsClamped(t, 12*n) })
)

func between(name string, measure func(a, b time.Time) int) NativeFunc {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("%s requires 2 arguments (date1, date2)", name)
		}
		a, ok1 := readDate(args[0])
		b, ok2 := readDate(args[1])
		if !ok1 || !ok2 {
			return nil, nil
		}
		return float64(measure(a, b)), nil
	}
}

func elapsed(unit int64) func(a, b time.Time) int {
	return func(a, b time.Time) int { return int((b.UnixMilli() - a.UnixMilli()) / unit) }
}

// monthsBetween counts whole calendar months from a to b.
func monthsBetween(a, b time.Time) int {
	if b.Before(a) {
		return -monthsBetween(b, a)
	}
	months := (b.Year()-a.Year())*12 + int(b.Month()) - int(a.Month())
	if b.Day() < a.Day() || (b.Day() == a.Day() && timeOfDay(b) < timeOfDay(a)) {
		months--
	}
	return months
}

func timeOfDay(t time.Time) time.Duration {
	return time.Duration(t.Hour())*time.Hour + time.Duration(t.Minute())*time.Minute +
		time.Duration(t.Second())*time.Second + time.Duration(t.Nanosecond())
}

var (
	nativeDiffDays    = between("diffDays", elapsed(msPerDay))
	nativeDiffHours   = between("diffHours", elapsed(3_600_000))
	nativeDiffMinutes = between("diffMinutes", elapsed(60_000))
	nativeDiffMonths  = between("diffMonths", monthsBetween)
	nativeDiffYears   = between("diffYears", func(a, b time.Time) int { return monthsBetween(a, b) / 12 })
)

func nativeAge(args ...interface{}) (interface{}, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("age requires 1 or 2 arguments (birth, [at])")
	}
	birth, ok := readDate(args[0])
	if !ok {
		return nil, nil
	}
	at, ok := dateOrNow(args[1:])
	if !ok {
		return nil, nil
	}
	return float64(monthsBetween(birth, at) / 12), nil
}

// ---- Periods ----

func period(name string, bound func(time.Time) time.Time) NativeFunc {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("%s requires 1 argument (date)", name)
		}
		t, ok := readDate(args[0])
		if !ok {
			return nil, nil
		}
		return dateValue(bound(t)), nil
	}
}

const lastMilli = 999 * time.Millisecond

var (
	nativeStartOfDay = period("startOfDay", func(t time.Time) time.Time {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	})
	nativeEndOfDay = period("endOfDay", func(t time.Time) time.Time {
		return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, int(lastMilli), time.UTC)
	})
	nativeStartOfMonth = period("startOfMonth", func(t time.Time) time.Time {
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	})
	nativeEndOfMonth = period("endOfMonth", func(t time.Time) time.Time {
		return time.Date(t.Year(), t.Month(), daysIn(t.Year(), t.Month()), 23, 59, 59, int(lastMilli), time.UTC)
	})
)
