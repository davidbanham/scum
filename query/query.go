package query

import (
	"fmt"
	"strings"

	"github.com/davidbanham/scum/filter"
	"github.com/davidbanham/scum/pagination"
	"github.com/lib/pq"
)

type Query interface {
	Construct(columns []string, table string, filters filter.Filters, pagination pagination.Pagination, order string) string
	Args() []any
}

type All struct {
	props []any
}

func (this *All) Construct(columns []string, table string, filters filter.Filters, pagination pagination.Pagination, order string) string {
	filterQuery, filterProps := filters.Query(1)
	this.props = append(this.props, filterProps...)

	return fmt.Sprintf(`SELECT %s FROM %s %s ORDER BY %s %s`, strings.Join(columns, ", "), table, filterQuery, order, pagination.PaginationQuery())
}
func (this *All) Args() []any {
	return this.props
}

type ByOrg struct {
	ID    string
	props []any
}

func (this *ByOrg) Construct(columns []string, table string, filters filter.Filters, pagination pagination.Pagination, order string) string {
	filterQuery, filterProps := filters.Query(2)
	this.props = append(this.props, filterProps...)

	tableOnly := strings.Split(table, " ")[0]

	return fmt.Sprintf(`SELECT %s FROM %s %s AND %s.organisation_id = $1 ORDER BY %s %s`, strings.Join(columns, ", "), table, filterQuery, tableOnly, order, pagination.PaginationQuery())
}
func (this *ByOrg) Args() []any {
	return append([]any{this.ID}, this.props...)
}

type ByUser struct {
	ID    string
	props []any
}

func (this *ByUser) Construct(columns []string, table string, filters filter.Filters, pagination pagination.Pagination, order string) string {
	filterQuery, filterProps := filters.Query(2)
	this.props = append(this.props, filterProps...)

	tableOnly := strings.Split(table, " ")[0]

	return fmt.Sprintf(`SELECT %s FROM %s %s AND %s.user_id = $1 ORDER BY %s %s`, strings.Join(columns, ", "), table, filterQuery, tableOnly, order, pagination.PaginationQuery())
}
func (this *ByUser) Args() []any {
	return append([]any{this.ID}, this.props...)
}

type ByIDs struct {
	IDs   []string
	props []any
}

func (this *ByIDs) Construct(columns []string, table string, filters filter.Filters, pagination pagination.Pagination, order string) string {
	filterQuery, filterProps := filters.Query(2)
	this.props = append(this.props, filterProps...)

	tableOnly := strings.Split(table, " ")[0]

	return fmt.Sprintf(`SELECT %s FROM %s %s AND %s.id = ANY($1) ORDER BY %s %s`, strings.Join(columns, ", "), table, filterQuery, tableOnly, order, pagination.PaginationQuery())
}
func (this *ByIDs) Args() []any {
	return append([]any{pq.Array(this.IDs)}, this.props...)
}
