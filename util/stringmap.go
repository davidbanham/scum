package util

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type StringMap map[string]string

func (p StringMap) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

func (p *StringMap) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}

	return json.Unmarshal(source, p)
}
