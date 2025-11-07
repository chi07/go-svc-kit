package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-resty/resty/v2"

	"github.com/chi07/go-svc-kit/httpx"
)

// helper tạo exec func dùng chính *resty.Request truyền vào
func makeExec(url string, status int, body string) (func(*resty.Request) (*resty.Response, error), func()) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	exec := func(r *resty.Request) (*resty.Response, error) {
		// dùng GET cho đơn giản
		return r.Get(srv.URL)
	}
	cleanup := func() { srv.Close() }
	return exec, cleanup
}

func newReq() *resty.Request {
	return resty.New().R()
}

func TestExecuteAndMapByField_Success_DefaultPath(t *testing.T) {
	// data có: id=a (2 lần, bản sau overwrite), id=b,
	// phần tử không hợp lệ: id rỗng, id là số, object thiếu id, số 5
	json := `{"data":[{"id":"a","v":1},{"id":"b","v":2},{"id":""},{"id":123},{"x":"noid"},5,{"id":"a","v":9}]}`
	exec, cleanup := makeExec("", http.StatusOK, json)
	defer cleanup()

	got, err := httpx.ExecuteAndMapByField(newReq(), exec, "id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}
	if v := got["a"]["v"]; v != float64(9) { // JSON numbers -> float64
		t.Fatalf(`expected "a".v = 9, got %v`, v)
	}
	if v := got["b"]["v"]; v != float64(2) {
		t.Fatalf(`expected "b".v = 2, got %v`, v)
	}
}

func TestExecuteAndMapByFieldAt_Success_NestedPath(t *testing.T) {
	json := `{"payload":{"items":[{"slug":"x","n":1},{"slug":"y","n":2}]}}`
	exec, cleanup := makeExec("", http.StatusOK, json)
	defer cleanup()

	got, err := httpx.ExecuteAndMapByFieldAt(newReq(), exec, []string{"payload", "items"}, "slug")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantKeys := []string{"x", "y"}
	if len(got) != len(wantKeys) {
		t.Fatalf("expected %d items, got %d", len(wantKeys), len(got))
	}
	for _, k := range wantKeys {
		if _, ok := got[k]; !ok {
			t.Fatalf("missing key %q", k)
		}
	}
}

func TestExecuteAndMapByField_HTTPError(t *testing.T) {
	exec, cleanup := makeExec("", http.StatusInternalServerError, `{"error":"boom"}`)
	defer cleanup()

	_, err := httpx.ExecuteAndMapByField(newReq(), exec, "id")
	if err == nil {
		t.Fatalf("expected error for HTTP 500")
	}
}

func TestExecuteAndMapByField_InvalidJSON(t *testing.T) {
	exec, cleanup := makeExec("", http.StatusOK, `{"data":[{]}`)
	defer cleanup()

	_, err := httpx.ExecuteAndMapByField(newReq(), exec, "id")
	if err == nil {
		t.Fatalf("expected JSON unmarshal error")
	}
}

func TestExecuteAndMapByFieldAt_PathSegmentNotObject(t *testing.T) {
	exec, cleanup := makeExec("", http.StatusOK, `{"data":[1,2,3]}`)
	defer cleanup()

	_, err := httpx.ExecuteAndMapByFieldAt(newReq(), exec, []string{"data", "items"}, "id")
	want := `invalid path segment "items" (expect object)`
	if err == nil || err.Error() != want {
		t.Fatalf("unexpected error: %v (want %q)", err, want)
	}
}

func TestExecuteAndMapByFieldAt_FirstSegmentNotObject(t *testing.T) {
	exec, cleanup := makeExec("", http.StatusOK, `[{"data":1}]`)
	defer cleanup()

	_, err := httpx.ExecuteAndMapByFieldAt(newReq(), exec, []string{"data"}, "id")
	want := `invalid path segment "data" (expect object)`
	if err == nil || err.Error() != want {
		t.Fatalf("unexpected error: %v (want %q)", err, want)
	}
}

func TestExecuteAndMapByFieldAt_PathSegmentNotFound(t *testing.T) {
	exec, cleanup := makeExec("", http.StatusOK, `{"data":{"other":[]}}`)
	defer cleanup()

	_, err := httpx.ExecuteAndMapByFieldAt(newReq(), exec, []string{"data", "items"}, "id")
	if err == nil || err.Error() != `path segment "items" not found` {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteAndMapByFieldAt_TargetNotArray(t *testing.T) {
	exec, cleanup := makeExec("", http.StatusOK, `{"data":{"items":{}}}`)
	defer cleanup()

	_, err := httpx.ExecuteAndMapByFieldAt(newReq(), exec, []string{"data", "items"}, "id")
	if err == nil || err.Error() != "target at path is not array" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteAndMapByField_SkipsNonObjectsOrMissingID(t *testing.T) {
	json := `{"data":[5, {"noid":1}, {"id":""}, {"id":123}, {"id":"ok","v":1}]}`
	exec, cleanup := makeExec("", http.StatusOK, json)
	defer cleanup()

	got, err := httpx.ExecuteAndMapByField(newReq(), exec, "id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[string]map[string]any{
		"ok": {"id": "ok", "v": float64(1)},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}
