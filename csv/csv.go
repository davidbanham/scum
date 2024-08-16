package csv

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

func CSVCols(model any) []string {
	ret := []string{}
	thing := reflect.TypeOf(model)
	elem := reflect.ValueOf(model)
	for _, field := range reflect.VisibleFields(thing) {
		if field.IsExported() {
			colName, ok := field.Tag.Lookup("csv")
			if colName == "-" {
				continue
			} else if !ok {
				colName = field.Name
			}
			inter := elem.FieldByIndex(field.Index).Interface()
			if field.Type.String() == "time.Time" {
				ret = append(ret, colName)
			} else if field.Type.Kind() == reflect.Struct {
				subs := CSVCols(inter)
				for _, sub := range subs {
					ret = append(ret, fmt.Sprintf("%s.%s", colName, sub))
				}
			} else {
				ret = append(ret, colName)
			}
		}
	}
	return ret
}

func CSVVals(model any) []string {
	ret := []string{}
	thing := reflect.TypeOf(model)
	elem := reflect.ValueOf(model)
	for _, field := range reflect.VisibleFields(thing) {
		if field.IsExported() {
			colName := field.Tag.Get("csv")
			if colName == "-" {
				continue
			}
			inter := elem.FieldByIndex(field.Index).Interface()
			if field.Type.String() == "time.Time" {
				ret = append(ret, inter.(time.Time).Format(time.RFC3339))
			} else if field.Type.Kind() == reflect.Struct {
				ret = append(ret, CSVVals(inter)...)
			} else {
				var val string
				switch v := inter.(type) {
				default:
					val = ""
				case *CSVStringAble:
					val = (*v).CSVString()
				case CSVStringAble:
					val = v.CSVString()
				case *fmt.Stringer:
					val = (*v).String()
				case fmt.Stringer:
					val = v.String()
				case *bool:
					if *v {
						val = "true"
					} else {
						val = "false"
					}
				case bool:
					if v {
						val = "true"
					} else {
						val = "false"
					}
				case *string:
					val = *v
				case string:
					val = v
				case *int:
					val = strconv.Itoa(*v)
				case int:
					val = strconv.Itoa(v)
				}
				ret = append(ret, val)
			}
		}
	}
	return ret
}

type CSVStringAble interface {
	CSVString() string
}
