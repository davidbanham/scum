package role

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type Roles []Role

func (roles Roles) ByName() map[string]Role {
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

//func (roles Roles) CanOver(name string, entityID string) bool {
//	for _, role := range roles {
//		for _, sub := range role.Over {
//			if sub == entityID {
//				return role.Can(name)
//			}
//		}
//	}
//	return false
//}

type Role struct {
	Name    string `json:"name"`
	Label   string `json:"label"`
	Implies Roles  `json:"-"`

	//Over       []string `json:"over"`
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
