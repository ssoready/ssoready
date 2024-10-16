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
				fmt.Println("top-level replace")
				f.Set(reflect.ValueOf(op.Value))
				continue
			case "add":
				return fmt.Errorf("unsupported top-level SCIM add operation")
			}
		}

		for _, segment := range segments[:len(segments)-1] {
			fmt.Println("mapindex", segment)
			f = f.Elem().MapIndex(reflect.ValueOf(segment))
		}
		switch op.Op {
		case "replace":
			fmt.Println("set map index", f.Kind(), f)
			f.Elem().SetMapIndex(reflect.ValueOf(segments[len(segments)-1]), reflect.ValueOf(op.Value))
		case "add":
			f = f.Elem().MapIndex(reflect.ValueOf(segments[len(segments)-1]))

			fmt.Println("set append", f.Kind(), f)
			f.Elem().Set(reflect.Append(f.Elem(), reflect.ValueOf(op.Value)))
		}
	}
	return nil
}
