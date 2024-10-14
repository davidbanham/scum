package components

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"strings"
	"time"

	"github.com/davidbanham/heroicons"
	"github.com/davidbanham/scum/util"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
	uuid "github.com/satori/go.uuid"
)

//go:embed *.html
var FS embed.FS

var heroIcons heroicons.Icons

var FuncMap = template.FuncMap{
	"uniq": func() string {
		return util.SelectorSafe(uuid.NewV4().String())
	},
	"selectorSafe": util.SelectorSafe,
	"heroIcon": func(name string) template.HTML {
		icon, err := heroIcons.ByName(name)
		if err != nil {
			return template.HTML(err.Error())
		}

		return template.HTML(icon)
	},
	"isoTime": func(t time.Time) string {
		return t.Format(time.RFC3339)
	},
	"humanTime": func(t time.Time) string {
		return t.Format(time.RFC822)
	},
	"humanDayDate": func(t time.Time) string {
		loc, err := time.LoadLocation("Australia/Sydney")
		if err != nil {
			loc, _ = time.LoadLocation("UTC")
		}
		return t.In(loc).Format("Mon 02 Jan 2006")
	},
	"dateonly": func(t time.Time) string {
		return t.Format(time.DateOnly)
	},
	"timeonly": func(t time.Time) string {
		return t.Format(time.TimeOnly)
	},
	"roundTime": func(input time.Time, minutes int) time.Time {
		return input.Round(time.Duration(minutes) * time.Minute)
	},
	"dict": func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, errors.New("invalid dict call")
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, errors.New("dict keys must be strings")
			}
			if key == "BaseDict" {
				for k, v := range values[i+1].(map[string]interface{}) {
					dict[k] = v
				}
			} else {
				dict[key] = values[i+1]
			}
		}
		return dict, nil
	},
	"noescape": func(str string) template.HTML {
		return template.HTML(bluemonday.UGCPolicy().Sanitize(str))
	},
	"queryString": func(vals url.Values) template.URL {
		return "?" + template.URL(vals.Encode())
	},
	"crumbs": func(values ...string) ([]Crumb, error) {
		if len(values)%2 != 0 {
			return nil, errors.New("invalid dict call")
		}
		crumbs := []Crumb{}
		for i := 0; i < len(values); i += 2 {
			crumbs = append(crumbs, Crumb{
				Title: values[i],
				Path:  values[i+1],
			})
		}
		path := ""
		for i, crumb := range crumbs {
			if crumb.Path != "" && crumb.Path != "#" {
				if path == "" {
					path = crumb.Path
				} else if crumb.Path[0] == '/' {
					path = crumb.Path
				} else {
					path = fmt.Sprintf("%s/%s", strings.Split(path, "?")[0], crumb.Path)
				}
				crumbs[i].Path = path
			}
		}
		return crumbs, nil
	},
	"contains": util.Contains,
	"markdown": func(str string) template.HTML {
		extensions := parser.CommonExtensions | parser.NoEmptyLineBeforeBlock
		p := parser.NewWithExtensions(extensions)
		md := []byte(str)
		md = markdown.NormalizeNewlines(md)
		output := markdown.ToHTML(md, p, nil)
		return template.HTML(bluemonday.UGCPolicy().Sanitize(string(output)))
	},
}

func Tmpl() (*template.Template, error) {
	t, err := template.New("components").Funcs(FuncMap).ParseFS(FS, "*")
	if err != nil {
		return nil, err
	}
	if err := heroicons.Extend(t); err != nil {
		return nil, err
	}
	return t, nil
}

type Crumb struct {
	Title string
	Path  string
}
