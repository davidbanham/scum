package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/davidbanham/scum/filter"
	"github.com/davidbanham/scum/pagination"
	"github.com/davidbanham/scum/query"
	"github.com/davidbanham/scum/util"
	uuid "github.com/satori/go.uuid"
)

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

func StandardSave(ctx context.Context, table string, cols []string, auditQuery string, props []any) (string, error) {
	db := ctx.Value("tx").(Querier)

	newRev := uuid.NewV4().String()

	posArgs := []string{}
	for i := 4; i < len(cols)+4; i++ {
		posArgs = append(posArgs, fmt.Sprintf("$%d", i))
	}

	if result, err := db.ExecContext(ctx, auditQuery+`
INSERT INTO `+table+` (
	updated_at,
	revision,
	id,
	`+strings.Join(cols, ",")+`
) VALUES (
	now(), $1, $3, `+strings.Join(posArgs, ", ")+`
) ON CONFLICT (id) DO UPDATE SET (
	updated_at,
	revision,
	`+strings.Join(cols, ",")+`
) = (
	now(), $1, `+strings.Join(posArgs, ", ")+`
) WHERE `+table+`.revision = $2`,
		append([]any{
			newRev,
		}, props...)...,
	); err != nil {
		return "", err
	} else {
		if num, affErr := result.RowsAffected(); affErr != nil {
			return "", affErr
		} else if num == 0 {
			return "", ErrWrongRev
		}
	}

	return newRev, nil
}

func FindByColumn(ctx context.Context, table string, cols []string, col string, val any, props []any) error {
	db := ctx.Value("tx").(Querier)

	return db.QueryRowContext(ctx, `
SELECT
	revision,
	id,
	created_at,
	updated_at,
	`+strings.Join(cols, ",")+`
FROM `+table+` WHERE `+col+` = $1`, val).Scan(props...)
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
