package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/davidbanham/scum/filter"
	"github.com/davidbanham/scum/pagination"
	"github.com/davidbanham/scum/query"
	"github.com/davidbanham/scum/util"
	uuid "github.com/satori/go.uuid"
)

type Colmap map[string]any

func (this *Colmap) Delete(keys ...string) *Colmap {
	ret := &Colmap{}
	for k, v := range *this {
		skip := false
		for _, key := range keys {
			if k == key {
				skip = true
			}
		}
		if !skip {
			(*ret)[k] = v
		}
	}
	return ret
}

func (this *Colmap) Strip(table string) *Colmap {
	ret := &Colmap{}
	for k, v := range *this {
		bits := strings.SplitN(k, ".", 2)
		if len(bits) == 2 && bits[0] != table {
			continue
		}
		(*ret)[bits[len(bits)-1]] = v
	}
	return ret
}

func (this *Colmap) Split() (cols []string, props []any) {
	for k, v := range *this {
		cols = append(cols, k)
		props = append(props, v)
	}
	return
}

func (this *Colmap) ByKeys(keys []string) (props []any) {
	for _, k := range keys {
		props = append(props, (*this)[k])
	}
	return
}

type Model interface {
	Save(context.Context) error
	FindByID(context.Context, string) error
	FindByColumn(context.Context, string, string) error
	NullDynamicValues()
	Blank() Model
	Id() string
	Tablename() string
	Query(Criteria) string
}

func StandardSave(table string, colmap *Colmap, auditQuery string) (string, []any, string) {
	stripped := colmap.Strip(table)
	if _, ok := (*stripped)["updated_at"]; ok {
		(*stripped)["updated_at"] = time.Now()
	}
	cols, props := stripped.Delete("revision").Split()

	posArgs := []string{}
	for i := range cols {
		posArgs = append(posArgs, fmt.Sprintf("$%d", i+3))
	}

	ret := auditQuery + `
INSERT INTO ` + table + ` (
	revision, ` + strings.Join(cols, ", ") + `
) VALUES (
	$1, ` + strings.Join(posArgs, ", ") + `
) ON CONFLICT (id) DO UPDATE SET (
	revision, ` + strings.Join(cols, ", ") + `
) = (
	$1, ` + strings.Join(posArgs, ", ") + `
) WHERE ` + table + `.revision = $2`

	newRev := uuid.NewV4().String()

	return ret, append([]any{newRev, (*stripped)["revision"]}, props...), newRev
}

func ExecSave(ctx context.Context, query string, props []any) error {
	db := ctx.Value("tx").(Querier)

	if result, err := db.ExecContext(ctx, query, props...); err != nil {
		return err
	} else {
		if num, affErr := result.RowsAffected(); affErr != nil {
			return affErr
		} else if num == 0 {
			return ErrWrongRev
		}
	}
	return nil
}

func FindByColumn(table string, colmap *Colmap, col string) (string, []any) {
	cols := []string{}
	props := []any{}
	for k, v := range *colmap {
		cols = append(cols, k)
		props = append(props, v)
	}

	return `
SELECT
	` + strings.Join(cols, ", ") + `
FROM ` + table + ` WHERE ` + col + ` = $1`, props
}

func ExecFindByColumn(ctx context.Context, query string, val any, props []any) error {
	db := ctx.Value("tx").(Querier)

	return db.QueryRowContext(ctx, query, val).Scan(props...)
}

type Criteria struct {
	Query      query.Query
	Filters    filter.Filters
	Pagination pagination.Pagination
}

type Querier interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

var ErrWrongRev = util.ClientSafeError{Message: "This record has been changed by another request since you loaded it. Review the changes by going back and refreshing, and try again if appropriate."}
