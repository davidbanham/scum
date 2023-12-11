package toggle

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type Toggles []Toggle

type Toggle struct {
	Label    string `json:"-"`
	Key      string `json:"key"`
	Category string `json:"-"`
	State    bool   `json:"state"`
	HelpText string `json:"-"`
}

type Gettable interface {
	// For example, url.Values
	Get(string) string
}

func (this *Toggles) FromForm(form Gettable) {
	for i, toggle := range *this {
		(*this)[i].State = form.Get(toggle.Key) == "true"
	}
}

func (this *Toggles) Populate(validToggles []Toggle) {
	m := this.toMap()
	(*this) = []Toggle{}
	for _, toggle := range validToggles {
		t := Toggle{}
		t.Category = toggle.Category
		t.Label = toggle.Label
		t.Key = toggle.Key
		t.HelpText = toggle.HelpText
		if existing, ok := m[toggle.Key]; ok {
			t.State = existing.State
		}
		(*this) = append((*this), t)
	}
}

func (this Toggles) toMap() map[string]Toggle {
	ret := map[string]Toggle{}
	for _, toggle := range this {
		ret[toggle.Key] = toggle
	}
	return ret
}

func (this Toggles) ByKey(key string) Toggle {
	for _, t := range this {
		if t.Key == key {
			return t
		}
	}
	return Toggle{}
}

func (this Toggles) List() []Toggle {
	return this
}

func (this Toggles) ByCategory(category string) []Toggle {
	ret := []Toggle{}
	for _, toggle := range this.List() {
		if toggle.Category == category {
			ret = append(ret, toggle)
		}
	}
	return ret
}

func (this Toggles) Categories() []string {
	ret := []string{}
	for _, toggle := range this.List() {
		ret = append(ret, toggle.Category)
	}
	return ret
}

func (this Toggles) Value() (driver.Value, error) {
	return json.Marshal(this)
}

func (this *Toggles) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &this)
}
