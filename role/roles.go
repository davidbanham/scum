package role

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type Roles []Role

func (roles Roles) Value() (driver.Value, error) {
	return json.Marshal(roles)
}

func (roles *Roles) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &roles)
}

func (roles Roles) Can(name string) bool {
	for _, role := range roles {
		if role.Can(name) {
			return true
		}
	}
	return false
}

type Role struct {
	Name       string
	Label      string
	Implies    Roles
	ValidRoles map[string]*Role
}

func (this *Role) Can(role string) bool {
	if role == this.Name {
		return true
	}
	this.Implies = this.ValidRoles[this.Name].Implies
	for _, sub := range this.Implies {
		if sub.Can(role) {
			return true
		}
	}
	return false
}

func (this Role) Implications() []string {
	ret := []string{}
	for name, _ := range this.ValidRoles {
		if name == this.Name {
			continue
		}
		if this.Can(name) {
			ret = append(ret, name)
		}
	}
	return ret
}
