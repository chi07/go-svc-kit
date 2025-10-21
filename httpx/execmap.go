package httpx

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
)

func ExecuteAndMapByFieldAt(
	req *resty.Request,
	exec func(*resty.Request) (*resty.Response, error),
	arrayPath []string,
	idField string,
) (map[string]map[string]any, error) {

	resp, err := exec(req)
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode(), resp.Status())
	}

	var root any
	if err := json.Unmarshal(resp.Body(), &root); err != nil {
		return nil, err
	}

	cur := root
	for _, k := range arrayPath {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid path segment %q (expect object)", k)
		}
		nxt, ok := m[k]
		if !ok {
			return nil, fmt.Errorf("path segment %q not found", k)
		}
		cur = nxt
	}

	arr, ok := cur.([]any)
	if !ok {
		return nil, fmt.Errorf("target at path is not array")
	}

	out := make(map[string]map[string]any, len(arr))
	for _, el := range arr {
		obj, ok := el.(map[string]any)
		if !ok {
			continue
		}
		if v, ok := obj[idField].(string); ok && v != "" {
			out[v] = obj
		}
	}
	return out, nil
}

func ExecuteAndMapByField(
	req *resty.Request,
	exec func(*resty.Request) (*resty.Response, error),
	field string,
) (map[string]map[string]any, error) {
	return ExecuteAndMapByFieldAt(req, exec, []string{"data"}, field)
}
