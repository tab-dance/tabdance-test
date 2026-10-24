package daemon

import "encoding/json"

type Request struct {
	Method string         `json:"method"`
	Params map[string]any `json:"params,omitempty"`
}

type Response struct {
	OK     bool   `json:"ok"`
	Result any    `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

func (r *Request) GetString(key string) string {
	if v, ok := r.Params[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (r *Request) GetInt(key string) int {
	if v, ok := r.Params[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		case json.Number:
			i, _ := n.Int64()
			return int(i)
		}
	}
	return 0
}

func (r *Request) GetFloat(key string) float64 {
	if v, ok := r.Params[key]; ok {
		switch n := v.(type) {
		case float64:
			return n
		case int:
			return float64(n)
		case json.Number:
			f, _ := n.Float64()
			return f
		}
	}
	return 0
}

func (r *Request) GetBool(key string) bool {
	if v, ok := r.Params[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// GetBoolDefault returns a bool param value; unlike GetBool it treats an
// absent key as def so callers can express tri-state defaults.
func (r *Request) GetBoolDefault(key string, def bool) bool {
	if _, ok := r.Params[key]; !ok {
		return def
	}
	return r.GetBool(key)
}

func (r *Request) GetStringSlice(key string) []string {
	if v, ok := r.Params[key]; ok {
		if arr, ok := v.([]any); ok {
			result := make([]string, 0, len(arr))
			for _, item := range arr {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		}
	}
	return nil
}

func SuccessResponse(result any) Response {
	return Response{OK: true, Result: result}
}

func ErrorResponse(err string) Response {
	return Response{OK: false, Error: err}
}
