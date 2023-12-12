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

type Gettable interface {
	// For example, url.Values
	Get(string) string
}

func DefaultedDatesFromForm(form Gettable, numDaysFromNowDefault int, weekBoundary bool) (startTime, endTime time.Time, err error) {
	start := form.Get("start")
	end := form.Get("end")

	format := "2006-01-02"
	begin := time.Now()

	if weekBoundary {
		begin = NextDay(begin, time.Sunday)
	}

	now := begin.Format(format)
	then := begin.Add(24 * time.Duration(numDaysFromNowDefault) * time.Hour).Format(format)

	startTime, _ = time.Parse(format, now)
	endTime, _ = time.Parse(format, then)

	if start != "" {
		parsed, err := time.Parse(format, start)
		if err != nil {
			return startTime, endTime, err
		}
		startTime = parsed
	}
	if end != "" {
		parsed, err := time.Parse(format, end)
		if err != nil {
			return startTime, endTime, err
		}
		endTime = parsed
	}

	return startTime, endTime, nil
}
