// Package propcopy deep-copies node and edge property values and
// rejects values that cannot be represented as JSON.
//
// Recorded entities form an audit trail, so once a value is recorded,
// later mutation by the caller must not reach the stored copy. A
// shallow map copy is not enough: nested maps and slices would still
// be shared. Restricting values to JSON-representable data also keeps
// the in-memory store's behavior consistent with what any persistent
// store can round-trip.
//
// Supported values are nil, bool, string, every integer and float
// kind (except NaN and ±Inf), time.Time, and maps with string keys,
// slices, arrays, and pointers composed of supported values. Concrete
// types are preserved: a []string is copied as a []string, not as a
// []any.
package propcopy

import (
	"fmt"
	"math"
	"reflect"
	"time"
)

// Error describes a value that cannot be copied, and where it was
// found within the property tree.
type Error struct {
	// Path locates the offending value, e.g. "metadata.limits[2]".
	Path string
	// Reason describes why the value was rejected.
	Reason string
}

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Path, e.Reason)
}

var timeType = reflect.TypeFor[time.Time]()

// CloneMap returns a deep copy of properties. A nil map is returned as
// nil.
func CloneMap(
	properties map[string]any,
) (map[string]any, error) {
	if properties == nil {
		return nil, nil
	}

	cloned := make(map[string]any, len(properties))

	for key, value := range properties {
		copied, err := Clone(value, key)
		if err != nil {
			return nil, err
		}

		cloned[key] = copied
	}

	return cloned, nil
}

// Clone returns a deep copy of value. path names value in any
// returned *Error.
func Clone(value any, path string) (any, error) {
	if value == nil {
		return nil, nil
	}

	copied, err := cloneValue(reflect.ValueOf(value), path)
	if err != nil {
		return nil, err
	}

	return copied.Interface(), nil
}

func cloneValue(
	value reflect.Value,
	path string,
) (reflect.Value, error) {
	switch value.Kind() {
	case reflect.Bool,
		reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64:
		return value, nil

	case reflect.Float32, reflect.Float64:
		f := value.Float()
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return reflect.Value{}, &Error{
				Path:   path,
				Reason: "NaN and infinite floats are not representable as JSON",
			}
		}

		return value, nil

	case reflect.Interface:
		if value.IsNil() {
			return value, nil
		}

		inner, err := cloneValue(value.Elem(), path)
		if err != nil {
			return reflect.Value{}, err
		}

		out := reflect.New(value.Type()).Elem()
		out.Set(inner)

		return out, nil

	case reflect.Pointer:
		if value.IsNil() {
			return value, nil
		}

		inner, err := cloneValue(value.Elem(), path)
		if err != nil {
			return reflect.Value{}, err
		}

		out := reflect.New(value.Type().Elem())
		out.Elem().Set(inner)

		return out, nil

	case reflect.Map:
		if value.Type().Key().Kind() != reflect.String {
			return reflect.Value{}, &Error{
				Path: path,
				Reason: fmt.Sprintf(
					"map key type %s is not a string",
					value.Type().Key(),
				),
			}
		}

		if value.IsNil() {
			return value, nil
		}

		out := reflect.MakeMapWithSize(value.Type(), value.Len())

		iter := value.MapRange()
		for iter.Next() {
			copied, err := cloneValue(
				iter.Value(),
				path+"."+iter.Key().String(),
			)
			if err != nil {
				return reflect.Value{}, err
			}

			out.SetMapIndex(iter.Key(), copied)
		}

		return out, nil

	case reflect.Slice:
		if value.IsNil() {
			return value, nil
		}

		out := reflect.MakeSlice(
			value.Type(),
			value.Len(),
			value.Len(),
		)

		if err := cloneElements(value, out, path); err != nil {
			return reflect.Value{}, err
		}

		return out, nil

	case reflect.Array:
		out := reflect.New(value.Type()).Elem()

		if err := cloneElements(value, out, path); err != nil {
			return reflect.Value{}, err
		}

		return out, nil

	case reflect.Struct:
		if value.Type() == timeType {
			return value, nil
		}
	}

	return reflect.Value{}, &Error{
		Path: path,
		Reason: fmt.Sprintf(
			"unsupported type %s; use JSON-compatible values such as maps, slices, strings, numbers, and bools",
			value.Type(),
		),
	}
}

func cloneElements(
	src reflect.Value,
	dst reflect.Value,
	path string,
) error {
	for i := range src.Len() {
		copied, err := cloneValue(
			src.Index(i),
			fmt.Sprintf("%s[%d]", path, i),
		)
		if err != nil {
			return err
		}

		dst.Index(i).Set(copied)
	}

	return nil
}
