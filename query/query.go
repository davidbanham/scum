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

type All struct{}

func (All) Construct(columns []string, table string, filters filter.Filters, pagination pagination.Pagination, order string) string {
	return fmt.Sprintf(`SELECT %s FROM %s %s ORDER BY %s %s`, strings.Join(columns, ", "), table, filters.Query(), order, pagination.PaginationQuery())
}
func (All) Args() []any {
	return []any{}
}

type ByOrg struct {
	ID string
}

func (this ByOrg) Construct(columns []string, table string, filters filter.Filters, pagination pagination.Pagination, order string) string {
	return fmt.Sprintf(`SELECT %s FROM %s %s AND organisation_id = $1 ORDER BY %s %s`, strings.Join(columns, ", "), table, filters.Query(), order, pagination.PaginationQuery())
}
func (this ByOrg) Args() []any {
	return []any{this.ID}
}

type ByUser struct {
	ID string
}

func (this ByUser) Construct(columns []string, table string, filters filter.Filters, pagination pagination.Pagination, order string) string {
	return fmt.Sprintf(`SELECT %s FROM %s %s AND user_id = $1 ORDER BY %s %s`, strings.Join(columns, ", "), table, filters.Query(), order, pagination.PaginationQuery())
}
func (this ByUser) Args() []any {
	return []any{this.ID}
}

type ByIDs struct {
	IDs []string
}

func (this ByIDs) Construct(columns []string, table string, filters filter.Filters, pagination pagination.Pagination, order string) string {
	return fmt.Sprintf(`SELECT %s FROM %s %s AND id = ANY($1) ORDER BY %s %s`, strings.Join(columns, ", "), table, filters.Query(), order, pagination.PaginationQuery())
}
func (this ByIDs) Args() []any {
	return []any{pq.Array(this.IDs)}
}

type ByEntityID struct {
	EntityID string
}

func (this ByEntityID) Construct(columns []string, table string, filters filter.Filters, pagination pagination.Pagination, order string) string {
	return fmt.Sprintf(`SELECT %s FROM %s %s AND entity_id = ($1) ORDER BY %s %s`, strings.Join(columns, ", "), table, filters.Query(), order, pagination.PaginationQuery())
}
func (this ByEntityID) Args() []any {
	return []any{this.EntityID}
}
