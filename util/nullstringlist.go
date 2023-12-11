package util

import (
	"database/sql/driver"
	"fmt"

	"github.com/lib/pq"
)

type NullStringList struct {
	Valid   bool
	Strings []string
}

func (n *NullStringList) Scan(value interface{}) error {
	if value == nil {
		n.Strings, n.Valid = []string{}, false
		return nil
	}
	n.Valid = true
	switch value.(type) {
	default:
		return fmt.Errorf("Not a list of strings")
	case []byte:
		return pq.Array(&n.Strings).Scan(value)
	case []string:
		n.Strings = value.([]string)
		return nil
	}
}

// Value implements the driver Valuer interface.
func (n NullStringList) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return pq.Array(n.Strings).Value()
}
