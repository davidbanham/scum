package flash

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strconv"

	uuid "github.com/satori/go.uuid"
)

type Flashable interface {
	PersistFlash(context.Context, Flash) error
}

type Flash struct {
	Persistent  bool          `json:"persistent"`
	Sticky      bool          `json:"sticky"`
	EntityKey   string        `json:"entity_key"`
	ID          string        `json:"id"`
	Text        string        `json:"text"`
	Actions     []FlashAction `json:"actions"`
	Type        FlashLevel    `json:"type"`
	OnceOnlyKey string        `json:"once_only_key"`
}

func (this Flash) Value() (driver.Value, error) {
	return json.Marshal(this)
}

func (this *Flash) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &this)
}

func (this *Flash) Add(ctx context.Context) error {
	this.ID = uuid.NewV4().String()

	flashes := Flashes{}
	unconv := ctx.Value("flashes")
	if unconv != nil {
		flashes = unconv.(Flashes)
	}
	flashes = append(flashes, *this)
	if this.Persistent {
		key := this.EntityKey
		if key == "" {
			key = "user"
		}
		unconv := ctx.Value(key)
		if unconv != nil {
			user := unconv.(Flashable)
			if err := user.PersistFlash(ctx, *this); err != nil {
				return err
			}
		}
	}
	return nil
}

type Flashes []Flash

func (this Flashes) Value() (driver.Value, error) {
	if len(this) == 0 {
		return "[]", nil
	}
	return json.Marshal(this)
}

func (this *Flashes) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &this)
}

type FlashAction struct {
	Url  string
	Text string
}

func (this *Flashes) Add(flash Flash) Flash {
	for _, existing := range *this {
		if flash.OnceOnlyKey != "" && flash.OnceOnlyKey == existing.OnceOnlyKey {
			return Flash{}
		}
	}
	if flash.ID == "" {
		flash.ID = uuid.NewV4().String()
	}
	(*this) = append((*this), flash)

	return flash
}

type FlashLevel int

const (
	Warn FlashLevel = 1 + iota
	Success
	Info
)

func (this FlashLevel) Levels() []FlashLevel {
	return []FlashLevel{
		Warn, Success, Info,
	}
}

func (this FlashLevel) Label() string {
	switch this {
	case Warn:
		return "Warning"
	case Success:
		return "Success"
	case Info:
		return "Information"
	}
	return "Unknown"
}

func (this *FlashLevel) FromString(in string) error {
	parsed, err := strconv.Atoi(in)
	if err != nil {
		return err
	}
	(*this) = FlashLevel(parsed)

	return nil
}
