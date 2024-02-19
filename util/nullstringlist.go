package util

import (
	"database/sql/driver"

	"github.com/lib/pq"
)

type NullStringList struct {
	Valid   bool
	Strings pq.StringArray
}

func (n *NullStringList) Scan(value interface{}) error {
	if value == nil {
		n.Strings, n.Valid = []string{}, false
		return nil
	}
	n.Valid = true
	return n.Strings.Scan(value)
}

// Value implements the driver Valuer interface.
func (n NullStringList) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Strings.Value()
}
