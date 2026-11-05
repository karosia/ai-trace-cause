package propcopy_test

import (
	"errors"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/karosia/ai-trace-cause/internal/propcopy"
)

func TestCloneMapIsDeep(t *testing.T) {
	nested := map[string]any{"replicas": 5}
	list := []string{"a", "b"}
	n := 7
	original := map[string]any{
		"cfg":  nested,
		"list": list,
		"ptr":  &n,
		"arr":  [2]any{map[string]any{"x": 1}, "y"},
	}

	cloned, err := propcopy.CloneMap(original)
	if err != nil {
		t.Fatalf("CloneMap() error = %v", err)
	}

	nested["replicas"] = 999
	list[0] = "mutated"
	n = 999
	original["arr"].([2]any)[0].(map[string]any)["x"] = 999

	if got := cloned["cfg"].(map[string]any)["replicas"]; got != 5 {
		t.Errorf("nested map shared: replicas = %v, want 5", got)
	}

	if got := cloned["list"].([]string)[0]; got != "a" {
		t.Errorf("slice shared: list[0] = %q, want a", got)
	}

	if got := *cloned["ptr"].(*int); got != 7 {
		t.Errorf("pointer shared: *ptr = %d, want 7", got)
	}

	if got := cloned["arr"].([2]any)[0].(map[string]any)["x"]; got != 1 {
		t.Errorf("array element shared: x = %v, want 1", got)
	}
}

func TestClonePreservesConcreteTypes(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	type label string

	values := []any{
		nil,
		true,
		"s",
		label("named"),
		int8(1),
		uint64(2),
		float32(1.5),
		now,
		[]string{"a"},
		map[string]int{"a": 1},
		map[label]any{"k": []any{1, "two"}},
	}

	for _, value := range values {
		cloned, err := propcopy.Clone(value, "v")
		if err != nil {
			t.Fatalf("Clone(%#v) error = %v", value, err)
		}

		if !reflect.DeepEqual(cloned, value) {
			t.Errorf("Clone(%#v) = %#v", value, cloned)
		}

		if value != nil &&
			reflect.TypeOf(cloned) != reflect.TypeOf(value) {
			t.Errorf(
				"Clone(%#v) type = %T, want %T",
				value,
				cloned,
				value,
			)
		}
	}
}

func TestCloneRejectsUnsupportedValues(t *testing.T) {
	type point struct{ X int }

	cases := map[string]struct {
		properties map[string]any
		wantPath   string
	}{
		"func": {
			properties: map[string]any{"f": func() {}},
			wantPath:   "f",
		},
		"channel nested in slice": {
			properties: map[string]any{
				"list": []any{1, make(chan int)},
			},
			wantPath: "list[1]",
		},
		"non-string map key": {
			properties: map[string]any{
				"m": map[int]string{1: "a"},
			},
			wantPath: "m",
		},
		"struct": {
			properties: map[string]any{
				"meta": map[string]any{"p": point{X: 1}},
			},
			wantPath: "meta.p",
		},
		"NaN": {
			properties: map[string]any{"v": math.NaN()},
			wantPath:   "v",
		},
		"Inf": {
			properties: map[string]any{"v": math.Inf(1)},
			wantPath:   "v",
		},
		"complex": {
			properties: map[string]any{"v": complex(1, 2)},
			wantPath:   "v",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := propcopy.CloneMap(tc.properties)

			var copyErr *propcopy.Error
			if !errors.As(err, &copyErr) {
				t.Fatalf(
					"CloneMap() error = %v, want *propcopy.Error",
					err,
				)
			}

			if copyErr.Path != tc.wantPath {
				t.Errorf(
					"Path = %q, want %q",
					copyErr.Path,
					tc.wantPath,
				)
			}
		})
	}
}

func TestCloneMapNil(t *testing.T) {
	cloned, err := propcopy.CloneMap(nil)
	if err != nil || cloned != nil {
		t.Fatalf("CloneMap(nil) = %v, %v; want nil, nil", cloned, err)
	}
}
