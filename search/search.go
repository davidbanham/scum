package search

import (
	"context"
	"strings"

	"github.com/davidbanham/scum/filter"
	"github.com/davidbanham/scum/model"
	"github.com/davidbanham/scum/pagination"
)

type Searchables []Searchable

func (this Searchables) ByTableName() map[string]Searchable {
	ret := map[string]Searchable{}
	for _, s := range this {
		ret[s.Tablename] = s
	}
	return ret
}

func (this Searchables) ByEntityType() map[string]Searchable {
	ret := map[string]Searchable{}
	for _, s := range this {
		ret[s.EntityType] = s
	}
	return ret
}

func (this Searchables) FilterByTableNames(tables []string) Searchables {
	ret := Searchables{}
	byTableName := this.ByTableName()
	for _, t := range tables {
		if s, ok := byTableName[t]; ok {
			ret = append(ret, s)
		}
	}
	return ret
}

func (this Searchables) FilterByEntityType(entities []string) Searchables {
	ret := Searchables{}
	byEntityType := this.ByEntityType()
	for _, t := range entities {
		if s, ok := byEntityType[t]; ok {
			ret = append(ret, s)
		}
	}
	return ret
}

func (this Searchables) FilterByRole(roles Roles, query SearchQuery) Searchables {
	ret := Searchables{}

	for _, searchable := range this {
		if searchable.Permitted(roles, query) {
			ret = append(ret, searchable)
		}
	}

	return ret
}

type SearchCriteria struct {
	Query      SearchQuery
	Entities   []string
	Filters    filter.Filters
	Pagination pagination.Pagination
}

type Roles interface {
	Can(string) bool
}

type Searchable struct {
	EntityType string
	Label      string
	Path       string
	Tablename  string
	Permitted  func(Roles, SearchQuery) bool
}

type SearchQuery interface {
	Construct(models Searchables, filters filter.Filters, pagination pagination.Pagination) string
	Args() []any
	UserInput() string
}

type SearchResult struct {
	Path       string
	EntityType string
	ID         string
	Label      string
	Rank       float64
}

type SearchResults struct {
	Data     []SearchResult
	Criteria SearchCriteria
}

type ByPhrase struct {
	OrganisationID string
	Phrase         string
}

func (this ByPhrase) Construct(entities Searchables, filters filter.Filters, pagination pagination.Pagination) string {
	parts := []string{}
	for _, entity := range entities {
		parts = append(parts, `
SELECT
	text '`+entity.EntityType+`' AS entity_type, text '`+entity.Path+`' AS uri_path, id AS id, `+entity.Label+` AS label, ts_rank_cd(ts, query) AS rank
FROM
`+entity.Tablename+`, plainto_tsquery('english', $2) query `+filters.Query()+` AND organisation_id = $1 AND query @@ ts`)
	}

	query := strings.Join(parts, " UNION ALL ")

	query += " ORDER BY rank DESC " + pagination.PaginationQuery()

	return query
}

func (this ByPhrase) Args() []any {
	return []any{this.OrganisationID, this.Phrase}
}

func (this ByPhrase) UserInput() string {
	return this.Phrase
}

func (results *SearchResults) FindAll(ctx context.Context, roles Roles, criteria SearchCriteria, searchables Searchables) error {
	results.Criteria = criteria

	filtered := searchables.FilterByRole(roles, criteria.Query)
	if len(filtered) == 0 {
		return nil
	}

	filtered = searchables.FilterByTableNames(criteria.Entities)
	if len(filtered) == 0 {
		return nil
	}

	query := criteria.Query.Construct(filtered, criteria.Filters, criteria.Pagination)

	db := ctx.Value("tx").(model.Querier)

	rows, err := db.QueryContext(ctx, query, criteria.Query.Args()...)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		result := SearchResult{}
		err = rows.Scan(
			&result.EntityType, &result.Path, &result.ID, &result.Label, &result.Rank,
		)
		if err != nil {
			return err
		}
		(*results).Data = append((*results).Data, result)
	}

	return err
}

func BasicRoleCheck(requiredRole string) func(roles Roles, query SearchQuery) bool {
	return func(roles Roles, query SearchQuery) bool {
		if roles.Can(requiredRole) {
			return true
		}
		return false
	}
}
