package scimpatch

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Operation struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value"`
}

type pathSegment struct {
	name   string
	filter *filterExpr
}

type filterExpr struct {
	attr  string
	op    string
	value string
}

func (p pathSegment) String() string {
	if p.filter == nil {
		return p.name
	}
	return fmt.Sprintf("%s[%s %s \"%s\"]", p.name, p.filter.attr, p.filter.op, p.filter.value)
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

	current := obj
	for i, segment := range segments {
		isLast := i == len(segments)-1

		if segment.filter != nil {
			arr, ok := (*current)[segment.name].([]any)
			if !ok {
				return fmt.Errorf("invalid path: array not found at %q", segment.String())
			}

			modified := false
			for j, item := range arr {
				if m, ok := item.(map[string]any); ok {
					if v, exists := m[segment.filter.attr]; exists {
						matches := false
						switch segment.filter.op {
						case "eq":
							matches = v == segment.filter.value
						case "ne":
							matches = v != segment.filter.value
						case "co":
							if str, ok := v.(string); ok {
								matches = strings.Contains(str, segment.filter.value)
							} else {
								return fmt.Errorf("'co' operator can only be used with string values")
							}
						case "sw":
							if str, ok := v.(string); ok {
								matches = strings.HasPrefix(str, segment.filter.value)
							} else {
								return fmt.Errorf("'sw' operator can only be used with string values")
							}
						case "ew":
							if str, ok := v.(string); ok {
								matches = strings.HasSuffix(str, segment.filter.value)
							} else {
								return fmt.Errorf("'ew' operator can only be used with string values")
							}
						case "pr":
							if str, ok := v.(string); ok {
								matches = str != ""
							} else {
								matches = v != nil
							}
						case "gt", "ge", "lt", "le":
							switch val := v.(type) {
							case string:
								// Try to parse as date first
								if date, err := time.Parse(time.RFC3339, val); err == nil {
									compareDate, err := time.Parse(time.RFC3339, segment.filter.value)
									if err != nil {
										return fmt.Errorf("invalid date format in comparison: %q", segment.filter.value)
									}
									matches = compareDates(date, compareDate, segment.filter.op)
								} else {
									matches = compare(val, segment.filter.value, segment.filter.op)
								}
							case float64:
								num, err := strconv.ParseFloat(segment.filter.value, 64)
								if err != nil {
									return fmt.Errorf("invalid number in comparison: %q", segment.filter.value)
								}
								matches = compare(val, num, segment.filter.op)
							case int:
								num, err := strconv.ParseFloat(segment.filter.value, 64)
								if err != nil {
									return fmt.Errorf("invalid number in comparison: %q", segment.filter.value)
								}
								matches = compare(float64(val), num, segment.filter.op)
							default:
								return fmt.Errorf("comparison operators can only be used with string or numeric values")
							}
						default:
							return fmt.Errorf("unsupported filter operator: %q", segment.filter.op)
						}

						if matches {
							modified = true
							if isLast {
								if strings.Contains(op.Path, "].") {
									// If we have a field after the filter, create a new operation for it
									parts := strings.Split(op.Path, "].")
									if len(parts) == 2 {
										newMap := make(map[string]any)
										for k, v := range m {
											newMap[k] = v
										}
										arr[j] = newMap
										if err := applyOp(Operation{
											Op:    op.Op,
											Path:  parts[1],
											Value: op.Value,
										}, &newMap); err != nil {
											return err
										}
									}
								} else {
									// If no field after the filter, replace the entire object
									arr[j] = op.Value
								}
							} else {
								// Not the last segment, continue with the rest of the path
								newMap := make(map[string]any)
								for k, v := range m {
									newMap[k] = v
								}
								arr[j] = newMap
								if err := applyOp(Operation{
									Op:    op.Op,
									Path:  strings.Join(segmentsToStrings(segments[i+1:]), "."),
									Value: op.Value,
								}, &newMap); err != nil {
									return err
								}
							}
						}
					}
				}
			}
			if !modified {
				return fmt.Errorf("no matching element found for filter %q", segment.String())
			}
			(*current)[segment.name] = arr
			return nil
		}

		if isLast {
			if opReplace {
				(*current)[segment.name] = op.Value
				return nil
			}
			if opAdd {
				if err := applyAdd(*current, segment.name, op.Value); err != nil {
					return err
				}
			}
			return nil
		}

		subV, ok := (*current)[segment.name].(map[string]any)
		if !ok {
			return fmt.Errorf("invalid path: %q", segment.String())
		}
		current = &subV
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
func splitPath(path string) []pathSegment {
	if path == "" {
		return nil
	}
	if path == enterpriseUserPrefix {
		return []pathSegment{{name: enterpriseUserPrefix}}
	}
	if strings.HasPrefix(path, enterpriseUserPrefix+":") {
		remainingPath := strings.TrimPrefix(path, enterpriseUserPrefix+":")
		return append([]pathSegment{{name: enterpriseUserPrefix}}, splitPath(remainingPath)...)
	}

	var segments []pathSegment
	for _, part := range strings.Split(path, ".") {
		if idx := strings.Index(part, "["); idx != -1 {
			if end := strings.Index(part, "]"); end != -1 {
				filter := parseFilter(part[idx+1 : end])
				segments = append(segments, pathSegment{
					name:   part[:idx],
					filter: filter,
				})
				continue
			}
		}
		segments = append(segments, pathSegment{name: part})
	}
	return segments
}

func parseFilter(expr string) *filterExpr {
	parts := strings.Split(expr, " ")
	if len(parts) == 2 && parts[1] == "pr" {
		return &filterExpr{
			attr: parts[0],
			op:   "pr",
		}
	}
	if len(parts) != 3 {
		return nil
	}
	// Remove quotes from value
	value := strings.Trim(parts[2], "\"")
	return &filterExpr{
		attr:  parts[0],
		op:    parts[1],
		value: value,
	}
}

// Helper function to convert pathSegments back to strings
func segmentsToStrings(segments []pathSegment) []string {
	result := make([]string, len(segments))
	for i, seg := range segments {
		result[i] = seg.String()
	}
	return result
}

// comparable is a constraint that permits ordered comparisons
type ordered interface {
	~string | ~float64 | ~int
}

func compare[T ordered](a, b T, op string) bool {
	switch op {
	case "gt":
		return a > b
	case "ge":
		return a >= b
	case "lt":
		return a < b
	case "le":
		return a <= b
	default:
		return false
	}
}

func compareDates(a, b time.Time, op string) bool {
	switch op {
	case "gt":
		return a.After(b)
	case "ge":
		return a.After(b) || a.Equal(b)
	case "lt":
		return a.Before(b)
	case "le":
		return a.Before(b) || a.Equal(b)
	default:
		return false
	}
}
