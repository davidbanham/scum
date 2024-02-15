package util

import (
	"database/sql/driver"
	"fmt"

	"github.com/lib/pq"
)

type StringList []string

func (this *StringList) Scan(value interface{}) error {
	switch value.(type) {
	default:
		return fmt.Errorf("Not a list of strings")
	case []byte:
		return pq.Array(this).Scan(value)
	case []string:
		(*this) = value.([]string)
		return nil
	}
}

func (this StringList) Value() (driver.Value, error) {
	return pq.Array(this).Value()
}
