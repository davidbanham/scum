package model

import "github.com/notbadsoftware/html2text"

type Markedup string

func (this *Markedup) FromString(input string) {
	(*this) = Markedup(input)
}

func (this Markedup) String() string {
	text, err := html2text.FromString(string(this), html2text.Options{PrettyTables: true})
	if err != nil {
		return string(this)
	}
	return text
}
