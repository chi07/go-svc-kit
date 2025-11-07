package mapx_test

import (
	"reflect"
	"testing"

	"github.com/chi07/go-svc-kit/mapx"
)

func TestDeepEqualMap_NilVsNil(t *testing.T) {
	if !mapx.DeepEqualMap(nil, nil) {
		t.Fatalf("nil vs nil should be equal")
	}
}

func TestDeepEqualMap_OneNil(t *testing.T) {
	a := map[string]any{"a": 1}
	if mapx.DeepEqualMap(a, nil) {
		t.Fatalf("a vs nil should be NOT equal")
	}
	if mapx.DeepEqualMap(nil, a) {
		t.Fatalf("nil vs a should be NOT equal")
	}
}

func TestDeepEqualMap_EqualSimple(t *testing.T) {
	a := map[string]any{"a": 1, "b": "x"}
	b := map[string]any{"b": "x", "a": 1}
	if !mapx.DeepEqualMap(a, b) {
		t.Fatalf("maps with same kv should be equal")
	}
}

func TestDeepEqualMap_DifferentValues(t *testing.T) {
	a := map[string]any{"a": 1}
	b := map[string]any{"a": 2}
	if mapx.DeepEqualMap(a, b) {
		t.Fatalf("maps with different values should NOT be equal")
	}
}

func TestDeepEqualMap_Nested(t *testing.T) {
	a := map[string]any{
		"n": map[string]any{"k": 1},
		"s": []int{1, 2, 3},
	}
	b := map[string]any{
		"n": map[string]any{"k": 1},
		"s": []int{1, 2, 3},
	}
	if !mapx.DeepEqualMap(a, b) {
		t.Fatalf("nested structures should be deeply equal")
	}
}

func TestCloneMap_NilInput(t *testing.T) {
	if got := mapx.CloneMap(nil); got != nil {
		t.Fatalf("expected nil clone for nil input, got %#v", got)
	}
}

func TestCloneMap_ShallowTopLevelIndependence(t *testing.T) {
	src := map[string]any{"a": 1, "b": "x"}
	clone := mapx.CloneMap(src)

	if clone == nil {
		t.Fatalf("clone is nil")
	}

	clone["a"] = 2

	if src["a"].(int) != 1 {
		t.Fatalf("top-level mutation on clone should not affect src; src[a]=%v", src["a"])
	}

	src2 := map[string]any{"a": 1, "b": "x"}
	if reflect.DeepEqual(clone, src2) {
		t.Fatalf("clone should differ after mutation")
	}
}

func TestCloneMap_ShallowForNested(t *testing.T) {
	src := map[string]any{
		"n": map[string]any{"k": 1},
	}
	clone := mapx.CloneMap(src)

	clone["n"].(map[string]any)["k"] = 9

	if src["n"].(map[string]any)["k"].(int) != 9 {
		t.Fatalf("expected shallow copy: nested mutation in clone should reflect in src")
	}
}
