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

func Patch(ops []Operation, obj *map[string]any) error {
	for _, op := range ops {
		if err := applyOp(op, obj); err != nil {
			return err
		}
	}
	return nil
}

func applyOp(op Operation, obj *map[string]any) error {
	opReplace := op.Op == "replace" || op.Op == "Replace"
	opAdd := op.Op == "add" || op.Op == "Add"

	if !opReplace && !opAdd {
		return fmt.Errorf("unsupported SCIM PATCH operation: %q", op.Op)
	}

	segments := splitPath(op.Path)

	if len(segments) == 0 {
		if opReplace {
			val, ok := op.Value.(map[string]any)
			if !ok {
				return fmt.Errorf("top-level 'replace' operation must have an object value")
			}
			*obj = val
			return nil
		}

		if opAdd {
			return fmt.Errorf("unsupported 'add' operation on top-level object")
		}
	}

	for _, segment := range segments[:len(segments)-1] {
		subV, ok := (*obj)[segment].(map[string]any)
		if !ok {
			return fmt.Errorf("invalid path: %q", op.Path)
		}

		obj = &subV
	}

	k := segments[len(segments)-1]
	if opReplace {
		(*obj)[k] = op.Value
		return nil
	}
	if opAdd {
		if err := applyAdd(*obj, k, op.Value); err != nil {
			return err
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
		v, ok := v.(map[string]any)
		if !ok {
			return fmt.Errorf("'add' operation pointing at object must be object-valued")
		}

		for k := range v {
			objVal[k] = v[k]
		}
		return nil
	case []any:
		v, ok := v.([]any)
		if !ok {
			return fmt.Errorf("'add' operation pointing at array must be array-valued")
		}

		obj[k] = append(objVal, v...)
		return nil
	default:
		obj[k] = v
		return nil
	}
}

var enterpriseUserPrefix = "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"

// splitPath splits an op's path into its segments
//
// splitPath has special-case behavior as a concession to Entra's non-conformant
// behavior; they do PATCHes with paths like:
//
//	urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:manager
//
// Entra intends this to mean the "manager" property under "urn:...:User", but
// the spec indicates this should mean the "urn:...:2" > "0:User:manager"
// property. The selective behavior around ":" and "." can't be made to make
// sense beyond just a straightforward special-casing.
func splitPath(path string) []string {
	if path == "" {
		return nil
	}
	if path == enterpriseUserPrefix {
		return []string{enterpriseUserPrefix}
	}
	if strings.HasPrefix(path, enterpriseUserPrefix+":") {
		return []string{enterpriseUserPrefix, strings.TrimPrefix(path, enterpriseUserPrefix+":")}
	}
	return strings.Split(path, ".")
}
