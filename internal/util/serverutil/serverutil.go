package serverutil

import (
	"os"
	"reflect"
	"strings"
)

func GetFieldValueByName(object interface{}, field string) string {
	vo := reflect.ValueOf(object)
	var v reflect.Value
	if vo.Kind() == reflect.Ptr {
		v = reflect.ValueOf(object).Elem()
	} else {
		v = reflect.ValueOf(object)
	}
	f := v.FieldByName(field)
	if f.Kind() == reflect.Ptr {
		f = f.Elem()
	}
	if f.IsValid() {
		return f.String()
	}
	return ""
}

func GetEnv() string {
	env := strings.ToLower(os.Getenv("env"))

	if env == "" {
		env = "test"
	}
	return env
}
