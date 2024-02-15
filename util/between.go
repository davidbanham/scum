package util

import (
	"time"
)

func IsBetween(target, start, end time.Time) bool {
	return target.After(start.Add(-time.Minute)) && target.Before(end)
}

func DaysBetween(start, end time.Time) []time.Time {
	allNights := []time.Time{}
	for d := start.Round(24 * time.Hour); d.Before(end); d = d.AddDate(0, 0, 1) {
		allNights = append(allNights, d)
	}
	return allNights
}

type Interval struct {
	Start time.Time
	End   time.Time
}

func Overlap(one, two Interval) bool {
	if IsBetween(one.Start, two.Start, two.End) {
		return true
	}
	if IsBetween(one.End, two.Start, two.End) {
		return true
	}
	if IsBetween(two.Start, one.Start, one.End) {
		return true
	}
	if IsBetween(two.End, one.Start, one.End) {
		return true
	}
	return false
}
