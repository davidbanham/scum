package role

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type Roles []Role

func (roles Roles) ByName(name string) Role {
	for _, r := range roles {
		if r.Name == name {
			return r
		}
	}
	return Role{}
}

func (roles Roles) NamedMap() map[string]Role {
	ret := map[string]Role{}
	for _, r := range roles {
		ret[r.Name] = r
	}
	return ret
}

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

func (roles Roles) CanOnly(name string) bool {
	if len(roles) == 1 && roles[0].Name == name && roles.Can(name) {
		return true
	}
	return false
}

func (roles Roles) CanOver(name string, entityID string) bool {
	for _, role := range roles {
		for _, sub := range role.Over {
			if sub == entityID {
				return role.Can(name)
			}
		}
	}
	return false
}

func (roles *Roles) AssignEntities(name string, entityIDs []string) {
	for i, role := range *roles {
		if role.Name == name {
			(*roles)[i].Over = entityIDs
		}
	}
}

func (this *Roles) Implications(validRoles Roles) {
	for i, role := range *this {
		role.Implies = validRoles.ByName(role.Name).Implies
		for j, sub := range role.Implies {
			sub.Implications(validRoles)
			role.Implies[j] = sub
		}
		(*this)[i] = role
	}
}

type Role struct {
	Name    string   `json:"name"`
	Label   string   `json:"label"`
	Implies Roles    `json:"-"`
	Over    []string `json:"over"`
}

func (this *Role) Implications(validRoles Roles) {
	this.Implies = validRoles.ByName(this.Name).Implies
	for _, role := range this.Implies {
		role.Implications(validRoles)
	}
}

func (this *Role) Can(role string) bool {
	if role == this.Name {
		return true
	}
	for _, sub := range this.Implies {
		if sub.Can(role) {
			return true
		}
	}
	return false
}
