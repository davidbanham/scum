package filter

import (
	"log"
	"time"

	"github.com/davidbanham/scum/util"
)

func CommonFilters(table string) Filters {
	ub := DateFilter{}
	if err := ub.Hydrate(DateFilterOpts{
		Label: "Updated Between",
		ID:    "updated-between",
		Table: table,
		Col:   "updated_at",
		Period: util.Period{
			Start: time.Now().Add(-24 * time.Hour),
			End:   time.Now().Add(24 * time.Hour),
		},
	}); err != nil {
		log.Fatal(err)
	}

	cb := DateFilter{}
	if err := cb.Hydrate(DateFilterOpts{
		Label: "Created Between",
		ID:    "created-between",
		Table: table,
		Col:   "created_at",
		Period: util.Period{
			Start: time.Now().Add(-24 * time.Hour),
			End:   time.Now().Add(24 * time.Hour),
		},
	}); err != nil {
		log.Fatal(err)
	}
	return Filters{
		&ub,
		&cb,
	}
}
