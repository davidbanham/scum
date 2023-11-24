package filter

import (
	"log"
	"testing"
)

func TestComposition(t *testing.T) {
	type SomeFilter struct {
		dateBase
	}

	foo := SomeFilter{}
	ugh := foo.table
	log.Printf(`DEBUG ugh: %+v`, ugh)
}
