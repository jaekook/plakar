package policy

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// test helpers

func mustLoc(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}

func id(i int) string { return fmt.Sprintf("id%03d", i) }

func ts(y int, m time.Month, d, hh, mm, ss int, loc *time.Location) time.Time {
	return time.Date(y, m, d, hh, mm, ss, 0, loc)
}

func contains(set map[string]struct{}, k string) bool {
	_, ok := set[k]
	return ok
}

// --- Tests ---

func TestSelect_DailyCap1(t *testing.T) {
	loc := time.UTC
	now := ts(2025, 8, 20, 12, 0, 0, loc)

	// Build 5 days of snapshots, newest day first, 2 per day (one minute apart).
	var snaps []Item
	n := 0
	for d := 20; d >= 16; d-- {
		snaps = append(snaps, Item{ItemID: id(n), Timestamp: ts(2025, 8, d, 10, 1, 0, loc)})
		n++
		snaps = append(snaps, Item{ItemID: id(n), Timestamp: ts(2025, 8, d, 10, 0, 0, loc)})
		n++
	}

	crit := Criteria{
		KeepDays:   3, // last 3 days
		KeepPerDay: 1, // keep top 1 per day
	}
	pol := NewPolicy(crit, now)
	keep, why := pol.Select(snaps)

	if len(keep) != 3 {
		t.Fatalf("expected 3 kept, got %d", len(keep))
	}
	// Expect the newest per-day representative (the xx:10:01) for days 20, 19, 18.
	wantIDs := []string{"id000", "id002", "id004"}
	for _, w := range wantIDs {
		if !contains(keep, w) {
			t.Fatalf("expected keep to contain %s", w)
		}
		r := why[w]
		if r.Action != "keep" || r.Rule != RuleDays || r.Rank != 1 || r.Cap != 1 {
			t.Fatalf("bad reason for %s: %+v", w, r)
		}
		if r.Bucket != dayKey(snaps[0].Timestamp) && r.Bucket == "" {
			t.Fatalf("unexpected bucket for %s: %+v", w, r)
		}
	}
	// An older day must be deleted "outside windows"
	delID := "id008" // day 16 top entry
	if _, ok := keep[delID]; ok {
		t.Fatalf("expected %s to be deleted", delID)
	}
	if why[delID].Action != "delete" {
		t.Fatalf("expected delete reason for %s", delID)
	}
}

func TestSelect_PerBucketCap2(t *testing.T) {
	loc := time.UTC
	now := ts(2025, 8, 20, 12, 0, 0, loc)

	// Same day, 5 snapshots at different minutes.
	var snaps []Item
	for i := 0; i < 5; i++ {
		snaps = append(snaps, Item{ItemID: id(i), Timestamp: ts(2025, 8, 20, 10, 59-i, 0, loc)})
	}

	crit := Criteria{KeepDays: 1, KeepPerDay: 2}
	pol := NewPolicy(crit, now)
	keep, why := pol.Select(snaps)

	if len(keep) != 2 {
		t.Fatalf("expected 2 kept, got %d", len(keep))
	}
	// Should keep the two newest: id000 (10:59) and id001 (10:58)
	for _, k := range []string{"id000", "id001"} {
		if !contains(keep, k) {
			t.Fatalf("missing kept %s", k)
		}
		if why[k].Rank != 1 && why[k].Rank != 2 {
			t.Fatalf("unexpected rank for %s: %+v", k, why[k])
		}
		if why[k].Cap != 2 {
			t.Fatalf("unexpected cap for %s: %+v", k, why[k])
		}
	}
	// The rest exceeded cap; ensure reason says so
	for _, k := range []string{"id002", "id003", "id004"} {
		if contains(keep, k) {
			t.Fatalf("unexpected keep %s", k)
		}
		if why[k].Action != "delete" || why[k].Note != "exceeds per-bucket cap" {
			t.Fatalf("expected 'exceeds per-bucket cap' delete for %s; got %+v", k, why[k])
		}
		if why[k].Rule != RuleDays {
			t.Fatalf("expected RuleDays for %s; got %+v", k, why[k])
		}
	}
}

func TestSelect_UnionAcrossRules(t *testing.T) {
	loc := time.UTC
	now := ts(2025, 8, 20, 12, 0, 0, loc)

	// Build snapshots across two hours with several minutes.
	snaps := []Item{
		{ItemID: "A", Timestamp: ts(2025, 8, 20, 12, 10, 0, loc)},
		{ItemID: "B", Timestamp: ts(2025, 8, 20, 12, 00, 0, loc)},
		{ItemID: "C", Timestamp: ts(2025, 8, 20, 11, 59, 0, loc)},
		{ItemID: "D", Timestamp: ts(2025, 8, 20, 11, 05, 0, loc)},
	}

	crit := Criteria{
		KeepHours:     1, // last hour window: 12:00..12:59
		KeepPerHour:   1, // keep top of that hour
		KeepMinutes:   2, // additionally: last 2 minutes windows
		KeepPerMinute: 1,
	}
	pol := NewPolicy(crit, now)
	keep, _ := pol.Select(snaps)

	// Hour rule should keep "A" (top of hour 12).
	// Minutes rule (2 buckets: now minute and previous minute) may keep "A" plus another minute bucket representative.
	if !contains(keep, "A") {
		t.Fatalf("expected A kept by hour rule")
	}
	// Expect at least 2 kept due to union (hour + minute); exact 2 if minutes pick A and B.
	if len(keep) < 2 {
		t.Fatalf("expected at least 2 kept due to union, got %d", len(keep))
	}
}

func TestSelect_OutsideRetentionWindows(t *testing.T) {
	loc := time.UTC
	now := ts(2025, 8, 20, 12, 0, 0, loc)

	// One very old snapshot
	snaps := []Item{
		{ItemID: "old", Timestamp: ts(2023, 8, 20, 12, 0, 0, loc)},
	}
	crit := Criteria{KeepDays: 1, KeepPerDay: 1}
	pol := NewPolicy(crit, now)
	keep, why := pol.Select(snaps)

	if len(keep) != 0 {
		t.Fatalf("expected no kept snapshots")
	}
	r := why["old"]
	if r.Action != "delete" || r.Note != "outside retention windows" || r.Rule != "" {
		t.Fatalf("expected outside-window delete reason; got %+v", r)
	}
}

func TestSelect_ISOWeeks(t *testing.T) {
	loc := time.UTC
	// ISO week boundary: 2025-12-29 (Mon) is ISO week 2025-W01; test around turn of year.
	mon := ts(2025, 12, 29, 10, 0, 0, loc) // Monday
	sun := ts(2025, 12, 28, 10, 0, 0, loc) // Sunday prior (belongs to ISO week 2025-W52)

	snaps := []Item{
		{ItemID: "sun", Timestamp: sun},
		{ItemID: "mon", Timestamp: mon},
	}
	crit := Criteria{KeepWeeks: 2, KeepPerWeek: 1}
	pol := NewPolicy(crit, mon)
	keep, why := pol.Select(snaps)

	if len(keep) != 2 {
		t.Fatalf("expected both weeks kept, got %d", len(keep))
	}
	if why["sun"].Bucket == why["mon"].Bucket {
		t.Fatalf("expected different ISO week buckets for sun and mon, got both %s", why["sun"].Bucket)
	}
}

func TestSelect_MonthsYears(t *testing.T) {
	loc := time.UTC
	now := ts(2025, 8, 20, 12, 0, 0, loc)

	snaps := []Item{
		{ItemID: "m8", Timestamp: ts(2025, 8, 10, 0, 0, 0, loc)},
		{ItemID: "m7", Timestamp: ts(2025, 7, 10, 0, 0, 0, loc)},
		{ItemID: "m6", Timestamp: ts(2025, 6, 10, 0, 0, 0, loc)},
		{ItemID: "y2024", Timestamp: ts(2024, 12, 31, 23, 59, 0, loc)},
	}

	crit := Criteria{
		KeepMonths: 2, KeepPerMonth: 1, // keep Aug & Jul representatives
		KeepYears: 2, KeepPerYear: 1, // keep representatives for 2025 and 2024
	}
	pol := NewPolicy(crit, now)
	keep, why := pol.Select(snaps)

	// Expect: m8 (Aug rep), m7 (Jul rep), and y2024 (year 2024 rep).
	// m6 may or may not be included depending on overlap; but with KeepMonths=2 (Aug, Jul),
	// month rule will not include June. Year rule should include one snapshot for 2025 and one for 2024.
	if !contains(keep, "m8") || !contains(keep, "m7") || !contains(keep, "y2024") {
		t.Fatalf("expected m8, m7, y2024 to be kept; keep=%v", keep)
	}
	if why["m8"].Rule != RuleMonths || why["m7"].Rule != RuleMonths {
		t.Fatalf("expected month rule for m8/m7; got %+v %+v", why["m8"], why["m7"])
	}
	if why["y2024"].Rule != RuleYears {
		t.Fatalf("expected year rule for y2024; got %+v", why["y2024"])
	}
}

func TestSelect_FixedNowDeterminism(t *testing.T) {
	loc := time.UTC
	now := ts(2025, 8, 20, 12, 0, 0, loc)

	snaps := []Item{
		{ItemID: "a", Timestamp: ts(2025, 8, 20, 11, 0, 0, loc)},
		{ItemID: "b", Timestamp: ts(2025, 8, 19, 11, 0, 0, loc)},
	}
	crit := Criteria{KeepDays: 1, KeepPerDay: 1}

	// With fixed "now" we should always keep only the representative within that day window.
	p1 := NewPolicy(crit, now)
	p2 := NewPolicy(crit, now)
	k1, _ := p1.Select(snaps)
	k2, _ := p2.Select(snaps)

	if len(k1) != len(k2) {
		t.Fatalf("non-deterministic keep set sizes: %d vs %d", len(k1), len(k2))
	}
	if (contains(k1, "a") && !contains(k2, "a")) || (contains(k1, "b") && !contains(k2, "b")) {
		t.Fatalf("non-deterministic membership: %v vs %v", k1, k2)
	}
}

func TestSelect_OrderIndependence(t *testing.T) {
	loc := time.UTC
	now := ts(2025, 8, 20, 12, 0, 0, loc)

	snaps := []Item{
		{ItemID: "1", Timestamp: ts(2025, 8, 20, 12, 10, 0, loc)},
		{ItemID: "2", Timestamp: ts(2025, 8, 20, 12, 9, 0, loc)},
		{ItemID: "3", Timestamp: ts(2025, 8, 20, 12, 8, 0, loc)},
	}
	crit := Criteria{KeepHours: 1, KeepPerHour: 2}
	p := NewPolicy(crit, now)

	keep1, _ := p.Select(snaps)

	// Shuffle input
	shuffled := append([]Item(nil), snaps...)
	rand.Seed(42)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

	keep2, _ := p.Select(shuffled)

	if len(keep1) != len(keep2) {
		t.Fatalf("order dependence detected (len %d vs %d)", len(keep1), len(keep2))
	}
	for k := range keep1 {
		if !contains(keep2, k) {
			t.Fatalf("order dependence: %s missing after shuffle", k)
		}
	}
}

func TestZeroCapTreatedAsOne(t *testing.T) {
	loc := time.UTC
	now := ts(2025, 8, 20, 12, 0, 0, loc)
	snaps := []Item{
		{ItemID: "a", Timestamp: ts(2025, 8, 20, 12, 10, 0, loc)},
		{ItemID: "b", Timestamp: ts(2025, 8, 20, 12, 9, 0, loc)},
	}
	crit := Criteria{KeepDays: 1, KeepPerDay: 0} // coerced to 1
	p := NewPolicy(crit, now)
	keep, _ := p.Select(snaps)

	if len(keep) != 1 || !contains(keep, "a") {
		t.Fatalf("expected only newest kept due to cap=1 coercion; keep=%v", keep)
	}
}

func TestReasonForDeleteWithinWindowExceedsCap(t *testing.T) {
	loc := time.UTC
	now := ts(2025, 8, 20, 12, 0, 0, loc)
	snaps := []Item{
		{ItemID: "a", Timestamp: ts(2025, 8, 20, 12, 10, 0, loc)},
		{ItemID: "b", Timestamp: ts(2025, 8, 20, 12, 9, 0, loc)},
		{ItemID: "c", Timestamp: ts(2025, 8, 20, 12, 8, 0, loc)},
	}
	crit := Criteria{KeepHours: 1, KeepPerHour: 2}
	p := NewPolicy(crit, now)
	keep, why := p.Select(snaps)

	if len(keep) != 2 {
		t.Fatalf("expected 2 kept, got %d", len(keep))
	}
	if contains(keep, "c") {
		t.Fatalf("did not expect c to be kept")
	}
	r := why["c"]
	if r.Action != "delete" || r.Note != "exceeds per-bucket cap" || r.Rule != RuleHours || r.Rank != 3 || r.Cap != 2 {
		t.Fatalf("unexpected delete reason for c: %+v", r)
	}
}

func TestKeyFormatHelpers(t *testing.T) {
	loc := time.UTC
	tm := ts(2025, 8, 9, 7, 5, 0, loc)

	if got := minuteKey(tm); got != "2025-08-09-07:05" {
		t.Fatalf("minuteKey format mismatch: %s", got)
	}
	if got := hourKey(tm); got != "2025-08-09-07" {
		t.Fatalf("hourKey format mismatch: %s", got)
	}
	if got := dayKey(tm); got != "2025-08-09" {
		t.Fatalf("dayKey format mismatch: %s", got)
	}
	if got := isoWeekKey(tm); got == "" || got[0] != 'W' {
		t.Fatalf("isoWeekKey format mismatch: %s", got)
	}
	if got := monthKey(tm); got != "2025-08" {
		t.Fatalf("monthKey format mismatch: %s", got)
	}
	if got := yearKey(tm); got != "2025" {
		t.Fatalf("yearKey format mismatch: %s", got)
	}
}

// A slightly larger scenario that mixes multiple windows with caps.
func TestMixedScenario(t *testing.T) {
	loc := time.UTC
	now := ts(2025, 8, 20, 12, 0, 0, loc)

	// 3 days, each with 3 hourly entries
	var snaps []Item
	idx := 0
	for d := 20; d >= 18; d-- {
		for h := 10; h >= 8; h-- {
			snaps = append(snaps, Item{
				ItemID:    id(idx),
				Timestamp: ts(2025, 8, d, h, 0, 0, loc),
			})
			idx++
		}
	}
	crit := Criteria{
		KeepDays: 2, KeepPerDay: 1,
		KeepHours: 4, KeepPerHour: 1,
		KeepMonths: 1, KeepPerMonth: 1,
	}
	p := NewPolicy(crit, now)
	keep, _ := p.Select(snaps)

	// Expect at least:
	// - 2 daily reps (for 20, 19)
	// - 1 monthly rep (for Aug)
	// - hourly reps from the last 4 hours (which may overlap with daily reps)
	if len(keep) < 3 {
		t.Fatalf("expected >=3 kept due to union; got %d", len(keep))
	}
}

func BenchmarkSelect_Scale(b *testing.B) {
	loc := time.UTC
	now := ts(2025, 8, 20, 12, 0, 0, loc)

	// Generate 100k items over ~120 days, 35 per day.
	const days = 120
	const perDay = 35
	items := make([]Item, 0, days*perDay)
	idc := 0
	for d := 0; d < days; d++ {
		day := now.AddDate(0, 0, -d)
		for k := 0; k < perDay; k++ {
			items = append(items, Item{
				ItemID:    id(idc),
				Timestamp: day.Add(-time.Duration(k) * time.Minute),
			})
			idc++
		}
	}
	crit := Criteria{
		KeepDays: 30, KeepPerDay: 1,
		KeepWeeks: 6, KeepPerWeek: 1,
		KeepMonths: 3, KeepPerMonth: 2,
		KeepYears: 2, KeepPerYear: 1,
		KeepHours: 12, KeepPerHour: 1,
		KeepMinutes: 10, KeepPerMinute: 1,
	}
	p := NewPolicy(crit, now)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Select(items)
	}
}

func TestUTCOnly(t *testing.T) {
	locParis, _ := time.LoadLocation("Europe/Paris")
	localNow := time.Date(2025, 8, 20, 12, 0, 0, 0, locParis)

	crit := Criteria{KeepDays: 1, KeepPerDay: 1}
	p := NewPolicy(crit, time.Time{}) // should default to time.Now().UTC() internally
	items := []Item{
		{ItemID: "x", Timestamp: localNow}, // non-UTC timestamp
	}
	_, why := p.Select(items)

	// Bucket must be computed as UTC day, not Paris day.
	want := dayKey(localNow.UTC())
	got := why["x"].Bucket
	if got != want {
		t.Fatalf("bucket not UTC: got %s want %s", got, want)
	}
}
