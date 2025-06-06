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
	Query(int) (string, []any)
	ID() string
	Populate(url.Values) error
	Inputs() []string
	TableName() string
	InputValues() []string
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

func (this *filterBase) InputValues() []string {
	return []string{}
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

func (this dateBase) Query(propIndex int) (string, []any) {
	return fmt.Sprintf("%s.%s BETWEEN $%d AND $%d", this.table, this.col, propIndex, propIndex+1), []any{this.period.Start, this.period.End}
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
	tz := form.Get(prefix + "-tz")
	var loc *time.Location
	if tz != "" {
		loc, _ = time.LoadLocation(tz)
	}
	if start == "" || end == "" {
		return nil
	}
	if loc != nil {
		s, err := time.ParseInLocation("2006-01-02", start, loc)
		if err != nil {
			return err
		}
		e, err := time.ParseInLocation("2006-01-02", end, loc)
		if err != nil {
			return err
		}
		this.period.Start = s
		this.period.End = e
	} else {
		s, err := time.Parse("2006-01-02", start)
		if err != nil {
			return err
		}
		e, err := time.Parse("2006-01-02", end)
		if err != nil {
			return err
		}
		this.period.Start = s
		this.period.End = e
	}
	if this.period.End == this.period.Start {
		this.period.End = this.period.End.AddDate(0, 0, 1)
	}
	return nil
}

type Filters []Filter

func (filters Filters) ByID(id string) Filter {
	for _, filter := range filters {
		if filter.ID() == id {
			return filter
		}
	}
	return &invalidFilter{}
}

func (filters Filters) Active(id string) bool {
	for _, filter := range filters {
		if filter.ID() == id {
			return true
		}
	}
	return false
}

func (filters Filters) Except(except ...string) Filters {
	ret := Filters{}
	for _, f := range filters {
		if !util.Contains(except, f.ID()) {
			ret = append(ret, f)
		}
	}
	return ret
}

func (filters Filters) Only(except ...string) Filters {
	ret := Filters{}
	for _, f := range filters {
		if util.Contains(except, f.ID()) {
			ret = append(ret, f)
		}
	}
	return ret
}

func (filters Filters) IDMap() map[string]Filter {
	ret := map[string]Filter{}
	for _, filter := range filters {
		ret[filter.ID()] = filter
	}
	return ret
}

func (filters Filters) Query(propIndex int) (string, []any) {
	if len(filters) == 0 {
		return " WHERE true = true ", []any{}
	}
	fragments := []string{}
	props := []any{}
	for _, filter := range filters {
		q, p := filter.Query(propIndex)
		propIndex += len(p)
		fragments = append(fragments, q)
		props = append(props, p...)
	}
	return " WHERE " + strings.Join(fragments, " AND "), props
}

func (filters *Filters) FromForm(form url.Values, availableFilters Filters, customFilters ...Filter) error {
	activeFilters := Filters{}

	availableFiltersByID := availableFilters.IDMap()
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

func (this Filters) Values() url.Values {
	ret := url.Values{}
	for _, filter := range this {
		ret[filter.ID()] = append(ret[filter.ID()], filter.InputValues()...)
	}
	return ret
}

type invalidFilter struct {
	filterBase
}

func (this invalidFilter) Query(int) (string, []any) {
	return "true = false", []any{}
}

type HasProp struct {
	hydrated bool
	filterBase
	col   string
	value string
}

func (this HasProp) Query(propIndex int) (string, []any) {
	return fmt.Sprintf("%s.%s = $%d", this.table, this.col, propIndex), []any{this.value}
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

func (this Custom) InputValues() []string {
	return this.Values
}

func (this Custom) Query(propIndex int) (string, []any) {
	if len(this.Values) == 0 {
		return "true = true", []any{}
	}
	props := []any{}
	placeholders := []string{}
	for i, val := range this.Values {
		props = append(props, val)
		placeholders = append(placeholders, fmt.Sprintf("$%d", propIndex+i))
	}
	return fmt.Sprintf("%s::text = ANY (ARRAY[%s])", this.Col, strings.Join(placeholders, ", ")), props
}

func (this *Custom) Populate(form url.Values) error {
	this.Values = form[this.CustomID]
	return nil
}

func (this Custom) Label() string {
	return this.CustomLabel
}

func (this Custom) ID() string {
	return this.CustomID
}

func (this Custom) Inputs() []string {
	return []string{"hidden"}
}

type FilterSet struct {
	filterBase
	Filters     Filters
	CustomID    string
	CustomLabel string
	IsAnd       bool
	Values      []string
}

func (this FilterSet) InputValues() []string {
	return this.Values
}

func (this FilterSet) Query(propIndex int) (string, []any) {
	queries := []string{}
	props := []any{}
	for _, filter := range this.Filters {
		q, p := filter.Query(propIndex)
		propIndex += len(p)
		queries = append(queries, q)
		props = append(props, p...)
	}
	operator := " OR "
	if this.IsAnd {
		operator = " AND "
	}
	return fmt.Sprintf("(%s)", strings.Join(queries, operator)), props
}

func (this FilterSet) Label() string {
	return this.CustomLabel
}

func (this FilterSet) ID() string {
	return this.CustomID
}

func (this FilterSet) Inputs() []string {
	return []string{"hidden"}
}
