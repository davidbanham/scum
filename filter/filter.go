package filter

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/davidbanham/scum/util"
)

type Filter interface {
	Label() string
	Query() string
	ID() string
	Populate(url.Values) error
	Inputs() []string
	TableName() string
}

type DateFilterOpts struct {
	Label  string
	ID     string
	Table  string
	Col    string
	Period util.Period
}

type DateFilter struct {
	id string
	dateBase
}

func (this *DateFilter) Populate(form url.Values) error {
	return this.dateBase.Populate(this.ID(), form)
}

type filterBase struct {
	label string
	id    string
	table string
}

func (this *filterBase) Populate(url.Values) error {
	return nil
}

func (this filterBase) TableName() string {
	return this.table
}

func (this filterBase) Label() string {
	return this.label
}

func (this filterBase) ID() string {
	return this.id
}

func (this filterBase) Inputs() []string {
	return []string{}
}

var ErrorAlreadyHydrated = fmt.Errorf("Filter already hydrated. This is a once-only operation to reduce the risk of injection attacks.")

type dateBase struct {
	filterBase
	hydrated bool
	table    string
	col      string
	period   util.Period
}

func (this dateBase) Query() string {
	start := strings.Replace(this.period.Start.Format(time.RFC3339), "T", " ", 1)
	end := strings.Replace(this.period.End.Format(time.RFC3339), "T", " ", 1)
	return fmt.Sprintf("%s.%s BETWEEN '%s' AND '%s'", this.table, this.col, start, end)
}

func (this dateBase) Inputs() []string {
	return []string{"start_end_date"}
}

func (this dateBase) Period() util.Period {
	return this.period
}

func (this *dateBase) Hydrate(opts DateFilterOpts) error {
	if this.hydrated {
		return ErrorAlreadyHydrated
	}
	this.id = opts.ID
	this.label = opts.Label
	this.table = opts.Table
	this.col = opts.Col
	this.period = opts.Period
	this.hydrated = true

	return nil
}

func (this *dateBase) Populate(prefix string, form url.Values) error {
	start := form.Get(prefix + "-start")
	end := form.Get(prefix + "-end")
	if start == "" || end == "" {
		return nil
	}
	s, err := time.Parse("2006-01-02", start)
	if err != nil {
		return err
	}
	e, err := time.Parse("2006-01-02", end)
	if err != nil {
		return err
	}
	p := this.period
	p.Start = s
	p.End = e
	this.period = p
	return nil
}

type Filters []Filter

func (filters Filters) ByID() map[string]Filter {
	ret := map[string]Filter{}
	for _, filter := range filters {
		ret[filter.ID()] = filter
	}
	return ret
}

func (filters Filters) Query() string {
	if len(filters) == 0 {
		return " WHERE true = true "
	}
	fragments := []string{}
	for _, filter := range filters {
		fragments = append(fragments, filter.Query())
	}
	return " WHERE " + strings.Join(fragments, " AND ")
}

func (filters *Filters) FromForm(form url.Values, availableFilters Filters, customFilters ...Filter) error {
	activeFilters := Filters{}

	availableFiltersByID := append(availableFilters, customFilters...).ByID()
	for _, k := range form["filter"] {
		f, ok := availableFiltersByID[k]
		if ok {
			if err := f.Populate(form); err != nil {
				return err
			}
			activeFilters = append(activeFilters, f)
		}
	}
	for _, cf := range customFilters {
		if err := cf.Populate(form); err != nil {
			return err
		}
		activeFilters = append(activeFilters, cf)
	}
	(*filters) = append((*filters), activeFilters...)
	return nil
}

type HasProp struct {
	hydrated bool
	filterBase
	col   string
	value string
}

func (this HasProp) Query() string {
	return fmt.Sprintf("%s.%s = '%s'", this.table, this.col, this.value)
}

type HasPropOpts struct {
	Label string
	ID    string
	Table string
	Col   string
	Value string
}

func (this *HasProp) Hydrate(opts HasPropOpts) error {
	if this.hydrated {
		return ErrorAlreadyHydrated
	}
	this.id = opts.ID
	this.label = opts.Label
	this.table = opts.Table
	this.col = opts.Col
	this.value = opts.Value
	this.hydrated = true

	return nil
}

type Custom struct {
	filterBase
	Col         string
	Values      []string
	CustomID    string
	CustomLabel string
}

func (this Custom) Query() string {
	vals := []string{}
	for _, val := range this.Values {
		vals = append(vals, fmt.Sprintf("'%s'", val))
	}
	return fmt.Sprintf("%s::text = ANY (ARRAY[%s])", this.Col, strings.Join(vals, ", "))
}

func (this Custom) Label() string {
	return this.CustomLabel
}

func (this Custom) ID() string {
	return this.CustomID
}

type FilterSet struct {
	filterBase
	Filters     Filters
	CustomID    string
	CustomLabel string
	IsAnd       bool
}

func (this FilterSet) Query() string {
	queries := []string{}
	for _, filter := range this.Filters {
		queries = append(queries, filter.Query())
	}
	operator := " OR "
	if this.IsAnd {
		operator = " AND "
	}
	return fmt.Sprintf("(%s)", strings.Join(queries, operator))
}

func (this FilterSet) Label() string {
	return this.CustomLabel
}

func (this FilterSet) ID() string {
	return this.CustomID
}
