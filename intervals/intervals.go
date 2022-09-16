// Time interval sums accounting for overlap.
package intervals

import (
	"fmt"
	"time"
)

type Interval struct {
	Start time.Time
	End   time.Time
}

type TimeTracker struct {
	noOverlap []Interval
}

func (tt *TimeTracker) Track(i Interval) error {
	if i.End.Before(i.Start) {
		return fmt.Errorf("Invalid negative interval: %v .. %v", i.Start, i.End)
	}

	type noOverlap struct {
		intervals []Interval
	}

	single := func(i Interval) noOverlap {
		return noOverlap{[]Interval{i}}
	}

	prepend := func(i Interval, rest noOverlap) noOverlap {
		return noOverlap{append(single(i).intervals, rest.intervals...)}
	}

	merge := func(a, b Interval) (Interval, bool) {
		if a.End.Before(b.Start) || b.End.Before(a.Start) {
			return Interval{}, false
		}
		i := Interval{
			Start: a.Start,
			End:   a.End,
		}
		if b.Start.Before(a.Start) {
			i.Start = b.Start
		}
		if b.End.After(i.End) {
			i.End = b.End
		}
		return i, true
	}

	var mergeInto func(acc noOverlap, i Interval) noOverlap
	mergeInto = func(acc noOverlap, i Interval) noOverlap {
		if len(acc.intervals) == 0 {
			return single(i)
		}

		head := acc.intervals[0]
		tail := noOverlap{acc.intervals[1:]}

		if m, ok := merge(head, i); ok {
			return mergeInto(tail, m)
		}

		mtail := mergeInto(tail, i)

		if len(mtail.intervals) < len(tail.intervals)+1 {
			return mergeInto(mtail, head)
		}

		return prepend(i, acc)
	}

	tt.noOverlap = mergeInto(noOverlap{tt.noOverlap}, i).intervals
	return nil
}

func (tt *TimeTracker) TimeTaken() time.Duration {
	var total time.Duration
	for _, i := range tt.noOverlap {
		total += i.End.Sub(i.Start)
	}
	return total
}
