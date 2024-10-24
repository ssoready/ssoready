package scimpatch

import (
	"fmt"
	"strings"
)

type Operation struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value"`
}

func Patch(ops []Operation, v *map[string]any) error {
	for _, op := range ops {
		if err := applyOp(op, v); err != nil {
			return err
		}
	}
	return nil
}

func applyOp(op Operation, v *map[string]any) error {
	opReplace := op.Op == "replace" || op.Op == "Replace"
	opAdd := op.Op == "add" || op.Op == "Add"

	if !opReplace && !opAdd {
		return fmt.Errorf("unsupported SCIM PATCH operation: %q", op.Op)
	}

	var segments []string
	if len(op.Path) > 0 {
		segments = strings.Split(op.Path, ".")
	}

	if len(segments) == 0 {
		if opReplace {
			val, ok := op.Value.(map[string]any)
			if !ok {
				return fmt.Errorf("top-level 'replace' operation must have an object value")
			}
			*v = val
			return nil
		}

		if opAdd {
			return fmt.Errorf("unsupported 'add' operation on top-level object")
		}
	}

	for _, segment := range segments[:len(segments)-1] {
		subV, ok := (*v)[segment].(map[string]any)
		if !ok {
			return fmt.Errorf("invalid path: %q", op.Path)
		}

		v = &subV
	}

	k := segments[len(segments)-1]
	if opReplace {
		(*v)[k] = op.Value
		return nil
	}
	if opAdd {
		kval, ok := (*v)[k]
		if !ok {
			(*v)[k] = op.Value
			return nil
		}

		switch kval := kval.(type) {
		case []any:
			// If the target location specifies a multi-valued attribute, a new
			// value is added to the attribute.
			(*v)[k] = append(kval, op.Value)
		case map[string]any:
			// If the target location specifies a complex attribute, a set of
			// sub-attributes SHALL be specified in the "value" parameter.
			//
			// which implies "write in the k/v pairs", given:
			//
			// If omitted, the target location is assumed to be the resource
			// itself.  The "value" parameter contains a set of attributes to be
			// added to the resource.
			mergeIn, ok := op.Value.(map[string]any)
			if !ok {
				// this is the SHALL
				return fmt.Errorf("'add' operation pointing at object must have an object value")
			}

			for k, v := range mergeIn {
				kval[k] = v
			}
		default:
			// If the target location specifies a single-valued attribute, the
			// existing value is replaced.
			//
			// only arrays and objects are multi-valued, so this branch is just
			// a replace
			(*v)[k] = op.Value
		}
	}
	return nil
}

func applyAdd(obj map[string]any, k string, v any) error {
	if _, ok := obj[k]; !ok {
		obj[k] = v
		return nil
	}

	switch objVal := obj[k].(type) {
	case map[string]any:
	case []any:
	default:
		obj[k] = v
		return nil
	}
}
