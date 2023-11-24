package components

import "testing"

func TestTmplParse(t *testing.T) {
	_, err := Tmpl()
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
}
