package policy

import (
	"sort"
	"time"
)

const (
	RuleMinutes = "minutes"
	RuleHours   = "hours"
	RuleDays    = "days"
	RuleWeeks   = "weeks"
	RuleMonths  = "months"
	RuleYears   = "years"
)

type Reason struct {
	Action string // "keep" or "delete"
	Rule   string // minutes/hours/days/weeks/months/years; empty if outside windows
	Bucket string // "2025-08-20" or "2025-08-20-14" or "W2025-34"
	Rank   int    // position within the bucket, newest-first, used for cap
	Cap    int    // cap applied for that bucket in the rule (max items per bucket)
	Note   string // human message: "outside retention windows"
}

type Item struct {
	ItemID    string
	Timestamp time.Time
}

type Criteria struct {
	// buckets
	KeepMinutes, KeepHours, KeepDays, KeepWeeks, KeepMonths, KeepYears int

	// cap
	KeepPerMinute, KeepPerHour, KeepPerDay, KeepPerWeek, KeepPerMonth, KeepPerYear int
}

type Policy struct {
	crit Criteria
	now  time.Time
}

func NewPolicy(c Criteria, now time.Time) *Policy {
	return &Policy{crit: c, now: now}
}

func (p *Policy) Criteria() Criteria { return p.crit }
func (p *Policy) Now() time.Time {
	if !p.now.IsZero() {
		return p.now.UTC()
	}
	return time.Now().UTC()
}

// Select computes the keep set and a reason for every snapshot.
//   - Input is arbitrary order; internally sorted newest-first.
//   - A snapshot kept by ANY rule is kept.
//   - Non-kept snapshots get the best available delete-reason (nearest miss),
//     or "outside retention windows" when not considered by any window.
//
// Returns:
//
//	keep:   map[ID]struct{} of snapshots to keep
//	reasons: map[ID]Reason explaining the decision for each snapshot
func (p *Policy) Select(snaps []Item) (map[string]struct{}, map[string]Reason) {
	now := p.Now()

	// copy + sort newest-first
	items := make([]Item, len(snaps))
	for i := range snaps {
		items[i] = snaps[i]
		items[i].Timestamp = items[i].Timestamp.UTC()
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Timestamp.After(items[j].Timestamp) })

	kept := make(map[string]struct{}, len(items))
	reasons := make(map[string]Reason, len(items))
	ruleKeepReasons := make(map[string]Reason, len(items)) // per-snapshot best keep reason
	ruleDropReasons := make(map[string]Reason, len(items)) // per-snapshot best delete reason

	// helper to run a single rule (minutes/hours/...)
	processRule := func(rule string, windowKeys map[string]struct{}, keyFn func(time.Time) string, capPer int) {
		if len(windowKeys) == 0 || capPer < 1 {
			return
		}
		// bucket indices (items already newest-first)
		buckets := make(map[string][]int)
		for idx, s := range items {
			k := keyFn(s.Timestamp)
			if _, ok := windowKeys[k]; ok {
				buckets[k] = append(buckets[k], idx)
			}
		}
		// assign reasons (top K kept; rest exceed cap)
		for bkey, idxs := range buckets {
			for rank, idx := range idxs {
				s := items[idx]
				id := s.ItemID
				if rank < capPer {
					r := Reason{
						Action: "keep", Rule: rule, Bucket: bkey,
						Rank: rank + 1, Cap: capPer,
					}
					if prev, ok := ruleKeepReasons[id]; !ok || r.Rank < prev.Rank {
						ruleKeepReasons[id] = r
					}
				} else {
					r := Reason{
						Action: "delete", Rule: rule, Bucket: bkey,
						Rank: rank + 1, Cap: capPer, Note: "exceeds per-bucket cap",
					}
					if prev, ok := ruleDropReasons[id]; !ok || r.Rank < prev.Rank {
						ruleDropReasons[id] = r
					}
				}
			}
		}
	}

	// windows
	c := p.crit
	if c.KeepMinutes > 0 {
		processRule(RuleMinutes, lastNMinutesKeys(now, c.KeepMinutes), minuteKey, capOr1(c.KeepPerMinute))
	}
	if c.KeepHours > 0 {
		processRule(RuleHours, lastNHoursKeys(now, c.KeepHours), hourKey, capOr1(c.KeepPerHour))
	}
	if c.KeepDays > 0 {
		processRule(RuleDays, lastNDaysKeys(now, c.KeepDays), dayKey, capOr1(c.KeepPerDay))
	}
	if c.KeepWeeks > 0 {
		processRule(RuleWeeks, lastNISOWeeksKeys(now, c.KeepWeeks), isoWeekKey, capOr1(c.KeepPerWeek))
	}
	if c.KeepMonths > 0 {
		processRule(RuleMonths, lastNMonthsKeys(now, c.KeepMonths), monthKey, capOr1(c.KeepPerMonth))
	}
	if c.KeepYears > 0 {
		processRule(RuleYears, lastNYearsKeys(now, c.KeepYears), yearKey, capOr1(c.KeepPerYear))
	}

	// finalize decision for each snapshot
	for _, s := range items {
		id := s.ItemID
		if kr, ok := ruleKeepReasons[id]; ok {
			kept[id] = struct{}{}
			reasons[id] = kr
			continue
		}
		if dr, ok := ruleDropReasons[id]; ok {
			reasons[id] = dr
		} else {
			reasons[id] = Reason{
				Action: "delete", Note: "outside retention windows",
			}
		}
	}
	return kept, reasons
}
