package util

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type IntMap map[string]int

func (p IntMap) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

func (p *IntMap) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}

	return json.Unmarshal(source, p)
}
