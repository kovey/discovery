package krpc

import "encoding/json"

type LoadBalance struct {
	Config []map[string]any `json:"loadBalancingConfig"`
}

func NewLoadBalance(name string) LoadBalance {
	l := LoadBalance{Config: make([]map[string]any, 1)}
	l.Config[0] = map[string]any{name: map[string]any{}}
	return l
}

func (l LoadBalance) encode() string {
	if buf, err := json.Marshal(l); err == nil {
		return string(buf)
	}

	return "{}"
}
