package restyx_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"

	"github.com/chi07/go-svc-kit/restyx"
)

func newReq() *resty.Request { return resty.New().R() }

// helper to build an ExecFunc backed by an httptest.Server
func makeExec(status int, body string) (restyx.ExecFunc, func()) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	exec := func(r *resty.Request) (*resty.Response, error) {
		return r.Get(s.URL)
	}
	cleanup := func() { s.Close() }
	return exec, cleanup
}

func TestExecuteAndMapByFieldAt_Success(t *testing.T) {
	json := `{"data":{"items":[{"id":"a","v":1},{"id":2,"v":3},{"id":"a","v":9},{"x":1}]}}`
	exec, cleanup := makeExec(200, json)
	defer cleanup()

	got, err := restyx.ExecuteAndMapByFieldAt(newReq(), exec, []string{"data", "items"}, "id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 mapped items, got %d", len(got))
	}

	// "a" should be overwritten by the later element (v=9)
	rawA, ok := got["a"]
	if !ok {
		t.Fatalf(`missing key "a"`)
	}
	objA, ok := rawA.(map[string]any)
	if !ok {
		t.Fatalf(`value for "a" not an object: %T`, rawA)
	}
	if v := objA["v"]; v != float64(9) { // JSON numbers -> float64
		t.Fatalf(`expected got["a"]["v"] = 9, got %v`, v)
	}

	// numeric id -> key stringified "2"
	if _, ok := got["2"]; !ok {
		t.Fatalf(`expected key "2" to exist`)
	}
}

func TestExecuteAndMapByFieldAt_BadStatus_Default2xxOnly(t *testing.T) {
	exec, cleanup := makeExec(500, `{"err":"x"}`)
	defer cleanup()

	_, err := restyx.ExecuteAndMapByFieldAt(newReq(), exec, []string{"data"}, "id")
	if err == nil || !strings.Contains(err.Error(), "restyx: bad status") {
		t.Fatalf("expected HTTP error, got %v", err)
	}
}

func TestExecuteAndMapByFieldAt_PathNotFound(t *testing.T) {
	exec, cleanup := makeExec(200, `{"data":{"other":[]}}`)
	defer cleanup()

	_, err := restyx.ExecuteAndMapByFieldAt(newReq(), exec, []string{"data", "items"}, "id")
	if err == nil || !strings.Contains(err.Error(), "path [data items] not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteAndMapByFieldAt_NodeNotArray(t *testing.T) {
	exec, cleanup := makeExec(200, `{"data":{"items":{}}}`)
	defer cleanup()

	_, err := restyx.ExecuteAndMapByFieldAt(newReq(), exec, []string{"data", "items"}, "id")
	if err == nil || !strings.Contains(err.Error(), "is not an array") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteAndMapByFieldAt_ExecNil(t *testing.T) {
	_, err := restyx.ExecuteAndMapByFieldAt(newReq(), nil, []string{"data"}, "id")
	if err == nil || !strings.Contains(err.Error(), "exec func is nil") {
		t.Fatalf("expected exec nil error, got %v", err)
	}
}

func TestExecuteAndMapByFieldAt_UnmarshalError(t *testing.T) {
	exec, cleanup := makeExec(200, `{"data":[{]}`)
	defer cleanup()

	_, err := restyx.ExecuteAndMapByFieldAt(newReq(), exec, []string{"data"}, "id")
	if err == nil || !strings.Contains(err.Error(), "restyx: unmarshal body") {
		t.Fatalf("expected unmarshal error, got %v", err)
	}
}

func TestExecuteAndMapByKeyAt_Success_WithValueFn(t *testing.T) {
	json := `{"payload":{"rows":[{"id":"x","v":1},{"id":"skip","v":2},{"id":3,"v":4}]}}`
	exec, cleanup := makeExec(200, json)
	defer cleanup()

	valFn := func(m map[string]any) (int, error) {
		if m["id"] == "skip" {
			return 0, errors.New("skip")
		}
		if v, ok := m["v"].(float64); ok {
			return int(v), nil
		}
		return 0, nil
	}

	got, err := restyx.ExecuteAndMapByKeyAt[int](newReq(), exec, []string{"payload", "rows"}, "id", valFn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// keys: "x" and "3" (stringified)
	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}
	if got["x"] != 1 {
		t.Fatalf(`expected got["x"]=1, got %v`, got["x"])
	}
	if got["3"] != 4 {
		t.Fatalf(`expected got["3"]=4, got %v`, got["3"])
	}
	// "skip" should be omitted because valueFn returns error
	if _, ok := got["skip"]; ok {
		t.Fatalf("did not expect key 'skip'")
	}
}

func TestExecuteAndMapByKeyAt_ValueFnNil(t *testing.T) {
	exec, cleanup := makeExec(200, `{"data":{"items":[]}}`)
	defer cleanup()

	_, err := restyx.ExecuteAndMapByKeyAt[int](newReq(), exec, []string{"data", "items"}, "id", nil)
	if err == nil || !strings.Contains(err.Error(), "valueFn is nil") {
		t.Fatalf("expected valueFn nil error, got %v", err)
	}
}

func TestExecuteAndMapByKeyAt_StatusAllowedList(t *testing.T) {
	// Returns 201; should pass only if we include 201 in expectedStatus
	exec, cleanup := makeExec(201, `{"data":{"items":[{"id":"ok","v":1}]}}`)
	defer cleanup()

	// Allowed -> success
	_, err := restyx.ExecuteAndMapByKeyAt[int](newReq(), exec, []string{"data", "items"}, "id",
		func(m map[string]any) (int, error) { return 1, nil },
		http.StatusCreated,
	)
	if err != nil {
		t.Fatalf("unexpected error for allowed 201: %v", err)
	}

	// Not allowed -> error
	exec2, cleanup2 := makeExec(201, `{"data":{"items":[]}}`)
	defer cleanup2()
	_, err = restyx.ExecuteAndMapByKeyAt[int](newReq(), exec2, []string{"data", "items"}, "id",
		func(m map[string]any) (int, error) { return 0, nil },
		http.StatusOK,
	)
	if err == nil || !strings.Contains(err.Error(), "restyx: bad status 201") {
		t.Fatalf("expected bad status for 201 not in allowed, got %v", err)
	}
}

func TestExecuteAndDecodeAt_Success(t *testing.T) {
	type payload struct {
		N int    `json:"n"`
		S string `json:"s"`
	}
	exec, cleanup := makeExec(200, `{"data":{"obj":{"n":7,"s":"x"}}}`)
	defer cleanup()

	got, err := restyx.ExecuteAndDecodeAt[payload](newReq(), exec, []string{"data", "obj"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.N != 7 || got.S != "x" {
		t.Fatalf("unexpected decoded payload: %#v", got)
	}
}

func TestExecuteAndDecodeAt_PathNotFound(t *testing.T) {
	exec, cleanup := makeExec(200, `{"data":{}}`)
	defer cleanup()

	_, err := restyx.ExecuteAndDecodeAt[any](newReq(), exec, []string{"data", "missing"})
	if err == nil || !strings.Contains(err.Error(), "path [data missing] not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteAndDecodeAt_StatusNotAllowed(t *testing.T) {
	exec, cleanup := makeExec(202, `{"ok":true}`)
	defer cleanup()

	_, err := restyx.ExecuteAndDecodeAt[map[string]any](newReq(), exec, nil, http.StatusOK)
	if err == nil || !strings.Contains(err.Error(), "restyx: bad status 202") {
		t.Fatalf("expected bad status, got %v", err)
	}
}
