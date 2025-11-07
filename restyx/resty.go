package restyx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type ExecFunc func(r *resty.Request) (*resty.Response, error)

const (
	msgUnmarshalBody = "restyx: unmarshal body: %w"
	msgPathNotFound  = "restyx: path %v not found or invalid"
)

func ExecuteAndMapByFieldAt(
	req *resty.Request,
	exec ExecFunc,
	at []string,
	keyField string,
	expectedStatus ...int,
) (map[string]any, error) {
	if exec == nil {
		return nil, errors.New("restyx: exec func is nil")
	}
	resp, err := exec(req)
	if err != nil {
		return nil, err
	}
	if !statusAllowed(resp.StatusCode(), expectedStatus) {
		return nil, newHTTPError(resp)
	}

	var root any
	if err := json.Unmarshal(resp.Body(), &root); err != nil {
		return nil, fmt.Errorf(msgUnmarshalBody, err)
	}
	node, ok := findAt(root, at)
	if !ok {
		return nil, fmt.Errorf(msgPathNotFound, at)
	}
	arr, ok := node.([]any)
	if !ok {
		return nil, fmt.Errorf("restyx: node at %v is not an array", at)
	}

	out := make(map[string]any, len(arr))
	for _, el := range arr {
		obj, ok := el.(map[string]any)
		if !ok {
			continue
		}
		k, ok := obj[keyField]
		if !ok {
			continue
		}
		key, ok := k.(string)
		if !ok {
			key = fmt.Sprint(k)
		}
		out[key] = obj
	}
	return out, nil
}

func ExecuteAndMapByKeyAt[T any](
	req *resty.Request,
	exec ExecFunc,
	at []string,
	keyField string,
	valueFn func(map[string]any) (T, error),
	expectedStatus ...int,
) (map[string]T, error) {
	if valueFn == nil {
		return nil, errors.New("restyx: valueFn is nil")
	}
	resp, err := exec(req)
	if err != nil {
		return nil, err
	}
	if !statusAllowed(resp.StatusCode(), expectedStatus) {
		return nil, newHTTPError(resp)
	}
	var root any
	if err := json.Unmarshal(resp.Body(), &root); err != nil {
		return nil, fmt.Errorf(msgUnmarshalBody, err)
	}
	node, ok := findAt(root, at)
	if !ok {
		return nil, fmt.Errorf(msgPathNotFound, at)
	}
	arr, ok := node.([]any)
	if !ok {
		return nil, fmt.Errorf("restyx: node at %v is not an array", at)
	}

	out := make(map[string]T, len(arr))
	for _, el := range arr {
		obj, ok := el.(map[string]any)
		if !ok {
			continue
		}
		k, ok := obj[keyField]
		if !ok {
			continue
		}
		key, ok := k.(string)
		if !ok {
			key = fmt.Sprint(k)
		}
		v, err := valueFn(obj)
		if err != nil {
			continue
		}
		out[key] = v
	}
	return out, nil
}

func ExecuteAndDecodeAt[T any](
	req *resty.Request,
	exec ExecFunc,
	at []string,
	expectedStatus ...int,
) (T, error) {
	var zero T
	resp, err := exec(req)
	if err != nil {
		return zero, err
	}
	if !statusAllowed(resp.StatusCode(), expectedStatus) {
		return zero, newHTTPError(resp)
	}
	var root any
	if err := json.Unmarshal(resp.Body(), &root); err != nil {
		return zero, fmt.Errorf(msgUnmarshalBody, err)
	}
	node, ok := findAt(root, at)
	if !ok {
		return zero, fmt.Errorf(msgPathNotFound, at)
	}

	b, err := json.Marshal(node)
	if err != nil {
		return zero, fmt.Errorf("restyx: marshal node: %w", err)
	}
	if err := json.Unmarshal(b, &zero); err != nil {
		return zero, fmt.Errorf("restyx: decode node to type: %w", err)
	}
	return zero, nil
}

func statusAllowed(code int, allowed []int) bool {
	if len(allowed) == 0 {
		return code >= 200 && code < 300
	}
	for _, c := range allowed {
		if code == c {
			return true
		}
	}
	return false
}

func newHTTPError(resp *resty.Response) error {
	code := http.StatusInternalServerError
	status := "unknown"
	if resp != nil {
		code = resp.StatusCode()
		status = resp.Status()
	}
	return fmt.Errorf("restyx: bad status %d %s", code, status)
}

func findAt(root any, path []string) (any, bool) {
	if len(path) == 0 {
		return root, true
	}
	cur := root
	for _, seg := range path {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		next, ok := m[seg]
		if !ok {
			return nil, false
		}
		cur = next
	}
	return cur, true
}
