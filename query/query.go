package query

import (
	"fmt"
	"strings"

	"github.com/davidbanham/scum/filter"
	"github.com/davidbanham/scum/pagination"
	"github.com/davidbanham/scum/util"
	"github.com/lib/pq"
)

type Query interface {
	Construct(columns []string, table string, filters filter.Filters, pagination pagination.Pagination, order Order) string
	Args() []any
}

type Order struct {
	Desc bool
	By   string
}

func (this Order) String(whitelist []string) string {
	direction := "ASC"
	if this.Desc {
		direction = "DESC"
	}
	if util.Contains(whitelist, this.By) {
		return fmt.Sprintf("ORDER BY %s %s", this.By, direction)
	}
	return ""
}

type All struct {
	props []any
}

func (this *All) Construct(columns []string, table string, filters filter.Filters, pagination pagination.Pagination, order Order) string {
	filterQuery, filterProps := filters.Query(1)
	this.props = append(this.props, filterProps...)

	return fmt.Sprintf(`SELECT %s FROM %s %s %s %s`, strings.Join(columns, ", "), table, filterQuery, order.String(columns), pagination.PaginationQuery())
}
func (this *All) Args() []any {
	return this.props
}

type ByOrg struct {
	ID    string
	props []any
}

func (this *ByOrg) Construct(columns []string, table string, filters filter.Filters, pagination pagination.Pagination, order Order) string {
	filterQuery, filterProps := filters.Query(2)
	this.props = append(this.props, filterProps...)

	tableOnly := strings.Split(table, " ")[0]

	return fmt.Sprintf(`SELECT %s FROM %s %s AND %s.organisation_id = $1 %s %s`, strings.Join(columns, ", "), table, filterQuery, tableOnly, order.String(columns), pagination.PaginationQuery())
}
func (this *ByOrg) Args() []any {
	return append([]any{this.ID}, this.props...)
}

type ByUser struct {
	ID    string
	props []any
}

func (this *ByUser) Construct(columns []string, table string, filters filter.Filters, pagination pagination.Pagination, order Order) string {
	filterQuery, filterProps := filters.Query(2)
	this.props = append(this.props, filterProps...)

	tableOnly := strings.Split(table, " ")[0]

	return fmt.Sprintf(`SELECT %s FROM %s %s AND %s.user_id = $1 %s %s`, strings.Join(columns, ", "), table, filterQuery, tableOnly, order.String(columns), pagination.PaginationQuery())
}
func (this *ByUser) Args() []any {
	return append([]any{this.ID}, this.props...)
}

type ByIDs struct {
	IDs   []string
	props []any
}

func (this *ByIDs) Construct(columns []string, table string, filters filter.Filters, pagination pagination.Pagination, order Order) string {
	filterQuery, filterProps := filters.Query(2)
	this.props = append(this.props, filterProps...)

	tableOnly := strings.Split(table, " ")[0]

	return fmt.Sprintf(`SELECT %s FROM %s %s AND %s.id = ANY($1) %s %s`, strings.Join(columns, ", "), table, filterQuery, tableOnly, order.String(columns), pagination.PaginationQuery())
}
func (this *ByIDs) Args() []any {
	return append([]any{pq.Array(this.IDs)}, this.props...)
}
