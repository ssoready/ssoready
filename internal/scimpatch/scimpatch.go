package scimpatch

import (
	"fmt"
	"reflect"
	"strings"
)

type Operation struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value"`
}

func Patch(ops []Operation, v any) error {
	for _, op := range ops {
		if op.Op != "replace" && op.Op != "add" {
			return fmt.Errorf("unsupported SCIM PATCH operation: %q", op.Path)
		}

		var segments []string
		if len(op.Path) > 0 {
			segments = strings.Split(op.Path, ".")
		}

		f := reflect.ValueOf(v).Elem()
		if len(segments) == 0 {
			switch op.Op {
			case "replace":
				f.Set(reflect.ValueOf(op.Value))
				continue
			case "add":
				return fmt.Errorf("unsupported top-level SCIM add operation")
			}
		}

		f = f.Elem()
		for _, segment := range segments[:len(segments)-1] {
			f = f.MapIndex(reflect.ValueOf(segment)).Elem()
		}
		key := reflect.ValueOf(segments[len(segments)-1])
		switch op.Op {
		case "replace":
			f.SetMapIndex(key, reflect.ValueOf(op.Value))
		case "add":
			f.SetMapIndex(key, reflect.Append(f.MapIndex(key).Elem(), reflect.ValueOf(op.Value)))
		}
	}
	return nil
}
