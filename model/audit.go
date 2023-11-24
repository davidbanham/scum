package model

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/davidbanham/scum/query"
	"github.com/kylelemons/godebug/diff"
)

type Audit struct {
	ID              string
	EntityID        string
	OrganisationID  string
	TableName       string
	Stamp           time.Time
	UserID          string
	UserName        string
	Action          string
	OldRowData      string
	maybeOldRowData sql.NullString
	NewRowData      string
	maybeNewRowData sql.NullString
	Diff            string
}

type Audits struct {
	Data     []Audit
	Criteria Criteria
}

func (this *Audits) FindAll(ctx context.Context, criteria Criteria) error {
	this.Criteria = criteria

	db := ctx.Value("tx").(Querier)

	var rows *sql.Rows
	var err error

	cols := append([]string{
		"audit_log.id",
		"entity_id",
		"organisation_id",
		"table_name",
		"stamp",
		"user_id",
		"action",
		"old_row_data - 'revision' - 'updated_at'",
		"users.email",
		"lead(old_row_data - 'revision' - 'updated_at', 1) OVER (PARTITION BY entity_id ORDER BY stamp) new_row_data",
	})

	switch v := criteria.Query.(type) {
	case query.Query:
		rows, err = db.QueryContext(ctx, v.Construct(cols, "audit_log LEFT JOIN users ON audit_log.user_id = users.id::text", criteria.Filters, criteria.Pagination, "stamp"), v.Args()...)
	}
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		audit := Audit{}
		maybeUserName := sql.NullString{}
		if err := rows.Scan(
			&audit.ID,
			&audit.EntityID,
			&audit.OrganisationID,
			&audit.TableName,
			&audit.Stamp,
			&audit.UserID,
			&audit.Action,
			&audit.maybeOldRowData,
			&maybeUserName,
			&audit.maybeNewRowData,
		); err != nil {
			return err
		}
		audit.OldRowData = "{}"
		if audit.maybeOldRowData.Valid {
			audit.OldRowData = audit.maybeOldRowData.String
		}
		audit.NewRowData = "{}"
		if audit.maybeNewRowData.Valid {
			audit.NewRowData = audit.maybeNewRowData.String
		}
		audit.UserName = audit.UserID
		if maybeUserName.Valid {
			audit.UserName = maybeUserName.String
		}

		(*this).Data = append((*this).Data, audit)
	}

	for i, audit := range (*this).Data {
		if !audit.maybeNewRowData.Valid && audit.Action != "D" {
			if err := db.QueryRowContext(ctx, `SELECT to_jsonb(`+audit.TableName+`) - 'ts' - 'revision' - 'updated_at' FROM `+audit.TableName+` WHERE id = $1`, audit.EntityID).Scan(&audit.NewRowData); err != nil && err != sql.ErrNoRows {
				return err
			}
		}

		if audit.Action == "D" {
			audit.Diff = "Deleted"
		} else if audit.maybeOldRowData.Valid {
			audit.OldRowData = prettyJsonString(audit.OldRowData)
			audit.NewRowData = prettyJsonString(audit.NewRowData)

			audit.Diff = diffOnly(diff.Diff(audit.OldRowData, audit.NewRowData))

		} else {
			audit.Diff = "Created"
		}

		(*this).Data[i] = audit
	}
	return err
}

func prettyJsonString(input string) string {
	var out bytes.Buffer
	json.Indent(&out, []byte(input), "", "  ")
	return out.String()
}

func diffOnly(input string) string {
	parts := strings.Split(input, "\n")
	relevant := []string{}
	for _, part := range parts {
		if strings.Index(part, "+") == 0 || strings.Index(part, "-") == 0 {
			if string(part[len(part)-1]) == "," {
				relevant = append(relevant, part[1:len(part)-1])
			} else {
				relevant = append(relevant, part[1:])
			}
		}
	}
	pairs := []string{}
	hold := ""
	for _, part := range relevant {
		if hold == "" {
			hold = part
		} else {
			sep := strings.Index(part, ":")
			pairs = append(pairs, fmt.Sprintf("%s -> %s", hold, part[sep+1:]))
			hold = ""
		}
	}
	if hold != "" {
		pairs = append(pairs, hold)
	}
	return strings.Join(pairs, " ")
}
