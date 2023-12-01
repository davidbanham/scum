package util

import "time"

func CombineDateTimePair(d, t time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

func ParseDateTimePair(d, t string, loc *time.Location) (time.Time, error) {
	parsedDate, err := time.ParseInLocation(time.DateOnly, d, loc)
	if err != nil {
		return time.Time{}, err
	}
	timeFormat := time.TimeOnly
	if len(t) == 5 {
		timeFormat = "15:04"
	}
	parsedTime, err := time.ParseInLocation(timeFormat, t, loc)
	if err != nil {
		return time.Time{}, err
	}
	return CombineDateTimePair(parsedDate, parsedTime).In(loc), nil
}
