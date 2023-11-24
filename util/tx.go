package util

import (
	"context"
	"database/sql"
	"fmt"
)

func GetTxCtx(db *sql.DB) (context.Context, *sql.Tx, error) {
	var tx *sql.Tx
	var err error

	ctx := context.Background()

	tx, err = db.BeginTx(ctx, nil)
	if err != nil {
		return ctx, nil, err
	}
	tx.ExecContext(ctx, "SET application_name = 'system_user'")

	ctx = context.WithValue(ctx, "tx", tx)

	return ctx, tx, err
}

func RollbackTx(ctx context.Context) error {
	tx := ctx.Value("tx")
	switch v := tx.(type) {
	case *sql.Tx:
		rollbackErr := v.Rollback()
		if rollbackErr != nil {
			return fmt.Errorf("Error rolling back tx: %w", rollbackErr)
		}
	default:
		//fmt.Printf("DEBUG no transaction on context\n")
	}
	return nil
}
