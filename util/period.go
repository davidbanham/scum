package util

import "time"

type Period struct {
	Start time.Time
	End   time.Time
}

func (this Period) Next() Period {
	dur := this.End.Sub(this.Start) + (24 * time.Hour)
	ret := Period{
		Start: this.Start.Add(dur),
		End:   this.End.Add(dur),
	}
	return ret
}

func (this Period) Prev() Period {
	dur := this.End.Sub(this.Start) + (24 * time.Hour)
	ret := Period{
		Start: this.Start.Add(-dur),
		End:   this.End.Add(-dur),
	}
	return ret
}
