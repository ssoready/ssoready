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
		opReplace := op.Op == "replace" || op.Op == "Replace"
		opAdd := op.Op == "add" || op.Op == "Add"

		if !opReplace && !opAdd {
			return fmt.Errorf("unsupported SCIM PATCH operation: %q", op.Op)
		}

		var segments []string
		if len(op.Path) > 0 {
			segments = strings.Split(op.Path, ".")
		}

		f := reflect.ValueOf(v).Elem()
		if len(segments) == 0 {
			switch {
			case opReplace:
				f.Set(reflect.ValueOf(op.Value))
				continue
			case opAdd:
				return fmt.Errorf("unsupported top-level SCIM add operation")
			}
		}

		f = f.Elem()
		for _, segment := range segments[:len(segments)-1] {
			f = f.MapIndex(reflect.ValueOf(segment)).Elem()
		}
		key := reflect.ValueOf(segments[len(segments)-1])
		switch {
		case opReplace:
			f.SetMapIndex(key, reflect.ValueOf(op.Value))
		case opAdd:
			f.SetMapIndex(key, reflect.Append(f.MapIndex(key).Elem(), reflect.ValueOf(op.Value)))
		}
	}
	return nil
}
